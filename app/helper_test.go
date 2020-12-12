package app

import (
	"bytes"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/store/sqlstore"
	"github.com/masterhung0112/hk_server/store/storetest/mocks"
	"github.com/masterhung0112/hk_server/testlib"
	"github.com/masterhung0112/hk_server/utils"
	"sync"
	"testing"
	"time"

	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
)

type TestHelper struct {
	App       *App
	Server    *Server
	LogBuffer *bytes.Buffer

	tempWorkspace string

	BasicTeam    *model.Team
	BasicUser    *model.User
	BasicUser2   *model.User
	BasicChannel *model.Channel
	// BasicPost    *model.Post

	SystemAdminUser *model.User
}

func setupTestHelper(dbStore store.Store, tb testing.TB, configSet func(*model.Config)) *TestHelper {

	memoryStore, err := config.NewMemoryStoreWithOptions(&config.MemoryStoreOptions{IgnoreEnvironmentOverrides: true})
	if err != nil {
		panic("failed to initialize memory store: " + err.Error())
	}
	config := memoryStore.Get()
	if configSet != nil {
		configSet(config)
	}
	// *config.PluginSettings.Directory = filepath.Join(tempWorkspace, "plugins")
	// *config.PluginSettings.ClientDirectory = filepath.Join(tempWorkspace, "webapp")
	// *config.LogSettings.EnableSentry = false // disable error reporting during tests
	memoryStore.Set(config)

	buffer := &bytes.Buffer{}

	var options []Option
	options = append(options, ConfigStore(memoryStore))
	options = append(options, StoreOverride(dbStore))
	// options = append(options, SetLogger(mlog.NewTestingLogger(tb, buffer)))

	s, err := NewServer(options...)
	if err != nil {
		panic(err)
	}

	th := &TestHelper{
		App:       New(ServerConnector(s)),
		Server:    s,
		LogBuffer: buffer,
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.MaxUsersPerTeam = 50 })
	// th.App.UpdateConfig(func(cfg *model.Config) { *cfg.RateLimitSettings.Enable = false })
	prevListenAddress := *th.App.Config().ServiceSettings.ListenAddress
	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.ListenAddress = ":0" })
	// Start HTTP Server and other stuff
	if err := th.Server.Start(); err != nil {
		panic(err)
	}

	th.App.UpdateConfig(func(cfg *model.Config) { *cfg.ServiceSettings.ListenAddress = prevListenAddress })

	// th.App.Srv().SearchEngine = mainHelper.SearchEngine

	// th.App.Srv().Store.MarkSystemRanUnitTests()

	// th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.EnableOpenServer = true })

	// Disable strict password requirements for test
	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PasswordSettings.MinimumLength = 5
		*cfg.PasswordSettings.Lowercase = false
		*cfg.PasswordSettings.Uppercase = false
		*cfg.PasswordSettings.Symbol = false
		*cfg.PasswordSettings.Number = false
	})

	th.App.InitServer()

	return th
}

func Setup(tb testing.TB) *TestHelper {
	if testing.Short() {
		tb.SkipNow()
	}
	dbStore := mainHelper.GetStore()
	dbStore.DropAllTables()
	dbStore.MarkSystemRanUnitTests()

	return setupTestHelper(dbStore, tb, nil)
}

var initBasicOnce sync.Once
var userCache struct {
	SystemAdminUser *model.User
	BasicUser       *model.User
	BasicUser2      *model.User
}

func (me *TestHelper) InitBasic() *TestHelper {
	// create users once and cache them because password hashing is slow
	initBasicOnce.Do(func() {
		me.SystemAdminUser = me.CreateUser()
		me.App.UpdateUserRoles(me.SystemAdminUser.Id, model.SYSTEM_USER_ROLE_ID+" "+model.SYSTEM_ADMIN_ROLE_ID, false)
		me.SystemAdminUser, _ = me.App.GetUser(me.SystemAdminUser.Id)
		userCache.SystemAdminUser = me.SystemAdminUser.DeepCopy()

		me.BasicUser = me.CreateUser()
		me.BasicUser, _ = me.App.GetUser(me.BasicUser.Id)
		userCache.BasicUser = me.BasicUser.DeepCopy()

		me.BasicUser2 = me.CreateUser()
		me.BasicUser2, _ = me.App.GetUser(me.BasicUser2.Id)
		userCache.BasicUser2 = me.BasicUser2.DeepCopy()
	})

	// restore cached users
	me.SystemAdminUser = userCache.SystemAdminUser.DeepCopy()
	me.BasicUser = userCache.BasicUser.DeepCopy()
	me.BasicUser2 = userCache.BasicUser2.DeepCopy()
	mainHelper.GetSQLStore().GetMaster().Insert(me.SystemAdminUser, me.BasicUser, me.BasicUser2)

	me.BasicTeam = me.CreateTeam()

	me.LinkUserToTeam(me.BasicUser, me.BasicTeam)
	me.LinkUserToTeam(me.BasicUser2, me.BasicTeam)
	me.BasicChannel = me.CreateChannel(me.BasicTeam)
	// me.BasicPost = me.CreatePost(me.BasicChannel)
	return me
}

func (me *TestHelper) CreateChannel(team *model.Team) *model.Channel {
	return me.createChannel(team, model.CHANNEL_OPEN)
}

func (me *TestHelper) CreatePrivateChannel(team *model.Team) *model.Channel {
	return me.createChannel(team, model.CHANNEL_PRIVATE)
}

func (me *TestHelper) createChannel(team *model.Team, channelType string) *model.Channel {
	id := model.NewId()

	channel := &model.Channel{
		DisplayName: "dn_" + id,
		Name:        "name_" + id,
		Type:        channelType,
		TeamId:      team.Id,
		CreatorId:   me.BasicUser.Id,
	}

	utils.DisableDebugLogForTest()
	var err *model.AppError
	if channel, err = me.App.CreateChannel(channel, true); err != nil {
		mlog.Error(err.Error())

		time.Sleep(time.Second)
		panic(err)
	}
	utils.EnableDebugLogForTest()
	return channel
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

func (me *TestHelper) TearDown() {
	// if me.IncludeCacheLayer {
	// 	// Clean all the caches
	// 	me.App.Srv().InvalidateAllCaches()
	// }
	me.ShutdownApp()
	// if me.tempWorkspace != "" {
	// 	os.RemoveAll(me.tempWorkspace)
	// }
}

func SetupWithStoreMock(tb testing.TB) *TestHelper {
	mockStore := testlib.GetMockStoreForSetupFunctions()
	th := setupTestHelper(mockStore, tb, nil) //setupTestHelper(mockStore, false, false, tb, nil)
	emptyMockStore := mocks.Store{}
	emptyMockStore.On("Close").Return(nil)
	th.App.Srv().Store = &emptyMockStore
	return th
}

func (me *TestHelper) GetSQLStore() *sqlstore.SqlStore {
	return mainHelper.GetSQLStore()
}

func (me *TestHelper) CreateUser() *model.User {
	return me.CreateUserOrGuest(false)
}

func (me *TestHelper) CreateUserOrGuest(guest bool) *model.User {
	id := model.NewId()

	user := &model.User{
		Email:    "success+" + id + "@simulator.amazonses.com",
		Username: "un_" + id,
		//TODO: Open this
		// Nickname:      "nn_" + id,
		Password:      "Password1",
		EmailVerified: true,
	}

	utils.DisableDebugLogForTest()
	var err *model.AppError
	if guest {
		if user, err = me.App.CreateGuest(user); err != nil {
			mlog.Error(err.Error())

			time.Sleep(time.Second)
			panic(err)
		}
	} else {
		if user, err = me.App.CreateUser(user); err != nil {
			mlog.Error(err.Error())

			time.Sleep(time.Second)
			panic(err)
		}
	}
	utils.EnableDebugLogForTest()
	return user
}

func (me *TestHelper) CreateTeam() *model.Team {
	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TEAM_OPEN,
	}

	utils.DisableDebugLogForTest()
	var err *model.AppError
	if team, err = me.App.CreateTeam(team); err != nil {
		mlog.Error(err.Error())

		time.Sleep(time.Second)
		panic(err)
	}
	utils.EnableDebugLogForTest()
	return team
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
