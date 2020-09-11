package model

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	HEADER_REQUEST_ID = "X-Request-ID"
	HEADER_VERSION_ID = "X-Version-ID"
	HEADER_CLUSTER_ID = "X-Cluster-ID"

	HEADER_ETAG_SERVER = "ETag"
	HEADER_ETAG_CLIENT = "If-None-Match"

	HEADER_TOKEN  = "token"
	HEADER_BEARER = "BEARER"
	HEADER_AUTH   = "Authorization"

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

type Client struct {
	Url    string // the location of the server
	ApiUrl string // The Api location of the server

	HttpClient *http.Client
	AuthToken  string
	AuthType   string
	HttpHeader map[string]string
}

func NewApiClient(url string) *Client {
	return &Client{
		url,
		url + API_URL_SUFFIX,
		&http.Client{},
		"",
		"",
		map[string]string{},
	}
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

func closeBody(r *http.Response) {
	if r.Body != nil {
		_, _ = io.Copy(ioutil.Discard, r.Body)
		_ = r.Body.Close()
	}
}

func (client *Client) GetUsersRoute() string {
	return "/users"
}

func (c *Client) DoApiGet(url string, etag string) (*http.Response, *AppError) {
	return c.DoApiRequest(http.MethodGet, c.ApiUrl+url, "", etag)
}

func (client *Client) DoApiPost(url string, data string) (*http.Response, *AppError) {
	return client.DoApiRequest(http.MethodPost, client.ApiUrl+url, data, "")
}

func (client *Client) DoApiRequest(method, url, data, etag string) (*http.Response, *AppError) {
	return client.doApiRequestReader(method, url, strings.NewReader(data), etag)
}

func (client *Client) doApiRequestReader(method, url string, data io.Reader, etag string) (*http.Response, *AppError) {
	rq, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), http.StatusBadRequest)
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
		return nil, NewAppError(url, "model.client.connecting.app_error", nil, err.Error(), 0)
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

func (client *Client) CreateUser(user *User) (*User, *Response) {
	r, err := client.DoApiPost(client.GetUsersRoute(), user.ToJson())
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	return UserFromJson(r.Body), BuildResponse(r)
}

// Login authenticates a user by login id, which can be username, email or some sort
// of SSO identifier based on server configuration, and a password.
func (c *Client) Login(loginId string, password string) (*User, *Response) {
	m := make(map[string]string)
	m["login_id"] = loginId
	m["password"] = password
	return c.login(m)
}

func (c *Client) login(m map[string]string) (*User, *Response) {
	r, err := c.DoApiPost("/users/login", MapToJson(m))
	if err != nil {
		return nil, BuildErrorResponse(r, err)
	}
	defer closeBody(r)
	c.AuthToken = r.Header.Get(HEADER_TOKEN)
	c.AuthType = HEADER_BEARER
	return UserFromJson(r.Body), BuildResponse(r)
}
