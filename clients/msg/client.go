package msg

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timerzz/itchatgo/clients/base"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/model"
	"github.com/timerzz/itchatgo/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

const (
	chunkSize = 524288
)

type Client struct {
	*base.Client
	c     chan struct{}
	stopC chan struct{}
}

func NewClient(base *base.Client) *Client {
	return &Client{
		Client: base,
		c:      make(chan struct{}),
		stopC:  make(chan struct{}),
	}
}

func (c *Client) Receive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) (startC, stopC chan struct{}) {
	go func() {
	OUT:
		for {
			select {
			case <-c.stopC:
				break OUT
			default:
				c.doReceive(msgHandler, errHandler)
			}
		}
		c.c <- struct{}{}
	}()
	return c.c, c.stopC
}

func (c *Client) doReceive(msgHandler func(*model.WxRecvMsg), errHandler func(error)) {
	retcode, selector, err := c.SyncCheck()
	if err != nil {
		errHandler(err)
		if retcode == 1101 {
			go func() {
				c.Logged = false
				c.stopC <- struct{}{}
			}()
		}
		return
	}

	if retcode == 0 && selector != 0 {
		wxRecvMsges, err := c.WebWxSync()
		if err != nil {
			errHandler(err)
			return
		}

		for i := 0; i < wxRecvMsges.MsgCount; i++ {
			msgHandler(wxRecvMsges.MsgList[i])
		}
	}
	return
}

func (c *Client) SyncCheck() (int64, int64, error) {
	timeStamp := fmt.Sprintf("%d", time.Now().UnixNano()/1000000)
	urlMap := map[string]string{
		enum.R:         timeStamp,
		enum.SKey:      c.LoginInfo.BaseRequest.SKey,
		enum.Sid:       c.LoginInfo.BaseRequest.Sid,
		enum.Uin:       c.LoginInfo.BaseRequest.Uin,
		enum.DeviceId:  c.LoginInfo.BaseRequest.DeviceID,
		enum.SyncKey:   c.LoginInfo.SyncKeyStr,
		enum.TimeStamp: timeStamp,
	}
	c.HttpClient.Timeout = 30 * time.Second
	syncUrl := fmt.Sprintf("%s/synccheck", c.LoginInfo.SyncUrl)
	resp, err := c.HttpClient.Get(syncUrl+utils.GetURLParams(urlMap), nil)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	/* 根据正则得到selector => window.synccheck={retcode:"0",selector:"0"}*/
	reg := regexp.MustCompile(`^window.synccheck={retcode:"(\d+)",selector:"(\d+)"}$`)
	matches := reg.FindStringSubmatch(string(respBytes))

	retcode, err := strconv.ParseInt(matches[1], 10, 64) /* 取第二个数据为retcode值 */
	if err != nil {
		return 0, 0, errors.New("解析微信心跳数据失败:\n" + err.Error() + "\n" + string(respBytes))
	}

	selector, err := strconv.ParseInt(matches[2], 10, 64) /* 取第三个数据为selector值 */
	if err != nil {
		return 0, 0, errors.New("解析微信心跳数据失败:\n" + err.Error() + "\n" + string(respBytes))
	}

	if retcode != 0 {
		return retcode, selector, errors.New(fmt.Sprintf("retcode异常：%d", retcode))
	}

	return retcode, selector, nil
}

func (c *Client) WebWxSync() (wxMsges model.WxRecvMsges, err error) {
	urlMap := map[string]string{
		enum.Sid:        c.LoginInfo.BaseRequest.Sid,
		enum.SKey:       c.LoginInfo.BaseRequest.SKey,
		enum.PassTicket: c.LoginInfo.PassTicket,
	}

	webWxSyncJsonData := map[string]interface{}{
		"BaseRequest": c.LoginInfo.BaseRequest,
		"SyncKey":     c.LoginInfo.SyncKeys,
		"rr":          -time.Now().Unix(),
	}

	err = c.HttpClient.PostJson(c.LoginInfo.Url+enum.WEB_WX_SYNC_URL+utils.GetURLParams(urlMap), webWxSyncJsonData, &wxMsges)
	if err != nil {
		return
	}

	/* 更新SyncKey */
	c.LoginInfo.SyncKeys = wxMsges.SyncKeys
	c.LoginInfo.SyncKeyStr = wxMsges.SyncKeys.ToString()

	return
}

func (c *Client) SendRawMsg(wxSendMsg model.WxSendMsg) (rsp model.SendResponse, err error) {
	urlMap := map[string]string{
		enum.Lang:       enum.LangValue,
		enum.PassTicket: c.LoginInfo.PassTicket,
	}
	wxSendMsgMap := map[string]interface{}{
		enum.BaseRequest: c.LoginInfo.BaseRequest,
		"Msg":            wxSendMsg,
		"Scene":          0,
	}
	urlPath := enum.WEB_WX_SENDMSG_URL
	switch wxSendMsg.Type {
	case 1:
	case 3:
		urlPath = enum.WEB_WX_SENDIMG_URL
		urlMap[enum.Fun], urlMap["f"] = "async", "json"
	case 6:
		urlPath = enum.WEB_WX_SENDFILE_URL
		urlMap[enum.Fun], urlMap["f"] = "async", "json"
	case 43:
		urlPath = enum.WEB_WX_SENDVIDEO_URL
		urlMap[enum.Fun], urlMap["f"] = "async", "json"
	}
	err = c.HttpClient.PostJson(c.LoginInfo.Url+urlPath+utils.GetURLParams(urlMap), wxSendMsgMap, &rsp)
	return
}

func (c *Client) SendMsg(msg, toUserName string) (rsp model.SendResponse, err error) {
	if toUserName == "" {
		toUserName = enum.FileHelper
	}
	var id = fmt.Sprintf("%d", time.Now().Unix())
	return c.SendRawMsg(model.WxSendMsg{
		Type:         1,
		Content:      msg,
		FromUserName: c.LoginInfo.SelfUserName,
		ToUserName:   toUserName,
		LocalID:      id,
		ClientMsgId:  id,
	})
}

type fileInfos struct {
	filePath string
	fileSize int64
	fileMD5  string
	file     *os.File
}

func (c *Client) UploadFile(filePath, toUserName string, isPic, isVideo bool) (rsp model.UploadResponse, err error) {
	f, err := prepareFile(filePath)
	defer f.file.Close()
	if err != nil {
		return rsp, err
	}
	symbol := "doc"
	if isPic {
		symbol = "pic"
	} else if isVideo {
		symbol = "video"
	}
	uploadMediaRequest := map[string]interface{}{
		"UploadType":    2,
		"BaseRequest":   c.LoginInfo.BaseRequest,
		"ClientMediaId": time.Now().Unix(),
		"TotalLen":      f.fileSize,
		"StartPos":      0,
		"DataLen":       f.fileSize,
		"MediaType":     4,
		"FromUserName":  c.LoginInfo.SelfUserName,
		"ToUserName":    toUserName,
		"FileMd5":       f.fileMD5,
	}
	uploadbyte, err := json.Marshal(&uploadMediaRequest)
	if err != nil {
		return rsp, err
	}
	chunks := (f.fileSize-1)/chunkSize + 1
	rsp = model.UploadResponse{BaseResponse: model.BaseResponse{Ret: -1005}}
	for chunk := int64(1); chunk <= chunks; chunk++ {
		if rsp, err = c.uploadChunkFile(symbol, f, chunk, chunks, uploadbyte); err != nil {
			return
		}
	}
	if rsp.Ret != 0 {
		err = errors.New(fmt.Sprintf("上传文件失败，Ret:%d, ErrMsg:%s", rsp.Ret, rsp.ErrMsg))
	}
	return
}

func (c *Client) uploadChunkFile(symbol string, f fileInfos, chunkNum, chunkTotal int64, uploadMediaRequest []byte) (rsp model.UploadResponse, err error) {
	fileName := path.Base(f.filePath)

	var body = &bytes.Buffer{}
	w := multipart.NewWriter(body)
	defer w.Close()
	contentType := w.FormDataContentType()
	var chunk = make([]byte, chunkSize)
	var n int
	if n, err = f.file.ReadAt(chunk, (chunkNum-1)*chunkSize); err != nil && err != io.EOF {
		return
	}
	pa, _ := w.CreateFormFile("filename", fileName)
	if _, err = pa.Write(chunk[:n]); err != nil {
		return
	}
	var cookies = c.HttpClient.Jar.Cookies(nil)
	var dataTicket = ""
	for _, cookie := range cookies {
		if cookie.Name == "webwx_data_ticket" {
			dataTicket = cookie.Value
			break
		}
	}
	if dataTicket == "" {
		err = errors.New("webwx_data_ticket is null")
		return
	}
	for k, v := range map[string]string{
		"id":                 "WU_FILE_0",
		"name":               fileName,
		"type":               "application/octet-stream",
		"lastModifiedDate":   time.Now().String(),
		"size":               fmt.Sprintf("%d", f.fileSize),
		"mediatype":          symbol,
		"uploadmediarequest": string(uploadMediaRequest),
		"webwx_data_ticket":  dataTicket,
		"pass_ticket":        c.LoginInfo.PassTicket,
	} {
		w.WriteField(k, v)
	}
	if chunkTotal > 1 {
		w.WriteField("chunk", fmt.Sprintf("%d", chunkNum))
		w.WriteField("chunks", fmt.Sprintf("%d", chunkTotal))
	}
	res, err := c.HttpClient.Post(c.LoginInfo.Url+"/webwxuploadmedia?f=json", body, &http.Header{"Content-Type": []string{contentType}})
	if err != nil {
		return
	}
	defer res.Body.Close()
	var b []byte
	if b, err = ioutil.ReadAll(res.Body); err != nil {
		return
	}
	err = json.Unmarshal(b, &rsp)
	return
}

func prepareFile(filePath string) (_file fileInfos, err error) {
	if _file.file, err = os.Open(filePath); err != nil {
		return
	}
	md5h := md5.New()
	if _file.fileSize, err = io.Copy(md5h, _file.file); err != nil {
		return
	}
	_file.fileMD5 = fmt.Sprintf("md5:%x", md5h.Sum([]byte{}))
	_, err = _file.file.Seek(0, 0)
	return
}

// SendImage
// 如果有mediaId，就优先使用mediaId
// 如果mediaId是空的，就会先上传
// 如果toUserName 为空， 那么默认会发送给文件助手
///**
func (c *Client) SendImage(filePath string, toUserName string, mediaId string) error {
	if toUserName == "" {
		toUserName = enum.FileHelper
	}
	if mediaId == "" {
		rsp, err := c.UploadFile(filePath, toUserName, true, false)
		if err != nil {
			return err
		}
		mediaId = rsp.MediaId
	}
	id := fmt.Sprintf("%d", time.Now().Unix())
	_, err := c.SendRawMsg(model.WxSendMsg{
		Type:         3,
		FromUserName: c.LoginInfo.SelfUserName,
		ToUserName:   toUserName,
		LocalID:      id,
		ClientMsgId:  id,
		MediaId:      mediaId,
	})
	return err
}
