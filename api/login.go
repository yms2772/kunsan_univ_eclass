package api

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Host string

const (
	HostPortal Host = "portal.kunsan.ac.kr"
	HostTKIS   Host = "tkis.kunsan.ac.kr"
	HostEclass Host = "eclass.kunsan.ac.kr"
)

func getHostURL(h Host) *url.URL {
	uri, _ := url.Parse(string("https://" + h))
	return uri
}

func (u *user) IsLoggedIn(h Host) bool {
	return u.cookie.Cookies(getHostURL(h)) != nil
}

func (u *user) LoginTKIS() error {
	req, err := http.NewRequest("GET", getHostURL(HostTKIS).String()+"/index.do", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Set("Referer", "https://portal.kunsan.ac.kr/")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return u.getTKIS()
}

func (u *user) LoginEclass() error {
	req, err := http.NewRequest("GET", getHostURL(HostEclass).String(), nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")
	req.Header.Set("Referer", "https://portal.kunsan.ac.kr/")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return u.getEclass()
}

func (u *user) LoginPortal(id, pw string) error {
	if id == "" || pw == "" {
		return errors.New("아이디 또는 비밀번호를 입력해주세요")
	}

	values := url.Values{
		"userId":     {id},
		"loginPwd":   {pw},
		"saveuserid": {"N"},
		"firstFlag":  {"Y"},
	}

	req, err := http.NewRequest(http.MethodPost, getHostURL(HostPortal).String()+"/index.do", strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://portal.kunsan.ac.kr/intro.do?sso=ok")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	u.cookie, _ = cookiejar.New(nil)
	lastRedirect := ""

	client := u.getClient()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		lastRedirect = req.URL.Path
		return nil
	}

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if lastRedirect == "/intro.do" {
		return errors.New("학번 또는 비밀번호가 올바르지 않습니다")
	}
	u.id = id
	return nil
}
