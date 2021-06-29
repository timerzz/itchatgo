# itchatgo

itchatgo受到[itchat](https://github.com/littlecodersh/ItChat) 项目的启发，可以通过golang处理微信信息。  

itchatgo参考UOS版本微信，绕过了微信的登录限制。 

itchatgo提供了更为易扩展的接口，通过这些接口，可以极大程度的开发出极具个性化的程序。

## 使用

在golang项目中使用itchatgo

```golang 
import(
    "github.com/timerzz/itchatgo"
)
```
## TODO
- [x] 登录、登出
- [x] 接收消息
- [x] 发送文字，图片
- [x] 获取联系人
- [x] 获取群成员详情
- [ ] 发送文件
- [ ] 发送视频
- [ ] 群、好友变更

## 接口

### 登录
```golang
import(
"github.com/timerzz/itchatgo"
)

func saveFile(bytes []byte) error {
    return ioutil.WriteFile("qr.jpg", bytes, 0644)
}

func errHandler(err error) {
    fmt.Println(err)
}
func func main() {
    clientSet := itchatgo.NewClientSet()
    loginWaitC, stopC := cs.LoginCtl().Login(saveFile, errHandler)
    <-loginWaitC
}
```
登录是```LoginCtl```的```Login```函数，函数的第一个参数用于处理登录时的二维码，这里用了
```saveFile```将二维码保存到本地。你也可以自己编写处理函数，比如将二维码发送给前端，在前端显示。```Login```的第二个函数用于处理登录过程中的错误。
登录返回的两个函数一个用于等待登录完成，一个用于停止登录。
```golang
stopC <- struct{}{}
```
像这样可以停止登录过程。可以配合超时使用。

### 退出登录
```go
clientSet.LoginCtl().Logout()
```
### 接收消息
```go
func msgHandler(msg *model.WxRecvMsg) {
    if msg.MsgType == 1 {
    fmt.Println(msg.FromUserName + ":  " + msg.Content)
    }
}
...
start, _ := clientSet.MsgCtl().Receive(msgHandler, errHandler)
<-start
```
和登录类似，```Receive```需要传入消息的处理函数和错误处理函数。两个返回值，一个用于等待消息接收完毕（比如手机上退出了登录的微信，Receive就会停止退出，不再轮询），
还有一个用于停止消息的接收。

如果你不需要一直进行消息的接收，也可以使用```MsgCtl().WebWxSync()```来接收消息（可以参考Receive的实现）。


### 发送消息
```go
clientSet.MsgCtl().SendMsg(msg, toUserName)  //发送文字消息
clientSet.MsgCtl().SendImage(filePath, toUserName, MediaId)  //发送图片
```
如果```toUserName```是""，默认会发给文件传输助手。

### 获取联系人
```go
clientSet.ContactCtl().GetAllContact()
clientSet.ContactCtl().GetUserByNickName()
```
```GetAllContact```会获取所有联系人，好友和群会混在一起。

### 获取群详情
```go
clientSet.ContactCtl().GetAllContact()
```

## 特别鸣谢
[itchat4go](https://github.com/newflydd/itchat4go)  
[ItChat](https://github.com/luvletter2333/ItChat)

## 参考
[微信网页版接口分析](https://inf.news/zh-hans/tech/6e1e407bcde81fae1b8357f3963d5599.html?__cf_chl_jschl_tk__=6cab055d555c12c5d18115d76c0ec0e65fd16ad6-1624937505-0-AU0rwDgz7Pd0NCVGvFVQp91KhGVjLNUcxDdcPaUDsshTsySQnpySmYnNjzsBEUBSG_gREo8c_cruNVwIpPod80Nh8HfKyY8KGYXelKsDf2iHdSBEbxwf1cxii2bw8J09gGVBeGpZRU0QJA84UQ7naUcc9twcPXhvKGXAMfiVzTpPF68iTd_UsQ2UEFb8swVowfDjc056D3zblJnKGGMGGDau1GmjOmD4G25otOjY9J6woDTFD81H4rfVGuy1IUoiTmDFjskVRKz_YdfAkGLrnEgbSQ5UWkU2Qp_5CAEnZBWvT-Ui0Qlyj5pL8FUByf0rjoJPIL1TzlOUhkoG7KiINt2ThHhj3ktPK0KEkrQ3e1_kKjyQ9P0igSyiL1CoXhUNsuPGk8ooIpjTApdFdQqINuCq7ETrfdfe40-2dwfoGZ3yzhcb-i1fTd7OFi9sHEn4WV7uph5fqiKtTEVtVg3N7x9tCaA0LKHCHjh2I6WtnWrYwTF9D9YwTvy0cMelt-eoGYm3MHWLqQbgzGRjc0zaRd8)

## 捐赠
如果觉得这个项目有所帮助， 可以请我喝杯奶茶  
<img src="https://z3.ax1x.com/2021/06/29/RafsgK.md.jpg" width = "200" height = "200" alt="图片名称" align=center />
<img src="https://z3.ax1x.com/2021/06/29/RafE1f.jpg" width = "200" height = "200" alt="图片名称" align=center />
