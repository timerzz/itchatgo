package enum

const (
	BASE_URL   = "https://login.weixin.qq.com"           /* API基准地址 */
	UUID_URL   = BASE_URL + "/jslogin"                   /* 获取uuid的地址 */
	QRCODE_URL = BASE_URL + "/qrcode/"                   /* 获取二维码的地址 */
	LOGIN_URL  = BASE_URL + "/cgi-bin/mmwebwx-bin/login" /* 登陆URL  */

	API_BASE_URL              = "https://wx.qq.com/cgi-bin/mmwebwx-bin" /* API基准地址 */
	INIT_URL                  = "/webwxinit"                            /* 初始化URL  */
	STATUS_NOTIFY_URL         = API_BASE_URL + "/webwxstatusnotify"     /* 通知微信状态变化 */
	GET_ALL_CONTACT_URL       = "/webwxgetcontact"                      /* 获取所有联系人信息 */
	WEB_WX_SYNC_URL           = "/webwxsync"                            /* 拉取同步消息 */
	WEB_WX_SENDMSG_URL        = "/webwxsendmsg"                         /* 发送消息 */
	WEB_WX_SENDIMG_URL        = "/webwxsendmsgimg"
	WEB_WX_SENDFILE_URL       = "/webwxsendappmsg"
	WEB_WX_SENDVIDEO_URL      = "/webwxsendvideomsg"
	WEB_WX_UPDATECHATROOM_URL = API_BASE_URL + "/webwxupdatechatroom" /* 群更新，主要用来邀请好友 */

)

var (
	WxURLs = map[string][]string{
		"wx2.qq.com":      {"file.wx2.qq.com", "webpush.wx2.qq.com"},
		"wx8.qq.com":      {"file.wx8.qq.com", "webpush.wx8.qq.com"},
		"qq.com":          {"file.wx.qq.com", "webpush.wx.qq.com"},
		"web2.wechat.com": {"file.web2.wechat.com", "webpush.web2.wechat.com"},
		"wechat.com":      {"file.web.wechat.com", "webpush.web.wechat.com"},
	}
)
