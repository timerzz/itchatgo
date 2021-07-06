package itchatgo

import (
	"github.com/timerzz/itchatgo/api"
	"github.com/timerzz/itchatgo/clients/contact"
	"github.com/timerzz/itchatgo/clients/login"
	"github.com/timerzz/itchatgo/clients/msg"
	"github.com/timerzz/itchatgo/model"
)

type ClientSet struct {
	msgCtl     *msg.Client
	loginCtl   *login.Client
	contactCtl *contact.Client
	api        *api.Api
}

func NewClientSet() *ClientSet {
	cs := &ClientSet{
		api: api.NewApi(),
	}
	cs.contactCtl = contact.NewClient(cs.api)
	cs.loginCtl = login.NewClient(cs.api, cs.contactCtl)
	cs.msgCtl = msg.NewClient(cs.api, func() {
		cs.LoginCtl().SetLogged(false)
	})
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

func (cs *ClientSet) Api() *api.Api {
	return cs.api
}

func (cs *ClientSet) Self() (u *model.User) {
	if cs != nil && cs.LoginCtl() != nil {
		u = cs.LoginCtl().Self()
	}
	return
}
