package itchatgo

import (
	"fmt"
	"github.com/timerzz/itchatgo/model"
	"io/ioutil"
)

func saveFile(bytes []byte) error {
	return ioutil.WriteFile("qr.jpg", bytes, 0644)
}

func errHandler(err error) {
	fmt.Println(err)
}

func msgHandler(msg *model.WxRecvMsg) {
	if msg.MsgType == 1 {
		fmt.Println(msg.FromUserName + ":  " + msg.Content)
	}
}

func main() {
	cs := NewClientSet()
	c, _ := cs.LoginCtl().Login(saveFile, errHandler)
	<-c
	if cons, err := cs.ContactCtl().GetAllContact(); err != nil {
		fmt.Println(err)
		fmt.Println(cons)
	}
	start, _ := cs.MsgCtl().Receive(msgHandler, errHandler)
	<-start
	fmt.Println(start)
}
