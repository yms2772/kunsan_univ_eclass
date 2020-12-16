package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/theme"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var (
	myID string
	myPW string
)

func ParseURL(urlStr string) *url.URL {
	link, _ := url.Parse(urlStr)

	return link
}

func CheckLogin(id, pw string) bool {
	client := &http.Client{}

	postData := url.Values{}
	postData.Set("cmd", "loginUser")
	postData.Set("userDTO.localeKey", "ko")
	postData.Set("userDTO.userId", id)
	postData.Set("userDTO.password", pw)

	req, _ := http.NewRequest("POST", "https://eclass.kunsan.ac.kr/MUser.do", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(req)

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	bodyStr := string(body)

	fmt.Println(bodyStr)
	fmt.Println(resp.StatusCode)

	return strings.Contains(bodyStr, "군산")
}

func GetPageHTML() string {
	client := &http.Client{}

	data := url.Values{}
	data.Add("cmd", "loginUser")
	data.Add("userDTO.localeKey", "ko")
	data.Add("userDTO.userId", myID)
	data.Add("userDTO.password", myPW)

	resp, err := client.PostForm("https://eclass.kunsan.ac.kr/MUser.do", data)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	jar, _ := cookiejar.New(nil)

	var cookies []*http.Cookie

	cookie := &http.Cookie{
		Name:  "WMONID",
		Value: resp.Cookies()[0].Value,
		Path:  "/",
	}

	cookies = append(cookies, cookie)

	cookie = &http.Cookie{
		Name:  "JSESSIONID",
		Value: resp.Cookies()[1].Value,
		Path:  "/",
	}

	cookies = append(cookies, cookie)

	u, _ := url.Parse("https://eclass.kunsan.ac.kr/MUser.do")

	jar.SetCookies(u, cookies)

	newClient := &http.Client{
		Jar: jar,
	}

	postData := url.Values{}
	postData.Set("cmd", "loginUser")
	postData.Set("userDTO.localeKey", "ko")
	postData.Set("userDTO.userId", myID)
	postData.Set("userDTO.password", myPW)

	newResp, err := newClient.PostForm("https://eclass.kunsan.ac.kr/MUser.do", postData)
	if err != nil {
		panic(err)
	}

	body, _ := ioutil.ReadAll(newResp.Body)
	newResp.Body.Close()

	bodyStr := string(body)

	return bodyStr
}

func Refresh() *widget.Group {
	myTodoRegexp, _ := regexp.Compile(`<span class="fcBlue">(.*)</span>.*\n.*>(.*)</a>.*\n.*moveContentsList\('(\d+).*title="(.*)".*\n.*\n.*<li>(.*)</li>`)
	myInfoRegexp, _ := regexp.Compile(`<li>(.* \(\d+\))</li>`)

	nowTime := time.Now()

	bodyStr := GetPageHTML()

	myTodos := myTodoRegexp.FindAllStringSubmatch(bodyStr, -1)
	myInfo := myInfoRegexp.FindStringSubmatch(bodyStr)

	queue := widget.NewGroup(fmt.Sprintf("%s님의 TODO (%s 기준)", myInfo[1], nowTime.Format("01월 02일 15시 04분")))

	for _, myTodo := range myTodos {
		classHyperLink := widget.NewHyperlink(strings.ReplaceAll(myTodo[4], " ", ""), ParseURL(fmt.Sprintf("https://eclass.kunsan.ac.kr/MCourse.do?cmd=viewStudyHome&courseDTO.courseId=%s&boardInfoDTO.boardInfoGubun=study_home&boardGubun=study_course&gubun=study_course", myTodo[3])))

		form := &widget.Form{}

		form.Append("분류:", widget.NewLabel(myTodo[1]))
		form.Append("일자:", classHyperLink)
		form.Append("남은 시간:", widget.NewLabel(strings.ReplaceAll(myTodo[2], " ", "")))

		queueLayout := widget.NewVBox(widget.NewGroup(fmt.Sprintf(myTodo[5])),
			form,
		)

		queue.Append(queueLayout)
	}

	return queue
}

func main() {
	var logout bool

	a := app.NewWithID("com.eclass.todo")
	a.Settings().SetTheme(NewCustomTheme())

	flag.BoolVar(&logout, "logout", false, "로그아웃")

	flag.Parse()

	if logout {
		a.Preferences().SetString("id", "")
	}

	savedID := a.Preferences().String("id")
	savedPW := a.Preferences().String("pw")

	w := a.NewWindow("내 할일")
	w.CenterOnScreen()
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(400, 600))

	w.SetOnClosed(func() {
		os.Exit(1)
	})

	log.Println("ACCOUNT ID:", len(savedID))
	log.Println("ACCOUNT PW:", len(savedPW))

	var queue *widget.Group

	if len(savedID) != 0 {
		myID = savedID
		myPW = savedPW

		queue = Refresh()
	} else {
		username := widget.NewEntry()
		password := widget.NewPasswordEntry()

		username.SetPlaceHolder("아이디")
		password.SetPlaceHolder("비밀번호")

		loginContent := widget.NewForm(widget.NewFormItem("사용자 ID", username),
			widget.NewFormItem("사용자 PW", password))

		queue = widget.NewGroup(fmt.Sprintf("%s님의 TODO", "정보 없음"))
		queue.Append(widget.NewLabelWithStyle("로그인 필요", fyne.TextAlignCenter, fyne.TextStyle{Bold: false}))
		queue.Append(fyne.NewContainerWithLayout(layout.NewCenterLayout(), widget.NewHBox(widget.NewButton("로그인 하기", func() {
			dialog.ShowCustomConfirm("E-Class 로그인", "로그인", "취소", loginContent, func(b bool) {
				if b {
					if CheckLogin(username.Text, password.Text) {
						myID = username.Text
						myPW = password.Text

						log.Println("MYID:", myID)
						log.Println("MYPW", myPW)

						a.Preferences().SetString("id", myID)
						a.Preferences().SetString("pw", myPW)
						dialog.ShowConfirm("E-Class TODO", "로그인 되었습니다\n앱을 다시 실행해주세요", func(b bool) {
							if b {
								os.Exit(1)
							}
						}, w)
					} else {
						dialog.ShowError(fmt.Errorf("존재하지 않는 계정입니다"), w)
					}
				}
			}, w)
		}))))
	}

	terminateBtn := widget.NewButtonWithIcon("종료", theme.CancelIcon(), func() {
		os.Exit(0)
	})

	//mainContent := widget.NewVScrollContainer(queue)
	mainContent := widget.NewVScrollContainer(widget.NewVBox(queue,
		layout.NewSpacer(),
		fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, terminateBtn), terminateBtn),
	))

	w.SetContent(mainContent)
	w.ShowAndRun()
}
