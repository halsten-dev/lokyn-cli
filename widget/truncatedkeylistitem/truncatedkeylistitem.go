package truncatedkeylistitem

import (
	"lokyn-cli/engine"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	data        engine.Key
	renderValue string
}

func Constructor(data engine.Key) widgetlist.ListItem[engine.Key] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.renderValue = "EMPTY"

	w.UpdateData(data)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 3

	w.BaseWidget.Resize(size)

	w.refreshKeyName()
}

func (w *Widget) refreshKeyName() {
	width := w.GetContentSize().Width
	value := string(w.data)

	if width > 0 {
		value = ansi.Truncate(value, width, "…")
	}

	w.renderValue = value
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (w *Widget) UpdateData(data engine.Key) {
	w.data = data

	w.refreshKeyName()
}

func (w *Widget) GetData() engine.Key {
	return w.data
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.renderValue)
}

func (w *Widget) FilterValue() string {
	return string(w.data)
}
