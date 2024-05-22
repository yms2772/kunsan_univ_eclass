package api

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"jaytaylor.com/html2text"
)

const (
	PostTypeNew int = iota
	PostTypeMy
)

type PostData struct {
	Content     string
	Attachments []Attachment
}

type Post struct {
	Type     int
	Subject  string
	Datetime time.Time
	Category string
	Title    string

	boardContentsID string
	boardInfoID     string
	courseID        string
	boardClass      string
	boardType       string
}

func (p Post) GetURL() *url.URL {
	href := "https://eclass.kunsan.ac.kr"

	if p.boardType == "course" {
		href += "/Course.do?" +
			"cmd=viewBoardContents" +
			"&gubun=myPage" +
			"&boardContentsDTO.boardContentsId=" + p.boardContentsID +
			"&boardInfoDTO.boardInfoGubun=" + p.boardClass +
			"&boardInfoDTO.boardInfoId=" + p.boardInfoID +
			"&boardInfoDTO.boardClass=" + p.boardClass +
			"&boardInfoDTO.boardType=" + p.boardType +
			"&courseDTO.courseId=" + p.courseID +
			"&boardGubun=study_course"
	} else {
		href += "/Board.do?" +
			"cmd=viewBoardContents" +
			"&gubun=myPage" +
			"&boardContentsDTO.boardContentsId=" + p.boardContentsID +
			"&boardInfoDTO.boardInfoGubun=" + p.boardClass +
			"&boardInfoDTO.boardInfoId=" + p.boardInfoID +
			"&boardInfoDTO.boardClass=" + p.boardClass +
			"&boardInfoDTO.boardType=" + p.boardType
	}

	u, _ := url.Parse(href)
	return u
}

func (u *user) GetPost(data Post) (PostData, error) {
	var post PostData

	req, err := http.NewRequest(http.MethodGet, data.GetURL().String(), nil)
	if err != nil {
		return post, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return post, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return post, err
	}

	content, _ := doc.Find(`div.cont`).Html()

	post.Content, _ = html2text.FromString(content, html2text.Options{
		PrettyTables: true,
		TextOnly:     true,
	})

	r := regexp.MustCompile(`fileDownload\((.*?),(.*?),(.*?)\);`)

	doc.Find(`dd.info li`).Each(func(i int, s *goquery.Selection) {
		var attachment Attachment
		attachment.Name = strings.TrimSpace(s.Find(`span`).Text())

		href := s.Find(`span`).AttrOr("onclick", "")
		href = strings.ReplaceAll(href, "&#39;", "")
		href = strings.ReplaceAll(href, "'", "")

		match := r.FindStringSubmatch(href)
		if len(match) != 4 {
			return
		}

		attachment.URL = fmt.Sprintf("https://eclass.kunsan.ac.kr/fileDownServlet?rFileName=%s&sFileName=%s&filePath=%s", match[1], match[2], match[3])
		post.Attachments = append(post.Attachments, attachment)
	})
	return post, nil
}
