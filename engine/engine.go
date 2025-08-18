// Package engine contains all the processing logic for the application.
package engine

const (
	PROJECT_DIR_NAME  string = ".lokyn"
	PROJECT_FILE_NAME string = "lokynproj.json"
)

type Keys []Key

type Key struct {
	// key holds the key found in source files of the project.
	key string

	// isPlural holds the fact that the key was used as a plural key.
	isPlural bool

	// err will hold the error message if the key was invalid for example.
	// Useful to report problems to the user.
	err error
}

type Project struct {
	exportDir        string
	managedLanguages []string
	translationMaps  []TranslationMap
}

type TranslationMap struct {
	language string
	keys     []TranslationKey
}

type TranslationKey struct {
	key        string
	isPlural   bool
	oneValue   string
	otherValue string
}
