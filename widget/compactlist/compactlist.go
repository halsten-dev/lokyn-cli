// Package compactlist keeps Orvyn's list pagination within one terminal row.
package compactlist

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Widget[T any] struct {
	*widgetlist.Widget[T]
}

func New[T any](constructor func(T) widgetlist.ListItem[T]) *Widget[T] {
	return &Widget[T]{Widget: widgetlist.New(constructor)}
}

func (w *Widget[T]) Render() string {
	// Orvyn's dot paginator has one dot for every page, but reserves only one
	// row. Shorten it BEFORE the surrounding style wraps/borders the content.
	// Otherwise a long pager also causes JoinVertical to indent every item.
	style := w.GetStyle()
	width := w.GetContentSize().Width
	w.SetStyle(style.Transform(func(body string) string {
		lines := strings.Split(body, "\n")
		isPager := func(line string) bool {
			plain := strings.TrimSpace(ansi.Strip(line))
			return strings.Contains(plain, "•") && strings.Trim(plain, "• ") == ""
		}
		hasPager := false
		for _, line := range lines {
			if isPager(line) {
				hasPager = true
				break
			}
		}
		if !hasPager {
			return body
		}
		// Undo centering introduced by an oversized pager. List items/filter
		// already have the allocated content width; only the dots can exceed it.
		indent := max(0, (lipgloss.Width(body)-width)/2)
		for i, line := range lines {
			if isPager(line) {
				count := len(w.GetItems())
				index := min(max(0, w.GetGlobalIndex()+1), count)
				text := ansi.Truncate(fmt.Sprintf("%d / %d", index, count), width, "…")
				lines[i] = lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(text)
			} else if indent > 0 {
				lines[i] = ansi.Cut(line, indent, indent+width)
			}
		}
		return strings.Join(lines, "\n")
	}))
	defer w.SetStyle(style)
	return w.Widget.Render()
}
