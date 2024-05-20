package main

import (
	"kunsan_univ_eclass/api"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type MainApp struct {
	user api.User

	w, h    float32
	app     fyne.App
	window  fyne.Window
	loading *dialog.CustomDialog
}

func NewMainApp() *MainApp {
	m := &MainApp{
		w: 1200,
		h: 720,
	}

	m.user = api.NewUser()

	m.app = app.NewWithID("kr.mokky.kunsan_univ_eclass")
	m.app.Settings().SetTheme(&myTheme{})

	m.window = m.app.NewWindow("군산대학교 eClass")
	m.window.Resize(fyne.NewSize(m.w, m.h))
	m.window.SetMaster()

	m.loading = dialog.NewCustomWithoutButtons("불러오는 중...", widget.NewProgressBarInfinite(), m.window)
	m.loading.Resize(fyne.NewSize(m.w/2, 0))
	return m
}

func (m *MainApp) ShowError(err error) {
	m.loading.Hide()
	dialog.ShowError(err, m.window)
}

func (m *MainApp) Run() {
	m.loading.Show()

	go func() {
		defer m.loading.Hide()

		id := m.app.Preferences().String("eclass_id")
		pw := m.app.Preferences().String("eclass_pw")

		if err := m.user.LoginPortal(id, pw); err != nil {
			m.window.SetContent(m.Login())
		} else {
			m.window.SetContent(m.AppTabs())
		}
	}()

	m.window.ShowAndRun()
}
