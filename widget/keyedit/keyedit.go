package keyedit

import (
	"lokyn-cli/internal/orvyn"
	"lokyn-cli/internal/orvyn/widget/list"
)

type Widget struct {
	orvyn.BaseWidget
	orvyn.BaseFocusable

	translationFieldsList *list.Widget[]
}
