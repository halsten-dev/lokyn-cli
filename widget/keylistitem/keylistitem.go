package keylistitem

import (
	"lokyn-cli/engine"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data engine.DiscoveredKey

	style lipgloss.Style
}

func Constructor(data engine.DiscoveredKey) widgetlist.ListItem[engine.DiscoveredKey] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.OnBlur()

	w.UpdateData(data)

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

func (w *Widget) UpdateData(data engine.DiscoveredKey) {
	w.data = data
}

func (w *Widget) GetData() engine.DiscoveredKey {
	return w.data
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) FilterValue() string {
	return string(w.data.Key)
}
