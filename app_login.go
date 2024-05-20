package main

import (
	"errors"
	"strings"

	"kunsan_univ_eclass/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func (m *MainApp) Login() fyne.CanvasObject {
	m.app.Preferences().SetString("eclass_id", "")
	m.app.Preferences().SetString("eclass_pw", "")

	idEntry := ui.NewEntryWithKeyPress()
	pwEntry := ui.NewEntryWithKeyPress()
	pwEntry.Password = true
	saveLoginBtn := widget.NewCheck("자동 로그인", nil)
	saveLoginBtn.Checked = true

	loginForm := widget.NewForm(
		widget.NewFormItem("학번", idEntry),
		widget.NewFormItem("비밀번호", pwEntry),
		widget.NewFormItem("", container.NewHBox(
			layout.NewSpacer(),
			saveLoginBtn,
		)),
	)
	loginForm.SubmitText = "로그인"
	loginForm.OnSubmit = func() {
		m.loading.Show()
		defer m.loading.Hide()

		id := idEntry.Text
		pw := pwEntry.Text

		if id == "" || pw == "" {
			m.ShowError(errors.New("학번 또는 비밀번호를 입력해주세요"))
			return
		}

		if err := m.user.LoginPortal(id, pw); err != nil {
			idEntry.SetText("")
			pwEntry.SetText("")
			m.ShowError(err)
			return
		}

		if saveLoginBtn.Checked {
			m.app.Preferences().SetString("eclass_id", id)
			m.app.Preferences().SetString("eclass_pw", pw)
		}

		m.window.SetContent(m.AppTabs())
	}

	login := []ui.KeyPress{
		{KeyName: fyne.KeyReturn, OnKeyPress: loginForm.OnSubmit},
	}

	idEntry.KeyPress = login
	pwEntry.KeyPress = login
	return container.NewCenter(
		widget.NewCard("로그인"+strings.Repeat("\t", 6), "", loginForm),
	)
}
