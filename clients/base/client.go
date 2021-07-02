package base

import (
	"github.com/timerzz/itchatgo/http_client"
	"github.com/timerzz/itchatgo/http_client/cookiejar"
	"github.com/timerzz/itchatgo/model"
	"strings"
	"sync"
)

type Client struct {
	HttpClient  *http_client.Client
	LoginInfo   *model.LoginMap
	Self        model.User
	Logging     bool
	Logged      bool
	friends     map[string]*model.User
	chatRooms   map[string]*model.User
	contactSync sync.Mutex
}

func NewClient(httpclt *http_client.Client, loginInfo *model.LoginMap) *Client {
	return &Client{
		HttpClient: httpclt,
		LoginInfo:  loginInfo,
		friends:    make(map[string]*model.User),
		chatRooms:  make(map[string]*model.User),
	}
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
	c.LoginInfo = &model.LoginMap{}
	c.Logged = false
	c.Logging = false
	c.friends = make(map[string]*model.User)
	c.chatRooms = make(map[string]*model.User)
	c.HttpClient.Jar, _ = cookiejar.New(nil)
}
