package keylistitem

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
	"github.com/halsten-dev/orvyn/widget/checkbox"
	"github.com/halsten-dev/orvyn/widget/textinput"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	srKeyName    *orvyn.SimpleRenderable
	tiKeyName    *textinput.Widget
	chbxIsPlural *checkbox.Widget

	data engine.DiscoveredKey

	value string

	layout     *layout.PileLayout
	editLayout *layout.HBoxGrowLayout
	keyLayout  *layout.VBoxLayout

	focusManager *orvyn.FocusManager
}

func Constructor(data engine.DiscoveredKey) widgetlist.ListItem[engine.DiscoveredKey] {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()
	w.BaseFocusable = orvyn.NewBaseFocusable(w)

	w.srKeyName = orvyn.NewSimpleRenderable("")
	w.srKeyName.SizeConstraint = true

	w.tiKeyName = textinput.New()
	w.tiKeyName.Placeholder = lokyn.L("Key name")
	w.chbxIsPlural = checkbox.New(lokyn.L("Is plural ?"))

	w.focusManager = orvyn.NewFocusManager()
	w.focusManager.Add(w.tiKeyName)
	w.focusManager.Add(w.chbxIsPlural)

	w.editLayout = layout.NewHBoxGrowLayout(0, 0,
		w.tiKeyName,
		w.chbxIsPlural,
	)

	w.keyLayout = layout.NewMaxWidthVBoxLayout(0,
		w.srKeyName)

	w.layout = layout.NewPileLayout(
		w.editLayout,
		w.keyLayout,
	)

	w.OnBlur()

	w.UpdateData(data)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = lipgloss.Height(w.value)

	if w.data.IsCreatedByUser {
		size.Height = 3
	}

	w.BaseWidget.Resize(size)

	w.layout.Resize(w.GetContentSize())
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	if w.data.IsCreatedByUser {
		cmd = w.focusManager.Update(msg)
	}

	if w.data.IsCreatedByUser {
		w.data.Key = engine.Key(w.tiKeyName.Value())
		w.data.IsPlural = w.chbxIsPlural.IsChecked()
	}

	return cmd
}

func (w *Widget) Render() string {
	contentSize := w.GetContentSize()

	return w.GetStyle().
		Width(contentSize.Width).
		Height(contentSize.Height).
		Render(w.layout.Render())
}

func (w *Widget) UpdateData(data engine.DiscoveredKey) {
	w.data = data

	w.value = string(data.Key)

	if data.Err != nil {
		errorStr := orvyn.GetTheme().
			Style(theme.StatusErrorTextStyleID).Render(
			fmt.Sprintf("(%s)", data.Err.Error()),
		)

		w.value = fmt.Sprintf("%s%s", w.value, errorStr)
	}

	w.srKeyName.SetValue(w.value)

	w.tiKeyName.SetValue(string(data.Key))
	w.chbxIsPlural.SetChecked(data.IsPlural)

	if data.IsCreatedByUser {
		w.editLayout.SetActive(true)
		w.keyLayout.SetActive(false)
	} else {
		w.editLayout.SetActive(false)
		w.keyLayout.SetActive(true)
	}
}

func (w *Widget) GetData() engine.DiscoveredKey {
	return w.data
}

func (w *Widget) GetEnterInputKeybind() *key.Binding {
	if w.data.IsCreatedByUser {
		return &keybind.EKey
	}

	return nil
}

func (w *Widget) OnEnterInput() {
	w.focusManager.FocusFirst()
}

func (w *Widget) OnExitInput() {
	w.focusManager.BlurCurrent()
}

func (w *Widget) FilterValue() string {
	return string(w.data.Key)
}
