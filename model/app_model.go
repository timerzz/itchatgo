package model

import (
	"net/http"
)

type Cookies []*http.Cookie

type LoginMap struct {
	PassTicket  string
	BaseRequest BaseRequest /* 将涉及登陆有关的验证数据封装成对象 */

	SelfNickName string
	SelfUserName string

	SyncKeys   SyncKeysJsonData /* 同步消息时需要验证的Keys */
	SyncKeyStr string           /* Keys组装成的字符串 */

	Url       string
	FileUrl   string
	SyncUrl   string
	LoginTime int64
}

func (c Cookies) Get(key string) string {
	for _, cookie := range c {
		if cookie.Name == key {
			return cookie.Value
		}
	}
	return ""
}
