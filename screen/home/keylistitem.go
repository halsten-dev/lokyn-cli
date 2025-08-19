package home

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"lokyn-cli/engine"
	"lokyn-cli/internal/style"
	"strings"
)

type KeyListItem struct {
	engine.DiscoveredKey
}

func (i KeyListItem) FilterValue() string {
	return ""
}

type KeyListItemDelegate struct {
}

func (l KeyListItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(KeyListItem)

	if !ok {
		return
	}

	var s lipgloss.Style
	var b strings.Builder
	var width int

	width = m.Width() - 2 // borders

	if index == m.Index() {
		s = style.FocusedStyle
	} else {
		s = style.BlurredStyle
	}

	b.WriteString(i.Key)

	tui := s.Width(width).Render(b.String())

	fmt.Fprint(w, tui)

}

func (l KeyListItemDelegate) Height() int {
	return 1
}

func (l KeyListItemDelegate) Spacing() int {
	return 0
}

func (l KeyListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	// selectedIndex := m.GlobalIndex()
	// selectedItem, ok := m.SelectedItem().(KeyListItem)
	//
	// if !ok {
	// 	return nil
	// }
	//
	// updateItem(m, selectedIndex, selectedItem)
	return nil
}

// func updateItem(m *list.Model, index int, item ListItem) {
// 	m.SetItem(index, item)
// }
