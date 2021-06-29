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
