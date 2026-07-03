package popup

import (
	"fmt"
	"lokyn-cli/internal/keybind"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/layout"
	"github.com/halsten-dev/orvyn/theme"
)

type Option struct {
	Keybind key.Binding
	Text    string
	Value   uint
}

type Config struct {
	Message string
	Options []Option
}

type Screen struct {
	config Config

	content *orvyn.SimpleRenderable
	options *orvyn.SimpleRenderable

	layout *layout.CenterLayout

	value uint
}

func NewYesNo(message string) *Screen {
	options := []Option{
		{
			Keybind: keybind.YKey,
			Text:    lokyn.L("Yes"),
			Value:   1,
		},
		{
			Keybind: keybind.NKey,
			Text:    lokyn.L("No"),
			Value:   2,
		},
	}

	config := Config{
		Message: message,
		Options: options,
	}

	return New(config)
}

func New(config Config) *Screen {
	var b strings.Builder

	s := new(Screen)

	t := orvyn.GetTheme()

	s.config = config

	b.WriteString(config.Message)
	b.WriteString("\n\n")

	s.content = orvyn.NewSimpleRenderable(b.String())
	s.content.Style = t.Style(theme.TitleStyleID).AlignHorizontal(lipgloss.Center)
	s.content.SizeConstraint = true

	b.Reset()

	for i, o := range config.Options {
		if i > 0 {
			fmt.Fprintf(&b, " %c ", '•')
		}

		fmt.Fprintf(&b, "%s %s",
			t.Style(theme.NormalTextStyleID).Render(o.Keybind.Help().Key),
			t.Style(theme.DimTextStyleID).Render(o.Text))
	}

	s.options = orvyn.NewSimpleRenderable(b.String())

	s.layout = layout.NewCenterLayout(
		layout.NewVBoxLayout(10,
			s.content,
			s.options,
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	return nil
}

func (s *Screen) OnExit() any {
	return s.value
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keybind.Quit) {
			return tea.Quit
		}

		for _, o := range s.config.Options {
			if key.Matches(msg, o.Keybind) {
				s.value = o.Value
				return orvyn.CloseDialog()
			}
		}
	}

	return nil
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
