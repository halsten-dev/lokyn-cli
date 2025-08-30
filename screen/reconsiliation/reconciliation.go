package reconsiliation

import (
	"errors"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
	"lokyn-cli/engine"
	"lokyn-cli/internal/helper"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/layout"
	"lokyn-cli/screen"
	"lokyn-cli/widget/keylistitem"
	"slices"
)

type Screen struct {
	discoveredKeyTitle *orvyn.SimpleRenderable
	discoveredKeyList  *list.Widget[engine.DiscoveredKey]

	dataKeyTitle *orvyn.SimpleRenderable
	dataKeyList  *list.Widget[engine.DiscoveredKey]

	discoveredKeys      engine.Keys
	translationKeys     engine.Keys
	alreadyExistingKeys []engine.Key
	data                engine.KeyLangMap

	layout *layout.HBoxGrowLayout

	focusManager *orvyn.FocusManager

	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	s.discoveredKeyTitle = orvyn.NewSimpleRenderable(lokyn.L("New discovered keys"))
	s.discoveredKeyList = list.New(keylistitem.Constructor)

	s.dataKeyTitle = orvyn.NewSimpleRenderable(lokyn.L("Unused keys"))
	s.dataKeyList = list.New(keylistitem.Constructor)

	discoveredListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			s.discoveredKeyTitle,
			s.discoveredKeyList,
		},
	)

	dataListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		[]orvyn.Renderable{
			s.dataKeyTitle,
			s.dataKeyList,
		},
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.discoveredKeyList)
	s.focusManager.Add(s.dataKeyList)

	s.layout = layout.NewHBoxGrowFullHeightLayout(1, 0,
		[]orvyn.Renderable{
			discoveredListLayout,
			dataListLayout,
		},
	)

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

	s.discoveredKeys = keys

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

	s.compareKeys()

	s.discoveredKeyList.SetItems(s.discoveredKeys)
	s.dataKeyList.SetItems(s.translationKeys)

	s.focusManager.Focus(0)

	return nil
}

func (s *Screen) OnExit() any {
	return engine.TranslationData{
		Project: s.project,
		Data:    s.data,
	}
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Enter):
			s.mergeData()

			return orvyn.SwitchScreen(screen.IDTranslation)
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) compareKeys() {
	var foundIndex int

	s.alreadyExistingKeys = make([]engine.Key, 0)

	for _, v := range s.translationKeys {
		foundIndex = findInKeyList(v.Key, &s.discoveredKeys)

		if foundIndex == -1 {
			continue
		}

		s.alreadyExistingKeys = append(s.alreadyExistingKeys, v.Key)
	}

	for i := len(s.discoveredKeys) - 1; i >= 0; i-- {
		foundIndex = findInKeyList(s.discoveredKeys[i].Key, &s.translationKeys)

		if foundIndex == -1 {
			continue
		}

		s.discoveredKeys = helper.SliceRemove(s.discoveredKeys, i)
	}

	for _, k := range s.alreadyExistingKeys {
		foundIndex = findInKeyList(k, &s.translationKeys)

		if foundIndex == -1 {
			continue
		}

		s.translationKeys = helper.SliceRemove(s.translationKeys, foundIndex)
	}
}

// mergeData merges remaining elements of both discovered and translations lists into KeyLangMap.
func (s *Screen) mergeData() {
	// Mashup the both list into the data KeyLangMap.

	var foundIndex int

	for k := range s.data {
		foundIndex = findInKeyList(k, &s.translationKeys)

		if foundIndex >= 0 {
			continue
		}

		if slices.Contains(s.alreadyExistingKeys, k) {
			continue
		}

		delete(s.data, k)
	}

	for _, v := range s.discoveredKeys {
		_, ok := s.data[v.Key]

		if ok {
			continue
		}

		s.data[v.Key] = make(map[engine.Lang]engine.Translation)

		for _, l := range s.project.ManagedLanguages {
			s.data[v.Key][l] = engine.Translation{
				Key:        v.Key,
				Lang:       l,
				OneValue:   "",
				OtherValue: "",
				IsPlural:   v.IsPlural,
			}
		}
	}
}

func findInKeyList(key engine.Key, keyList *engine.Keys) int {
	for i, k := range *keyList {
		if k.Key == key {
			return i
		}
	}

	return -1
}
