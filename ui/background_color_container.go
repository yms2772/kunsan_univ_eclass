package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func NewBackgroundColorVBox(color Color, obj ...fyne.CanvasObject) *fyne.Container {
	c := container.NewVBox(obj...)
	r := canvas.NewRectangle(color.color)
	r.SetMinSize(c.Size())
	return container.NewStack(r, c)
}
