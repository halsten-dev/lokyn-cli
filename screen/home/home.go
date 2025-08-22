package home

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/orvyn/widget/list"
	"lokyn-cli/widget/keylistitem"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	keyList *list.Widget[engine.DiscoveredKey]

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	keys    engine.Keys
	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("home"))
	s.title.SizeConstraint = true

	s.keyList = list.New[engine.DiscoveredKey](keylistitem.Constructor)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.keyList)

	s.layout = layout.NewCenterLayout(
		layout.NewHBoxFixedRatioLayout(
			10, 2, 0,
			[]layout.FixedRatioRenderable{
				layout.NewFixedRatioRenderable(0.30, s.keyList),
				layout.NewFixedRatioRenderable(0.70, s.title),
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	keys, err := engine.DiscoverKeys()

	if err != nil {
		panic(err)
	}

	s.keys = keys

	project, ok := i.(engine.Project)

	if !ok {
		panic(errors.New("invalid project loaded"))
	}

	s.project = project

	s.title.SetValue(s.project.ExportDir)

	s.updateKeyList()

	s.focusManager.Focus(0)

	return nil
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) updateKeyList() {
	s.keyList.SetItems(s.keys)
}
