package translation

import (
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

func TestEmptyReconciledDataDoesNotPanic(t *testing.T) {
	lokyn.Init()
	lokyn.SetLanguage("en")
	keybind.Init()
	bubblehelp.Init()
	bubblehelp.RegisterContext(keybind.ContextTranslation, bubblehelp.NewKeymap(2))
	orvyn.Init()
	s := New()
	s.OnEnter(engine.TranslationData{Project: engine.ProjectNew(t.TempDir(), ".", []string{"en"}), Data: make(engine.KeyLangMap)})
	s.Render().Resize(orvyn.NewSize(100, 35))
	for _, char := range []rune{'c', 't', 'T', 'x'} {
		s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
	}
	if len(s.data) != 0 {
		t.Fatalf("empty reconciliation created bogus keys: %+v", s.data)
	}
}
