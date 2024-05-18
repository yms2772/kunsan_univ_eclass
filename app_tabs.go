package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"time"

	"kunsan_univ_eclass/api"
	"kunsan_univ_eclass/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/skratchdot/open-golang/open"
)

func (m *MainApp) AppTabs() fyne.CanvasObject {
	// 일정
	scheduleBox := widget.NewAccordion()

	// 게시물
	postBoxData := []*widget.Accordion{
		widget.NewAccordion(),
		widget.NewAccordion(),
	}
	postBox := container.NewGridWithRows(2,
		widget.NewCard("새로운 게시물", "", postBoxData[api.PostTypeNew]),
		widget.NewCard("내 게시물", "", postBoxData[api.PostTypeMy]),
	)

	// 쪽지
	messageBoxData := []*widget.Accordion{
		widget.NewAccordion(),
		widget.NewAccordion(),
	}
	messageBox := container.NewGridWithRows(2,
		widget.NewCard("받은 쪽지", "", messageBoxData[api.MessageTypeReceived]),
		widget.NewCard("보낸 쪽지", "", messageBoxData[api.MessageTypeSend]),
	)

	// 탭
	appTabs := container.NewAppTabs(
		container.NewTabItem("내 정보", container.NewVScroll(container.NewVBox())),
		container.NewTabItem("일정", container.NewVScroll(scheduleBox)),
		container.NewTabItem("게시물", container.NewVScroll(postBox)),
		container.NewTabItem("쪽지", container.NewVScroll(messageBox)),
	)

	appTabs.OnSelected = func(tabItem *container.TabItem) {
		m.loading.Show()
		defer m.loading.Hide()

		classroom, err := m.user.GetClassroom()
		if err != nil {
			m.window.SetContent(m.Login())
			return
		}

		switch tabItem.Text {
		case "내 정보":
			var subjects []fyne.CanvasObject

			for i, subject := range classroom.Subjects {

				subjects = append(subjects, ui.NewRichTextTappable(fmt.Sprintf("%d. %s", i+1, subject.Name), ui.ColorHyperLink, func() {
					m.loading.Show()
					defer m.loading.Hide()

					_ = open.Run(subject.GetURL().String())
				}))
			}

			profileImg := m.user.GetProfileImg()
			profileImgCanvas := canvas.NewImageFromImage(profileImg)
			profileImgCanvas.FillMode = canvas.ImageFillOriginal

			profileBox := container.NewHBox(
				profileImgCanvas,
				widget.NewForm(
					widget.NewFormItem("학번", container.NewHBox(
						widget.NewLabel(m.user.GetID()),
						ui.NewColorfulHyperlink("로그아웃", ui.ColorLogout, func() {
							m.window.SetContent(m.Login())
						}),
					)),
					widget.NewFormItem("이름", widget.NewLabel(m.user.GetName())),
					widget.NewFormItem("수강", container.NewVBox(subjects...)),
				))

			timetableData, err := m.user.GetTimetable()
			if err != nil {
				m.ShowError(errors.New("내용을 불러올 수 없습니다"))
				return
			}

			dayStart := time.Date(0, 0, 0, 8, 10, 0, 0, time.UTC)

			timetable := widget.NewTable(
				func() (rows int, cols int) {
					return 16, 8
				},
				func() fyne.CanvasObject {
					return &widget.Label{
						Text:      "",
						Alignment: fyne.TextAlignCenter,
						Wrapping:  fyne.TextTruncate,
					}
				},
				func(id widget.TableCellID, object fyne.CanvasObject) {
					label := object.(*widget.Label)

					if id.Row == 0 {
						switch id.Col {
						case 0:
							label.SetText("")
						case 1:
							label.SetText("월")
						case 2:
							label.SetText("화")
						case 3:
							label.SetText("수")
						case 4:
							label.SetText("목")
						case 5:
							label.SetText("금")
						case 6:
							label.SetText("토")
						case 7:
							label.SetText("일")
						}
					} else if id.Col == 0 {
						classTime := id.Row - 1
						classStart := dayStart.Add(time.Duration(classTime) * time.Hour)
						classEnd := classStart.Add(50 * time.Minute)

						label.SetText(fmt.Sprintf("%d교시\n"+
							"%s ~ %s",
							classTime,
							classStart.Format("15:04"),
							classEnd.Format("15:04"),
						))
					} else {
						label.SetText(fmt.Sprintf("%s\n"+
							"%s",
							timetableData[id.Row-1][id.Col-1].Subject,
							timetableData[id.Row-1][id.Col-1].Professor,
						))
					}
				},
			)

			timetable.StickyRowCount = 1
			timetable.StickyColumnCount = 1

			for row := 0; row <= 14; row++ {
				if row == 0 {
					timetable.SetRowHeight(row, 30)
				} else {
					timetable.SetRowHeight(row, 50)
				}
			}

			for col := 0; col <= 7; col++ {
				if col == 0 {
					timetable.SetColumnWidth(col, 120)
				} else {
					timetable.SetColumnWidth(col, 150)
				}
			}

			appTabs.Items[0].Content = container.NewVScroll(
				container.NewBorder(profileBox, nil, nil, nil, timetable),
			)
			appTabs.Items[0].Content.Refresh()
		case "일정":
			now := time.Now().UTC().Add(9 * time.Hour)

			classroom.Schedules = slices.DeleteFunc(classroom.Schedules, func(a api.Schedule) bool {
				return now.After(a.End)
			})

			slices.SortStableFunc(classroom.Schedules, func(a, b api.Schedule) int {
				return cmp.Compare(a.End.Unix(), b.End.Unix())
			})

			scheduleBox.Items = nil

			for _, schedule := range classroom.Schedules {
				checkContentBtn := widget.NewButton("내용 확인", nil)
				contentBox := container.NewVBox(
					widget.NewForm(
						widget.NewFormItem("과목", widget.NewLabel(schedule.Subject)),
						widget.NewFormItem("시작", widget.NewLabel(schedule.Start.Format("2006년 01월 02일 15시 04분"))),
						widget.NewFormItem("종료", widget.NewLabel(schedule.End.Format("2006년 01월 02일 15시 04분"))),
					),
					checkContentBtn,
				)

				checkContentBtn.OnTapped = func() {
					m.loading.Show()
					defer m.loading.Hide()

					scheduleData, err := m.api.GetSchedule(schedule)
					if err != nil {
						m.ShowError(errors.New("내용을 불러올 수 없습니다"))
						return
					}

					var objects []fyne.CanvasObject

					for _, item := range scheduleData.Attachments {
						objects = append(objects, ui.NewColorfulHyperlink(item.Name, ui.ColorHyperLink, func() {
							m.loading.Show()
							defer m.loading.Hide()

							body, err := item.Download()
							if err != nil {
								m.ShowError(errors.New("다운로드 실패"))
								return
							}

							dir := path.Join("downloaded")

							if err := os.MkdirAll(dir, os.ModePerm); err != nil {
								m.ShowError(errors.New("폴더를 생성할 수 없습니다"))
								return
							}

							_ = os.WriteFile(path.Join(dir, item.Name), body, os.ModePerm)
							_ = open.Run(dir)
						}))
					}

					content := widget.NewLabel(scheduleData.Content)
					content.Wrapping = fyne.TextWrapBreak

					objects = append(objects,
						content,
						ui.NewColorfulHyperlink("원본 보기", ui.ColorHyperLink, func() {
							_ = open.Run(schedule.GetURL().String())
						}),
					)

					contentBox.Remove(checkContentBtn)
					contentBox.Add(ui.NewBackgroundColorVBox(ui.ColorBackground, objects...))
				}

				scheduleBox.Items = append(scheduleBox.Items,
					widget.NewAccordionItem(schedule.Title, contentBox),
				)
			}

			scheduleBox.Refresh()
		case "게시물":
			postBoxData[api.PostTypeNew].Items = nil
			postBoxData[api.PostTypeMy].Items = nil

			for _, post := range classroom.Posts {
				checkContentBtn := widget.NewButton("내용 확인", nil)
				contentBox := container.NewVBox(
					widget.NewForm(
						widget.NewFormItem("카테고리", widget.NewLabel(post.Category)),
						widget.NewFormItem("과목", widget.NewLabel(post.Subject)),
						widget.NewFormItem("업로드 날짜", widget.NewLabel(post.Datetime.Format("2006년 01월 02일"))),
					),
					checkContentBtn,
				)

				checkContentBtn.OnTapped = func() {
					m.loading.Show()
					defer m.loading.Hide()

					postData, err := m.api.GetPost(post)
					if err != nil {
						m.ShowError(errors.New("내용을 불러올 수 없습니다"))
						return
					}

					var objects []fyne.CanvasObject

					for _, item := range postData.Attachments {
						objects = append(objects, ui.NewRichTextTappable(item.Name, ui.ColorHyperLink, func() {
							m.loading.Show()
							defer m.loading.Hide()

							body, err := item.Download()
							if err != nil {
								m.ShowError(errors.New("다운로드 실패"))
								return
							}

							dir := path.Join("downloaded")

							if err := os.MkdirAll(dir, os.ModePerm); err != nil {
								m.ShowError(errors.New("폴더를 생성할 수 없습니다"))
								return
							}

							_ = os.WriteFile(path.Join(dir, item.Name), body, os.ModePerm)
							_ = open.Run(dir)
						}))
					}

					if len(objects) > 0 {
						objects = append(objects, widget.NewSeparator())
					}

					content := widget.NewLabel(postData.Content)
					content.Wrapping = fyne.TextWrapBreak

					objects = append(objects, content)

					contentBox.Remove(checkContentBtn)
					contentBox.Add(ui.NewBackgroundColorVBox(ui.ColorBackground, objects...))
				}

				postBoxData[post.Type].Items = append(postBoxData[post.Type].Items,
					widget.NewAccordionItem(post.Title, contentBox),
				)
			}

			postBoxData[api.PostTypeNew].Refresh()
			postBoxData[api.PostTypeMy].Refresh()
		case "쪽지":
			messageBoxData[api.MessageTypeReceived].Items = nil
			messageBoxData[api.MessageTypeSend].Items = nil

			for _, message := range classroom.Messages {
				checkContentBtn := widget.NewButton("내용 확인", nil)
				contentBox := container.NewVBox(
					widget.NewForm(
						widget.NewFormItem("보낸 사람", widget.NewLabel(message.From)),
						widget.NewFormItem("받은 사람", widget.NewLabel(message.To)),
						widget.NewFormItem("날짜", widget.NewLabel(message.Datetime.Format("2006년 01월 02일 15시 04분 05초"))),
					),
					checkContentBtn,
				)

				checkContentBtn.OnTapped = func() {
					m.loading.Show()
					defer m.loading.Hide()

					messageData, err := m.api.GetMessage(message)
					if err != nil {
						m.ShowError(errors.New("내용을 불러올 수 없습니다"))
						return
					}

					var objects []fyne.CanvasObject

					for _, item := range messageData.Attachments {
						objects = append(objects, ui.NewColorfulHyperlink(item.Name, ui.ColorHyperLink, func() {
							m.loading.Show()
							defer m.loading.Hide()

							body, err := item.Download()
							if err != nil {
								m.ShowError(errors.New("다운로드 실패"))
								return
							}

							dir := path.Join("downloaded")

							if err := os.MkdirAll(dir, os.ModePerm); err != nil {
								m.ShowError(errors.New("폴더를 생성할 수 없습니다"))
								return
							}

							_ = os.WriteFile(path.Join(dir, item.Name), body, os.ModePerm)
							_ = open.Run(dir)
						}))
					}

					content := widget.NewLabel(messageData.Content)
					content.Wrapping = fyne.TextWrapBreak

					objects = append(objects, content)

					contentBox.Remove(checkContentBtn)
					contentBox.Add(ui.NewBackgroundColorVBox(ui.ColorBackground, objects...))
				}

				messageBoxData[message.Type].Items = append(messageBoxData[message.Type].Items,
					widget.NewAccordionItem(message.Title, contentBox),
				)
			}

			messageBoxData[api.MessageTypeReceived].Refresh()
			messageBoxData[api.MessageTypeSend].Refresh()
		}
	}

	appTabs.SelectIndex(2)
	appTabs.OnSelected(appTabs.Selected())
	return container.NewStack(
		appTabs,
		container.NewHBox(
			layout.NewSpacer(),
			container.NewVBox(
				widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
					appTabs.OnSelected(appTabs.Selected())
				}),
			),
		),
	)
}
