package help

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	return w
}

func (w *Widget) Render() string {
	return bubblehelp.View(w.GetSize().Width)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.GetRenderSize(lipgloss.NewStyle(), bubblehelp.View(w.GetSize().Width))
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.GetRenderSize(lipgloss.NewStyle(), bubblehelp.View(w.GetSize().Width))
}
