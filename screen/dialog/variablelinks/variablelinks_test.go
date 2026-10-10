package variablelinks

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

func keyMsg(char rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}} }

func newTestDialog() *Screen {
	lokyn.Init()
	lokyn.SetLanguage("en")
	keybind.Init()
	orvyn.Init()
	s := New(engine.DiscoveredVar{Var: "id", Expression: "status", File: "main.go", Scope: "render"}, engine.Keys{"pending", "approved"}, nil)
	orvyn.OpenDialog("variableLinks", s, nil)
	s.Render().Resize(orvyn.NewSize(100, 35))
	return s
}

func TestPickerLayoutUsesListHeightAndFitsTerminal(t *testing.T) {
	s := newTestDialog()
	var keys engine.Keys
	for i := 0; i < 200; i++ {
		keys = append(keys, engine.Key(fmt.Sprintf("key-%03d %s", i, strings.Repeat("long label ", 12))))
	}
	s.available.SetItems(keys)
	s.linked.SetItems(keys)
	s.status.SetError(fmt.Errorf("%s", strings.Repeat("Long validation error. ", 40)))
	for _, size := range []orvyn.Size{orvyn.NewSize(80, 24), orvyn.NewSize(120, 40), orvyn.NewSize(240, 60)} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			s.Render().Resize(size)
			view := s.Render().Render()
			if lipgloss.Width(view) > size.Width || lipgloss.Height(view) > size.Height {
				t.Fatalf("rendered %dx%d exceeds terminal %dx%d", lipgloss.Width(view), lipgloss.Height(view), size.Width, size.Height)
			}
			if s.available.GetSize().Height != s.linked.GetSize().Height {
				t.Fatal("list headings/rows are not aligned")
			}
			if size.Height >= 40 && strings.Count(view, "key-") < 12 {
				t.Fatalf("list area is not growing: only %d visible rows", strings.Count(view, "key-"))
			}
		})
	}
}

func TestPickerLinksCreatesUnlinksAndSaves(t *testing.T) {
	s := newTestDialog()
	s.Update(tea.KeyMsg{Type: tea.KeySpace})
	s.Update(tea.KeyMsg{Type: tea.KeySpace})
	if len(s.linked.GetItems()) != 1 {
		t.Fatal("duplicate link was added")
	}
	s.Update(keyMsg('a'))
	for _, char := range "new key" {
		s.Update(keyMsg(char))
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if s.adding || len(s.linked.GetItems()) != 2 || s.linked.GetItems()[1] != "new key" {
		t.Fatalf("failed to add new key: %+v", s.linked.GetItems())
	}
	s.Update(tea.KeyMsg{Type: tea.KeyTab})
	s.Update(keyMsg('d'))
	if len(s.linked.GetItems()) != 1 || s.linked.GetSelectedItem() != "new key" {
		t.Fatalf("failed unlink: %+v", s.linked.GetItems())
	}
	cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || !s.result.Accepted || len(s.result.Keys) != 1 || s.result.Keys[0] != "new key" {
		t.Fatalf("bad result: %+v", s.result)
	}
	if msg, ok := cmd().(orvyn.DialogExitMsg); !ok || msg.DialogID != "variableLinks" {
		t.Fatal("did not close dialog")
	}
}

func TestPickerRejectsEmptyAndCancelsInput(t *testing.T) {
	s := newTestDialog()
	if cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil || s.result.Accepted {
		t.Fatal("empty links must not save")
	}
	s.Update(keyMsg('a'))
	if cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil || !s.adding {
		t.Fatal("empty new key must remain in editor")
	}
	s.Update(keyMsg('x'))
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if s.adding || len(s.linked.GetItems()) != 0 {
		t.Fatal("cancelled input created a key")
	}
	if cmd := s.Update(tea.KeyMsg{Type: tea.KeyEsc}); cmd == nil || s.result.Accepted {
		t.Fatal("cancelled picker accepted changes")
	}
}

func TestPickerFilteringDoesNotTriggerActions(t *testing.T) {
	s := newTestDialog()
	s.Update(keyMsg('/'))
	for _, char := range "approved" {
		s.Update(keyMsg(char))
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if s.result.Accepted || len(s.linked.GetItems()) != 0 {
		t.Fatal("filtering triggered actions")
	}
	s.Update(tea.KeyMsg{Type: tea.KeySpace})
	if len(s.linked.GetItems()) != 1 || s.linked.GetSelectedItem() != "approved" {
		t.Fatalf("filtered selection used wrong key: %+v", s.linked.GetItems())
	}
}
