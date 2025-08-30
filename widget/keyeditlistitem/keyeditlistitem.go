package keyeditlistitem

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/textinput"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/layout"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	contentSize orvyn.Size

	srLanguage   *orvyn.SimpleRenderable
	tiOneValue   *textinput.Widget
	tiOtherValue *textinput.Widget

	data engine.Translation

	style lipgloss.Style

	focusManager *orvyn.FocusManager

	layout *layout.VBoxFullLayout
}

func Constructor(data engine.Translation) list.IListItem[engine.Translation] {
	w := new(Widget)

	w.data = data

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

	w.OnBlur()

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(orvyn.NewSize(size.Width, 8))

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height = 6

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	w.data.OneValue = w.tiOneValue.Value()
	w.data.OtherValue = w.tiOtherValue.Value()

	return cmd
}

func (w *Widget) Render() string {
	return w.style.
		Width(w.contentSize.Width).
		Height(w.contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) OnFocus() {
	w.style = orvyn.GetTheme().Style(theme.FocusedWidgetStyleID)
}

func (w *Widget) OnBlur() {
	w.style = orvyn.GetTheme().Style(theme.BlurredWidgetStyleID)
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	return &keybind.EKey
}

func (w *Widget) OnEnterInput() {
	w.focusManager.Focus(0)
}

func (w *Widget) OnExitInput() {
	w.focusManager.BlurCurrent()
}

func (w *Widget) GetData() engine.Translation {
	return w.data
}

func (w *Widget) FilterValue() string {
	return ""
}
