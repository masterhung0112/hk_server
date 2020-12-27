package model

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	HEADER_REQUEST_ID         = "X-Request-ID"
	HEADER_VERSION_ID         = "X-Version-ID"
	HEADER_CLUSTER_ID         = "X-Cluster-ID"
	HEADER_ETAG_SERVER        = "ETag"
	HEADER_ETAG_CLIENT        = "If-None-Match"
	HEADER_FORWARDED          = "X-Forwarded-For"
	HEADER_REAL_IP            = "X-Real-IP"
	HEADER_FORWARDED_PROTO    = "X-Forwarded-Proto"
	HEADER_TOKEN              = "token"
	HEADER_CSRF_TOKEN         = "X-CSRF-Token"
	HEADER_BEARER             = "BEARER"
	HEADER_AUTH               = "Authorization"
	HEADER_CLOUD_TOKEN        = "X-Cloud-Token"
	HEADER_REQUESTED_WITH     = "X-Requested-With"
	HEADER_REQUESTED_WITH_XML = "XMLHttpRequest"
	STATUS                    = "status"
	STATUS_OK                 = "OK"
	STATUS_FAIL               = "FAIL"
	STATUS_UNHEALTHY          = "UNHEALTHY"
	STATUS_REMOVE             = "REMOVE"

	CLIENT_DIR = "client"

	API_URL_SUFFIX = "/api/v1"
)

type Response struct {
	StatusCode    int
	Error         *AppError
	RequestId     string
	Etag          string
	ServerVersion string
	Header        http.Header
}

type Client1 struct {
	Url        string       // The location of the server, for example  "http://localhost:8065"
	ApiUrl     string       // The api location of the server, for example "http://localhost:8065/api/v4"
	HttpClient *http.Client // The http client
	AuthToken  string
	AuthType   string
	HttpHeader map[string]string // Headers to be copied over for each request

	// TrueString is the string value sent to the server for true boolean query parameters.
	trueString string

	// FalseString is the string value sent to the server for false boolean query parameters.
	falseString string
}

// SetBoolString is a helper method for overriding how true and false query string parameters are
// sent to the server.
//
// This method is only exposed for testing. It is never necessary to configure these values
// in production.
func (c *Client1) SetBoolString(value bool, valueStr string) {
	if value {
		c.trueString = valueStr
	} else {
		c.falseString = valueStr
	}
}

// boolString builds the query string parameter for boolean values.
func (c *Client1) boolString(value bool) string {
	if value && c.trueString != "" {
		return c.trueString
	} else if !value && c.falseString != "" {
		return c.falseString
	}

	if value {
		return "true"
	}
	return "false"
}

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}
}

// Must is a convenience function used for testing.
func (c *Client1) Must(result interface{}, resp *Response) interface{} {
	if resp.Error != nil {
		time.Sleep(time.Second)
		panic(resp.Error)
	}

	return result
}

func NewApiClient(url string) *Client1 {
	url = strings.TrimRight(url, "/")
	return &Client1{url, url + API_URL_SUFFIX, &http.Client{}, "", "", map[string]string{}, "", ""}
}

func BuildErrorResponse(r *http.Response, err *AppError) *Response {
	var statusCode int
	var header http.Header
	if r != nil {
		statusCode = r.StatusCode
		header = r.Header
	} else {
		statusCode = 0
		header = make(http.Header)
	}

	return &Response{
		StatusCode: statusCode,
		Error:      err,
		Header:     header,
	}
}

func BuildResponse(r *http.Response) *Response {
	return &Response{
		StatusCode:    r.StatusCode,
		RequestId:     r.Header.Get(HEADER_REQUEST_ID),
		Etag:          r.Header.Get(HEADER_ETAG_SERVER),
		ServerVersion: r.Header.Get(HEADER_VERSION_ID),
		Header:        r.Header,
	}
}

func (client *Client1) GetUsersRoute() string {
	return "/users"
}

func (c *Client1) GetUserRoute(userId string) string {
	return fmt.Sprintf(c.GetUsersRoute()+"/%v", userId)
}

func (c *Client1) GetTeamsRoute() string {
	return "/teams"
}

func (c *Client1) DoApiGet(url string, etag string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodGet, c.ApiUrl+url, "", etag)
}

func (client *Client1) DoApiPost(url string, data string) (*http.Response, *AppError) {
	return client.DoApiRequest(http.MethodPost, client.ApiUrl+url, data, "")
}

func (client *Client1) DoApiRequest(method, url, data, etag string) (*http.Response, *AppError) {
	return client.doApiRequestReader(method, url, strings.NewReader(data), etag)
}

func (client *Client1) doApiRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *AppError) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, NewAppError(url, "model.Client1.connecting.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	if len(etag) > 0 {
		rq.Header.Set(HEADER_ETAG_CLIENT, etag)
	}

	if len(client.AuthToken) > 0 {
		rq.Header.Set(HEADER_AUTH, client.AuthType+" "+client.AuthToken)
	}

	if client.HttpHeader != nil && len(client.HttpHeader) > 0 {
		for k, v := range client.HttpHeader {
			rq.Header.Set(k, v)
		}
	}

	rp, err := client.HttpClient.Do(rq)
	if err != nil || rp == nil {
		return nil, NewAppError(url, "model.Client1.connecting.app_error", nil, err.Error(), 0)
	}

	if rp.StatusCode == 304 {
		return rp, nil
	}

	if rp.StatusCode >= 300 {
		defer closeBody(rp)
		return rp, AppErrorFromJson(rp.Body)
	}

	return rp, nil
}

func (client *Client1) CreateUser(user *User) (*User, *Response) {
	r, err := client.DoApiPost(client.GetUsersRoute(), user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// Login authenticates a user by login id, which can be username, email or some sort
// of SSO identifier based on server configuration, and a password.
func (c *Client1) Login(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	return c.login(m)
}

func (c *Client1) login(m map[string]string) (*User, *Response) {
	r, err := c.DoApiPost("/users/login", MapToJson(m))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	c.AuthToken = r.Header.Get(HEADER_TOKEN)
	c.AuthType = HEADER_BEARER
	return UserFromJson(r.Body), BuildResponse(r)
}

// GetUsers returns a page of users on the system. Page counting starts at 0.
func (c *Client1) GetUsers(page int, perPage int, etag string) ([]*User, *Response) {
	query := fmt.Sprintf("?page=%v&per_page=%v", page, perPage)
	r, err := c.DoApiGet(c.GetUsersRoute()+query, etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserListFromJson(r.Body), BuildResponse(r)
}

// Logout terminates the current user's session.
func (c *Client1) Logout() (bool, *Response) {
	r, err := c.DoApiPost("/users/logout", "")
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	c.AuthToken = ""
	c.AuthType = HEADER_BEARER
	return CheckStatusOK(r), BuildResponse(r)
}

// CheckStatusOK is a convenience function for checking the standard OK response
// from the web service.
func CheckStatusOK(r *http.Response) bool {
	m := MapFromJson(r.Body)
	defer closeBody(r)

	if m != nil && m[STATUS] == STATUS_OK {
		return true
	}

	return false
}

// Team Section

// CreateTeam creates a team in the system based on the provided team struct.
func (c *Client1) CreateTeam(team *Team) (*Team, *Response) {
	r, err := c.DoApiPost(c.GetTeamsRoute(), team.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return TeamFromJson(r.Body), BuildResponse(r)
}

// GetMe returns the logged in user.
func (c *Client1) GetMe(etag string) (*User, *Response) {
	r, err := c.DoApiGet(c.GetUserRoute(ME), etag)
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// CreateChannel creates a channel based on the provided channel struct.
func (c *Client1) CreateChannel(channel *Channel) (*Channel, *Response) {
	r, err := c.DoApiPost(c.GetChannelsRoute(), channel.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ChannelFromJson(r.Body), BuildResponse(r)
}

// DeleteChannel deletes channel based on the provided channel id string.
func (c *Client1) DeleteChannel(channelId string) (bool, *Response) {
	r, err := c.DoApiDelete(c.GetChannelRoute(channelId))
	if err != nil {
		return false, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return CheckStatusOK(r), BuildResponse(r)
}

func (c *Client1) DoApiDelete(url string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodDelete, c.ApiUrl+url, "", "")
}

func (c *Client1) GetChannelsRoute() string {
	return "/channels"
}

func (c *Client1) GetChannelRoute(channelId string) string {
	return fmt.Sprintf(c.GetChannelsRoute()+"/%v", channelId)
}

func (c *Client1) GetConfigRoute() string {
	return "/config"
}

// GetConfig will retrieve the server config with some sanitized items.
func (c *Client1) GetConfig() (*Config, *Response) {
	r, err := c.DoApiGet(c.GetConfigRoute(), "")
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return ConfigFromJson(r.Body), BuildResponse(r)
}
