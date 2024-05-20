package api

import (
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Timetable struct {
	Subject   string
	Professor string
}

func (u *user) GetTimetable() ([15][7]Timetable, error) {
	timetable := [15][7]Timetable{}

	req, err := http.NewRequest(http.MethodGet, "https://eclass.kunsan.ac.kr/Study.do?cmd=viewMyTimetable", nil)
	if err != nil {
		return timetable, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return timetable, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return timetable, err
	}

	doc.Find(`table.boardListInfo tbody tr`).Each(func(i int, s *goquery.Selection) {
		s.Find(`td[onclick]`).Each(func(j int, s2 *goquery.Selection) {
			total := s2.Text()

			timetable[i][j].Subject = s2.Find(`div`).Text()
			timetable[i][j].Professor = strings.ReplaceAll(total, timetable[i][j].Subject, "")
		})
	})
	return timetable, nil
}
