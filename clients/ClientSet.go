package clients

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
}

func NewClientSet() *ClientSet {
	baseClt := base.NewClient(http_client.NewHttpClient(enum.DefaultHeader), &model.LoginMap{})
	cs := &ClientSet{
		msgCtl:     msg.NewClient(baseClt),
		loginCtl:   login.NewClient(baseClt),
		contactCtl: contact.NewClient(baseClt),
	}
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
