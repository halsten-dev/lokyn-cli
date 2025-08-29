package main

import (
	"embed"
	"log"
	"lokyn-cli/internal/config"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/translate"
	"lokyn-cli/screen"
	"lokyn-cli/screen/projectloading"
	"lokyn-cli/screen/reconsiliation"
	"lokyn-cli/screen/translation"

	"github.com/halsten-dev/orvyn"

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

	config.Init()

	translate.Init(viper.GetString(config.DEEPL_API_KEY))

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

	orvyn.RegisterScreen(screen.IDProjectLoading, projectloading.New())
	orvyn.RegisterScreen(screen.IDTranslation, translation.New())
	orvyn.RegisterScreen(screen.IDReconsiliation, reconsiliation.New())

	orvyn.SwitchScreen(screen.IDProjectLoading)

	p := tea.NewProgram(&App{}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
