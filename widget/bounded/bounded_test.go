package bounded

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
)

func TestTextRemainsWithinAllocatedRowsAndColumns(t *testing.T) {
	for _, text := range []string{"", strings.Repeat("Long footer text. ", 100), strings.Repeat("你好🌍\n", 20)} {
		child := orvyn.NewSimpleRenderable(text)
		child.SizeConstraint = true
		w := New(child, 2)
		w.Resize(orvyn.NewSize(30, 2))
		view := w.Render()
		if lipgloss.Width(view) > 30 || lipgloss.Height(view) != 2 {
			t.Fatalf("text exceeds allocation: %dx%d", lipgloss.Width(view), lipgloss.Height(view))
		}
	}
}
