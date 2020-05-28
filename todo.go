package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"fyne.io/fyne/dialog"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"

	"github.com/tidwall/gjson"
)

type LoginInfo struct {
	ID string `json:"id"`
	PW string `json:"pw"`
}

var (
	loginInfo = os.Getenv("LOCALAPPDATA") + "/todo_login.json"
	myID      string
	myPW      string
)

func init() {
	_ = os.Setenv("FYNE_FONT", "./bin/AppleSDGothicNeoB.ttf")
}

func ParseURL(urlStr string) *url.URL {
	link, _ := url.Parse(urlStr)

	return link
}

func RunAgain() {
	path, _ := os.Executable()

	exec.Command(path).Start()

	os.Exit(1)
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

	req, _ := http.NewRequest("POST", "https://eclass.kunsan.ac.kr/MUser.do", strings.NewReader(data.Encode()))

	resp, _ := client.Do(req)
	defer resp.Body.Close()

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

	client = &http.Client{
		Jar: jar,
	}

	postData := url.Values{}
	postData.Set("cmd", "loginUser")
	postData.Set("userDTO.localeKey", "ko")
	postData.Set("userDTO.userId", myID)
	postData.Set("userDTO.password", myPW)

	req, _ = http.NewRequest("POST", "https://eclass.kunsan.ac.kr/MUser.do", strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		panic(nil)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	bodyStr := string(body)

	return bodyStr
}

func main() {
	myTodoRegexp, _ := regexp.Compile(`<span class="fcBlue">(.*)</span>.*\n.*>(.*)</a>.*\n.*moveContentsList\('(\d+).*title="(.*)".*\n.*\n.*<li>(.*)</li>`)
	myInfoRegexp, _ := regexp.Compile(`<li>(.* \(\d+\))</li>`)

	a := app.New()

	w := a.NewWindow("내 할일")
	w.CenterOnScreen()
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(400, 600))

	waitLogin := dialog.NewProgressInfinite("인증", "로그인 기다리는 중...", w)
	waitLogin.Show()

	loginW := a.NewWindow("E-Class 로그인")
	loginW.CenterOnScreen()
	loginW.SetFixedSize(true)
	loginW.Resize(fyne.NewSize(550, 300))

	loginW.SetOnClosed(func() {
		os.Exit(1)
	})

	username := widget.NewEntry()
	password := widget.NewPasswordEntry()

	username.SetPlaceHolder("아이디")
	password.SetPlaceHolder("비밀번호")

	loginContent := widget.NewForm(widget.NewFormItem("사용자 ID", username),
		widget.NewFormItem("사용자 PW", password))

	loginContent.SubmitText = "로그인"
	loginContent.CancelText = "취소"

	loginContent.OnSubmit = func() {
		loginProg := dialog.NewProgressInfinite("인증", "로그인 중...", loginW)
		loginProg.Show()

		if CheckLogin(username.Text, password.Text) {
			myID = username.Text
			myPW = password.Text

			loginJSON := LoginInfo{
				ID: myID,
				PW: myPW,
			}

			file, _ := json.MarshalIndent(loginJSON, "", " ")

			ioutil.WriteFile(loginInfo, file, 0777)

			fmt.Println(string(file))

			RunAgain()
		} else {
			loginProg.Hide()

			dialog.ShowError(fmt.Errorf("존재하지 않는 계정입니다"), loginW)
		}
	}

	loginContent.OnCancel = func() {
		os.Exit(1)
	}

	loginW.SetContent(widget.NewGroup("E-Class 로그인",
		loginContent,
	))

	if _, err := os.Stat(loginInfo); err == nil {
		loginJSON, err := ioutil.ReadFile(loginInfo)
		if err != nil {
			RunAgain()
		}

		_, isJSON := gjson.Parse(string(loginJSON)).Value().(map[string]interface{})
		if !isJSON {
			os.Remove(loginInfo)

			RunAgain()
		}

		myID = gjson.Get(string(loginJSON), "id").String()
		myPW = gjson.Get(string(loginJSON), "pw").String()

		waitLogin.Hide()
	} else {
		loginW.ShowAndRun()
	}

	bodyStr := GetPageHTML()

	myTodos := myTodoRegexp.FindAllStringSubmatch(bodyStr, -1)
	myInfo := myInfoRegexp.FindStringSubmatch(bodyStr)

	queue := widget.NewGroup(fmt.Sprintf("%s님의 TODO", myInfo[1]))

	for _, myTodo := range myTodos {
		classHyperLink := widget.NewHyperlink(strings.ReplaceAll(myTodo[4], " ", ""), ParseURL(fmt.Sprintf("https://eclass.kunsan.ac.kr/Course.do?cmd=viewStudyHome&courseDTO.courseId=%s&boardInfoDTO.boardInfoGubun=study_home&gubun=study_course", myTodo[3])))

		form := &widget.Form{}

		form.Append("분류:", widget.NewLabel(myTodo[1]))
		form.Append("일자:", classHyperLink)
		form.Append("남은 시간:", widget.NewLabel(strings.ReplaceAll(myTodo[2], " ", "")))

		queueLayout := widget.NewVBox(widget.NewGroup(fmt.Sprintf(myTodo[5])),
			form,
		)

		queue.Append(queueLayout)
	}

	mainContent := widget.NewVScrollContainer(queue)

	w.SetContent(mainContent)

	w.ShowAndRun()
}
