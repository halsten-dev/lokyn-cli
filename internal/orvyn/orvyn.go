// Package orvyn is a layer on top of BubbleTea to help building complex tui applications.
package orvyn

import (
	tea "github.com/charmbracelet/bubbletea"
	"log"
)

var (
	// WindowSize hold the size of the Window.
	WindowSize Size

	// screens is the map holding all Screen that are registered in orvyn.
	screens map[ScreenID]Screen

	// currentScreenID holds the active ScreenID.
	currentScreenID ScreenID

	// previousScreenID holds the previously active ScreenID.
	previousScreenID ScreenID

	activeDialog *Dialog
)

func Init() {
	WindowSize = NewSize(0, 0)
	screens = make(map[ScreenID]Screen)
}

func Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		WindowSize.Width = msg.Width
		WindowSize.Height = msg.Height
	}

	if currentScreenID == "" {
		return nil
	}

	// TODO: If a dialog is active, update it and not the current screen.
	if activeDialog != nil {
		return activeDialog.screen.Update(msg)
	} else {
		return screens[currentScreenID].Update(msg)
	}

}

func Render() string {
	var layout Layout

	if currentScreenID == "" {
		return "Orvyn : No Current Screen"
	}

	// TODO: If a dialog is active, render it and not the current screen.
	if activeDialog != nil {
		layout = activeDialog.screen.Render()
	} else {
		layout = screens[currentScreenID].Render()
	}

	layout.Resize(WindowSize)
	return layout.Render()
}

// Screen management

// RegisterScreen allows to register a Screen with the given ScreenID.
func RegisterScreen(id ScreenID, screen Screen) {
	screens[id] = screen
}

// SwitchScreen change the currently active screen and called OnExit and OnEnter.
func SwitchScreen(id ScreenID) tea.Cmd {
	var param interface{}

	_, ok := screens[id]

	if !ok {
		log.Fatalf("Orvyn : Screen with ID %s does not exist", id)
		return nil
	}

	if currentScreenID != "" {
		param = screens[currentScreenID].OnExit()
	}

	previousScreenID = currentScreenID

	currentScreenID = id

	return screens[currentScreenID].OnEnter(param)
}

func SwitchToPreviousScreen() tea.Cmd {
	if previousScreenID == "" {
		return nil
	}

	return SwitchScreen(previousScreenID)
}

func SetPreviousScreen(id ScreenID) {
	previousScreenID = id
}

// GetScreen returns the Screen for the given registered ScreenID.
func GetScreen(id ScreenID) Screen {
	_, ok := screens[id]

	if !ok {
		return nil
	}

	return screens[id]
}

// GetCurrentScreenID returns the currently active ScreenID.
func GetCurrentScreenID() ScreenID {
	return currentScreenID
}

// Dialog API

func OpenDialog(dialogID ScreenID, dialog Screen, param interface{}) {
	activeDialog = new(Dialog)

	activeDialog.dialogID = dialogID
	activeDialog.screen = dialog

	activeDialog.screen.OnEnter(param)
}

func CloseDialog() tea.Cmd {
	param := activeDialog.screen.OnExit()
	id := activeDialog.dialogID

	activeDialog = nil

	return DialogExitCmd(id, param)
}
