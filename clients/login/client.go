package login

import (
	"fmt"
	"github.com/timerzz/itchatgo/api"
	"github.com/timerzz/itchatgo/clients/contact"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/model"
	"time"
)

type Client struct {
	api        *api.Api
	contactCtl *contact.Client
	//当前登录的用户
	self     *model.User
	uuidInfo *model.UUidInfo
	//登录流程控制
	loginC     chan struct{}
	loginStopC chan struct{}
	//回调函数
	loggedCall    func() //登录成功的回调函数
	logoutCall    func() //退出登录的回调函数
	stopLoginCall func() //停止登录的回调
	//超时时间
	timeout time.Duration

	//登录状态
	logged  bool
	logging bool
}

func NewClient(api *api.Api, contact *contact.Client) *Client {
	return &Client{
		api:        api,
		contactCtl: contact,
		loginC:     make(chan struct{}),
		loginStopC: make(chan struct{}),
		timeout:    time.Minute * 10,
	}
}

func (c *Client) Login() (info *model.UUidInfo, err error) {
	if c.uuidInfo != nil {
		info = c.uuidInfo
	}
	if !c.logging && !c.logged {
		c.loginC = make(chan struct{})
		c.logging = true
		info, err = c.ReLoadUUid()
		go c.waitLogin()
	}
	return
}

//登录的超时时间
func (c *Client) SetTimeout(t time.Duration) *Client {
	c.timeout = t
	return c
}

func (c *Client) ReLoadUUid() (info *model.UUidInfo, err error) {
	c.uuidInfo = &model.UUidInfo{}
	c.uuidInfo.UUid, err = c.api.GetUUID()
	if err != nil {
		return
	}
	c.uuidInfo.QrUrl = enum.QRCODE + c.uuidInfo.UUid
	c.uuidInfo.QrImg, err = c.api.GetQR(c.uuidInfo.UUid)
	info = c.uuidInfo
	return
}

func (c *Client) WaitLogin() {
	if !c.logged && c.logging {
		<-c.loginC
	}
	return
}

func (c *Client) StopLogin() {
	if !c.logged && c.logging {
		c.loginStopC <- struct{}{}
	}
}

func (c *Client) UUidInfo() *model.UUidInfo {
	return c.uuidInfo
}

func (c *Client) waitLogin() {
	defer close(c.loginC)

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	timer := time.NewTimer(c.timeout)
	for ; c.logging; <-ticker.C {
		select {
		case <-c.loginStopC:
			c.logging = false
			if c.stopLoginCall != nil {
				c.stopLoginCall()
			}
			return
		case <-timer.C:
			c.logging = false
			if c.stopLoginCall != nil {
				c.stopLoginCall()
			}
			return
		default:
			status, _err := c.api.CheckLogin(c.uuidInfo.UUid)
			switch status {
			case 200:
				_ = c.api.NotifyStatus()
				initInfo, _ := c.api.InitWX()
				if c.contactCtl != nil {
					c.contactCtl.UpdateContacts(initInfo.ContactList...)
				}
				c.self = &initInfo.User
				c.logged, c.logging = true, false
				if c.loggedCall != nil {
					c.loggedCall()
				}
				return
			default:
				if _err != nil {
					fmt.Println(_err)
				}
			}
		}
	}
}

//设置登录成功时的回调
func (c *Client) SetLoggedCall(f func()) *Client {
	c.loggedCall = f
	return c
}

//设置退出登陆的回调
func (c *Client) SetLogoutCall(f func()) *Client {
	c.logoutCall = f
	return c
}

//设置停止登录时的回调，停止登录或者超时会调用
func (c *Client) SetStopLoginCall(f func()) *Client {
	c.stopLoginCall = f
	return c
}

func (c *Client) Self() *model.User {
	return c.self
}

func (c *Client) SetLogged(logged bool) {
	c.logged = logged
}

func (c *Client) Logout() (err error) {
	if c.logged {
		err = c.api.Logout()
		if err == nil {
			c.logged = false
			if c.contactCtl != nil {
				c.contactCtl.Clear()
			}
			if c.logoutCall != nil {
				c.logoutCall()
			}
		}
	}
	return
}
