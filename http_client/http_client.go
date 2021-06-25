package http_client

import (
	"io"
	"net/http"
	"net/http/cookiejar"
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
	if header != nil {
		req.Header = *header
	} else {
		req.Header = *c.header
	}
	return c.Do(req)
}
