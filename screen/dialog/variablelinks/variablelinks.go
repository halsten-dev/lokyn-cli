package variablelinks

import (
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/widget/bounded"
	"lokyn-cli/widget/compactlist"
	"lokyn-cli/widget/truncatedkeylistitem"
	"slices"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Result struct {
	Variable engine.Var
	Keys     engine.Keys
	Accepted bool
}

type Screen struct {
	available *compactlist.Widget[engine.Key]
	linked    *compactlist.Widget[engine.Key]
	input     *textinput.Widget
	status    *statusmessage.Widget
	focus     *orvyn.FocusManager
	layout    *layout.CenterLayout
	adding    bool
	result    Result
}

func New(variable engine.DiscoveredVar, available, linked engine.Keys) *Screen {
	s := &Screen{result: Result{Variable: variable.Var}}
	s.available = compactlist.New(truncatedkeylistitem.Constructor)
	s.linked = compactlist.New(truncatedkeylistitem.Constructor)
	s.input = textinput.New()
	s.input.CharLimit = 0
	s.input.Placeholder = "New concrete key (Enter adds, Esc cancels)"
	s.input.SetActive(false)
	s.status = statusmessage.New()
	s.focus = orvyn.NewFocusManager()
	s.focus.Add(s.available)
	s.focus.Add(s.linked)
	s.available.SetItems(available)
	s.linked.SetItems(slices.Clone(linked))
	kind := "singular"
	if variable.IsPlural {
		kind = "plural"
	}
	title := orvyn.NewSimpleRenderable(fmt.Sprintf("Link %s (%s) — %s:%d (%s)", variable.Expression, kind, variable.File, variable.Line, variable.Scope))
	title.SizeConstraint = true
	instructions := orvyn.NewSimpleRenderable("Tab: change list · Space: link · d: unlink · a: new key · Enter: save links · Esc: cancel")
	instructions.SizeConstraint = true
	s.layout = layout.NewCenterLayout(layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 1), 1,
		bounded.New(title, 2),
		layout.NewHBoxGrowFullHeightLayout(1, 0,
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				bounded.New(orvyn.NewSimpleRenderable("Available keys (filter with /)"), 1), s.available),
			layout.NewMaxWidthVBoxFullLayout(orvyn.NewSize(0, 0), 1,
				bounded.New(orvyn.NewSimpleRenderable("Linked keys ("+kind+")"), 1), s.linked)),
		s.input, bounded.New(s.status, 2), bounded.New(instructions, 2)))
	return s
}

func (s *Screen) OnEnter(any) tea.Cmd {
	s.focus.Focus(0)
	s.available.FocusFirst()
	s.linked.FocusFirst()
	return nil
}
func (s *Screen) OnExit() any          { return s.result }
func (s *Screen) Render() orvyn.Layout { return s.layout }

func (s *Screen) add(key engine.Key) {
	if err := engine.ValidateLinkedKeys(engine.Keys{key}); err != nil {
		s.status.SetError(err)
		return
	}
	if !slices.Contains(s.linked.GetItems(), key) {
		s.linked.AppendItem(key)
	}
	s.status.Reset()
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && key.Matches(k, keybind.Quit) {
		return tea.Quit
	}
	if s.adding {
		if k, ok := msg.(tea.KeyMsg); ok {
			switch {
			case key.Matches(k, keybind.Enter):
				if err := engine.ValidateLinkedKeys(engine.Keys{engine.Key(s.input.Value())}); err != nil {
					s.status.SetError(err)
					return nil
				}
				s.add(engine.Key(s.input.Value()))
				fallthrough
			case key.Matches(k, keybind.Esc):
				s.adding = false
				s.input.OnBlur()
				s.input.SetActive(false)
				s.focus.Focus(s.focus.TabIndex())
				return nil
			}
		}
		return s.input.Update(msg)
	}
	if s.available.FilterState() == widgetlist.Filtering || s.linked.FilterState() == widgetlist.Filtering {
		return s.focus.Update(msg)
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(k, keybind.Esc):
			// First Esc clears an applied list filter; the next cancels the dialog.
			if (s.available.IsFocused() && s.available.FilterState() == widgetlist.FilterApplied) ||
				(s.linked.IsFocused() && s.linked.FilterState() == widgetlist.FilterApplied) {
				return s.focus.Update(msg)
			}
			return orvyn.CloseDialog()
		case key.Matches(k, keybind.Enter):
			keys := engine.Keys(s.linked.GetItems())
			if err := engine.ValidateLinkedKeys(keys); err != nil {
				s.status.SetError(err)
				return nil
			}
			s.result.Keys, s.result.Accepted = keys, true
			return orvyn.CloseDialog()
		case key.Matches(k, keybind.Space):
			if s.available.IsFocused() && s.available.GetGlobalIndex() >= 0 {
				s.add(s.available.GetSelectedItem())
			}
			return nil
		case key.Matches(k, keybind.DKey):
			if s.linked.IsFocused() && s.linked.GetGlobalIndex() >= 0 {
				s.linked.RemoveItem(s.linked.GetGlobalIndex())
			}
			return nil
		case key.Matches(k, keybind.AKey):
			s.focus.BlurCurrent()
			s.adding = true
			s.input.SetActive(true)
			s.input.OnFocus()
			return s.input.Init()
		}
	}
	return s.focus.Update(msg)
}
