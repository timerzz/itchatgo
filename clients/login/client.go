package login

import (
	"encoding/json"
	"encoding/xml"
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
	base.Client
	logging    bool
	logged     bool
	loginC     chan struct{}
	loginStopC chan struct{}
}

func NewClient(base *base.Client) *Client {
	return &Client{
		Client:     *base,
		loginC:     make(chan struct{}),
		loginStopC: make(chan struct{}),
	}
}

func (c *Client) Login(qrHandler func([]byte) error, errHandler func(error)) (loginC, stopC chan struct{}) {
	go func() {
		defer func() {
			c.loginC <- struct{}{}
		}()
		c.logging = true
		uuid, err := c.GetUUID()
		if err != nil {
			errHandler(err)
			return
		}
		qr, err := c.GetQR(uuid)
		if err != nil {
			errHandler(err)
			return
		}
		go func() {
			if err = qrHandler(qr); err != nil {
				errHandler(err)
				return
			}
		}()
		ticker := time.NewTicker(time.Second * 2)
		defer ticker.Stop()
		for ; c.logging; <-ticker.C {
			select {
			case <-c.loginStopC:
				c.logging = false
				return
			default:
				status, _err := c.CheckLogin(uuid)
				switch status {
				case 200:
					_ = c.NotifyStatus()
					_ = c.InitWX()
					c.logging, c.logged = false, true
					return
				default:
					if _err != nil {
						errHandler(_err)
					}
				}
			}
		}
	}()
	return c.loginC, c.loginStopC
}

func (c *Client) GetUUID() (string, error) {
	resp, err := c.HttpClient.Get(enum.UUID_URL+utils.GetURLParams(enum.UuidParaEnum), nil)
	if err != nil {
		return "", errors.New("Faild to access the WeChat API:" + err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Failed to get WeChat API feedback UUID data:" + err.Error())
	}

	reg := regexp.MustCompile(`^window.QRLogin.code = (\d+); window.QRLogin.uuid = "(\S+)";$`)
	matches := reg.FindStringSubmatch(string(bodyBytes))
	if len(matches) != 3 {
		return "", errors.New("Failed to parse WeChat UUID API data:" + string(bodyBytes))
	}
	status, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return "", errors.New("Failed to parse WeChat UUID API data:" + err.Error())
	}

	if status != 200 {
		return "", errors.New(fmt.Sprintf("WeChat return status error, please troubleshoot network failure. status:%d", status))
	}

	return matches[2], nil
}

func (c *Client) GetQR(uuid string) ([]byte, error) {
	resp, err := c.HttpClient.Get(enum.QRCODE_URL+uuid, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	qr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return qr, nil
}

func (c *Client) CheckLogin(uuid string) (int64, error) {
	var timestamp = time.Now().UnixNano() / 1000000
	paraMap := enum.LoginParaEnum
	paraMap[enum.UUID] = uuid
	paraMap[enum.TimeStamp] = fmt.Sprintf("%d", timestamp)
	paraMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))

	resp, err := c.HttpClient.Get(enum.LOGIN_URL+utils.GetURLParams(paraMap), nil)
	if err != nil {
		return 0, errors.New("访问微信服务器API失败:" + err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.New("获取微信API反馈登陆数据失败:" + err.Error())
	}
	var body = string(bodyBytes)
	reg := regexp.MustCompile(`^window.code=(\d+);`)
	matches := reg.FindStringSubmatch(body)
	if len(matches) < 2 {
		return 0, errors.New("预期的返回结果格式不匹配")
	}

	status, err := strconv.ParseInt(matches[1], 10, 64)
	if status == 200 {
		if err = c.ProcessLoginInfo(body); err != nil {
			status = 400
		}
	}
	return status, err
}

func (c *Client) ProcessLoginInfo(loginContent string) (err error) {
	reg := regexp.MustCompile(`window.redirect_uri="(\S+)";`)
	groups := reg.FindStringSubmatch(loginContent)
	if len(groups) < 1 {
		return errors.New("process login  regexp match err")
	}
	c.LoginInfo.Url = groups[1]

	resp, err := c.HttpClient.Get(c.LoginInfo.Url+"&fun=new&version=v2", enum.ProcessLoginHeader)
	if err != nil {
		return errors.New("process login request err:" + err.Error())
	}
	c.LoginInfo.Url = c.LoginInfo.Url[:strings.LastIndex(c.LoginInfo.Url, "/")]
	c.LoginInfo.FileUrl, c.LoginInfo.SyncUrl = c.LoginInfo.Url, c.LoginInfo.Url
	for indexUrl, detailUrl := range enum.WxURLs {
		if strings.Contains(c.LoginInfo.Url, indexUrl) {
			urls := utils.Map(detailUrl, func(s string) string {
				return fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin", s)
			})
			c.LoginInfo.FileUrl, c.LoginInfo.SyncUrl = urls[0], urls[1]
			break
		}
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("获取微信登陆回调URL数据失败:" + err.Error())
	}

	loginCallbackXMLResult := model.LoginCallbackXMLResult{}
	err = xml.Unmarshal(bodyBytes, &loginCallbackXMLResult)

	c.LoginInfo.BaseRequest.SKey = loginCallbackXMLResult.SKey
	c.LoginInfo.BaseRequest.Sid = loginCallbackXMLResult.WXSid
	c.LoginInfo.BaseRequest.Uin = loginCallbackXMLResult.WXUin
	c.LoginInfo.BaseRequest.DeviceID = "e" + utils.GetRandomString(10, 15)

	c.LoginInfo.PassTicket = loginCallbackXMLResult.PassTicket
	if c.LoginInfo.BaseRequest.SKey == "" && c.LoginInfo.BaseRequest.Sid == "" && c.LoginInfo.BaseRequest.Uin == "" {
		return errors.New("Your wechat account may be LIMITED to log in WEB wechat")
	}
	return nil
}

func (c *Client) NotifyStatus() error {
	fmt.Println("通知状态变更")
	urlMap := map[string]string{enum.PassTicket: c.LoginInfo.PassTicket}

	statusNotifyJsonData := make(map[string]interface{}, 5)
	statusNotifyJsonData["BaseRequest"] = c.LoginInfo.BaseRequest
	statusNotifyJsonData["Code"] = 3
	statusNotifyJsonData["FromUserName"] = c.LoginInfo.SelfUserName
	statusNotifyJsonData["ToUserName"] = c.LoginInfo.SelfUserName
	statusNotifyJsonData["ClientMsgId"] = time.Now().UnixNano() / 1000000

	jsonBytes, err := json.Marshal(statusNotifyJsonData)
	if err != nil {
		return err
	}

	_, err = http.Post(enum.STATUS_NOTIFY_URL+utils.GetURLParams(urlMap), enum.JSON_HEADER, strings.NewReader(string(jsonBytes)))

	return err
}

func (c *Client) InitWX() error {
	/* post URL */
	var urlMap = enum.InitParaEnum
	var timestamp = time.Now().UnixNano() / 1000000
	urlMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))
	urlMap[enum.PassTicket] = c.LoginInfo.PassTicket

	/* post数据 */
	initPostJsonData := map[string]interface{}{}
	initPostJsonData["BaseRequest"] = c.LoginInfo.BaseRequest

	jsonBytes, err := json.Marshal(initPostJsonData)
	if err != nil {
		return err
	}
	initUrl := c.LoginInfo.Url + "/webwxinit"
	resp, err := http.Post(initUrl+utils.GetURLParams(urlMap), enum.JSON_HEADER, strings.NewReader(string(jsonBytes)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	initInfo := model.InitInfo{}
	err = json.Unmarshal(bodyBytes, &initInfo)
	if err != nil {
		return errors.New("无法解析JSON至InitInfo对象:" + err.Error())
	}

	c.LoginInfo.SelfNickName = initInfo.User.NickName
	c.LoginInfo.SelfUserName = initInfo.User.UserName

	/* 组装synckey */
	if initInfo.SyncKeys.Count < 1 {
		fmt.Println(string(bodyBytes))
		return errors.New("微信返回的数据有误")
	}
	c.LoginInfo.SyncKeys = initInfo.SyncKeys
	c.LoginInfo.SyncKeyStr = initInfo.SyncKeys.ToString()

	return nil
}

func (c *Client) Logging() bool {
	return c.logging
}

func (c *Client) Logged() bool {
	return c.logged
}
