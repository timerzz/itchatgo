package login

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/timerzz/itchatgo/clients/base"
	"github.com/timerzz/itchatgo/clients/contact"
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
	contactCtl *contact.Client
	self       *model.User
	uuidInfo   *model.UUidInfo
	loginC     chan struct{}
	loginStopC chan struct{}
	loggedCall func() //登录成功的回调函数
	logoutCall func() //退出登录的回调函数
}

func NewClient(base *base.Client, contact *contact.Client) *Client {
	return &Client{
		Client:     base,
		contactCtl: contact,
		loginC:     make(chan struct{}),
		loginStopC: make(chan struct{}),
	}
}

func (c *Client) Login() (info *model.UUidInfo, err error) {
	if c.uuidInfo != nil {
		info = c.uuidInfo
	}
	if !c.Logging() && !c.Logged() {
		c.SetLogging(true)
		err = c.ReLoadUUid()
		info = c.uuidInfo
		go c.waitLogin()
	}
	return
}

func (c *Client) ReLoadUUid() (err error) {
	c.uuidInfo = &model.UUidInfo{}
	c.uuidInfo.UUid, err = c.GetUUID()
	if err != nil {
		return
	}
	c.uuidInfo.QrUrl = enum.QRCODE + c.uuidInfo.UUid
	c.uuidInfo.QrImg, err = c.GetQR(c.uuidInfo.UUid)
	if err != nil {
		return
	}
	return
}

func (c *Client) WaitLogin() {
	if !c.Logged() && c.Logging() {
		<-c.loginC
	}
	return
}

func (c *Client) StopLogin() {
	if !c.Logged() && c.Logging() {
		c.loginStopC <- struct{}{}
	}
}

func (c *Client) waitLogin() {
	defer func() {
		c.loginC <- struct{}{}
	}()
	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()
	for ; c.Logging(); <-ticker.C {
		select {
		case <-c.loginStopC:
			c.SetLogging(false)
			return
		default:
			status, _err := c.CheckLogin(c.uuidInfo.UUid)
			switch status {
			case 200:
				_ = c.NotifyStatus()
				_ = c.InitWX()
				c.SetLogging(false)
				c.SetLogged(true)
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

func (c *Client) GetUUID() (string, error) {
	var httpClient = c.HttpClient()

	resp, err := httpClient.Get(enum.UUID_URL+utils.GetURLParams(enum.UuidParaEnum), nil)
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
	var httpClient = c.HttpClient()
	resp, err := httpClient.Get(enum.QRCODE_URL+uuid, nil)
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
	var httpClient = c.HttpClient()
	var timestamp = time.Now().UnixNano() / 1000000
	paraMap := enum.LoginParaEnum
	paraMap[enum.UUID] = uuid
	paraMap[enum.TimeStamp] = fmt.Sprintf("%d", timestamp)
	paraMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))

	resp, err := httpClient.Get(enum.LOGIN_URL+utils.GetURLParams(paraMap), nil)
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
	var loginInfo, httpClient = c.LoginInfo(), c.HttpClient()
	reg := regexp.MustCompile(`window.redirect_uri="(\S+)";`)
	groups := reg.FindStringSubmatch(loginContent)
	if len(groups) < 1 {
		return errors.New("process login  regexp match err")
	}
	loginInfo.Url = groups[1]

	resp, err := httpClient.Get(loginInfo.Url+"&fun=new&version=v2", enum.ProcessLoginHeader)
	if err != nil {
		return errors.New("process login request err:" + err.Error())
	}
	loginInfo.Url = loginInfo.Url[:strings.LastIndex(loginInfo.Url, "/")]
	loginInfo.FileUrl, loginInfo.SyncUrl = loginInfo.Url, loginInfo.Url
	for indexUrl, detailUrl := range enum.WxURLs {
		if strings.Contains(loginInfo.Url, indexUrl) {
			urls := utils.Map(detailUrl, func(s string) string {
				return fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin", s)
			})
			loginInfo.FileUrl, loginInfo.SyncUrl = urls[0], urls[1]
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

	loginInfo.BaseRequest.SKey = loginCallbackXMLResult.SKey
	loginInfo.BaseRequest.Sid = loginCallbackXMLResult.WXSid
	loginInfo.BaseRequest.Uin = loginCallbackXMLResult.WXUin
	loginInfo.BaseRequest.DeviceID = "e" + utils.GetRandomString(10, 15)

	loginInfo.PassTicket = loginCallbackXMLResult.PassTicket
	if loginInfo.BaseRequest.SKey == "" && loginInfo.BaseRequest.Sid == "" && loginInfo.BaseRequest.Uin == "" {
		return errors.New("Your wechat account may be LIMITED to log in WEB wechat")
	}
	return nil
}

func (c *Client) NotifyStatus() error {
	var loginInfo = c.LoginInfo()
	urlMap := map[string]string{enum.PassTicket: loginInfo.PassTicket}

	statusNotifyJsonData := make(map[string]interface{}, 5)
	statusNotifyJsonData["BaseRequest"] = loginInfo.BaseRequest
	statusNotifyJsonData["Code"] = 3
	statusNotifyJsonData["FromUserName"] = loginInfo.SelfUserName
	statusNotifyJsonData["ToUserName"] = loginInfo.SelfUserName
	statusNotifyJsonData["ClientMsgId"] = time.Now().UnixNano() / 1000000

	jsonBytes, err := json.Marshal(statusNotifyJsonData)
	if err != nil {
		return err
	}
	_, err = http.Post(enum.STATUS_NOTIFY_URL+utils.GetURLParams(urlMap), enum.JSON_HEADER, bytes.NewReader(jsonBytes))

	return err
}

func (c *Client) InitWX() error {
	var loginInfo, httpClient = c.LoginInfo(), c.HttpClient()
	/* post URL */
	var urlMap = enum.InitParaEnum
	var timestamp = time.Now().UnixNano() / 1000000
	urlMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))
	urlMap[enum.PassTicket] = loginInfo.PassTicket

	/* post数据 */
	initPostJsonData := map[string]interface{}{
		"BaseRequest": loginInfo.BaseRequest,
	}
	var initInfo model.InitInfo
	err := httpClient.PostJson(loginInfo.Url+enum.INIT_URL+utils.GetURLParams(urlMap), initPostJsonData, &initInfo)
	if err != nil {
		return err
	}

	loginInfo.SelfNickName = initInfo.User.NickName
	loginInfo.SelfUserName = initInfo.User.UserName
	c.self = &initInfo.User

	/* 组装synckey */
	if initInfo.SyncKeys.Count < 1 {
		return errors.New("微信返回的数据有误")
	}
	loginInfo.SyncKeys = initInfo.SyncKeys
	loginInfo.SyncKeyStr = initInfo.SyncKeys.ToString()

	c.contactCtl.UpdateContacts(initInfo.ContactList...)
	return nil
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

func (c *Client) Logout() (err error) {
	var loginInfo, httpClient = c.LoginInfo(), c.HttpClient()
	if c.Logged() {
		url := fmt.Sprintf("%s/webwxlogout", loginInfo.Url)
		params := map[string]string{
			"redirect": "1",
			"type":     "1",
			"skey":     loginInfo.BaseRequest.SKey,
		}
		_, err = httpClient.Get(url+utils.GetURLParams(params), nil)
		if err != nil {
			return
		}
		c.uuidInfo = nil
		c.contactCtl.Clear()
		if c.logoutCall != nil {
			c.logoutCall()
		}
	}
	return
}

func (c *Client) Self() *model.User {
	return c.self
}
