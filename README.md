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
[入门Demo](https://github.com/timerzz/itchatgo/blob/main/demo.go)

## TODO
- [x] 登录、登出
- [x] 接收消息
- [x] 发送文字，图片
- [x] 获取联系人
- [x] 获取群成员详情
- [x] 获取头像
- [x] 登录回调，退出登录回调，停止监听回调  
- [ ] 发送文件
- [ ] 发送视频
- [ ] 群、好友变更
- [ ] 创建群聊
- [ ] 邀请入群  
- [ ] 撤回消息

## 接口

### 登录
```golang
import(
"github.com/timerzz/itchatgo"
)

func func main() {
    clientSet := itchatgo.NewClientSet()
	uuidInfo, err := cs.LoginCtl().Login()
	if err != nil{
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
}
```
登录是```LoginCtl```的```Login```函数，函数返回uuidInfo和error。其中uuidInfo包含uuid、二维码图片
的[]byte以及二维码url。可以通过类似github.com/qianlnk/qrcode的库直接通过uuidInfo中的QrUrl生成二维码。  

除此之外，还可以通过SetLoggedCall和SetLogoutCall用来设置登录成功和退出登录的回调函数。

### 更新二维码
```go
	cs.LoginCtl().ReLoadUUid()
```

### 停止登录
```go
	cs.LoginCtl().StopLogin()
```
在登录成功前，如果不停止登录，会一直轮询有没有登录成功
### 退出登录
```go
clientSet.LoginCtl().Logout()
```
### 接收消息
```go
var msgHandler = func(msg *model.WxRecvMsg) {
    if strings.Contains(msg.Content, "exit"){
        cs.LoginCtl().Logout()
    }else{
        fmt.Println(msg.Content)
    }
}

//错误处理函数
var errHandler = func(err error) {
    fmt.Println(err)
}

//开始监听
cs.MsgCtl().Receive(msgHandler, errHandler)
```
```Receive```需要传入消息的处理函数和错误处理函数

如果你不需要一直进行消息的接收，也可以使用```MsgCtl().WebWxSync()```来接收消息（可以参考Receive的实现）。

也可以通过```SetExitCall```设置退出监听时的回调函数

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

### 获取头像
```go
clientSet.ContactCtl().GetHeadImg(userName, chatRoomName,filePath) 
clientSet.ContactCtl().GetHeadImgByUser(User,filePath)
```
获取头像提供了两个方法。  
```GetHeadImg```接收三个参数：userName, chatRoomUserName, filePath。  
```如果有username而没有chatRoomUname， 就是获取用户的头像   
如果没有username而有chatRoomUname， 就是获取群的头像   
如果两个都有，就是获取群中某个用户的头像   
如果有picPath, 还会把头像保存到filePath路径
返回值是头像的[]byte和error
```
另一个方法是```GetHeadImgByUser(User,filePath)```, 可以直接传入model.User对象，这个方法
会使用User中的HedImgUrl字段去获取头像。  
**需要注意的是，这样个方法获取的头像大小不一样，GetHeadImg获取的图大，GetHeadImgByUser获取的图小**
### 获取群详情
```go
clientSet.ContactCtl().GetContactDetail()
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
