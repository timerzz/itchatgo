package enum

import "net/http"

const (
	JSON_HEADER = "application/json;charset=UTF-8"
)

var (
	ProcessLoginHeader = &http.Header{
		"User-Agent":     []string{USER_AGENT},
		"client-version": []string{UOS_PATCH_CLIENT_VERSION},
		"extspam":        []string{UOS_PATCH_EXTSPAM},
		"referer":        []string{"https://wx.qq.com/?&lang=zh_CN&target=t"},
	}
	DefaultHeader = &http.Header{
		"User-Agent": []string{USER_AGENT},
	}
)
