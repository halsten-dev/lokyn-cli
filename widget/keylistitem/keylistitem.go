package keylistitem

import (
	"github.com/charmbracelet/lipgloss"
	"lokyn-cli/engine"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/orvyn/widget/list"
	"lokyn-cli/internal/style"
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
	w.style = style.BlurredStyle

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Width -= style.BlurredStyle.GetHorizontalFrameSize()
	size.Height = lipgloss.Height(style.BlurredStyle.Render(string(w.data.Key)))

	w.BaseWidget.Resize(size)
}

func (w *Widget) Render() string {
	size := w.GetSize()

	return w.style.
		Width(size.Width).
		Render(string(w.data.Key))
}

func (w *Widget) OnFocus() {
	w.style = style.FocusedStyle
}

func (w *Widget) OnBlur() {
	w.style = style.BlurredStyle
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}
