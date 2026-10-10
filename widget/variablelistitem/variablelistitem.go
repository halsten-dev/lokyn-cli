package variablelistitem

import (
	"fmt"
	"lokyn-cli/engine"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/widgetlist"
)

type Data struct {
	Variable engine.DiscoveredVar
	Status   engine.VariableStatus
	Keys     engine.Keys
}

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable
	data Data
}

func Constructor(data Data) widgetlist.ListItem[Data] {
	w := &Widget{BaseWidget: orvyn.NewBaseWidget(), data: data}
	w.BaseFocusable = orvyn.NewBaseFocusable(w)
	return w
}

func (w *Widget) Resize(size orvyn.Size) {
	size.Height = 6
	w.BaseWidget.Resize(size)
}

func (w *Widget) Update(tea.Msg) tea.Cmd { return nil }
func (w *Widget) UpdateData(data Data)   { w.data = data }
func (w *Widget) GetData() Data          { return w.data }
func (w *Widget) FilterValue() string {
	return w.data.Variable.Expression + " " + w.data.Variable.File + " " + w.data.Variable.Scope
}
func (w *Widget) Render() string {
	v := w.data.Variable
	kind := "singular"
	if v.IsPlural {
		kind = "plural"
	}
	status := string(w.data.Status)
	if v.Err != nil {
		status += " / conflict"
	}
	keys := make([]string, len(w.data.Keys))
	for i, key := range w.data.Keys {
		keys[i] = string(key)
	}
	lines := []string{
		v.Expression,
		fmt.Sprintf("%s:%d", v.File, v.Line),
		v.Scope + " · " + kind,
		status + " · " + strings.Join(keys, ", "),
	}
	size := w.GetContentSize()
	for i, line := range lines {
		line = strings.ReplaceAll(strings.ReplaceAll(line, "\n", " "), "\r", " ")
		lines[i] = ansi.Truncate(line, max(0, size.Width), "…")
	}
	return w.GetStyle().Width(size.Width).Height(size.Height).Render(strings.Join(lines, "\n"))
}
