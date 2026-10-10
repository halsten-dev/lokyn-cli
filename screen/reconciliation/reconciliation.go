package reconciliation

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/screen"
	"lokyn-cli/screen/dialog/variablelinks"
	"lokyn-cli/widget/bounded"
	"lokyn-cli/widget/compactlist"
	"lokyn-cli/widget/help"
	"lokyn-cli/widget/keylistitem"
	"lokyn-cli/widget/variablelistitem"
	"slices"
	"strings"

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
	discoveredKeyList *compactlist.Widget[engine.DiscoveredKey]
	dataKeyList       *compactlist.Widget[engine.DiscoveredKey]
	variableList      *compactlist.Widget[variablelistitem.Data]
	variableDetail    *orvyn.SimpleRenderable
	status            *statusmessage.Widget
	help              *help.Widget
	layout            *layout.VBoxFullLayout
	focusManager      *orvyn.FocusManager
	project           engine.Project
	data              engine.KeyLangMap
	variables         engine.DiscoveredVars
	literals          engine.DiscoveredKeys
	common            engine.DiscoveredKeys
	removedUnused     map[engine.Key]bool
	loadErr           error
}

func New() *Screen {
	s := &Screen{}
	s.discoveredKeyList = compactlist.New(keylistitem.Constructor)
	s.discoveredKeyList.AutoFocusNewItem = true
	s.dataKeyList = compactlist.New(keylistitem.Constructor)
	s.variableList = compactlist.New(variablelistitem.Constructor)
	s.variableList.CursorMovedCallback = func(int) { s.refreshVariableDetail() }
	s.variableDetail = orvyn.NewSimpleRenderable("")
	s.variableDetail.SizeConstraint = true
	s.status = statusmessage.New()
	s.help = help.New()
	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.discoveredKeyList)
	s.focusManager.Add(s.dataKeyList)
	s.focusManager.Add(s.variableList)
	column := func(title string, list orvyn.Renderable) *layout.VBoxFullLayout {
		return layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
			bounded.New(orvyn.NewSimpleRenderable(title), 1), list)
	}
	explanation := orvyn.NewSimpleRenderable("Resolve or explicitly ignore every dynamic expression. Linked keys are protected; remaining unused keys are retained unless removed.")
	explanation.SizeConstraint = true
	s.layout = layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 1), 0,
		layout.NewHBoxGrowFullHeightLayout(1, 0,
			column(lokyn.L("New discovered keys"), s.discoveredKeyList),
			column(lokyn.L("Unused keys"), s.dataKeyList),
			column("Dynamic expressions", s.variableList)),
		bounded.New(s.variableDetail, 2), bounded.New(explanation, 2), bounded.New(s.status, 2), bounded.New(s.help, 2))
	return s
}

func sortKeys(keys engine.DiscoveredKeys) {
	slices.SortFunc(keys, func(a, b engine.DiscoveredKey) int { return strings.Compare(string(a.Key), string(b.Key)) })
}

func (s *Screen) OnEnter(input any) tea.Cmd {
	s.loadErr = nil
	s.status.Reset()
	s.common = nil
	s.variables = nil
	s.literals = nil
	s.data = make(engine.KeyLangMap)
	s.removedUnused = make(map[engine.Key]bool)
	s.discoveredKeyList.Init()
	s.dataKeyList.Init()
	s.variableList.Init()
	s.discoveredKeyList.SetItems(nil)
	s.dataKeyList.SetItems(nil)
	s.variableList.SetItems(nil)
	project, ok := input.(engine.Project)
	if !ok {
		s.loadErr = fmt.Errorf("invalid project loaded")
		s.status.SetError(s.loadErr)
		return nil
	}
	s.project = project.CloneVariableLinks()
	keys, variables, err := engine.DiscoverKeys(project.LocationPath)
	if err != nil {
		s.loadErr = err
		s.status.SetError(err)
		return nil
	}
	data, err := engine.ImportTranslations(&s.project)
	if err != nil {
		s.loadErr = err
		s.status.SetError(err)
		return nil
	}
	s.data, s.variables, s.literals = data, variables, keys
	s.project.MigrateLegacyLinks(variables)
	var discovered engine.DiscoveredKeys
	for _, key := range keys {
		if _, exists := data[key.Key]; exists {
			s.common = append(s.common, key)
		} else {
			discovered = append(discovered, key)
		}
	}
	sortKeys(discovered)
	s.discoveredKeyList.SetItems(discovered)
	s.refreshVariables()
	s.refreshUnused()
	s.focusManager.Focus(0)
	s.discoveredKeyList.FocusFirst()
	s.dataKeyList.FocusFirst()
	s.variableList.FocusFirst()
	s.refreshVariableDetail()
	if count := s.project.DormantVariableCount(variables); count > 0 {
		s.status.SetMessage(fmt.Sprintf("%d saved mapping(s) are no longer active; preserved for review in .lokyn/lokynproj.json", count), statusmessage.WarningMessage)
	}
	bubblehelp.SwitchContext(keybind.ContextReconciliation)
	return nil
}

func (s *Screen) OnExit() any          { return engine.TranslationData{Project: s.project, Data: s.data} }
func (s *Screen) Render() orvyn.Layout { return s.layout }

func (s *Screen) refreshVariables() {
	rows := make([]variablelistitem.Data, 0, len(s.variables))
	for _, variable := range s.variables {
		rows = append(rows, variablelistitem.Data{Variable: variable, Status: s.project.VariableStatus(variable.Var), Keys: s.project.VarsKeysLink[variable.Var]})
	}
	s.variableList.Init() // Clear filters before replacing their backing items.
	s.variableList.SetItems(rows)
	s.refreshVariableDetail()
}

func (s *Screen) refreshVariableDetail() {
	v := s.variableList.GetSelectedItem().Variable
	if v.Var == "" {
		s.variableDetail.SetValue("")
		return
	}
	links := s.project.VarsKeysLink[v.Var]
	s.variableDetail.SetValue(fmt.Sprintf("%s · %s:%d · %s · %s · links: %q", v.Expression, v.File, v.Line, v.Scope, s.project.VariableStatus(v.Var), links))
}

func (s *Screen) importedKey(key engine.Key) engine.DiscoveredKey {
	plural := false
	for _, lang := range s.project.ManagedLanguages {
		if translation, ok := s.data[key][lang]; ok {
			plural = translation.IsPlural
			break
		}
	}
	return engine.DiscoveredKey{Key: key, IsPlural: plural}
}

func (s *Screen) refreshUnused() {
	linked := s.project.LinkedDiscoveredKeys(s.variables)
	var unused engine.DiscoveredKeys
	for key := range s.data {
		if s.literals.ContainsKey(key) || linked.ContainsKey(key) || s.removedUnused[key] {
			continue
		}
		unused = append(unused, s.importedKey(key))
	}
	sortKeys(unused)
	s.dataKeyList.Init() // Avoid stale filtered indices after linked keys disappear.
	s.dataKeyList.SetItems(unused)
	s.dataKeyList.FocusFirst()
}

func (s *Screen) availableKeys() engine.Keys {
	seen := make(map[engine.Key]bool)
	for key := range s.data {
		seen[key] = true
	}
	for _, key := range s.literals {
		seen[key.Key] = true
	}
	for _, key := range s.discoveredKeyList.GetItems() {
		if key.Key != "" {
			seen[key.Key] = true
		}
	}
	for _, keys := range s.project.VarsKeysLink {
		for _, key := range keys {
			seen[key] = true
		}
	}
	keys := make(engine.Keys, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func (s *Screen) selectedKeys() engine.DiscoveredKeys {
	selected := slices.Clone(s.common)
	selected = append(selected, s.discoveredKeyList.GetItems()...)
	selected = append(selected, s.dataKeyList.GetItems()...)
	return selected
}

func (s *Screen) validate() error {
	if s.loadErr != nil {
		return s.loadErr
	}
	data, err := engine.ReconcileTranslations(s.project, s.variables, s.selectedKeys(), s.data)
	if err != nil {
		return err
	}
	// Do not advance or replace translations unless the associations were saved.
	if err := engine.ProjectSave(&s.project); err != nil {
		return fmt.Errorf("save variable associations: %w", err)
	}
	s.data = data
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if exit, ok := msg.(orvyn.DialogExitMsg); ok && exit.DialogID == "variableLinks" {
		if result, ok := exit.Param.(variablelinks.Result); ok && result.Accepted {
			if err := s.project.LinkVariable(result.Variable, result.Keys); err != nil {
				s.status.SetError(err)
				return nil
			}
			s.status.Reset()
			s.refreshVariables()
			s.refreshUnused()
		}
		bubblehelp.SwitchContext(keybind.ContextReconciliation)
		return nil
	}
	if s.discoveredKeyList.FilterState() != widgetlist.Filtering &&
		s.dataKeyList.FilterState() != widgetlist.Filtering && s.variableList.FilterState() != widgetlist.Filtering &&
		!s.discoveredKeyList.IsInputting() && !s.dataKeyList.IsInputting() && !s.variableList.IsInputting() {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(k, keybind.Enter):
				if err := s.validate(); err != nil {
					s.status.SetError(err)
					return nil
				}
				return orvyn.SwitchScreen(screen.IDTranslation)
			case key.Matches(k, keybind.EKey):
				if s.variableList.IsFocused() {
					variable := s.variableList.GetSelectedItem().Variable
					if variable.Var != "" {
						return orvyn.OpenDialog("variableLinks", variablelinks.New(variable, s.availableKeys(), s.project.VarsKeysLink[variable.Var]), nil)
					}
					return nil
				}
			case key.Matches(k, keybind.IKey), key.Matches(k, keybind.RKey):
				if s.variableList.IsFocused() {
					variable := s.variableList.GetSelectedItem().Variable
					if variable.Var == "" {
						return nil
					}
					if key.Matches(k, keybind.IKey) {
						s.project.IgnoreVariable(variable.Var)
					} else {
						s.project.ResetVariable(variable.Var)
					}
					s.status.Reset()
					s.refreshVariables()
					s.refreshUnused()
					return nil
				}
			case key.Matches(k, keybind.AKey):
				s.focusManager.Focus(0)
				s.discoveredKeyList.InsertItem(max(0, s.discoveredKeyList.GetGlobalIndex()), engine.DiscoveredKey{IsCreatedByUser: true})
				return nil
			case key.Matches(k, keybind.DKey):
				if s.discoveredKeyList.IsFocused() {
					s.discoveredKeyList.RemoveItem(s.discoveredKeyList.GetGlobalIndex())
					return nil
				}
				if s.dataKeyList.IsFocused() {
					if key := s.dataKeyList.GetSelectedItem().Key; key != "" {
						s.removedUnused[key] = true
					}
					s.dataKeyList.RemoveItem(s.dataKeyList.GetGlobalIndex())
					return nil
				}
			}
		}
	}
	return s.focusManager.Update(msg)
}
