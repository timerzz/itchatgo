package enum

const (
	APPID       = "appid"
	FUN         = "fun"
	Lang        = "lang"
	LangValue   = "zh_CN"
	TimeStamp   = "_"
	UUID        = "uuid"
	R           = "r"
	RedirectUrl = "redirect_uri"

	/* 以下信息会存储在loginMap中 */
	Ret          = "ret"
	Message      = "message"
	SKey         = "skey"
	WXSid        = "wxsid"
	WXUin        = "wxuin"
	PassTicket   = "pass_ticket"
	IsGrayscale  = "isgrayscale"
	DeviceID     = "DeviceID"
	SelfUserName = "UserName"
	SelfNickName = "NickName"
	SyncKeyStr   = "synckeystr"
	Fun          = "fun"

	Sid         = "sid"
	Uin         = "uin"
	DeviceId    = "deviceid"
	SyncKey     = "synckey"
	BaseRequest = "BaseRequest"
)

var (
	UuidParaEnum = map[string]string{
		APPID:       "wx782c26e4c19acffb",
		FUN:         "new",
		Lang:        LangValue,
		RedirectUrl: "https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?mod=desktop",
	}

	LoginParaEnum = map[string]string{
		"loginicon": "true",
		"tip":       "0",
		UUID:        "",
		R:           "",
		TimeStamp:   ""}

	InitParaEnum = map[string]string{
		R:          "",
		Lang:       LangValue,
		PassTicket: "",
	}
)
