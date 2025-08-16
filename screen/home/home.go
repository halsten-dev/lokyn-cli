package home

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/orvyn"
)

type Screen struct {
	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.layout = layout.NewCenterLayout(
		orvyn.NewSimpleRenderable(lokyn.L("home")),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	engine.DiscoverKeys()

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
