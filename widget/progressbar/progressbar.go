package progressbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"
)

type Widget struct {
	orvyn.BaseWidget

	progress.Model

	TitleStyle lipgloss.Style

	title string

	MaxValue     int
	CurrentValue int
}

func New(title string, color lipgloss.Color) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.Model = progress.New(progress.WithSolidFill(string(color)))
	w.Model.ShowPercentage = false

	w.TitleStyle = orvyn.GetTheme().Style(theme.TitleStyleID).
		AlignHorizontal(lipgloss.Center)
	w.title = title

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case progress.FrameMsg:
		progressModel, cmd := w.Model.Update(msg)
		w.Model = progressModel.(progress.Model)
		return cmd
	}

	return nil
}

func (w *Widget) Render() string {
	var b strings.Builder
	// var percent float64

	// size := w.GetContentSize()

	// if w.MaxValue > 0 {
	// 	percent = float64(100*w.CurrentValue/w.MaxValue) / 100
	// } else {
	// 	percent = 0
	// }

	if len(w.title) > 0 {
		b.WriteString(w.TitleStyle.Render(
			fmt.Sprintf("%s (%d/%d)",
				w.title, w.CurrentValue, w.MaxValue)))
	} else {
		b.WriteString(w.TitleStyle.Render(fmt.Sprintf("(%d/%d)",
			w.CurrentValue, w.MaxValue)))
	}
	// b.WriteString(strings.Repeat(
	// 	fmt.Sprintf("\n%s", w.Model.View()), size.Height-1),
	// )
	b.WriteString(fmt.Sprintf("\n%s", w.Model.View()))

	return b.String()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	w.Model.Width = size.Width

	w.TitleStyle = w.TitleStyle.Width(size.Width)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return orvyn.NewSize(10, 2)
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return orvyn.NewSize(30, 2)
}
