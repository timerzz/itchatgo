package itchatgo

import (
	"github.com/timerzz/itchatgo/clients/base"
	"github.com/timerzz/itchatgo/clients/contact"
	"github.com/timerzz/itchatgo/clients/login"
	"github.com/timerzz/itchatgo/clients/msg"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/http_client"
	"github.com/timerzz/itchatgo/model"
)

type ClientSet struct {
	msgCtl     *msg.Client
	loginCtl   *login.Client
	contactCtl *contact.Client
	base       *base.Client
}

func NewClientSet() *ClientSet {
	baseClt := base.NewClient(http_client.NewHttpClient(enum.DefaultHeader), &model.LoginMap{})
	cs := &ClientSet{
		msgCtl:     msg.NewClient(baseClt),
		contactCtl: contact.NewClient(baseClt),
		base:       baseClt,
	}
	cs.loginCtl = login.NewClient(baseClt, cs.contactCtl)
	return cs
}

func (cs *ClientSet) LoginCtl() *login.Client {
	return cs.loginCtl
}

func (cs *ClientSet) MsgCtl() *msg.Client {
	return cs.msgCtl
}

func (cs *ClientSet) ContactCtl() *contact.Client {
	return cs.contactCtl
}

func (cs *ClientSet) Self() (u *model.User) {
	if cs != nil && cs.base != nil {
		u = cs.LoginCtl().Self()
	}
	return
}
