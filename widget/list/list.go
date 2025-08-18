package list

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/style"
)

// Widget is a very simple list without filter or any feature.
type Widget struct {
	orvyn.BaseWidget

	list.Model

	MinSize       orvyn.Size
	PreferredSize orvyn.Size

	delegate list.ItemDelegate
}

func New(delegate list.ItemDelegate, items []list.Item) *Widget {
	w := new(Widget)

	w.BaseWidget = orvyn.NewBaseWidget()

	w.delegate = delegate

	w.Model = list.New(items, delegate, 0, 0)
	w.Model.DisableQuitKeybindings()

	w.Model.SetShowStatusBar(false)
	w.Model.SetShowFilter(false)
	w.Model.SetShowHelp(false)
	w.Model.SetShowTitle(false)
	w.Model.SetShowPagination(true)

	return w
}

func (w *Widget) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	w.Model, cmd = w.Model.Update(msg)

	return cmd
}

func (w *Widget) Render() string {
	return w.Model.View()
}

func (w *Widget) Resize(size orvyn.Size) {
	w.BaseWidget.Resize(size)

	listItemHeight := w.delegate.Height()
	itemHeight := listItemHeight + style.FocusedStyle.GetVerticalFrameSize()
	itemCount := size.Height / itemHeight

	w.Model.SetWidth(size.Width)
	w.Model.SetHeight(itemCount * listItemHeight)
}

func (w *Widget) GetMinSize() orvyn.Size {
	return w.MinSize
}

func (w *Widget) GetPreferredSize() orvyn.Size {
	return w.PreferredSize
}
