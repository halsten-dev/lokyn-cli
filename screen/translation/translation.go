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
	"slices"
	"strings"
)

type Screen struct {
	title *orvyn.SimpleRenderable

	keyList *list.Widget[string]
	keyEdit *keyedit.Widget

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
	s.keyEdit.SetTranslations(
		s.data[engine.Key(s.keys[index])])
}

func (s *Screen) updateKeyList() {
	slices.SortFunc(s.keys, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	s.keyList.SetItems(s.keys)
}
