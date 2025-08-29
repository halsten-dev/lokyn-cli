package translation

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/widget/keyedit"
	"lokyn-cli/widget/keylistitem"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	keyList *list.Widget[engine.DiscoveredKey]
	keyEdit *keyedit.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	keys    engine.Keys
	data    engine.KeyLangMap
	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("home"))
	s.title.SizeConstraint = true

	s.keyList = list.New(keylistitem.Constructor)
	s.keyList.CursorMovedCallback = s.keyListCursorMoved

	s.keyEdit = keyedit.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.keyList)
	s.focusManager.Add(s.keyEdit)

	s.layout = layout.NewCenterLayout(
		layout.NewHBoxFixedRatioLayout(
			10, 2, 0,
			[]layout.FixedRatioRenderable{
				layout.NewFixedRatioRenderable(0.30, s.keyList),
				layout.NewFixedRatioRenderable(0.70, s.keyEdit),
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	keys, err := engine.DiscoverKeys()

	if err != nil {
		panic(err)
	}

	s.keys = keys

	s.data = make(engine.KeyLangMap)

	project, ok := i.(engine.Project)

	if !ok {
		panic(errors.New("invalid project loaded"))
	}

	s.project = project

	s.keyEdit.InitTranslations(project)

	for _, k := range keys {
		s.data[k.Key] = make(map[engine.Lang]engine.Translation)

		for _, ml := range project.ManagedLanguages {
			s.data[k.Key][ml] = engine.Translation{
				Key:        k.Key,
				Lang:       ml,
				OneValue:   "",
				OtherValue: "",
				IsPlural:   k.IsPlural,
			}
		}
	}

	s.updateKeyList()

	s.focusManager.Focus(0)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) keyListCursorMoved(index int) {
	s.keyEdit.SetTranslations(s.data[s.keys[index].Key])
}

func (s *Screen) updateKeyList() {
	s.keyList.SetItems(s.keys)
}
