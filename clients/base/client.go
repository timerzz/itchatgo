package base

import (
	"github.com/timerzz/itchatgo/http_client"
	"github.com/timerzz/itchatgo/model"
)

type Client struct {
	httpClient *http_client.Client
	loginInfo  *model.LoginMap
	logging    bool
	logged     bool
}

func (c *Client) HttpClient() *http_client.Client {
	return c.httpClient
}

func (c *Client) LoginInfo() *model.LoginMap {
	return c.loginInfo
}

func (c *Client) Logging() bool {
	return c.logging
}

func (c *Client) SetLogging(logging bool) {
	c.logging = logging
}

func (c *Client) Logged() bool {
	return c.logged
}

func (c *Client) SetLogged(logged bool) {
	c.logged = logged
}

func NewClient(httpclt *http_client.Client, loginInfo *model.LoginMap) *Client {
	return &Client{
		httpClient: httpclt,
		loginInfo:  loginInfo,
	}
}

func (c *Client) Clear() {
	c.logged = false
	c.logging = false
}
