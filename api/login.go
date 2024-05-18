package api

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (a *API) checkCookie() (User, error) {
	req, err := http.NewRequest(http.MethodGet, "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome&userDTO.localeKey=ko", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := a.Client().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	if doc.Find(`form#loginForm fieldset legend`).First().Text() != "접속자 정보" {
		return nil, errors.New("옳지 않은 계정입니다")
	}

	nameInfo := doc.Find(`form#loginForm div.mem_info span.info`).Text()
	r := regexp.MustCompile(`(\d+)\((.*?)\)님`)
	match := r.FindStringSubmatch(nameInfo)
	if len(match) != 3 {
		return nil, errors.New("접속자 정보를 가져올 수 없습니다")
	}

	userdata := &user{
		API:    a,
		id:     match[1],
		name:   match[2],
		imgURL: "https://eclass.kunsan.ac.kr" + doc.Find(`img[alt="userPhoto"]`).AttrOr("src", ""),
	}
	return userdata, nil
}

func (a *API) Login(id, pw string) (User, error) {
	if id == "" || pw == "" {
		return nil, errors.New("아이디 또는 비밀번호를 입력해주세요")
	}

	values := url.Values{
		"cmd":      {"loginUser"},
		"userId":   {id},
		"password": {pw},
		"id_save":  {"on"},
	}

	req, err := http.NewRequest(http.MethodPost, "https://eclass.kunsan.ac.kr/User.do", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	a.cookie, _ = cookiejar.New(nil)

	resp, err := a.Client().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return a.checkCookie()
}
