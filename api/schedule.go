package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type ScheduleData struct {
	Content     string
	Attachments []Attachment
}

type Schedule struct {
	Subject string
	Title   string
	Content string
	Start   time.Time
	End     time.Time

	etestInfoID string
	courseID    string
	titleName   string
}

func (s Schedule) GetURL() *url.URL {
	href := "https://eclass.kunsan.ac.kr"

	switch s.titleName {
	case "A_Etest":
		href += "/Etest.do?" +
			"cmd=viewEtestInfoList" +
			"&gubun=myPage" +
			"&etestInfoDTO.etestInfoId=" + s.etestInfoID +
			"&courseDTO.courseId=" + s.courseID +
			"&boardGubun=study_course" +
			"&boardInfoDTO.boardInfoGubun=etest&pageGubun=study_home"
	case "C_Report":
		href += "/Report.do?" +
			"cmd=viewReportInfoPageList" +
			"&gubun=myPage" +
			"&reportInfoDTO.reportInfoId=" + s.etestInfoID +
			"&courseDTO.courseId=" + s.courseID +
			"&boardGubun=study_course" +
			"&boardInfoDTO.boardInfoGubun=report&pageGubun=study_home"
	case "D_Forum":
		href += "/Forum.do?" +
			"cmd=viewForumInfoList" +
			"&gubun=myPage" +
			"&boardGubun=study_course" +
			"&forumInfoDTO.forumInfoId=" + s.etestInfoID +
			"&courseDTO.courseId=" + s.courseID +
			"&boardInfoDTO.boardInfoGubun=forum&pageGubun=study_home"
	case "F_Teamact":
		href += "/Teamact.do?" +
			"cmd=viewTeamactList" +
			"&gubun=myPage" +
			"&teamactInfoDTO.teamactId=" + s.etestInfoID +
			"&boardGubun=study_course" +
			"&courseDTO.courseId=" + s.courseID +
			"&boardInfoDTO.boardInfoGubun=teamact&pageGubun=study_home"
	}

	u, _ := url.Parse(href)
	return u
}

func (a *API) GetSchedule(data Schedule) (ScheduleData, error) {
	var schedule ScheduleData

	req, err := http.NewRequest(http.MethodGet, data.GetURL().String(), nil)
	if err != nil {
		return schedule, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := a.Client().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return schedule, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return schedule, err
	}

	content, _ := doc.Find(`div.cont`).Html()

	schedule.Content, _ = html2text.FromString(content, html2text.Options{
		PrettyTables: true,
		TextOnly:     true,
	})

	r := regexp.MustCompile(`fileDownload\((.*?),(.*?),(.*?)\);`)

	doc.Find(`div#listBox dd li span[onclick]`).Each(func(i int, s *goquery.Selection) {
		var attachment Attachment
		attachment.Name = strings.TrimSpace(s.Text())

		href := s.AttrOr("onclick", "")
		href = strings.ReplaceAll(href, "&#39;", "")
		href = strings.ReplaceAll(href, "'", "")

		match := r.FindStringSubmatch(href)
		if len(match) != 4 {
			return
		}

		attachment.URL = fmt.Sprintf("https://eclass.kunsan.ac.kr/fileDownServlet?rFileName=%s&sFileName=%s&filePath=%s", match[1], match[2], match[3])
		schedule.Attachments = append(schedule.Attachments, attachment)
	})
	return schedule, nil
}
