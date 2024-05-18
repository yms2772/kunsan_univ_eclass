package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type notPaddedLayout struct {
}

func (n *notPaddedLayout) Layout(o []fyne.CanvasObject, s fyne.Size) {
	pad := theme.InnerPadding() * -2
	pad2 := pad * -2

	o[0].Move(fyne.NewPos(pad+10, pad))
	o[0].Resize(s.Add(fyne.NewSize(pad2, pad2)))
}

func (n *notPaddedLayout) MinSize(o []fyne.CanvasObject) fyne.Size {
	pad := theme.InnerPadding() * 4
	return o[0].MinSize().Subtract(fyne.NewSize(pad, pad))
}
