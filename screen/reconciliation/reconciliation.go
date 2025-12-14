package reconciliation

import (
	"errors"
	"lokyn-cli/engine"
	"lokyn-cli/internal/helper"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/screen"
	"lokyn-cli/widget/help"
	"lokyn-cli/widget/keylistitem"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	discoveredKeyTitle *orvyn.SimpleRenderable
	discoveredKeyList  *widgetlist.Widget[engine.DiscoveredKey]

	dataKeyTitle *orvyn.SimpleRenderable
	dataKeyList  *widgetlist.Widget[engine.DiscoveredKey]

	help *help.Widget

	alreadyExistingKeys []engine.Key
	data                engine.KeyLangMap

	layout *layout.VBoxFullLayout

	focusManager *orvyn.FocusManager

	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	s.discoveredKeyTitle = orvyn.NewSimpleRenderable(lokyn.L("New discovered keys"))
	s.discoveredKeyList = widgetlist.New(keylistitem.Constructor)

	s.dataKeyTitle = orvyn.NewSimpleRenderable(lokyn.L("Unused keys"))
	s.dataKeyList = widgetlist.New(keylistitem.Constructor)

	s.help = help.New()

	discoveredListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		s.discoveredKeyTitle,
		s.discoveredKeyList,
	)

	dataListLayout := layout.NewMaxWidthVBoxFullLayout(
		orvyn.NewSize(0, 0), 1,
		s.dataKeyTitle,
		s.dataKeyList,
	)

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.discoveredKeyList)
	s.focusManager.Add(s.dataKeyList)

	s.layout = layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 1),
		0,
		layout.NewHBoxGrowFullHeightLayout(1, 0,
			discoveredListLayout,
			dataListLayout,
		),
		s.help,
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	project, ok := i.(engine.Project)

	if !ok {
		panic(errors.New("invalid project loaded"))
	}

	s.project = project

	discoveredKeys, err := engine.DiscoverKeys(project.LocationPath)

	if err != nil {
		panic(err)
	}

	data, err := engine.ImportTranslations(&project)

	if err != nil {
		panic(err)
	}

	s.data = data

	translationKeys := make(engine.DiscoveredKeys, 0)

	firstLang := project.ManagedLanguages[0]

	for k, v := range data {
		translationKeys = append(translationKeys, engine.DiscoveredKey{
			Key:      k,
			IsPlural: v[firstLang].IsPlural,
		})
	}

	s.compareKeys()

	s.discoveredKeyList.SetItems(discoveredKeys)
	s.dataKeyList.SetItems(translationKeys)

	s.focusManager.Focus(0)
	s.discoveredKeyList.FocusFirst()
	s.dataKeyList.FocusFirst()

	bubblehelp.SwitchContext(keybind.ContextReconciliation)

	return nil
}

func (s *Screen) OnExit() any {
	return engine.TranslationData{
		Project: s.project,
		Data:    s.data,
	}
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {

	if s.discoveredKeyList.FilterState() != widgetlist.Filtering &&
		s.dataKeyList.FilterState() != widgetlist.Filtering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keybind.Enter):
				s.mergeData()

				return orvyn.SwitchScreen(screen.IDTranslation)
			case key.Matches(msg, keybind.DKey):
				switch {
				case s.discoveredKeyList.IsFocused():
					s.discoveredKeyList.RemoveItem(s.discoveredKeyList.GetGlobalIndex())
					return nil
				case s.dataKeyList.IsFocused():
					s.dataKeyList.RemoveItem(s.dataKeyList.GetGlobalIndex())
					return nil
				}
			}
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

	translationKeys := engine.DiscoveredKeys(s.dataKeyList.GetItems())
	discoveredKeys := engine.DiscoveredKeys(s.discoveredKeyList.GetItems())

	for _, v := range translationKeys {
		if !discoveredKeys.ContainsKey(v.Key) {
			continue
		}

		s.alreadyExistingKeys = append(s.alreadyExistingKeys, v.Key)
	}

	for i := len(discoveredKeys) - 1; i >= 0; i-- {
		if !translationKeys.ContainsKey(discoveredKeys[i].Key) {
			continue
		}

		discoveredKeys = helper.SliceRemove(discoveredKeys, i)
	}

	for _, k := range s.alreadyExistingKeys {
		foundIndex = translationKeys.KeyIndex(k)

		if foundIndex == -1 {
			continue
		}

		translationKeys = helper.SliceRemove(translationKeys, foundIndex)
	}

	s.dataKeyList.SetItems(translationKeys)
	s.discoveredKeyList.SetItems(discoveredKeys)
}

// mergeData merges remaining elements of both discovered and translations lists into KeyLangMap.
func (s *Screen) mergeData() {
	// Mashup the both list into the data KeyLangMap.
	translationKeys := engine.DiscoveredKeys(s.dataKeyList.GetItems())
	discoveredKeys := engine.DiscoveredKeys(s.discoveredKeyList.GetItems())

	for k := range s.data {
		if translationKeys.ContainsKey(k) {
			continue
		}

		if slices.Contains(s.alreadyExistingKeys, k) {
			continue
		}

		delete(s.data, k)
	}

	for _, v := range discoveredKeys {
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

func findInKeyList(key engine.Key, keyList *[]engine.DiscoveredKey) int {
	for i, k := range *keyList {
		if k.Key == key {
			return i
		}
	}

	return -1
}
