package translation

import (
	"errors"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/translate"
	"lokyn-cli/screen"
	"lokyn-cli/screen/dialog/progress"
	"lokyn-cli/widget/help"
	"lokyn-cli/widget/keyedit"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Screen struct {
	keyList *widgetlist.Widget[string]
	keyEdit *keyedit.Widget

	statusMessage *statusmessage.Widget

	help *help.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	data    engine.KeyLangMap
	project engine.Project

	progressDialog *progress.Screen
}

func New() *Screen {
	s := new(Screen)

	s.keyList = widgetlist.New(widgetlist.SimpleListItemConstructor)
	s.keyList.CursorMovedCallback = s.keyListCursorMoved

	s.keyEdit = keyedit.New()

	s.statusMessage = statusmessage.New()

	s.help = help.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.keyList)
	s.focusManager.Add(s.keyEdit)

	s.progressDialog = progress.New()

	keyLayout := []layout.FixedRatioRenderable{
		layout.NewFixedRatioRenderable(0.30, s.keyList),
		layout.NewFixedRatioRenderable(0.70, s.keyEdit),
	}

	s.layout = layout.NewCenterLayout(
		layout.NewMaxWidthVBoxFullLayout(
			orvyn.NewSize(0, 1), 0,
			layout.NewHBoxFixedRatioLayout(
				0, 2, 0, keyLayout...,
			),
			s.statusMessage,
			s.help,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	s.data = make(engine.KeyLangMap)

	s.statusMessage.Reset()

	data, ok := i.(engine.TranslationData)

	if !ok {
		panic(errors.New("invalid translation data passed to translation screen"))
	}

	s.project = data.Project
	s.data = data.Data
	keys := make([]string, 0)

	for k := range s.data {
		keys = append(keys, string(k))
	}

	s.keyEdit.InitTranslations(data.Project)

	s.updateKeyList(keys)

	s.focusManager.Focus(0)
	s.keyList.FocusFirst()
	s.keyListCursorMoved(0)

	s.statusMessage.Reset()

	bubblehelp.SwitchContext(keybind.ContextTranslation)
	bubblehelp.SetKeybindVisible(keybind.EKey, false)

	return nil
}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if !s.keyEdit.IsInputting() &&
		s.keyList.FilterState() != widgetlist.Filtering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			s.statusMessage.Reset()
			switch {
			case key.Matches(msg, keybind.TKey):
				s.translateAll()

			case key.Matches(msg, keybind.ShiftTKey):
				return s.translateAllKeys()

			case key.Matches(msg, keybind.CKey):
				s.getCurrentKey()

			case key.Matches(msg, keybind.XKey):
				s.updateData()

				err := engine.ExportAllTranslations(&s.project, s.data)

				if err != nil {
					s.statusMessage.SetError(err)
					return nil
				}

				s.statusMessage.SetMessage(lokyn.L("Successfully exported translations"),
					statusmessage.SuccessMessage)

			case key.Matches(msg, keybind.Esc):
				if s.keyList.FilterState() == widgetlist.Unfiltered {
					return orvyn.SwitchScreen(screen.IDProjectLoading)
				}
			}
		}
	}

	switch msg := msg.(type) {
	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "progressBar":
			s.keyEdit.SetTranslations(s.data[engine.Key(s.keyList.GetSelectedItem())])
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

	if index == -1 {
		s.keyEdit.SetTranslations(nil)
		return
	}

	key := s.keyList.GetItem(index)

	s.keyEdit.SetTranslations(
		s.data[engine.Key(key)])
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

func (s *Screen) updateKeyList(keys []string) {
	slices.SortFunc(keys, func(a, b string) int {
		return strings.Compare(strings.ToLower(a), strings.ToLower(b))
	})

	s.keyList.SetItems(keys)
}

func (s *Screen) translateAllLangs(key engine.Key) {
	var err error
	var trans engine.Translation

	mainLang := s.project.ManagedLanguages[0]
	currentKey := engine.Key(key)

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
}

func (s *Screen) translateAll() {
	s.updateData()

	key := engine.Key(s.keyList.GetSelectedItem())

	s.translateAllLangs(key)

	s.keyEdit.SetTranslations(s.data[key])
}

func (s *Screen) translateAllKeys() tea.Cmd {
	// Loop through every keys

	mainLang := s.project.ManagedLanguages[0]

	go func(dial *progress.Screen) {
		count := 0
		maxSteps := len(s.data)

		dial.UpdateProgress(count, maxSteps)

		for _, k := range s.data {
			count++
			dial.UpdateProgress(count, maxSteps)

			key := k[mainLang].Key
			trans := s.data[key][mainLang]

			if trans.IsPlural {
				continue
			}

			if trans.OneValue != "" {
				continue
			}

			trans.OneValue = string(trans.Key)

			s.data[key][mainLang] = trans

			s.translateAllLangs(key)

			// Delay to avoid too much call to DeepL API
			time.Sleep(800 * time.Millisecond)
		}
	}(s.progressDialog)

	return orvyn.OpenDialog("progressBar", s.progressDialog, nil)
}

func (s *Screen) getCurrentKey() {
	mainLang := s.project.ManagedLanguages[0]
	key := s.keyList.GetSelectedItem()
	currentKey := engine.Key(key)

	mainLangTrans := s.data[currentKey][mainLang]

	mainLangTrans.OneValue = string(currentKey)

	s.data[currentKey][mainLang] = mainLangTrans

	s.keyEdit.SetTranslations(s.data[currentKey])
}
