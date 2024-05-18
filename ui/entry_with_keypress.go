package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"slices"
)

type KeyPress struct {
	KeyName    fyne.KeyName
	OnKeyPress func()
}

type EntryWithKeyPress struct {
	*widget.Entry

	KeyPress []KeyPress
}

func (e *EntryWithKeyPress) TypedKey(key *fyne.KeyEvent) {
	idx := slices.IndexFunc(e.KeyPress, func(item KeyPress) bool {
		return key.Name == item.KeyName
	})
	if idx == -1 {
		return
	}

	if e.KeyPress[idx].OnKeyPress != nil {
		e.KeyPress[idx].OnKeyPress()
	}
}

func NewEntryWithKeyPress(key ...KeyPress) *EntryWithKeyPress {
	e := &EntryWithKeyPress{
		Entry: &widget.Entry{
			Wrapping: fyne.TextTruncate,
		},
		KeyPress: key,
	}
	e.ExtendBaseWidget(e)
	return e
}
