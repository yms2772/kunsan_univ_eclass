package api

import (
	"io"
	"net/http"
	"net/url"
)

type Attachment struct {
	Name string
	URL  string
}

func (a Attachment) GetURL() *url.URL {
	u, _ := url.Parse(a.URL)
	return u
}

func (a Attachment) Download() ([]byte, error) {
	resp, err := http.Get(a.URL)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return io.ReadAll(resp.Body)
}
