package keylistitem

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"lokyn-cli/engine"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data engine.DiscoveredKey

	style lipgloss.Style
}

func Constructor(data engine.DiscoveredKey) list.IListItem {
	w := new(Widget)

	w.data = data

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height = lipgloss.Height(w.style.Render(string(w.data.Key)))

	w.BaseWidget.Resize(size)
}

func (w *Widget) Render() string {
	size := w.GetSize()

	return w.style.
		Width(size.Width).
		Render(string(w.data.Key))
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}
