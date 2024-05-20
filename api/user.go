package api

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/http/cookiejar"
	"strings"

	"github.com/nfnt/resize"
)

type User interface {
	IsLoggedIn(h Host) bool
	LoginTKIS() error
	LoginEclass() error
	LoginPortal(id string, pw string) error

	GetID() string
	GetName() string
	GetProfileImg() image.Image
	GetClassroom() (Classroom, error)
	GetPost(data Post) (PostData, error)
	GetMessage(data Message) (MessageData, error)
	GetSchedule(data Schedule) (ScheduleData, error)
	GetTimetable() ([15][7]Timetable, error)

	getClient() *http.Client
	getEclass() error
	getTKIS() error
}

type user struct {
	cookie *cookiejar.Jar
	id     string
	name   string
	img    image.Image
}

func NewUser() User {
	u := &user{}
	u.cookie, _ = cookiejar.New(nil)
	return u
}

func (u *user) getClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				},
			},
		},
		Jar: u.cookie,
	}
}

func (u *user) getEclass() error {
	return nil
}

func (u *user) getTKIS() error {
	reqStr := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><Root xmlns="http://www.nexacroplatform.com/platform/dataset"><Parameters><Parameter id="WMONID"></Parameter><Parameter id="fsp_action">xDefaultAction</Parameter><Parameter id="fsp_cmd">execute</Parameter><Parameter id="GV_USER_ID">%s</Parameter><Parameter id="GV_IP_ADDRESS"></Parameter><Parameter id="fsp_logId">all</Parameter></Parameters><Dataset id="ds_cond"><ColumnInfo><Column id="HAKBUN" type="STRING" size="255"  /></ColumnInfo><Rows><Row><Col id="HAKBUN">%s</Col></Row></Rows></Dataset><Dataset id="fsp_ds_cmd"><ColumnInfo><Column id="TX_NAME" type="string" size="100"  /><Column id="TYPE" type="string" size="10"  /><Column id="SQL_ID" type="string" size="200"  /><Column id="KEY_SQL_ID" type="string" size="200"  /><Column id="KEY_INCREMENT" type="int" size="10"  /><Column id="CALLBACK_SQL_ID" type="string" size="200"  /><Column id="INSERT_SQL_ID" type="string" size="200"  /><Column id="UPDATE_SQL_ID" type="string" size="200"  /><Column id="DELETE_SQL_ID" type="string" size="200"  /><Column id="SAVE_FLAG_COLUMN" type="string" size="200"  /><Column id="USE_INPUT" type="string" size="1"  /><Column id="USE_ORDER" type="string" size="1"  /><Column id="KEY_ZERO_LEN" type="int" size="10"  /><Column id="BIZ_NAME" type="string" size="100"  /><Column id="PAGE_NO" type="int" size="10"  /><Column id="PAGE_SIZE" type="int" size="10"  /><Column id="READ_ALL" type="string" size="1"  /><Column id="EXEC_TYPE" type="string" size="2"  /><Column id="EXEC" type="string" size="1"  /><Column id="FAIL" type="string" size="1"  /><Column id="FAIL_MSG" type="string" size="200"  /><Column id="EXEC_CNT" type="int" size="1"  /><Column id="MSG" type="string" size="200"  /></ColumnInfo><Rows><Row><Col id="TX_NAME" /><Col id="TYPE">N</Col><Col id="SQL_ID">common:DIV_INFO_STUD_S01</Col><Col id="KEY_SQL_ID" /><Col id="KEY_INCREMENT">0</Col><Col id="CALLBACK_SQL_ID" /><Col id="INSERT_SQL_ID" /><Col id="UPDATE_SQL_ID" /><Col id="DELETE_SQL_ID" /><Col id="SAVE_FLAG_COLUMN" /><Col id="USE_INPUT" /><Col id="USE_ORDER" /><Col id="KEY_ZERO_LEN">0</Col><Col id="BIZ_NAME" /><Col id="PAGE_NO" /><Col id="PAGE_SIZE" /><Col id="READ_ALL" /><Col id="EXEC_TYPE">B</Col><Col id="EXEC" /><Col id="FAIL" /><Col id="FAIL_MSG" /><Col id="EXEC_CNT">0</Col><Col id="MSG" /></Row></Rows></Dataset></Root>`,
		u.id,
		u.id,
	)

	req, err := http.NewRequest(http.MethodPost, "https://tkis.kunsan.ac.kr/XMain", strings.NewReader(reqStr))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("Referer", "https://tkis.kunsan.ac.kr/index.do")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	var data tkisResponse
	if err := xml.NewDecoder(resp.Body).Decode(&data); err != nil {
		return err
	}

	if len(data.Dataset) == 0 {
		return errors.New("정보를 가져올 수 없습니다")
	}

	for _, col := range data.Dataset[0].Rows.Row.Col {
		switch col.ID {
		case "IRUM":
			u.name = col.Text
		case "PIC":
			body, _ := base64.StdEncoding.DecodeString(col.Text)
			img, _, err := image.Decode(bytes.NewReader(body))
			if err == nil {
				img = resize.Resize(120, 0, img, resize.Lanczos3)
				u.img = img
			}
		}
	}
	return nil
}

func (u *user) GetID() string {
	return u.id
}

func (u *user) GetName() string {
	return u.name
}

func (u *user) GetProfileImg() image.Image {
	return u.img
}
