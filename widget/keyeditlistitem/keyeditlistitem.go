package keyeditlistitem

import (
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	bti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/textinput"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
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

func Constructor(data engine.Translation) widgetlist.ListItem[engine.Translation] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.srLanguage = orvyn.NewSimpleRenderable(string(data.Lang))
	w.tiOneValue = textinput.New()
	w.tiOneValue.Placeholder = lokyn.L("One")
	w.tiOtherValue = textinput.New()
	w.tiOtherValue.Placeholder = lokyn.L("Multiple")

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiOneValue)
	w.focusManager.Add(w.tiOtherValue)

	w.layout = layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		w.srLanguage,
		w.tiOneValue,
		w.tiOtherValue,
	)

	w.OnBlur()

	w.UpdateData(data)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	height := 9

	if !w.data.IsPlural {
		height = 6
	}

	w.BaseWidget.Resize(orvyn.NewSize(size.Width, height))

	size.Width -= w.style.GetHorizontalFrameSize()
	size.Height = height - w.style.GetHorizontalFrameSize()

	w.contentSize = size
	w.layout.Resize(size)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.focusManager.Update(msg)

	w.data.OneValue = w.tiOneValue.Value()
	w.data.OtherValue = w.tiOtherValue.Value()

	return cmd
}

func (w *Widget) UpdateData(data engine.Translation) {
	w.data = data

	w.tiOneValue.SetValue(data.OneValue)

	if data.IsPlural {
		w.tiOneValue.Placeholder = lokyn.L("One")
		w.tiOtherValue.SetValue(data.OtherValue)
		w.tiOtherValue.SetActive(true)
	} else {
		w.tiOneValue.Placeholder = lokyn.L("Value")
		w.tiOtherValue.SetValue("")
		w.tiOtherValue.SetActive(false)
	}
}

func (w *Widget) GetData() engine.Translation {
	return w.data
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

func (w *Widget) OnEnterInput() tea.Cmd {
	w.focusManager.Focus(0)
	bubblehelp.SwitchContext(keybind.ContextInputMode)

	return bti.Blink
}

func (w *Widget) OnExitInput() tea.Cmd {
	w.focusManager.BlurCurrent()
	bubblehelp.SwitchContext(keybind.ContextTranslation)
	bubblehelp.SetKeybindVisible(keybind.EKey, true)

	return nil
}

func (w *Widget) FilterValue() string {
	return ""
}
