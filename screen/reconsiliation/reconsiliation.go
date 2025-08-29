package reconsiliation

import (
	"lokyn-cli/engine"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/orvyn"
)

type Screen struct {
	discoverdKeyList *list.Widget[engine.DiscoveredKey]
}

func (s *Screen) OnEnter(param any) {

}

func (s *Screen) OnExit() any {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return nil
}
