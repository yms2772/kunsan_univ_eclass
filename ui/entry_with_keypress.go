package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"unicode/utf8"
)

type KeyPress struct {
	KeyName    fyne.KeyName
	OnKeyPress func()
}

type EntryWithKeyPress struct {
	*widget.Entry

	keyPress []KeyPress
}

func (e *EntryWithKeyPress) defaultKeyPress() []KeyPress {
	return []KeyPress{
		{
			KeyName: fyne.KeyBackspace,
			OnKeyPress: func() {
				text := e.Entry.Text
				if len(text) != 0 {
					_, size := utf8.DecodeLastRuneInString(text)
					e.SetText(text[:len(text)-size])
				}
			},
		},
	}
}

func (e *EntryWithKeyPress) AddKeyPress(key ...KeyPress) {
	e.keyPress = append(e.keyPress, key...)
}

func (e *EntryWithKeyPress) TypedKey(key *fyne.KeyEvent) {
	for _, item := range e.keyPress {
		if item.KeyName == key.Name && item.OnKeyPress != nil {
			item.OnKeyPress()
		}
	}
}

func NewEntryWithKeyPress(key ...KeyPress) *EntryWithKeyPress {
	e := &EntryWithKeyPress{
		Entry: &widget.Entry{
			Wrapping: fyne.TextTruncate,
		},
	}
	e.AddKeyPress(e.defaultKeyPress()...)
	e.AddKeyPress(key...)
	e.ExtendBaseWidget(e)
	return e
}
