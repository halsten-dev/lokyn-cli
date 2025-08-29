package label

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

type Widget struct {
	orvyn.BaseWidget

	Style lipgloss.Style

	value string
}

func New(value string) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Style = lipgloss.NewStyle().Italic(true)
	w.value = value

	return w
}

func (w *Widget) SetValue(value string) {
	w.value = value
}

func (w *Widget) Render() string {
	size := w.GetSize()

	return w.Style.
		Width(size.Width).
		Height(size.Height).
		Render(w.value)
}
