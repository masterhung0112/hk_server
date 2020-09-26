package api

import (
	"fmt"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/testlib"
	"github.com/masterhung0112/hk_server/utils"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	App                  *app.App
	Server               *app.Server
	ConfigStore          config.Store
	Client               *model.Client
	BasicUser            *model.User
	BasicUser2           *model.User
	TeamAdminUser        *model.User
	BasicTeam            *model.Team
	BasicChannel         *model.Channel
	BasicPrivateChannel  *model.Channel
	BasicPrivateChannel2 *model.Channel
	BasicDeletedChannel  *model.Channel
	BasicChannel2        *model.Channel
	// BasicPost            *model.Post
	// Group                *model.Group

	SystemAdminClient *model.Client
	SystemAdminUser   *model.User
	LocalClient       *model.Client
}

var mainHelper *testlib.MainHelper

func SetMainHelper(mh *testlib.MainHelper) {
	mainHelper = mh
}

func setupTestHelper(dbStore store.Store) *TestHelper {
	var options []app.Option

	memoryStore, err := config.NewMemoryStoreWithOptions(&config.MemoryStoreOptions{IgnoreEnvironmentOverrides: true})
	if err != nil {
		panic("failed to initialize memory store: " + err.Error())
	}

	config := memoryStore.Get()
	// *config.PluginSettings.Directory = filepath.Join(tempWorkspace, "plugins")
	// *config.PluginSettings.ClientDirectory = filepath.Join(tempWorkspace, "webapp")
	// config.ServiceSettings.EnableLocalMode = model.NewBool(true)
	// *config.ServiceSettings.LocalModeSocketLocation = filepath.Join(tempWorkspace, "mattermost_local.sock")
	// if updateConfig != nil {
	// 	updateConfig(config)
	// }
	memoryStore.Set(config)

	options = append(options, app.ConfigStore(memoryStore))
	options = append(options, app.StoreOverride(dbStore))

	s, err := app.NewServer(options...)
	if err != nil {
		panic(err)
	}

	th := &TestHelper{
		App:    app.New(app.ServerConnector(s)),
		Server: s,
	}

	// Initialize the router URL
	ApiInit(th.Server.AppOptions, th.App.Srv().Router)
	prevListenAddress := *th.App.Config().ServiceSettings.ListenAddress
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.ListenAddress = ":0" })

	// Start HTTP Server and other stuff
	if err := th.Server.Start(); err != nil {
		panic(err)
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.ListenAddress = prevListenAddress })

	// Disable strict password requirements for test
	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PasswordSettings.MinimumLength = 5
		*cfg.PasswordSettings.Lowercase = false
		*cfg.PasswordSettings.Uppercase = false
		*cfg.PasswordSettings.Symbol = false
		*cfg.PasswordSettings.Number = false
	})

	th.Client = th.CreateClient()
	th.SystemAdminClient = th.CreateClient()
	//TODO: Open this
	// th.LocalClient = th.CreateLocalClient(*config.ServiceSettings.LocalModeSocketLocation)

	th.App.InitServer()

	return th
}

func (me *TestHelper) InitBasic() *TestHelper {
	me.BasicTeam = me.CreateTeam()
	me.BasicChannel = me.CreatePublicChannel()
	me.BasicPrivateChannel = me.CreatePrivateChannel()
	me.BasicPrivateChannel2 = me.CreatePrivateChannel()
	me.BasicDeletedChannel = me.CreatePublicChannel()
	me.BasicChannel2 = me.CreatePublicChannel()
	// me.BasicPost = me.CreatePost()
	me.LinkUserToTeam(me.BasicUser, me.BasicTeam)
	me.LinkUserToTeam(me.BasicUser2, me.BasicTeam)
	me.App.AddUserToChannel(me.BasicUser, me.BasicChannel)
	me.App.AddUserToChannel(me.BasicUser2, me.BasicChannel)
	me.App.AddUserToChannel(me.BasicUser, me.BasicChannel2)
	me.App.AddUserToChannel(me.BasicUser2, me.BasicChannel2)
	me.App.AddUserToChannel(me.BasicUser, me.BasicPrivateChannel)
	me.App.AddUserToChannel(me.BasicUser2, me.BasicPrivateChannel)
	me.App.AddUserToChannel(me.BasicUser, me.BasicDeletedChannel)
	me.App.AddUserToChannel(me.BasicUser2, me.BasicDeletedChannel)
	me.App.UpdateUserRoles(me.BasicUser.Id, model.SYSTEM_USER_ROLE_ID, false)
	me.Client.DeleteChannel(me.BasicDeletedChannel.Id)
	me.LoginBasic()
	// me.Group = me.CreateGroup()

	return me
}

func (th *TestHelper) CreateClient() *model.Client {
	return model.NewApiClient(fmt.Sprintf("http://localhost:%v", th.App.Srv().ListenAddr.Port))
}

func Setup(tb testing.TB) *TestHelper {
	if testing.Short() {
		tb.SkipNow()
	}

	if mainHelper == nil {
		tb.SkipNow()
	}

	dbStore := mainHelper.GetStore()
	dbStore.DropAllTables()
	dbStore.MarkSystemRanUnitTests()

	th := setupTestHelper(dbStore)
	th.InitLogin()
	return th
}

func (me *TestHelper) TearDown() {
	// utils.DisableDebugLogForTest()
	// if me.IncludeCacheLayer {
	// 	// Clean all the caches
	// 	me.App.Srv().InvalidateAllCaches()
	// }

	me.ShutdownApp()

	// utils.EnableDebugLogForTest()
}

func (me *TestHelper) ShutdownApp() {
	done := make(chan bool)
	go func() {
		me.Server.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		// panic instead of fatal to terminate all tests in this package, otherwise the
		// still running App could spuriously fail subsequent tests.
		panic("failed to shutdown App within 30 seconds")
	}
}

// ToDo: maybe move this to NewAPIv4SocketClient and reuse it in mmctl
func (me *TestHelper) CreateLocalClient(socketPath string) *model.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return &model.Client{
		ApiUrl:     "http://_" + model.API_URL_SUFFIX,
		HttpClient: httpClient,
	}
}

func (th *TestHelper) GenerateTestEmail() string {
	return strings.ToLower(model.NewId() + "@localhost")
}

func (th *TestHelper) GenerateTestUsername() string {
	return "fakeuser" + model.NewRandomString(10)
}

func checkHTTPStatus(t *testing.T, resp *model.Response, expectedStatus int, expectError bool) {
	t.Helper()

	require.NotNilf(t, resp, "Unexpected nil response, expected http:%v, expectError:%v", expectedStatus, expectError)
	if expectError {
		require.NotNil(t, resp.Error, "Expected a non-nil error and http status:%v, got nil, %v, error: %v", expectedStatus, resp.StatusCode)
	} else {
		require.Nil(t, resp.Error, "Expected no error and http status:%v, got %q, http:%v", expectedStatus, resp.Error, resp.StatusCode)
	}
	require.Equalf(t, expectedStatus, resp.StatusCode, "Expected http status:%v, got %v (err: %q)", expectedStatus, resp.StatusCode, resp.Error)
}

func CheckNoError(t *testing.T, resp *model.Response) {
	t.Helper()

	require.Nil(t, resp.Error, "expected no error")
}

func CheckCreatedStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusCreated, false)
}

func CheckBadRequestStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusBadRequest, true)
}

func CheckUserSanitization(t *testing.T, user *model.User) {
	t.Helper()

	require.Equal(t, "", user.Password, "password wasn't blank")
	//TODO: Open
	// require.Empty(t, user.AuthData, "auth data wasn't blank")
	// require.Equal(t, "", user.MfaSecret, "mfa secret wasn't blank")
}

func CheckErrorMessage(t *testing.T, resp *model.Response, errorId string) {
	t.Helper()

	require.NotNilf(t, resp.Error, "should have errored with message: %s", errorId)
	require.Equalf(t, resp.Error.Id, errorId, "incorrect error message, actual: %s, expected: %s", resp.Error.Id, errorId)
}

func GenerateTestUsername() string {
	return "fakeuser" + model.NewRandomString(10)
}

func GenerateTestTeamName() string {
	return "faketeam" + model.NewRandomString(6)
}

func GenerateTestChannelName() string {
	return "fakechannel" + model.NewRandomString(10)
}

func GenerateTestAppName() string {
	return "fakeoauthapp" + model.NewRandomString(10)
}

func GenerateTestId() string {
	return model.NewId()
}

// TestForAllClients runs a test function for all the clients
// registered in the TestHelper
func (me *TestHelper) TestForAllClients(t *testing.T, f func(*testing.T, *model.Client), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"Client", func(t *testing.T) {
		f(t, me.Client)
	})

	t.Run(testName+"SystemAdminClient", func(t *testing.T) {
		f(t, me.SystemAdminClient)
	})

	//TODO: Open this
	// t.Run(testName+"LocalClient", func(t *testing.T) {
	// 	f(t, me.LocalClient)
	// })
}

func CheckUnauthorizedStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusUnauthorized, true)
}

func (me *TestHelper) waitForConnectivity() {
	for i := 0; i < 1000; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%v", me.App.Srv().ListenAddr.Port))
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(time.Millisecond * 20)
	}
	panic("unable to connect")
}

var initBasicOnce sync.Once

var userCache struct {
	SystemAdminUser *model.User
	TeamAdminUser   *model.User
	BasicUser       *model.User
	BasicUser2      *model.User
}

func (me *TestHelper) CreateUser() *model.User {
	return me.CreateUserWithClient(me.Client)
}

func (me *TestHelper) CreateTeam() *model.Team {
	return me.CreateTeamWithClient(me.Client)
}

func (me *TestHelper) CreateUserWithClient(client *model.Client) *model.User {
	id := model.NewId()

	user := &model.User{
		Email:     me.GenerateTestEmail(),
		Username:  GenerateTestUsername(),
		Nickname:  "nn_" + id,
		FirstName: "f_" + id,
		LastName:  "l_" + id,
		Password:  "Pa$$word11",
	}

	utils.DisableDebugLogForTest()
	ruser, response := client.CreateUser(user)
	if response.Error != nil {
		panic(response.Error)
	}

	ruser.Password = "Pa$$word11"
	//TODO: Open
	// _, err := me.App.Srv().Store.User().VerifyEmail(ruser.Id, ruser.Email)
	// if err != nil {
	// 	return nil
	// }
	utils.EnableDebugLogForTest()
	return ruser
}

func (me *TestHelper) InitLogin() *TestHelper {
	me.waitForConnectivity()

	// create users once and cache them because password hashing is slow
	initBasicOnce.Do(func() {
		var err *model.AppError = nil
		me.SystemAdminUser = me.CreateUser()
		me.App.UpdateUserRoles(me.SystemAdminUser.Id, model.SYSTEM_USER_ROLE_ID+" "+model.SYSTEM_ADMIN_ROLE_ID, false)
		me.SystemAdminUser, err = me.App.GetUser(me.SystemAdminUser.Id)
		if err != nil {
			panic(err)
		}
		userCache.SystemAdminUser = me.SystemAdminUser.DeepCopy()

		me.TeamAdminUser = me.CreateUser()
		me.App.UpdateUserRoles(me.TeamAdminUser.Id, model.SYSTEM_USER_ROLE_ID, false)
		me.TeamAdminUser, err = me.App.GetUser(me.TeamAdminUser.Id)
		if err != nil {
			panic(err)
		}
		userCache.TeamAdminUser = me.TeamAdminUser.DeepCopy()

		me.BasicUser = me.CreateUser()
		me.BasicUser, err = me.App.GetUser(me.BasicUser.Id)
		if err != nil {
			panic(err)
		}
		userCache.BasicUser = me.BasicUser.DeepCopy()

		me.BasicUser2 = me.CreateUser()
		me.BasicUser2, err = me.App.GetUser(me.BasicUser2.Id)
		if err != nil {
			panic(err)
		}
		userCache.BasicUser2 = me.BasicUser2.DeepCopy()
	})
	// restore cached users
	me.SystemAdminUser = userCache.SystemAdminUser.DeepCopy()
	me.TeamAdminUser = userCache.TeamAdminUser.DeepCopy()
	me.BasicUser = userCache.BasicUser.DeepCopy()
	me.BasicUser2 = userCache.BasicUser2.DeepCopy()
	mainHelper.GetSQLSupplier().GetMaster().Insert(me.SystemAdminUser, me.TeamAdminUser, me.BasicUser, me.BasicUser2)
	// restore non hashed password for login
	me.SystemAdminUser.Password = "Pa$$word11"
	me.TeamAdminUser.Password = "Pa$$word11"
	me.BasicUser.Password = "Pa$$word11"
	me.BasicUser2.Password = "Pa$$word11"

	//TODO: Open
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		me.LoginSystemAdmin()
		wg.Done()
	}()
	go func() {
		me.LoginTeamAdmin()
		wg.Done()
	}()
	wg.Wait()
	return me
}

func (me *TestHelper) CreateTeamWithClient(client *model.Client) *model.Team {
	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        GenerateTestTeamName(),
		Email:       me.GenerateTestEmail(),
		Type:        model.TEAM_OPEN,
	}

	utils.DisableDebugLogForTest()
	rteam, resp := client.CreateTeam(team)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
	return rteam
}

func (me *TestHelper) LoginBasic() {
	me.LoginBasicWithClient(me.Client)
}

func (me *TestHelper) LoginSystemAdmin() {
	me.LoginSystemAdminWithClient(me.SystemAdminClient)
}

func (me *TestHelper) LoginTeamAdmin() {
	me.LoginTeamAdminWithClient(me.Client)
}

func (me *TestHelper) LoginBasicWithClient(client *model.Client) {
	utils.DisableDebugLogForTest()
	_, resp := client.Login(me.BasicUser.Email, me.BasicUser.Password)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
}

func (me *TestHelper) LoginTeamAdminWithClient(client *model.Client) {
	utils.DisableDebugLogForTest()
	_, resp := client.Login(me.TeamAdminUser.Email, me.TeamAdminUser.Password)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
}

func (me *TestHelper) LoginSystemAdminWithClient(client *model.Client) {
	utils.DisableDebugLogForTest()
	_, resp := client.Login(me.SystemAdminUser.Email, me.SystemAdminUser.Password)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
}

func (me *TestHelper) CreatePublicChannel() *model.Channel {
	return me.CreateChannelWithClient(me.Client, model.CHANNEL_OPEN)
}

func (me *TestHelper) CreatePrivateChannel() *model.Channel {
	return me.CreateChannelWithClient(me.Client, model.CHANNEL_PRIVATE)
}

func (me *TestHelper) CreateChannelWithClient(client *model.Client, channelType string) *model.Channel {
	return me.CreateChannelWithClientAndTeam(client, channelType, me.BasicTeam.Id)
}

func (me *TestHelper) CreateChannelWithClientAndTeam(client *model.Client, channelType string, teamId string) *model.Channel {
	id := model.NewId()

	channel := &model.Channel{
		DisplayName: "dn_" + id,
		Name:        GenerateTestChannelName(),
		Type:        channelType,
		TeamId:      teamId,
	}

	utils.DisableDebugLogForTest()
	rchannel, resp := client.CreateChannel(channel)
	if resp.Error != nil {
		panic(resp.Error)
	}
	utils.EnableDebugLogForTest()
	return rchannel
}

func (me *TestHelper) LinkUserToTeam(user *model.User, team *model.Team) {
	utils.DisableDebugLogForTest()

	err := me.App.JoinUserToTeam(team, user, "")
	if err != nil {
		mlog.Error(err.Error())

		time.Sleep(time.Second)
		panic(err)
	}

	utils.EnableDebugLogForTest()
}

func (me *TestHelper) AddUserToChannel(user *model.User, channel *model.Channel) *model.ChannelMember {
	utils.DisableDebugLogForTest()

	member, err := me.App.AddUserToChannel(user, channel)
	if err != nil {
		mlog.Error(err.Error())

		time.Sleep(time.Second)
		panic(err)
	}

	utils.EnableDebugLogForTest()

	return member
}
