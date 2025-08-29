package statusmessage

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

type messageType int

const (
	ErrorMessage messageType = iota
	SuccessMessage
	WarningMessage
	InformationMessage
	NeutralMessage
)

type Style struct {
	Error       lipgloss.Style
	Success     lipgloss.Style
	Warning     lipgloss.Style
	Information lipgloss.Style
	Neutral     lipgloss.Style
}

type Widget struct {
	orvyn.BaseWidget

	Style Style

	message      string
	messageType  messageType
	messageStyle lipgloss.Style
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	baseStyle := lipgloss.NewStyle().AlignHorizontal(lipgloss.Center)

	w.Style = Style{
		Error: baseStyle.
			Foreground(lipgloss.Color("#FF1824")),
		Success: baseStyle.
			Foreground(lipgloss.Color("#18DB24")),
		Warning: baseStyle.
			Foreground(lipgloss.Color("#F36318")),
		Information: baseStyle.
			Foreground(lipgloss.Color("#035AFF")),
		Neutral: baseStyle,
	}

	return w
}

func (w *Widget) Init() tea.Cmd {
	w.Reset()

	return nil
}

func (w *Widget) Render() string {
	size := w.GetSize()

	s := ""

	if w.message != "" {
		s = w.messageStyle.
			Width(size.Width).
			Render(w.message)
	}

	return s
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.GetRenderSize(w.Style.Neutral, w.message)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.GetMinSize()
}

func (w *Widget) SetMessage(msg string, msgType messageType) {
	w.message = msg
	w.messageType = msgType
	w.updateStyle()
}

func (w *Widget) SetError(err error) {
	w.message = err.Error()
	w.messageType = ErrorMessage
	w.messageStyle = w.Style.Error
}

func (w *Widget) Reset() {
	w.message = ""
	w.messageType = NeutralMessage
	w.messageStyle = w.Style.Neutral
}

func (w *Widget) updateStyle() {
	switch w.messageType {
	case ErrorMessage:
		w.messageStyle = w.Style.Error
	case SuccessMessage:
		w.messageStyle = w.Style.Success
	case WarningMessage:
		w.messageStyle = w.Style.Warning
	case InformationMessage:
		w.messageStyle = w.Style.Information
	case NeutralMessage:
		w.messageStyle = w.Style.Neutral
	}
}
