package api

import (
	"net/http"
	"net/http/cookiejar"
)

type API struct {
	cookie *cookiejar.Jar
}

func New() *API {
	return &API{}
}

func (a *API) Client() *http.Client {
	return &http.Client{
		Jar: a.cookie,
	}
}
