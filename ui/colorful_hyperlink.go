package ui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type ColorfulHyperlink struct {
	*widget.Hyperlink

	color Color
}

func (c *ColorfulHyperlink) CreateRenderer() fyne.WidgetRenderer {
	renderer := c.Hyperlink.CreateRenderer()

	richText := renderer.Objects()[0].(*widget.RichText)
	richText.Segments[0].(*widget.TextSegment).Style.ColorName = c.color.colorName

	under := renderer.Objects()[2].(*canvas.Rectangle)
	under.FillColor = c.color.color
	return renderer
}

func NewColorfulHyperlink(text string, color Color, uri *url.URL) *ColorfulHyperlink {
	ch := &ColorfulHyperlink{
		Hyperlink: &widget.Hyperlink{
			Text: text,
			URL:  uri,
		},
		color: color,
	}
	ch.ExtendBaseWidget(ch)
	return ch
}
