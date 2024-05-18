package api

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

type User interface {
	GetID() string
	GetName() string
	GetProfileImg() image.Image
	GetClassroom() (Classroom, error)
	GetTimetable() ([15][7]Timetable, error)
}

type user struct {
	*API

	id     string
	name   string
	imgURL string
}

func (u *user) GetID() string {
	return u.id
}

func (u *user) GetName() string {
	return u.name
}

func (u *user) GetProfileImg() image.Image {
	req, err := http.NewRequest(http.MethodGet, u.imgURL, nil)
	if err != nil {
		return nil
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome&userDTO.localeKey=ko")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.Client().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil
	}
	return img
}
