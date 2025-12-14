package keyedit

import (
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/widget/keyeditlistitem"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	translationFieldsList *widgetlist.Widget[engine.Translation]

	translations []engine.Translation

	layout *layout.CenterLayout

	project engine.Project
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.translationFieldsList = widgetlist.New(keyeditlistitem.Constructor)
	w.translationFieldsList.SetFilterable(false)

	w.layout = layout.NewCenterLayout(
		w.translationFieldsList,
	)

	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)
	w.layout.Resize(size)
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	cmd := w.translationFieldsList.Update(msg)

	w.translations = w.translationFieldsList.GetItems()

	return cmd
}

func (w *Widget) Render() string {
	return w.layout.Render()
}

func (w *Widget) OnFocus() {
	w.translationFieldsList.OnFocus()
	bubblehelp.SetKeybindVisible(keybind.EKey, true)
}

func (w *Widget) OnBlur() {
	w.translationFieldsList.OnBlur()
	bubblehelp.SetKeybindVisible(keybind.EKey, false)
}

func (w *Widget) IsInputting() bool {
	return w.translationFieldsList.IsInputting()
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) SetTranslations(langMap map[engine.Lang]engine.Translation) {
	if langMap == nil {
		w.translationFieldsList.SetItems(make([]engine.Translation, 0))
		return
	}

	for lang, trans := range langMap {
		for i, t := range w.translations {
			if t.Lang == lang {
				w.translations[i] = trans
				break
			}
		}
	}

	w.translationFieldsList.SetItems(w.translations)
}

func (w *Widget) GetTranslations() []engine.Translation {
	return w.translations
}

func (w *Widget) InitTranslations(p engine.Project) {
	w.project = p

	w.translations = make([]engine.Translation, len(p.ManagedLanguages))

	for i, l := range p.ManagedLanguages {
		w.translations[i] = engine.Translation{
			Key:        "",
			Lang:       l,
			OneValue:   "",
			OtherValue: "",
			IsPlural:   false,
		}
	}

	w.translationFieldsList.SetItems(w.translations)
}
