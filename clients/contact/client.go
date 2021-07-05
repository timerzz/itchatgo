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
	"strings"
	"sync"
	"time"
)

type Client struct {
	base        *base.Client
	friends     map[string]*model.User
	chatRooms   map[string]*model.User
	contactSync sync.Mutex
}

func NewClient(base *base.Client) *Client {
	return &Client{
		base:      base,
		friends:   make(map[string]*model.User),
		chatRooms: make(map[string]*model.User),
	}
}

func (c *Client) GetAllContact() (contactMap map[string]*model.User, err error) {
	if c.base == nil && !c.base.Logged() {
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
	var loginInfo, httpClient = c.base.LoginInfo(), c.base.HttpClient()
	urlMap := enum.InitParaEnum
	urlMap[enum.PassTicket] = loginInfo.PassTicket
	urlMap[enum.R] = fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	urlMap["seq"] = strconv.Itoa(seq)
	urlMap[enum.SKey] = loginInfo.BaseRequest.SKey

	resp, err := httpClient.Get(loginInfo.Url+enum.GET_ALL_CONTACT_URL+utils.GetURLParams(urlMap), nil)
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
// 主要需要Username字段
///**/
func (c *Client) GetContactDetail(users ...*model.User) (rsp model.ContactDetailResponse, err error) {
	var loginInfo, httpClient = c.base.LoginInfo(), c.base.HttpClient()
	var list = make([]contactDetailRequest, 0, len(users))
	for _, u := range users {
		list = append(list, contactDetailRequest{UserName: u.UserName, EncryChatRoomId: u.EncryChatRoomId})
	}
	urlMap := map[string]string{
		enum.PassTicket: loginInfo.PassTicket,
		"type":          "ex",
		enum.R:          fmt.Sprintf("%d", time.Now().UnixNano()/1000000),
		enum.Lang:       enum.LangValue,
	}

	params := map[string]interface{}{
		"BaseRequest": loginInfo.BaseRequest,
		"Count":       len(users),
		"List":        list,
	}
	err = httpClient.PostJson(loginInfo.Url+enum.WEB_WX_BATCH_GET_CONTACT+utils.GetURLParams(urlMap), params, &rsp)
	if err != nil {
		c.UpdateContacts(rsp.ContactList...)
	}
	return
}

//如果有username而没有chatRoomUname， 就是获取用户的头像
//如果没有username而有chatRoomUname， 就是获取群的头像
//如果两个都有，就是获取群中某个用户的头像
//如果有picPath, 还会把头像保存到这个路径
func (c *Client) GetHeadImg(username, chatRoomUname, picPath string) (pic []byte, err error) {
	var loginInfo, httpClient = c.base.LoginInfo(), c.base.HttpClient()
	var uname = ""
	if username != "" {
		uname = username
	} else if chatRoomUname != "" {
		uname = chatRoomUname
	} else {
		uname = loginInfo.SelfUserName
	}

	var params = map[string]string{
		"userName": uname,
		"skey":     loginInfo.BaseRequest.SKey,
		"type":     "big",
	}
	var url = loginInfo.Url + enum.WEB_WX_GETICON
	if username == "" && chatRoomUname != "" {
		url = loginInfo.Url + enum.WEB_WX_HEADIMG
	}
	if chatRoomUname != "" && username != "" {
		chatRoom := c.GetChatRoomByUname(chatRoomUname)
		if chatRoom != nil {
			if chatRoom.EncryChatRoomId != "" {
				params["chatroomid"] = chatRoom.EncryChatRoomId
			} else {
				params["chatroomid"] = chatRoomUname
			}
		}
	}

	rsp, err := httpClient.Get(url+utils.GetURLParams(params), nil)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	if pic, err = ioutil.ReadAll(rsp.Body); err != nil {
		return
	}
	if picPath != "" {
		err = ioutil.WriteFile(picPath, pic, 0644)
	}
	return
}

func (c *Client) GetHeadImgByUser(user *model.User, picPath string) (pic []byte, err error) {
	var httpClient = c.base.HttpClient()
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if user.HeadImgUrl != "" {
		rsp, _err := httpClient.Get("https://wx.qq.com"+user.HeadImgUrl, nil)
		if _err != nil {
			return nil, _err
		}

		defer rsp.Body.Close()
		if pic, err = ioutil.ReadAll(rsp.Body); err != nil {
			return
		}
		if picPath != "" {
			err = ioutil.WriteFile(picPath, pic, 0644)
		}
		return
	}
	if user.UserName != "" {
		return c.GetHeadImg(user.UserName, "", picPath)
	}
	return nil, errors.New("user has no HeadImgUrl or UserName")
}

func (c *Client) GetUserByUname(uname string) *model.User {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	if user, ok := c.friends[uname]; ok {
		return user
	}
	return nil
}

func (c *Client) GetUserByNickName(nickName string) (u *model.User) {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	for _, user := range c.friends {
		if user.NickName == nickName {
			return user
		}
	}
	return
}

func (c *Client) GetChatRoomByUname(uname string) *model.User {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	if user, ok := c.chatRooms[uname]; ok {
		return user
	}
	return nil
}

func (c *Client) GetChatRoomByNickName(nickName string) *model.User {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	for _, user := range c.chatRooms {
		if user.NickName == nickName {
			return user
		}
	}
	return nil
}

func (c *Client) SearchChatRoom(name string) (rsp []*model.User) {
	for _, r := range c.chatRooms {
		if strings.Contains(r.NickName, name) {
			rsp = append(rsp, r)
			continue
		}
	}
	return
}

func (c *Client) SearchCFriends(name string) (rsp []*model.User) {
	for _, r := range c.friends {
		if strings.Contains(r.NickName, name) {
			rsp = append(rsp, r)
			continue
		}
	}
	return
}

func (c *Client) UpdateContacts(contacts ...*model.User) {
	for _, con := range contacts {
		if strings.HasPrefix(con.UserName, "@@") {
			c.UpdateChatRoom(con)
		} else if strings.HasPrefix(con.UserName, "@") {
			c.UpdateUser(con)
		}
	}
}

func (c *Client) UpdateUser(user *model.User) {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	if _, ok := c.friends[user.UserName]; ok {
		c.friends[user.UserName] = user
		return
	}
	c.friends[user.UserName] = user
}

func (c *Client) UpdateChatRoom(room *model.User) {
	c.contactSync.Lock()
	defer c.contactSync.Unlock()
	if r, ok := c.chatRooms[room.UserName]; ok {
		var m = r.MemberMap
		for _, u := range room.MemberList {
			if _, ok = m[u.UserName]; ok {
				if u.NickName != "" {
					m[u.UserName] = u
				}
				continue
			}
			m[u.UserName] = u
		}
		return
	}
	room.MemberMap = room.GenMemberMap()
	c.chatRooms[room.UserName] = room
}

func (c *Client) Clear() {
	c.base.Clear()
	c.friends = make(map[string]*model.User)
	c.chatRooms = make(map[string]*model.User)
}
