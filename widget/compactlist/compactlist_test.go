package compactlist

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

func TestLongPaginatorDoesNotWrapOrIndentListContent(t *testing.T) {
	orvyn.Init()
	w := New(widgetlist.SimpleListItemConstructor)
	items := make([]string, 300)
	for i := range items {
		items[i] = fmt.Sprintf("entry-%03d", i)
	}
	w.SetItems(items)
	w.Resize(orvyn.NewSize(30, 18))
	style := w.GetStyle()
	view := w.Render()
	if lipgloss.Width(view) != 30 || lipgloss.Height(view) != 18 {
		t.Fatalf("rendered %dx%d", lipgloss.Width(view), lipgloss.Height(view))
	}
	if !strings.Contains(view, "1 / 300") || !strings.Contains(view, "entry-000") {
		t.Fatal("missing items or compact pagination")
	}
	// Normal borders and filtering/navigation must still work.
	if !strings.Contains(strings.Split(view, "\n")[1], "╭") {
		t.Fatal("filter border was shifted/wrapped")
	}
	w.Update(tea.KeyMsg{Type: tea.KeyDown})
	if w.GetGlobalIndex() != 1 {
		t.Fatal("navigation broke")
	}
	w.Render()
	if w.GetStyle().GetTransform() != nil || style.GetTransform() != nil {
		t.Fatal("render transformation leaked into stored style")
	}
}
