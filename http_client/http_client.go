package http_client

import (
	"bytes"
	"encoding/json"
	"github.com/timerzz/itchatgo/enum"
	"github.com/timerzz/itchatgo/http_client/cookiejar"
	"io"
	"io/ioutil"
	"net/http"
)

type Client struct {
	http.Client
	header *http.Header
}

func NewHttpClient(header *http.Header) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{http.Client{Jar: jar}, header}
}

func (c *Client) Get(url string, header *http.Header) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Close = true
	if header != nil {
		req.Header = *header
	} else {
		req.Header = *c.header
	}
	return c.Do(req)
}

func (c *Client) Post(url string, body io.Reader, header *http.Header) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Close = true
	if header != nil {
		req.Header = *header
	} else {
		req.Header = *c.header
	}
	return c.Do(req)
}

func (c *Client) PostJson(url string, params, response interface{}) error {
	body, err := json.Marshal(&params)
	if err != nil {
		return err
	}
	res, err := c.Client.Post(url, enum.JSON_HEADER, bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if body, err = ioutil.ReadAll(res.Body); err != nil {
		return err
	}
	return json.Unmarshal(body, &response)
}
