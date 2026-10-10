package reconciliation

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/screen/dialog/variablelinks"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
)

func setup(t *testing.T) (*Screen, engine.Project) {
	t.Helper()
	lokyn.Init()
	lokyn.SetLanguage("en")
	keybind.Init()
	bubblehelp.Init()
	keymap := bubblehelp.NewKeymap(2)
	for _, binding := range []struct {
		binding key.Binding
		text    string
	}{
		{keybind.AKey, "insert key"}, {keybind.DKey, "delete key"}, {keybind.EKey, "edit/link"},
		{keybind.IKey, "ignore"}, {keybind.RKey, "reset"}, {keybind.Enter, "validate"},
		{keybind.Tab, "next focus"}, {keybind.Quit, "quit"},
	} {
		keymap.NewKeyBinding(binding.binding, true)
		keymap.SetHelpDesc(binding.binding, binding.text)
	}
	bubblehelp.RegisterContext(keybind.ContextReconciliation, keymap)
	orvyn.Init()
	root := t.TempDir()
	project := engine.ProjectNew(root, ".", []string{"en", "fr"})
	if err := engine.ProjectSave(&project); err != nil {
		t.Fatal(err)
	}
	source := `package example
import "github.com/halsten-dev/lokyn"
var literal = lokyn.L("literal")
func render(status string) { lokyn.L(status) }
`
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	translations := `{"literal":"Literal translated","pending":"Waiting","old":"Old"}`
	if err := os.WriteFile(filepath.Join(root, "en.json"), []byte(translations), 0644); err != nil {
		t.Fatal(err)
	}
	s := New()
	s.OnEnter(project)
	s.Render().Resize(orvyn.NewSize(110, 40))
	if s.loadErr != nil {
		t.Fatal(s.loadErr)
	}
	return s, project
}

func runeKey(char rune) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}} }

func TestLinkIgnoreResetAndPersistence(t *testing.T) {
	s, project := setup(t)
	if len(s.variables) != 1 || len(s.common) != 1 || len(s.dataKeyList.GetItems()) != 2 {
		t.Fatalf("unexpected initial state: vars=%+v common=%+v unused=%+v", s.variables, s.common, s.dataKeyList.GetItems())
	}
	if err := s.validate(); err == nil {
		t.Fatal("unresolved variable must block validation")
	}
	variable := s.variables[0].Var
	s.focusManager.Focus(2)
	s.Update(orvyn.DialogExitMsg{DialogID: "variableLinks", Param: variablelinks.Result{Variable: variable, Keys: engine.Keys{"pending", "approved"}, Accepted: true}})
	if s.project.VariableStatus(variable) != engine.VariableLinked || len(s.dataKeyList.GetItems()) != 1 || s.dataKeyList.GetSelectedItem().Key != "old" {
		t.Fatalf("linked keys not protected: %+v", s.dataKeyList.GetItems())
	}
	if project.VariableStatus(variable) != engine.VariableUnresolved {
		t.Fatal("draft edits changed original project")
	}
	s.Update(runeKey('i'))
	if s.project.VariableStatus(variable) != engine.VariableIgnored || len(s.dataKeyList.GetItems()) != 2 {
		t.Fatal("ignoring should restore old linked translations to unused list")
	}
	s.Update(runeKey('r'))
	if s.project.VariableStatus(variable) != engine.VariableUnresolved {
		t.Fatal("reset did not clear decision")
	}
	s.Update(orvyn.DialogExitMsg{DialogID: "variableLinks", Param: variablelinks.Result{Variable: variable, Keys: engine.Keys{"pending", "approved"}, Accepted: true}})
	s.focusManager.Focus(1)
	s.Update(runeKey('d'))
	if err := s.validate(); err != nil {
		t.Fatal(err)
	}
	if len(s.data) != 3 || s.data["pending"]["en"].OneValue != "Waiting" || s.data["literal"]["en"].OneValue != "Literal translated" || s.data["approved"]["fr"].Lang != "fr" {
		t.Fatalf("translations changed/lost: %+v", s.data)
	}
	loaded, err := engine.DiscoverProject(project.LocationPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.VariableStatus(variable) != engine.VariableLinked || len(loaded.VarsKeysLink[variable]) != 2 {
		t.Fatal("links not saved")
	}
	// Reopening applies saved links before classifying unused keys.
	s.OnEnter(loaded)
	if len(s.dataKeyList.GetItems()) != 1 || s.dataKeyList.GetSelectedItem().Key != "old" {
		t.Fatalf("saved links not applied: %+v", s.dataKeyList.GetItems())
	}
	if output := s.Render().Render(); output == "" {
		t.Fatal("empty render")
	}
}

func TestLayoutUsesAvailableListHeight(t *testing.T) {
	s, _ := setup(t)
	var discovered, unused engine.DiscoveredKeys
	s.variables = nil
	for i := 0; i < 50; i++ {
		discovered = append(discovered, engine.DiscoveredKey{Key: engine.Key(fmt.Sprintf("new-%02d %s", i, strings.Repeat("long key ", 12)))})
		unused = append(unused, engine.DiscoveredKey{Key: engine.Key(fmt.Sprintf("unused-%02d", i))})
		s.variables = append(s.variables, engine.DiscoveredVar{Var: engine.Var(fmt.Sprintf("var-%02d", i)), Expression: fmt.Sprintf("dynamic-%02d", i), File: strings.Repeat("long-directory/", 8) + "main.go", Scope: "render", Line: i + 1})
	}
	s.discoveredKeyList.SetItems(discovered)
	s.dataKeyList.SetItems(unused)
	s.refreshVariables()
	s.status.SetError(fmt.Errorf("%s", strings.Repeat("A lengthy validation error. ", 50)))
	for _, size := range []orvyn.Size{orvyn.NewSize(80, 24), orvyn.NewSize(120, 40), orvyn.NewSize(240, 60)} {
		t.Run(fmt.Sprintf("%dx%d", size.Width, size.Height), func(t *testing.T) {
			s.Render().Resize(size)
			view := s.Render().Render()
			if lipgloss.Width(view) > size.Width || lipgloss.Height(view) > size.Height {
				t.Fatalf("rendered %dx%d exceeds terminal %dx%d; lists %v/%dx%d %v/%dx%d %v/%dx%d\n%s", lipgloss.Width(view), lipgloss.Height(view), size.Width, size.Height,
					s.discoveredKeyList.GetSize(), lipgloss.Width(s.discoveredKeyList.Render()), lipgloss.Height(s.discoveredKeyList.Render()),
					s.dataKeyList.GetSize(), lipgloss.Width(s.dataKeyList.Render()), lipgloss.Height(s.dataKeyList.Render()),
					s.variableList.GetSize(), lipgloss.Width(s.variableList.Render()), lipgloss.Height(s.variableList.Render()), view)
			}
			first := strings.Split(view, "\n")[0]
			for _, heading := range []string{"New discovered keys", "Unused keys", "Dynamic expressions"} {
				if !strings.Contains(first, heading) {
					t.Fatalf("headings are not top aligned: %q", first)
				}
			}
			if s.discoveredKeyList.GetSize().Height != s.dataKeyList.GetSize().Height || s.dataKeyList.GetSize().Height != s.variableList.GetSize().Height {
				t.Fatalf("unequal list heights: %v %v %v", s.discoveredKeyList.GetSize(), s.dataKeyList.GetSize(), s.variableList.GetSize())
			}
			if size.Height >= 40 {
				minimum := 3
				if size.Height >= 60 {
					minimum = 6
				}
				for prefix, listView := range map[string]string{
					"new-": s.discoveredKeyList.Render(), "unused-": s.dataKeyList.Render(), "dynamic-": s.variableList.Render(),
				} {
					if count := strings.Count(listView, prefix); count < minimum {
						t.Fatalf("only %d %s rows visible at height %d; want at least %d", count, prefix, size.Height, minimum)
					}
				}
			}
		})
	}
}

func TestManualKeyEditingPreservesValues(t *testing.T) {
	s, _ := setup(t)
	s.Update(runeKey('a'))
	s.Update(runeKey('e'))
	for _, char := range "created" {
		s.Update(runeKey(char))
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if key := s.discoveredKeyList.GetSelectedItem().Key; key != "created" {
		t.Fatalf("manual key=%q", key)
	}
	s.Update(runeKey('e'))
	s.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if key := s.discoveredKeyList.GetSelectedItem().Key; key != "created" {
		t.Fatalf("re-entering edit erased key: %q", key)
	}
}

func TestCancelledDialogDoesNotChangeMappings(t *testing.T) {
	s, _ := setup(t)
	variable := s.variables[0].Var
	s.Update(orvyn.DialogExitMsg{DialogID: "variableLinks", Param: variablelinks.Result{Variable: variable, Keys: engine.Keys{"pending"}}})
	if s.project.VariableStatus(variable) != engine.VariableUnresolved {
		t.Fatal("cancelled dialog changed mapping")
	}
}

func TestSaveFailureDoesNotAdvanceOrReplaceData(t *testing.T) {
	s, project := setup(t)
	s.project.IgnoreVariable(s.variables[0].Var)
	blocker := filepath.Join(project.LocationPath, "not-a-directory")
	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatal(err)
	}
	s.project.LocationPath = blocker
	s.dataKeyList.SetItems(nil)
	if err := s.validate(); err == nil {
		t.Fatal("expected save failure")
	}
	if len(s.data) != 3 || s.data["old"]["en"].OneValue != "Old" {
		t.Fatal("save failure replaced translations")
	}
	loaded, err := engine.DiscoverProject(project.LocationPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.VariableStatus(s.variables[0].Var) != engine.VariableUnresolved {
		t.Fatal("failed save persisted draft")
	}
}

func TestReplacingFilteredUnusedItemsIsSafe(t *testing.T) {
	s, _ := setup(t)
	s.focusManager.Focus(1)
	s.Update(runeKey('/'))
	for _, char := range "pending" {
		s.Update(runeKey(char))
	}
	s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	s.Update(orvyn.DialogExitMsg{DialogID: "variableLinks", Param: variablelinks.Result{Variable: s.variables[0].Var, Keys: engine.Keys{"pending"}, Accepted: true}})
	if len(s.dataKeyList.GetItems()) != 1 || s.dataKeyList.GetSelectedItem().Key != "old" {
		t.Fatalf("stale filtered indices: %+v", s.dataKeyList.GetItems())
	}
	for _, size := range []orvyn.Size{orvyn.NewSize(80, 24), orvyn.NewSize(110, 40)} {
		s.Render().Resize(size)
		if s.Render().Render() == "" {
			t.Fatal("empty render")
		}
	}
}

func TestRemovalIsRememberedAcrossVariableEdits(t *testing.T) {
	s, _ := setup(t)
	s.focusManager.Focus(1)
	// Sorted unused keys start with old.
	s.Update(runeKey('d'))
	s.focusManager.Focus(2)
	s.Update(runeKey('i'))
	if len(s.dataKeyList.GetItems()) != 1 || s.dataKeyList.GetSelectedItem().Key != "pending" {
		t.Fatalf("removed unused key reappeared: %+v", s.dataKeyList.GetItems())
	}
}
