package reconsiliation

import (
	"errors"
	"lokyn-cli/engine"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
)

type Screen struct {
	discoverdKeyList *list.Widget[engine.DiscoveredKey]
	dataKeyList      *list.Widget[engine.DiscoveredKey]

	keys            engine.Keys
	translationKeys engine.Keys
	data            engine.KeyLangMap

	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	project, ok := i.(engine.Project)

	if !ok {
		panic(errors.New("invalid project loaded"))
	}

	s.project = project

	keys, err := engine.DiscoverKeys()

	if err != nil {
		panic(err)
	}

	s.keys = keys

	data, err := engine.ImportTranslations(&project)

	if err != nil {
		panic(err)
	}

	s.data = data

	s.translationKeys = make(engine.Keys, 0)

	firstLang := project.ManagedLanguages[0]

	for k, v := range data {
		s.translationKeys = append(s.translationKeys, engine.DiscoveredKey{
			Key:      k,
			IsPlural: v[firstLang].IsPlural,
		})
	}

	s.discoverdKeyList.SetItems(s.keys)
	s.dataKeyList.SetItems(s.translationKeys)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return nil
}

func (s *Screen) compareKeys() {
	// TODO: Compare keys between both discovered and translation keys
}
