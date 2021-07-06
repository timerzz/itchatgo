package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
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

func (a *Api) GetUUID() (string, error) {
	resp, err := a.httpClient.Get(enum.UUID_URL+utils.GetURLParams(enum.UuidParaEnum), nil)
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

func (a *Api) GetQR(uuid string) ([]byte, error) {

	resp, err := a.httpClient.Get(enum.QRCODE_URL+uuid, nil)
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

func (a *Api) CheckLogin(uuid string) (int64, error) {

	var timestamp = time.Now().UnixNano() / 1000000
	paraMap := enum.LoginParaEnum
	paraMap[enum.UUID] = uuid
	paraMap[enum.TimeStamp] = fmt.Sprintf("%d", timestamp)
	paraMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))

	resp, err := a.httpClient.Get(enum.LOGIN_URL+utils.GetURLParams(paraMap), nil)
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
		if err = a.ProcessLoginInfo(body); err != nil {
			status = 400
		}
	}
	return status, err
}

func (a *Api) ProcessLoginInfo(loginContent string) (err error) {
	reg := regexp.MustCompile(`window.redirect_uri="(\S+)";`)
	groups := reg.FindStringSubmatch(loginContent)
	if len(groups) < 1 {
		return errors.New("process login  regexp match err")
	}
	a.loginInfo.Url = groups[1]

	resp, err := a.httpClient.Get(a.loginInfo.Url+"&fun=new&version=v2", enum.ProcessLoginHeader)
	if err != nil {
		return errors.New("process login request err:" + err.Error())
	}
	a.loginInfo.Url = a.loginInfo.Url[:strings.LastIndex(a.loginInfo.Url, "/")]
	a.loginInfo.FileUrl, a.loginInfo.SyncUrl = a.loginInfo.Url, a.loginInfo.Url
	for indexUrl, detailUrl := range enum.WxURLs {
		if strings.Contains(a.loginInfo.Url, indexUrl) {
			urls := utils.Map(detailUrl, func(s string) string {
				return fmt.Sprintf("https://%s/cgi-bin/mmwebwx-bin", s)
			})
			a.loginInfo.FileUrl, a.loginInfo.SyncUrl = urls[0], urls[1]
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

	a.loginInfo.BaseRequest.SKey = loginCallbackXMLResult.SKey
	a.loginInfo.BaseRequest.Sid = loginCallbackXMLResult.WXSid
	a.loginInfo.BaseRequest.Uin = loginCallbackXMLResult.WXUin
	a.loginInfo.BaseRequest.DeviceID = "e" + utils.GetRandomString(10, 15)

	a.loginInfo.PassTicket = loginCallbackXMLResult.PassTicket
	if a.loginInfo.BaseRequest.SKey == "" && a.loginInfo.BaseRequest.Sid == "" && a.loginInfo.BaseRequest.Uin == "" {
		return errors.New("Your wechat account may be LIMITED to log in WEB wechat")
	}
	return nil
}

func (a *Api) NotifyStatus() error {
	urlMap := map[string]string{enum.PassTicket: a.loginInfo.PassTicket}

	statusNotifyJsonData := make(map[string]interface{}, 5)
	statusNotifyJsonData["BaseRequest"] = a.loginInfo.BaseRequest
	statusNotifyJsonData["Code"] = 3
	statusNotifyJsonData["FromUserName"] = a.loginInfo.SelfUserName
	statusNotifyJsonData["ToUserName"] = a.loginInfo.SelfUserName
	statusNotifyJsonData["ClientMsgId"] = time.Now().UnixNano() / 1000000

	jsonBytes, err := json.Marshal(statusNotifyJsonData)
	if err != nil {
		return err
	}
	_, err = http.Post(enum.STATUS_NOTIFY_URL+utils.GetURLParams(urlMap), enum.JSON_HEADER, bytes.NewReader(jsonBytes))

	return err
}

func (a *Api) InitWX() (initInfo model.InitInfo, err error) {
	/* post URL */
	var urlMap = enum.InitParaEnum
	var timestamp = time.Now().UnixNano() / 1000000
	urlMap[enum.R] = fmt.Sprintf("%d", ^(int32)(timestamp))
	urlMap[enum.PassTicket] = a.loginInfo.PassTicket

	/* post数据 */
	initPostJsonData := map[string]interface{}{
		"BaseRequest": a.loginInfo.BaseRequest,
	}
	err = a.httpClient.PostJson(a.loginInfo.Url+enum.INIT_URL+utils.GetURLParams(urlMap), initPostJsonData, &initInfo)
	if err != nil {
		return
	}

	a.loginInfo.SelfNickName = initInfo.User.NickName
	a.loginInfo.SelfUserName = initInfo.User.UserName

	/* 组装synckey */
	if initInfo.SyncKeys.Count < 1 {
		return initInfo, errors.New("微信返回的数据有误")
	}
	a.loginInfo.SyncKeys = initInfo.SyncKeys
	a.loginInfo.SyncKeyStr = initInfo.SyncKeys.ToString()

	return
}

func (a *Api) Logout() (err error) {
	url := fmt.Sprintf("%s/webwxlogout", a.loginInfo.Url)
	params := map[string]string{
		"redirect": "1",
		"type":     "1",
		"skey":     a.loginInfo.BaseRequest.SKey,
	}
	_, err = a.httpClient.Get(url+utils.GetURLParams(params), nil)
	return
}
