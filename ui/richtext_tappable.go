package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type TappableSegment struct {
	Text     string
	Color    Color
	OnTapped func()
}

func (t *TappableSegment) Inline() bool {
	return true
}

func (t *TappableSegment) Textual() string {
	return t.Text
}

func (t *TappableSegment) Update(object fyne.CanvasObject) {
	link := object.(*fyne.Container).Objects[0].(*ColorfulHyperlink)
	link.Text = t.Text
	link.OnTapped = t.OnTapped
	link.Refresh()
}

func (t *TappableSegment) Visual() fyne.CanvasObject {
	link := NewColorfulHyperlink(t.Text, t.Color, nil)
	link.OnTapped = t.OnTapped
	return &fyne.Container{
		Layout:  &notPaddedLayout{},
		Objects: []fyne.CanvasObject{link},
	}
}

func (t *TappableSegment) Select(_, _ fyne.Position) {}

func (t *TappableSegment) SelectedText() string { return "" }

func (t *TappableSegment) Unselect() {}

func NewRichTextTappable(text string, color Color, tapped func()) *widget.RichText {
	return widget.NewRichText(&TappableSegment{
		Text:     text,
		Color:    color,
		OnTapped: tapped,
	})
}
