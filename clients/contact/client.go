package contact

import (
	"github.com/timerzz/itchatgo/api"
	"github.com/timerzz/itchatgo/model"
	"strings"
	"sync"
)

type Client struct {
	friends     map[string]*model.User
	chatRooms   map[string]*model.User
	contactSync sync.Mutex
	api         *api.Api
}

func NewClient(api *api.Api) *Client {
	return &Client{
		api:       api,
		friends:   make(map[string]*model.User),
		chatRooms: make(map[string]*model.User),
	}
}

func (c *Client) GetAllContacts() (contacts []*model.User, err error) {
	contacts, err = c.api.GetAllContact()
	if err != nil {
		return
	}
	c.UpdateContacts(contacts...)
	return
}

func (c *Client) GetContactDetail(contacts ...*model.User) ([]*model.User, error) {
	rsp, err := c.api.GetContactDetail(contacts...)
	if err != nil {
		return nil, err
	}
	c.UpdateContacts(rsp.ContactList...)
	return rsp.ContactList, err
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
	c.friends = make(map[string]*model.User)
	c.chatRooms = make(map[string]*model.User)
}
