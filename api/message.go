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

type MessageType int

const (
	MessageTypeReceived MessageType = iota
	MessageTypeSend
)

type MessageData struct {
	Content     string
	Attachments []Attachment
}

type Message struct {
	Type     MessageType
	From     string
	To       string
	Datetime time.Time
	Title    string

	messageID     string
	messageSendID string
	receiverID    string
}

func (m Message) GetURL() *url.URL {
	href := "https://eclass.kunsan.ac.kr"

	if m.Type == MessageTypeReceived {
		href += "/Message.do?" +
			"cmd=viewMessageContents" +
			"&gubun=myPage" +
			"&messageBoxDTO.messageId=" + m.messageID +
			"&messageBoxDTO.messageSendId=" + m.messageSendID +
			"&messageType=RECEIVE"
	} else {
		href += "/Message.do?" +
			"cmd=viewMessageContents" +
			"&gubun=myPage" +
			"&messageSendDTO.messageSendId=" + m.messageSendID +
			"&receiverId=" + m.receiverID +
			"&messageType=SEND"
	}

	u, _ := url.Parse(href)
	return u
}

func (a *API) GetMessage(data Message) (MessageData, error) {
	var message MessageData

	req, err := http.NewRequest(http.MethodGet, data.GetURL().String(), nil)
	if err != nil {
		return message, err
	}

	req.Header.Set("Referer", "https://eclass.kunsan.ac.kr/Main.do?cmd=viewHome")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := a.Client().Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return message, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return message, err
	}

	content, _ := doc.Find(`div.cont`).Html()

	message.Content, _ = html2text.FromString(content, html2text.Options{
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
		message.Attachments = append(message.Attachments, attachment)
	})
	return message, nil
}
