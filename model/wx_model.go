package model

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type LoginCallbackXMLResult struct {
	XMLName     xml.Name `xml:"error"` /* 根节点定义 */
	Ret         string   `xml:"ret"`
	Message     string   `xml:"message"`
	SKey        string   `xml:"skey"`
	WXSid       string   `xml:"wxsid"`
	WXUin       string   `xml:"wxuin"`
	PassTicket  string   `xml:"pass_ticket"`
	IsGrayscale string   `xml:"isgrayscale"`
}

type BaseRequest struct {
	Uin      string `json:"Uin"`
	Sid      string `json:"Sid"`
	SKey     string `json:"Skey"`
	DeviceID string `json:"DeviceID"`
}

/* 微信初始化时返回的大JSON，选择性地提取一些关键数据 */
type InitInfo struct {
	User     User             `json:"User"`
	SyncKeys SyncKeysJsonData `json:"SyncKey"`
}

/* 微信获取所有联系人列表时返回的大JSON */
type ContactList struct {
	MemberCount int     `json:"MemberCount"`
	MemberList  []*User `json:"MemberList"`
	Seq         int     `json:"Seq"`
}

/* 微信通用User结构，可根据需要扩展 */
type User struct {
	Uin        int64  `json:"Uin"`
	UserName   string `json:"UserName"`
	NickName   string `json:"NickName"`
	RemarkName string `json:"RemarkName"`
	Sex        int8   `json:"Sex"`
	Province   string `json:"Province"`
	City       string `json:"City"`
}

type SyncKeysJsonData struct {
	Count    int        `json:"Count"`
	SyncKeys []*SyncKey `json:"List"`
}

type SyncKey struct {
	Key int64 `json:"Key"`
	Val int64 `json:"Val"`
}

/* 设计一个构造成字符串的结构体方法 */
func (sks SyncKeysJsonData) ToString() string {
	res := make([]string, 0, sks.Count)
	for i := 0; i < sks.Count; i++ {
		res = append(res, fmt.Sprintf("%d_%d", sks.SyncKeys[i].Key, sks.SyncKeys[i].Val))
	}

	return strings.Join(res, "|")
}

type WxRecvMsges struct {
	MsgCount int              `json:"AddMsgCount"`
	MsgList  []*WxRecvMsg     `json:"AddMsgList"`
	SyncKeys SyncKeysJsonData `json:"SyncKey"`
}

type WxRecvMsg struct {
	MsgId        string `json:"MsgId"`
	FromUserName string `json:"FromUserName"`
	ToUserName   string `json:"ToUserName"`
	MsgType      int    `json:"MsgType"`
	Content      string `json:"Content"`
	CreateTime   int64  `json:"CreateTime"`
}

type WxSendMsg struct {
	Type         int    `json:"Type"`
	Content      string `json:"Content"`
	FromUserName string `json:"FromUserName"`
	ToUserName   string `json:"ToUserName"`
	LocalID      string `json:"LocalID"`
	ClientMsgId  string `json:"ClientMsgId"`
}
