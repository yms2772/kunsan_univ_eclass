package api

import (
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Classroom struct {
	Posts     []Post
	Messages  []Message
	Subjects  []Subject
	Schedules []Schedule
}

func (u *user) GetClassroom() (Classroom, error) {
	var classroom Classroom

	req, err := http.NewRequest(http.MethodGet, "https://eclass.kunsan.ac.kr/Study.do?cmd=viewStudyMyClassroom&boardInfoDTO.boardInfoGubun=myclassroom", nil)
	if err != nil {
		return classroom, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := u.getClient().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return classroom, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return classroom, err
	}

	doc.Find(`div#listBox div.portlet`).Each(func(i int, s *goquery.Selection) {
		portletName := s.Find(`div.portlet-header h3`).Text()

		switch portletName {
		case "일 정":
			var schedules []Schedule
			r1 := regexp.MustCompile(`\((.*?) ~(.*?)\)`)
			r2 := regexp.MustCompile(`ViewEtestContent\((.*?),(.*?),(.*?)\);`)

			s.Find(`ul#timeline li`).Each(func(j int, s2 *goquery.Selection) {
				var schedule Schedule
				href := s2.Find(`div[onclick]`).AttrOr("onclick", "")
				href = strings.ReplaceAll(href, "&#39;", "")
				href = strings.ReplaceAll(href, "'", "")

				if href == "" {
					return
				}

				s2.Find(`div[onclick] span`).Each(func(k int, s3 *goquery.Selection) {
					switch k {
					case 0:
						schedule.Subject = strings.TrimSpace(s3.Text())
					case 1:
						schedule.Content = strings.TrimSpace(s3.Text())
					case 2:
						match := r1.FindStringSubmatch(s3.Text())
						if len(match) != 3 {
							return
						}

						schedule.Start, _ = time.Parse("2006.01.02 15:04", strings.TrimSpace(match[1]))
						schedule.End, _ = time.Parse("2006.01.02 15:04", strings.TrimSpace(match[2]))
					}
				})

				match := r2.FindStringSubmatch(href)
				if len(match) == 4 {
					schedule.etestInfoID = match[1]
					schedule.courseID = match[2]
					schedule.titleName = match[3]
				}

				schedule.Title = strings.TrimSpace(s2.Find(`div[onclick] label`).Text())

				if !slices.ContainsFunc(schedules, func(item Schedule) bool {
					return item.etestInfoID == schedule.etestInfoID
				}) {

					schedules = append(schedules, schedule)
				}
			})

			classroom.Schedules = schedules
		case "수강 과목 정보":
			var subjects []Subject
			r := regexp.MustCompile(`viewStudyHome\((.*?)\);`)

			s.Find(`table.listBoard`).Each(func(j int, s2 *goquery.Selection) {
				href := s2.Find(`a`).AttrOr("href", "")
				href = strings.ReplaceAll(href, "&#39;", "")
				href = strings.ReplaceAll(href, "'", "")

				match := r.FindStringSubmatch(href)
				if len(match) != 2 {
					return
				}

				subjects = append(subjects, Subject{
					Name:     strings.TrimSpace(s2.Find(`tbody a`).Text()),
					courseID: match[1],
				})
			})

			classroom.Subjects = subjects
		case "받은 쪽지 보낸 쪽지":
			var messages []Message
			r1 := regexp.MustCompile(`moveRecvMessageView\((.*?),(.*?)\);`)
			r2 := regexp.MustCompile(`moveSendMessageView\((.*?),(.*?)\);`)

			s.Find(`div[id^="recv"]`).Each(func(j int, s2 *goquery.Selection) {
				id := s2.AttrOr("id", "")

				s2.Find(`dl.listTable`).Each(func(k int, s3 *goquery.Selection) {
					var message Message
					href := s3.Find(`a`).AttrOr("href", "")
					href = strings.ReplaceAll(href, "&#39;", "")
					href = strings.ReplaceAll(href, "'", "")

					switch id {
					case "recv":
						from := s3.Find(`span.fcOlive`).Text()
						datetimeStr := s3.Find(`dd`).Text()
						datetimeStr = strings.ReplaceAll(datetimeStr, from, "")
						datetimeStr = strings.ReplaceAll(datetimeStr, "보낸 쪽지", "")
						datetimeStr = strings.TrimSpace(datetimeStr)

						match := r1.FindStringSubmatch(href)
						if len(match) == 3 {
							message.messageID = match[1]
							message.messageSendID = match[2]
						}

						message.Type = MessageTypeReceived
						message.Title = strings.TrimSpace(s3.Find(`dt a`).Text())
						message.From = strings.TrimSpace(strings.Replace(from, "님 이", "", 1))
						message.To = u.GetName()
						message.Datetime, _ = time.Parse("2006-01-02 15:04:05", datetimeStr)
					case "recvoff":
						to := s3.Find(`span.fcOlive`).Text()
						datetimeStr := s3.Find(`dd`).Text()
						datetimeStr = strings.ReplaceAll(datetimeStr, to, "")
						datetimeStr = strings.ReplaceAll(datetimeStr, "에 보낸 쪽지", "")
						datetimeStr = strings.TrimSpace(datetimeStr)

						match := r2.FindStringSubmatch(href)
						if len(match) == 3 {
							message.messageSendID = match[1]
							message.receiverID = match[2]
						}

						message.Type = MessageTypeSend
						message.Title = strings.TrimSpace(s3.Find(`dt a`).Text())
						message.From = u.GetName()
						message.To = to
						message.Datetime, _ = time.Parse("2006-01-02 15:04:05", datetimeStr)
					}

					messages = append(messages, message)
				})
			})

			classroom.Messages = messages
		case "새로운 게시물", "내 글 목록내 댓글 목록":
			var posts []Post
			r := regexp.MustCompile(`ViewBoardContent\((.*?),(.*?),(.*?),(.*?),(.*?)\);`)

			s.Find(`dl.listTable`).Each(func(j int, s2 *goquery.Selection) {
				var post Post

				post.Category = strings.TrimSpace(s2.Find(`span.category`).Text())
				post.Title = strings.TrimSpace(s2.Find(`a`).Text())
				post.Subject = strings.TrimSpace(s2.Find(`dd span`).Text())

				datetimeStr := s2.Find(`dd`).Text()
				datetimeStr = strings.ReplaceAll(datetimeStr, post.Subject, "")
				datetimeStr = strings.TrimSpace(datetimeStr)

				post.Datetime, _ = time.Parse("2006-01-02", datetimeStr)

				href := s2.Find(`a`).AttrOr("href", "")
				href = strings.ReplaceAll(href, "&#39;", "")
				href = strings.ReplaceAll(href, "'", "")

				match := r.FindStringSubmatch(href)
				if len(match) == 6 {
					post.boardContentsID = match[1]
					post.boardInfoID = match[2]
					post.courseID = match[3]
					post.boardClass = match[4]
					post.boardType = match[5]
				}

				if portletName == "새로운 게시물" {
					post.Type = PostTypeNew
				} else {
					post.Type = PostTypeMy
				}

				posts = append(posts, post)
			})

			classroom.Posts = append(classroom.Posts, posts...)
		}
	})
	return classroom, nil
}
