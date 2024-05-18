package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type Color struct {
	color     color.RGBA
	colorName fyne.ThemeColorName
}

var (
	ColorBackground = Color{
		color: color.RGBA{R: 236, G: 239, B: 241, A: 255},
	}
	ColorHyperLink = Color{
		color:     color.RGBA{R: 25, G: 118, B: 210, A: 255},
		colorName: theme.ColorNameHyperlink,
	}
	ColorLogout = Color{
		color:     color.RGBA{R: 244, G: 67, B: 54, A: 255},
		colorName: theme.ColorNameError,
	}
)
