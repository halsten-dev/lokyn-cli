package translation

import (
	"errors"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/translate"
	"lokyn-cli/widget/keyedit"
	"slices"
	"strings"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	keyList *list.Widget[string]
	keyEdit *keyedit.Widget

	statusMessage *statusmessage.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	keys    []string
	data    engine.KeyLangMap
	project engine.Project
}

func New() *Screen {
	s := new(Screen)

	s.title = orvyn.NewSimpleRenderable(lokyn.L("Translation"))
	s.title.SizeConstraint = true

	s.keyList = list.New(list.SimpleListItemConstructor)
	s.keyList.CursorMovedCallback = s.keyListCursorMoved

	s.keyEdit = keyedit.New()

	s.statusMessage = statusmessage.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.keyList)
	s.focusManager.Add(s.keyEdit)

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(0, 0),
			0,
			[]orvyn.Renderable{
				layout.NewHBoxFixedRatioLayout(
					0, 2, 0,
					[]layout.FixedRatioRenderable{
						layout.NewFixedRatioRenderable(0.30, s.keyList),
						layout.NewFixedRatioRenderable(0.70, s.keyEdit),
					},
				),
				s.statusMessage,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.data = make(engine.KeyLangMap)

	data, ok := i.(engine.TranslationData)

	if !ok {
		panic(errors.New("invalid translation data passed to translation screen"))
	}

	s.project = data.Project
	s.data = data.Data
	s.keys = make([]string, 0)

	for k := range s.data {
		s.keys = append(s.keys, string(k))
	}

	s.keyEdit.InitTranslations(data.Project)

	s.updateKeyList()

	s.focusManager.Focus(0)
	s.keyList.FocusItem(0)
	s.keyListCursorMoved(0)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if !s.keyEdit.IsInputting() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, keybind.TKey):
				s.translateAll()

			case key.Matches(msg, keybind.KKey):
				s.getCurrentKey()

			case key.Matches(msg, keybind.XKey):
				err := engine.ExportAllTranslations(&s.project, s.data)

				if err != nil {
					s.statusMessage.SetError(err)
					return nil
				}

				s.statusMessage.SetMessage(lokyn.L("Successfully exported translations"),
					statusmessage.SuccessMessage)
			}
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) keyListCursorMoved(index int) {
	s.updateData()

	s.keyEdit.SetTranslations(
		s.data[engine.Key(s.keys[index])])
}

func (s *Screen) updateData() {
	translations := s.keyEdit.GetTranslations()

	for _, t := range translations {
		if t.Key == "" {
			continue
		}

		s.data[t.Key][t.Lang] = t
	}
}

func (s *Screen) updateKeyList() {
	slices.SortFunc(s.keys, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	s.keyList.SetItems(s.keys)
}

func (s *Screen) translateAll() {
	var err error
	var trans engine.Translation

	s.updateData()

	mainLang := s.project.ManagedLanguages[0]
	currentKey := engine.Key(s.keys[s.keyList.GetGlobalIndex()])

	mainLangTrans := s.data[currentKey][mainLang]

	for _, l := range s.project.ManagedLanguages {
		if l == mainLang {
			continue
		}

		trans = s.data[currentKey][l]

		if trans.Lang != l {
			trans = engine.Translation{
				Key:        currentKey,
				Lang:       l,
				OneValue:   "",
				OtherValue: "",
				IsPlural:   mainLangTrans.IsPlural,
			}
		}

		trans.OneValue, err = translate.Get(mainLangTrans.OneValue, string(mainLang), string(l))

		if err != nil {
			s.statusMessage.SetError(err)
			return
		}

		if trans.IsPlural {
			trans.OtherValue, err = translate.Get(mainLangTrans.OtherValue, string(mainLang), string(l))

			if err != nil {
				s.statusMessage.SetError(err)
				return
			}
		}

		s.data[currentKey][l] = trans
	}

	s.keyEdit.SetTranslations(s.data[currentKey])
}

func (s *Screen) getCurrentKey() {
	mainLang := s.project.ManagedLanguages[0]
	currentKey := engine.Key(s.keys[s.keyList.GetGlobalIndex()])

	mainLangTrans := s.data[currentKey][mainLang]

	mainLangTrans.OneValue = string(currentKey)

	s.data[currentKey][mainLang] = mainLangTrans

	s.keyEdit.SetTranslations(s.data[currentKey])
}
