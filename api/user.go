package api

import (
	"crypto/tls"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/http/cookiejar"
)

type User interface {
	IsLoggedIn(h string) bool
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
	GetScore() (ScoreData, error)

	getClient() *http.Client
	updateEclass() error
	updateTKIS() error
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

func (u *user) GetID() string {
	return u.id
}

func (u *user) GetName() string {
	return u.name
}

func (u *user) GetProfileImg() image.Image {
	return u.img
}
