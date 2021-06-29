package contact

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timerzz/itchatgo/clients/base"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/model"
	"github.com/timerzz/itchatgo/utils"
	"io/ioutil"
	"strconv"
	"time"
)

type Client struct {
	*base.Client
}

func NewClient(base *base.Client) *Client {
	return &Client{
		base,
	}
}

func (c *Client) GetAllContact() (contactMap map[string]*model.User, err error) {
	if c.Client == nil && !c.Logged {
		return nil, errors.New("未登录")
	}
	contactMap = make(map[string]*model.User)
	var seq = 0
	for {
		var users []*model.User
		users, seq, err = c.getContact(seq)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			contactMap[u.UserName] = u
		}
		c.UpdateContacts(users...)
		if seq == 0 {
			break
		}
	}
	return contactMap, nil
}

func (c *Client) getContact(seq int) (users []*model.User, reSeq int, err error) {
	urlMap := enum.InitParaEnum
	urlMap[enum.PassTicket] = c.LoginInfo.PassTicket
	urlMap[enum.R] = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	urlMap["seq"] = strconv.Itoa(seq)
	urlMap[enum.SKey] = c.LoginInfo.BaseRequest.SKey

	resp, err := c.HttpClient.Get(c.LoginInfo.Url+enum.GET_ALL_CONTACT_URL+utils.GetURLParams(urlMap), nil)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	contactList := model.ContactList{}
	err = json.Unmarshal(bodyBytes, &contactList)
	if err != nil {
		return nil, 0, err
	}
	return contactList.MemberList, seq, err
}

type contactDetailRequest struct {
	UserName        string `json:"UserName"`
	EncryChatRoomId string `json:"EncryChatRoomId"`
}

// GetContactDetail
// 获取联系人详情
// 获取群的详情主要是获取群内联系人列表
///**/
func (c *Client) GetContactDetail(users ...*model.User) (rsp model.ContactDetailResponse, err error) {
	var list = make([]contactDetailRequest, 0, len(users))
	for _, u := range users {
		list = append(list, contactDetailRequest{UserName: u.UserName, EncryChatRoomId: u.EncryChatRoomId})
	}
	urlMap := map[string]string{
		enum.PassTicket: c.LoginInfo.PassTicket,
		"type":          "ex",
		enum.R:          fmt.Sprintf("%d", time.Now().UnixNano()/1000000),
		enum.Lang:       enum.LangValue,
	}

	params := map[string]interface{}{
		"BaseRequest": c.LoginInfo.BaseRequest,
		"Count":       len(users),
		"List":        list,
	}
	err = c.HttpClient.PostJson(c.LoginInfo.Url+enum.WEB_WX_BATCH_GET_CONTACT+utils.GetURLParams(urlMap), params, &rsp)
	if err != nil {
		c.UpdateContacts(rsp.ContactList...)
	}
	return
}
