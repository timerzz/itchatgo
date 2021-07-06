package msg

import (
	"github.com/timerzz/itchatgo/api"
	"github.com/timerzz/itchatgo/model"
	"time"
)

type Client struct {
	api        *api.Api
	stopC      chan struct{}
	receiving  bool   // 是否正在监听
	exitCall   func() //退出时的回调
	logoutCall func() //退出登录的回调
}

func NewClient(api *api.Api, logoutCall func()) *Client {
	return &Client{
		api:        api,
		logoutCall: logoutCall,
		stopC:      make(chan struct{}),
	}
}

//设置退出时的回调函数
func (c *Client) SetExitCall(f func()) {
	c.exitCall = f
}

func (c *Client) Receive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) {
	if !c.receiving {
		c.receiving = true
		var ticker = time.NewTicker(time.Second)
		defer ticker.Stop()
	OUT:
		for ; ; <-ticker.C {
			select {
			case <-c.stopC:
				c.receiving = false
				break OUT
			default:
				c.doReceive(msgHandler, errHandler)
			}
		}
		if c.exitCall != nil {
			c.exitCall()
		}
	}
}

//停止监听
func (c *Client) StopReceive() {
	if c.receiving {
		c.stopC <- struct{}{}
	}
}

func (c *Client) doReceive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) {
	retcode, selector, err := c.api.SyncCheck()
	if err != nil {
		errHandler(err)
		if retcode == 1101 {
			go func() {
				if c.logoutCall != nil {
					c.logoutCall()
				}
				c.stopC <- struct{}{}
			}()
		}
		return
	}
	if retcode == 0 && selector != 0 {
		wxRecvMsges, err := c.api.WebWxSync()
		if err != nil {
			errHandler(err)
			return
		}

		for i := 0; i < wxRecvMsges.MsgCount; i++ {
			msgHandler(wxRecvMsges.MsgList[i])
		}
	}
	return
}
