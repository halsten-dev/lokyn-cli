package progress

import (
	"lokyn-cli/widget/progressbar"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Screen struct {
	progressBar *progressbar.Widget

	maxSteps int
	steps    int
	percent  float64

	layout *layout.CenterLayout

	tickTag uint
}

func New() *Screen {
	s := &Screen{
		progressBar: progressbar.New("On going", orvyn.GetTheme().Color(theme.NormalFontColorID)),
	}

	s.layout = layout.NewCenterLayout(
		s.progressBar,
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	return orvyn.TickCmd(1, s.tickTag)
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.progressBar.Update(msg)

	switch msg := msg.(type) {
	case orvyn.TickMsg:
		if msg.Tag != s.tickTag {
			return nil
		}

		s.progressBar.MaxValue = s.maxSteps
		s.progressBar.CurrentValue = s.steps

		cmd := s.progressBar.SetPercent(s.percent)

		s.tickTag++
		return tea.Batch(cmd, orvyn.TickCmd(1, s.tickTag))
	}

	if s.percent >= 1 {
		return orvyn.CloseDialog()
	}

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) UpdateProgress(steps, maxSteps int) {
	var percent float64

	if maxSteps > 0 {
		percent = float64(100*steps/maxSteps) / 100
	} else {
		percent = 0
	}

	s.maxSteps = maxSteps
	s.steps = steps
	s.percent = percent
}
