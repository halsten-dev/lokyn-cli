// Package bounded constrains informational text to its allocated terminal area.
package bounded

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
)

// Widget is a fixed-height wrapper for labels/status messages whose native
// renderers wrap text but do not enforce a maximum height.
type Widget struct {
	orvyn.BaseRenderable
	child orvyn.Renderable
}

func New(child orvyn.Renderable, height int) *Widget {
	w := &Widget{BaseRenderable: orvyn.NewBaseRenderable(), child: child}
	w.SetMinSize(orvyn.NewSize(1, height))
	w.SetPreferredSize(orvyn.NewSize(1, height))
	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Width, size.Height = max(0, size.Width), max(0, size.Height)
	w.BaseRenderable.Resize(size)
	w.child.Resize(size)
}

func (w *Widget) Render() string {
	size := w.GetSize()
	if size.Width == 0 || size.Height == 0 {
		return ""
	}
	lines := strings.Split(ansi.Hardwrap(w.child.Render(), size.Width, false), "\n")
	if len(lines) > size.Height {
		lines = lines[:size.Height]
		lines[len(lines)-1] = ansi.Truncate(lines[len(lines)-1], max(0, size.Width-1), "") + "…"
	}
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, size.Width, "…")
	}
	return lipgloss.NewStyle().Width(size.Width).Height(size.Height).
		MaxWidth(size.Width).MaxHeight(size.Height).Render(strings.Join(lines, "\n"))
}
