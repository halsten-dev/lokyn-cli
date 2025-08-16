package main

import (
	"embed"
	"log"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/orvyn"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/halsten-dev/bubblehelp"
	"github.com/halsten-dev/lokyn"
	"github.com/spf13/viper"
)

//go:embed translations
var translations embed.FS

func main() {
	f, err := tea.LogToFile("debug.log", "debug")

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	lokyn.Init()
	err = lokyn.AddTranslationFS(translations, "translations")

	if err != nil {
		log.Fatal(err)
	}

	lokyn.SetLanguage(viper.GetString("language"))

	keybind.Init()

	bubblehelp.Init()

	// registerKeymapContexts()

	// Orvyn
	orvyn.Init()

	// orvyn.RegisterScreen(screen.IDLogin, login.New())
	// orvyn.RegisterScreen(screen.IDCharacterSelection, characterselection.New())
	// orvyn.SwitchScreen(screen.IDLogin)

	p := tea.NewProgram(&App{}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
