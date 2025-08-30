// Package engine contains all the processing logic for the application.
package engine

const (
	PROJECT_DIR_NAME  string = ".lokyn"
	PROJECT_FILE_NAME string = "lokynproj.json"
)

type Key string
type Lang string

type Keys []DiscoveredKey

type DiscoveredKey struct {
	// Key holds the key found in source files of the project.
	Key Key

	// Err will hold the error message if the key was invalid for example.
	// Useful to report problems to the user.
	Err error

	// IsPlural holds the fact that the key was used as a plural key.
	IsPlural bool
}

type Project struct {
	ExportDir string
	// MainLanguage     Lang
	ManagedLanguages []Lang
}

type LangKeyMap map[Lang]map[Key]Translation

type KeyLangMap map[Key]map[Lang]Translation

type Translation struct {
	Key        Key
	Lang       Lang
	OneValue   string
	OtherValue string
	IsPlural   bool
}

type TranslationData struct {
	Project Project
	Data    KeyLangMap
}
