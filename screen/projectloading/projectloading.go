package projectloading

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/layout"
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/screen/dialog/popup"
	"lokyn-cli/widget/help"
	"lokyn-cli/widget/label"
	"lokyn-cli/widget/textinput"
	"os"
)

var (
	keymapContext = bubblehelp.NewKeymap(2)
)

type Screen struct {
	project engine.Project

	label     *label.Widget
	languages *textinput.Widget

	help *help.Widget

	layout *layout.CenterLayout
}

func New() *Screen {
	s := new(Screen)

	keymapContext.NewKeyBinding(keybind.Enter, true)
	keymapContext.SetHelpDesc(keybind.Enter, "create project")
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.SetHelpDesc(keybind.Esc, "cancel")

	bubblehelp.RegisterContext(keybind.ContextProjectLoading, keymapContext)

	s.label = label.New("wanted languages (separated by a coma)")
	s.languages = textinput.New()
	s.languages.Focus()
	s.help = help.New()

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(30, 100, 10,
			[]orvyn.Renderable{
				s.label,
				s.languages,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i interface{}) tea.Cmd {
	var err error
	var projectFound bool

	bubblehelp.SwitchContext(keybind.ContextProjectLoading)

	projectFound = false

	// Is there a valid Lokyn project on the current directory ?
	s.project, err = engine.DiscoverProject()

	if err != nil {
		s.project = engine.Project{}
	} else {
		projectFound = true
	}

	// Ask for project creation or exit
	if !projectFound {
		currentDir, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		orvyn.OpenDialog("AskProjectCreation", popup.NewYesNo(
			fmt.Sprintf("Do you want to create a lokyn project in : %s",
				currentDir)), nil)
	}

	return s.languages.Init()
}

func (s *Screen) OnExit() interface{} {
	return nil
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Enter):
			// Validate
			return nil

		case key.Matches(msg, keybind.Esc):
			return tea.Quit
		}

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "AskProjectCreation":
			val := msg.Param.(uint)

			switch val {
			case 1:
				return s.languages.Init()
			default:
				return tea.Quit
			}
		}
	}

	cmd := s.languages.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}
