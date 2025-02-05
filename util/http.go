package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var transport = &http.Transport{
	MaxIdleConns:        250,
	MaxIdleConnsPerHost: 250,
	IdleConnTimeout:     120 * time.Second,
	DisableKeepAlives:   false,
	TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
}

type Request struct {
	c *http.Client
}

func NewRequest(t *http.Transport) *Request {
	if t == nil {
		t = transport
	}

	return &Request{c: &http.Client{Transport: t}}
}

func (r *Request) CustomRequest(url, method string, bytesData interface{}, headers map[string]string, u url.Values, data interface{}) ([]byte, error) {
	var (
		bys []byte
		err error
	)
	bys, err = r.noTryRequest(r.getUrl(url, u), strings.ToUpper(method), bytesData, headers)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return bys, nil
	}
	err = json.NewDecoder(bytes.NewReader(bys)).Decode(data)
	return bys, err
}

// noTryRequest 所有公用的http请求无重试
func (r *Request) noTryRequest(url, method string, bytesData interface{}, headers map[string]string) (resByte []byte, err error) {
	req, err := http.NewRequest(method, url, r.getBody(bytesData))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for key, item := range headers {
		req.Header.Set(key, item)
	}
	resp, err := r.c.Do(req)
	if err != nil || resp == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode > http.StatusMultipleChoices || resp.Body == nil {
		return nil, fmt.Errorf("request err %s", resp.Status)
	}
	resByte, err = io.ReadAll(resp.Body)
	return
}

func (r *Request) getBody(bytesData interface{}) io.Reader {
	var body = io.Reader(nil)
	switch t := bytesData.(type) {
	case []byte:
		body = bytes.NewReader(t)
	case string:
		body = strings.NewReader(t)
	case *strings.Reader:
		body = t
	case *bytes.Buffer:
		body = t
	default:
		body = nil
	}
	return body
}

func (r *Request) getUrl(u string, values url.Values) string {
	path, err := url.Parse(u)
	if err != nil {
		return u
	}

	if len(values) > 0 {
		q := path.Query()

		for s, value := range values {
			for _, v := range value {
				q.Add(s, v)
			}
		}
		path.RawQuery = q.Encode()
	}
	return path.String()
}
