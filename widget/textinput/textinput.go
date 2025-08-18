package textinput

import (
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	textinput.Model
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = textinput.New()
	style.SetTextInputStyle(&w.Model)

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.Model.SetValue("")
	return textinput.Blink
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	return cmd
}

func (w *Widget) OnFocus() {
	w.Model.Focus()
}

func (w *Widget) OnBlur() {
	w.Model.Blur()
}

func (w *Widget) Render() string {
	var border lipgloss.Style

	if w.IsFocused() {
		border = style.FocusedStyle
	} else {
		border = style.BlurredStyle
	}

	return border.Render(w.Model.View())
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	// Take borders into account
	w.Model.Width = size.Width - style.BlurredStyle.GetHorizontalFrameSize()
	w.Model.Width -= max(1, len(w.Model.Prompt))

	// For the Bubbles textinput to process the update
	focused := w.Model.Focused()
	if !focused {
		w.Model.Focus()
	}

	w.Model, _ = w.Model.Update(nil)

	if !focused {
		w.Model.Blur()
	}
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(26, 3)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(46, 3)
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}
