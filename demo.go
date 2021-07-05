package itchatgo

import (
	"fmt"
	"github.com/qianlnk/qrcode"
	"github.com/timerzz/itchatgo/model"
	"strings"
	"time"
)

func main() {
	cs := NewClientSet()
	//登录并获得登录的QR相关信息
	uuidInfo, err := cs.LoginCtl().SetTimeout(3 * time.Minute).Login()
	if err != nil {
		fmt.Println(err)
		return
	}
	//设置登录成功和退出登录的回调函数
	cs.LoginCtl().SetLoggedCall(func() {
		fmt.Println("登录成功")
	})
	cs.LoginCtl().SetLogoutCall(func() {
		fmt.Println("退出登录")
	})

	//将二维码显示在终端
	qr := qrcode.NewQRCode(uuidInfo.QrUrl, true)
	qr.Output()

	//等待登录成功
	cs.LoginCtl().WaitLogin()

	//设置退出监听时的回调
	cs.MsgCtl().SetExitCall(func() {
		fmt.Println("停止监听")
	})

	//监听时的数据处理函数
	var msgHandler = func(msg *model.WxRecvMsg) {
		if strings.Contains(msg.Content, "exit") {
			cs.LoginCtl().Logout()
		} else {
			fmt.Println(msg.Content)
		}
	}

	//错误处理函数
	var errHandler = func(err error) {
		fmt.Println(err)
	}

	//开始监听
	cs.MsgCtl().Receive(msgHandler, errHandler)

}
