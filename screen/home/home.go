package home

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/orvyn"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("home"))

	s.layout = layout.NewCenterLayout(
		s.title,
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	keys, err := engine.DiscoverKeys()

	if err != nil {
		panic(err)
	}

	s.title.SetValue(fmt.Sprintf("Found keys count : %d", len(keys)))

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
