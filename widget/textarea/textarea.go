package textarea

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/style"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	textarea.Model

	MinHeight       int
	PreferredHeight int
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = textarea.New()
	style.SetTextAreaStyle(&w.Model)
	w.Model.Blur()

	w.MinHeight = 1
	w.PreferredHeight = 5

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.Model.SetValue("")
	return textarea.Blink
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
	return w.Model.View()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	// size.Width -= w.BlurredStyle.Base.GetHorizontalFrameSize()
	size.Height -= w.BlurredStyle.Base.GetVerticalFrameSize()

	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(size.Height)

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
	return orvyn.NewSize(20, w.MinHeight)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(50, w.PreferredHeight)
}

func (w *Widget) OnEnterInput() {
}

func (w *Widget) OnExitInput() {
}
