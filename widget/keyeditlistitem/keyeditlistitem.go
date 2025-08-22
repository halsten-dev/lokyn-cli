package keyeditlistitem

import (
	"github.com/charmbracelet/lipgloss"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/orvyn/widget/list"
	"lokyn-cli/internal/style"
	"lokyn-cli/widget/textinput"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	contentSize orvyn.Size

	srLanguage   *orvyn.SimpleRenderable
	tiOneValue   *textinput.Widget
	tiOtherValue *textinput.Widget

	data engine.Translation

	widgetStyle lipgloss.Style

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout
}

func Constructor(data engine.Translation) list.IListItem {
	w := new(Widget)

	w.data = data
	w.widgetStyle = style.BlurredStyle

	w.srLanguage = orvyn.NewSimpleRenderable(string(data.Lang))
	w.tiOneValue = textinput.New()
	w.tiOtherValue = textinput.New()

	w.tiOneValue.SetValue(data.OneValue)

	if data.IsPlural {
		w.tiOtherValue.SetValue(data.OtherValue)
		w.tiOtherValue.SetActive(true)
	} else {
		w.tiOtherValue.SetValue("")
		w.tiOtherValue.SetActive(false)
	}

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiOneValue)
	w.focusManager.Add(w.tiOtherValue)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			w.srLanguage,
			w.tiOneValue,
			w.tiOtherValue,
		},
	)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(orvyn.NewSize(size.Width, 8))

	size.Width -= w.widgetStyle.GetHorizontalFrameSize()
	size.Height = 6

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *Widget) Render() string {
	return w.widgetStyle.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.widgetStyle = style.FocusedStyle
}

func (w *Widget) OnBlur() {
	w.widgetStyle = style.BlurredStyle
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}
