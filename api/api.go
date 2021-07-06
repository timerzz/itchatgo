package api

import (
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/http_client"
	"github.com/timerzz/itchatgo/model"
)

type Api struct {
	httpClient *http_client.Client
	loginInfo  *model.LoginMap
}

func NewApi() *Api {
	return &Api{
		httpClient: http_client.NewHttpClient(enum.DefaultHeader),
		loginInfo:  &model.LoginMap{},
	}
}
