package projectloading

import (
	"errors"
	"fmt"
	"lokyn-cli/engine"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/screen"
	"lokyn-cli/screen/dialog/popup"
	"lokyn-cli/widget/help"
	"os"
	"strings"

	"github.com/halsten-dev/orvyn/layout"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/widget/label"
	"github.com/halsten-dev/orvyn/widget/statusmessage"
	"github.com/halsten-dev/orvyn/widget/textinput"
)

var (
	keymapContext = bubblehelp.NewKeymap(2)
)

type Screen struct {
	project engine.Project

	labelExportPath *label.Widget
	exportPath      *textinput.Widget

	labelLanguages *label.Widget
	languages      *textinput.Widget

	help *help.Widget

	statusMessage *statusmessage.Widget

	focusManager *orvyn.FocusManager

	layout *layout.CenterLayout

	currentDir string
}

func New() *Screen {
	var err error

	s := new(Screen)

	s.currentDir, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	keymapContext.NewKeyBinding(keybind.Enter, true)
	keymapContext.SetHelpDesc(keybind.Enter, "create project")
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.SetHelpDesc(keybind.Esc, "cancel")

	bubblehelp.RegisterContext(keybind.ContextProjectLoading, keymapContext)

	s.labelExportPath = label.New(
		fmt.Sprintf("translation export path (relative to %s)", s.currentDir))
	s.exportPath = textinput.New()

	s.labelLanguages = label.New("wanted languages (separated by a coma)")
	s.languages = textinput.New()

	s.help = help.New()

	s.statusMessage = statusmessage.New()

	s.focusManager = orvyn.NewFocusManager()
	s.focusManager.Add(s.exportPath)
	s.focusManager.Add(s.languages)

	s.layout = layout.NewCenterLayout(
		layout.NewDefinedWidthVerticalLayout(30, 100, 10,
			[]orvyn.Renderable{
				s.labelExportPath,
				s.exportPath,
				orvyn.VGap,
				s.labelLanguages,
				s.languages,
				orvyn.VGap,
				s.statusMessage,
				s.help,
			},
		),
	)

	return s
}

func (s *Screen) OnEnter(i any) tea.Cmd {
	var err error

	bubblehelp.SwitchContext(keybind.ContextProjectLoading)

	s.project, err = engine.DiscoverProject()

	if err != nil {
		s.project = engine.Project{}
	} else {
		return orvyn.SwitchScreen(screen.IDReconciliation)
	}

	orvyn.OpenDialog("AskProjectCreation", popup.NewYesNo(
		fmt.Sprintf("Do you want to create a lokyn project in : %s",
			s.currentDir)), nil)

	s.focusManager.Focus(0)

	return s.exportPath.Init()
}

func (s *Screen) OnExit() any {
	return s.project
}

func (s *Screen) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keybind.Enter):
			if s.projectCreation() {
				return orvyn.SwitchScreen(screen.IDReconciliation)
			}

		case key.Matches(msg, keybind.Esc):
			return tea.Quit
		}

	case orvyn.DialogExitMsg:
		switch msg.DialogID {
		case "AskProjectCreation":
			val := msg.Param.(uint)

			switch val {
			case 1:
				// For blinking cursor
				return s.exportPath.Init()
			default:
				return tea.Quit
			}
		}
	}

	cmd := s.focusManager.Update(msg)

	return cmd
}

func (s *Screen) Render() orvyn.Layout {
	return s.layout
}

func (s *Screen) projectCreation() bool {
	var exportPath string
	var languages []string

	exportPath = s.exportPath.Value()

	if len(exportPath) == 0 {
		s.statusMessage.SetError(errors.New(lokyn.L("Translation export path is mandatory")))
		return false
	}

	languages = strings.Split(
		strings.TrimSpace(s.languages.Value()), ",")

	if len(languages[0]) == 0 {
		s.statusMessage.SetError(errors.New(lokyn.L("Languages list is mandatory")))
		return false
	}

	s.project = engine.ProjectNew(exportPath, languages)

	err := engine.ProjectSave(&s.project)

	if err != nil {
		s.statusMessage.SetError(err)
		return false
	}

	return true
}
