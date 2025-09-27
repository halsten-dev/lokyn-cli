package dialog

import (
	"lokyn-cli/internal/keybind"

	"github.com/halsten-dev/lokyn"
	"github.com/halsten-dev/orvyn"
	"github.com/halsten-dev/orvyn/dialog"
)

func YesNoPopup(message string) orvyn.Screen {
	options := []dialog.Option{
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

	config := dialog.Config{
		Message: message,
		Options: options,
	}

	return dialog.NewPopup(config)
}
