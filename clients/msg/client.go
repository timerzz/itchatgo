package msg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timerzz/itchatgo/clients/base"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/model"
	"github.com/timerzz/itchatgo/utils"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	*base.Client
	c     chan struct{}
	stopC chan struct{}
}

func NewClient(base *base.Client) *Client {
	return &Client{
		Client: base,
		c:      make(chan struct{}),
		stopC:  make(chan struct{}),
	}
}

func (c *Client) Receive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) (startC, stopC chan struct{}) {
	go func() {
	OUT:
		for {
			select {
			case <-c.stopC:
				break OUT
			default:
				c.doReceive(msgHandler, errHandler)
			}
		}
		c.c <- struct{}{}
	}()
	return c.c, c.stopC
}

func (c *Client) doReceive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) {
	retcode, selector, err := c.SyncCheck()
	if err != nil {
		errHandler(err)
		if retcode == 1101 {
			go func() {
				c.Logged = false
				c.stopC <- struct{}{}
			}()
		}
		return
	}

	if retcode == 0 && selector != 0 {
		wxRecvMsges, err := c.WebWxSync()
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

func (c *Client) SyncCheck() (int64, int64, error) {
	timeStamp := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	urlMap := map[string]string{
		enum.R:         timeStamp,
		enum.SKey:      c.LoginInfo.BaseRequest.SKey,
		enum.Sid:       c.LoginInfo.BaseRequest.Sid,
		enum.Uin:       c.LoginInfo.BaseRequest.Uin,
		enum.DeviceId:  c.LoginInfo.BaseRequest.DeviceID,
		enum.SyncKey:   c.LoginInfo.SyncKeyStr,
		enum.TimeStamp: timeStamp,
	}
	c.HttpClient.Timeout = 30 * time.Second
	syncUrl := fmt.Sprintf("%s/synccheck", c.LoginInfo.SyncUrl)
	resp, err := c.HttpClient.Get(syncUrl+utils.GetURLParams(urlMap), nil)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	/* 根据正则得到selector => window.synccheck={retcode:"0",selector:"0"}*/
	reg := regexp.MustCompile(`^window.synccheck={retcode:"(\d+)",selector:"(\d+)"}$`)
	matches := reg.FindStringSubmatch(string(respBytes))

	retcode, err := strconv.ParseInt(matches[1], 10, 64) /* 取第二个数据为retcode值 */
	if err != nil {
		return 0, 0, errors.New("解析微信心跳数据失败:\n" + err.Error() + "\n" + string(respBytes))
	}

	selector, err := strconv.ParseInt(matches[2], 10, 64) /* 取第三个数据为selector值 */
	if err != nil {
		return 0, 0, errors.New("解析微信心跳数据失败:\n" + err.Error() + "\n" + string(respBytes))
	}

	if retcode != 0 {
		return retcode, selector, errors.New(fmt.Sprintf("retcode异常：%d", retcode))
	}

	return retcode, selector, nil
}

func (c *Client) WebWxSync() (wxMsges model.WxRecvMsges, err error) {
	urlMap := map[string]string{}
	urlMap[enum.Sid] = c.LoginInfo.BaseRequest.Sid
	urlMap[enum.SKey] = c.LoginInfo.BaseRequest.SKey
	urlMap[enum.PassTicket] = c.LoginInfo.PassTicket

	webWxSyncJsonData := map[string]interface{}{}
	webWxSyncJsonData["BaseRequest"] = c.LoginInfo.BaseRequest
	webWxSyncJsonData["SyncKey"] = c.LoginInfo.SyncKeys
	webWxSyncJsonData["rr"] = -time.Now().Unix()

	jsonBytes, err := json.Marshal(webWxSyncJsonData)
	if err != nil {
		return wxMsges, err
	}

	resp, err := http.Post(c.LoginInfo.Url+"/webwxsync"+utils.GetURLParams(urlMap), enum.JSON_HEADER, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return wxMsges, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return wxMsges, err
	}
	/* 解析组装消息对象 */
	err = json.Unmarshal(bodyBytes, &wxMsges)
	if err != nil {
		return wxMsges, err
	}
	for _, wx := range wxMsges.MsgList {
		if wx.MsgType == 1 {
			var body = string(bodyBytes)
			fmt.Println(body)
		}
	}

	/* 更新SyncKey */
	c.LoginInfo.SyncKeys = wxMsges.SyncKeys
	c.LoginInfo.SyncKeyStr = wxMsges.SyncKeys.ToString()

	return wxMsges, nil
}

func (c *Client) SendMsg(wxSendMsg model.WxSendMsg) error {
	urlMap := map[string]string{}
	urlMap[enum.Lang] = enum.LangValue
	urlMap[enum.PassTicket] = c.LoginInfo.PassTicket

	wxSendMsgMap := map[string]interface{}{}
	wxSendMsgMap[enum.BaseRequest] = c.LoginInfo.BaseRequest
	wxSendMsgMap["Msg"] = wxSendMsg
	wxSendMsgMap["Scene"] = 0

	jsonBytes, err := json.Marshal(wxSendMsgMap)
	if err != nil {
		return err
	}

	// TODO: 发送微信消息时暂不处理返回值
	_, err = http.Post(enum.WEB_WX_SENDMSG_URL+utils.GetURLParams(urlMap), enum.JSON_HEADER, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return err
	}

	return nil
}
