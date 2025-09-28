package main

import (
	"embed"
	"log"
	"lokyn-cli/internal/config"
	"lokyn-cli/internal/keybind"
	"lokyn-cli/internal/translate"
	"lokyn-cli/screen"
	"lokyn-cli/screen/projectloading"
	"lokyn-cli/screen/reconciliation"
	"lokyn-cli/screen/translation"

	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/theme"

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

	// Orvyn
	orvyn.Init()

	t := orvyn.GetTheme()
	ns := t.Style(theme.NormalTextStyleID)
	ds := t.Style(theme.DimTextStyleID)
	nds := t.Style(theme.NeutralDimTextStyleID)

	// Use orvyn's theme to set bubblehelps style
	helpStyle := bubblehelp.Style{
		EssentialKey:               ns,
		EssentialKeyDescription:    ds,
		EssentialKeySeparator:      ds,
		EssentialKeySeparatorValue: " ",
		EssentialColSeparator:      nds,
		EssentialColSeparatorValue: " • ",
		FullKey:                    ns,
		FullKeyDescription:         ds,
		FullKeySeparator:           ds,
		FullKeySeparatorValue:      " ",
		FullColSeparator:           nds,
		FullColSeparatorValue:      " • ",
	}

	bubblehelp.SetDefaultStyle(helpStyle)

	registerKeymapContexts()

	// Screens registering
	orvyn.RegisterScreen(screen.IDProjectLoading, projectloading.New())
	orvyn.RegisterScreen(screen.IDTranslation, translation.New())
	orvyn.RegisterScreen(screen.IDReconciliation, reconciliation.New())

	orvyn.SwitchScreen(screen.IDProjectLoading)

	p := tea.NewProgram(&App{}, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func registerKeymapContexts() {
	keymapContext := bubblehelp.NewKeymap(2)
	keymapContext.NewKeyBinding(keybind.Enter, true)
	keymapContext.SetHelpDesc(keybind.Enter, lokyn.L("validate"))
	keymapContext.NewKeyBinding(keybind.Tab, true)
	keymapContext.NewKeyBinding(keybind.Up, false)
	keymapContext.NewKeyBinding(keybind.Down, false)
	keymapContext.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextReconciliation, keymapContext)

	keymapContext = bubblehelp.NewKeymap(2)
	keymapContext.NewKeyBinding(keybind.TKey, true)
	keymapContext.SetHelpDesc(keybind.TKey, lokyn.L("translate"))
	keymapContext.NewKeyBinding(keybind.KKey, true)
	keymapContext.SetHelpDesc(keybind.KKey, lokyn.L("copy key"))
	keymapContext.NewKeyBinding(keybind.EKey, true)
	keymapContext.SetHelpDesc(keybind.EKey, lokyn.L("edit value"))
	keymapContext.NewKeyBinding(keybind.XKey, true)
	keymapContext.SetHelpDesc(keybind.XKey, lokyn.L("export"))
	keymapContext.NewKeyBinding(keybind.Tab, true)
	keymapContext.NewKeyBinding(keybind.Up, true)
	keymapContext.NewKeyBinding(keybind.Down, true)
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextTranslation, keymapContext)

	keymapContext = bubblehelp.NewKeymap(2)
	keymapContext.NewKeyBinding(keybind.Esc, true)
	keymapContext.NewKeyBinding(keybind.Quit, true)

	bubblehelp.RegisterContext(keybind.ContextInputMode, keymapContext)
}
