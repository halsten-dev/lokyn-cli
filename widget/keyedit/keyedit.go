package keyedit

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/list"
	"lokyn-cli/engine"
	"lokyn-cli/internal/layout"
	"lokyn-cli/widget/keyeditlistitem"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	translationFieldsList *list.Widget[engine.Translation]

	translations []engine.Translation

	layout *layout.CenterLayout

	project engine.Project
}

func New() *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.translationFieldsList = list.New(keyeditlistitem.Constructor)

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
	w.translationFieldsList.FocusItem(0)
}

func (w *Widget) OnBlur() {
	w.translationFieldsList.OnBlur()
	w.translationFieldsList.BlurCurrent()
}

func (w *Widget) IsInputting() bool {
	return w.translationFieldsList.IsInputting()
}

func (w *Widget) OnEnterInput() {}

func (w *Widget) OnExitInput() {}

func (w *Widget) SetTranslations(langMap map[engine.Lang]engine.Translation) {
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
