package api

import "net/url"

type Subject struct {
	Name string

	courseID string
}

func (s *Subject) GetURL() *url.URL {
	href := "https://eclass.kunsan.ac.kr/Course.do?" +
		"cmd=viewStudyHome" +
		"&gubun=myPage" +
		"&courseDTO.courseId=" + s.courseID

	u, _ := url.Parse(href)
	return u
}
