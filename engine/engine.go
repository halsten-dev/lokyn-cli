// Package engine contains all the processing logic for the application.
package engine

const (
	PROJECT_DIR_NAME  string = ".lokyn"
	PROJECT_FILE_NAME string = "lokynproj.json"
)

type Keys []DiscoveredKey

type DiscoveredKey struct {
	// key holds the key found in source files of the project.
	Key string

	// isPlural holds the fact that the key was used as a plural key.
	IsPlural bool

	// err will hold the error message if the key was invalid for example.
	// Useful to report problems to the user.
	Err error
}

type Project struct {
	ExportDir        string
	ManagedLanguages []string
	TranslationMaps  []TranslationMap
}

type TranslationMap struct {
	Language string
	Keys     []TranslationKey
}

type TranslationKey struct {
	Key        string
	IsPlural   bool
	OneValue   string
	OtherValue string
}
