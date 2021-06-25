package base

import (
	"github.com/timerzz/itchatgo/http_client"
	"github.com/timerzz/itchatgo/model"
)

type Client struct {
	HttpClient *http_client.Client
	LoginInfo  *model.LoginMap
	Logging    bool
	Logged     bool
}

func NewClient(httpclt *http_client.Client, loginInfo *model.LoginMap) *Client {
	return &Client{HttpClient: httpclt, LoginInfo: loginInfo}
}
