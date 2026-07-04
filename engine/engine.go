// Package engine contains all the processing logic for the application.
package engine

const (
	PROJECT_DIR_NAME  string = ".lokyn"
	PROJECT_FILE_NAME string = "lokynproj.json"
)

type Key string
type Keys []Key
type Var string
type Lang string

type DiscoveredKeys []DiscoveredKey

func (k DiscoveredKeys) ContainsKey(key Key) bool {
	for _, v := range k {
		if v.Key == key {
			return true
		}
	}

	return false
}

func (k DiscoveredKeys) KeyIndex(key Key) int {
	for i, v := range k {
		if v.Key == key {
			return i
		}
	}

	return -1
}

type DiscoveredVars []DiscoveredVar

func (v DiscoveredVars) ContainsVar(varName Var) bool {
	for _, v := range v {
		if v.Var == varName {
			return true
		}
	}

	return false
}

func (v DiscoveredVars) VarIndex(varName Var) int {
	for i, v := range v {
		if v.Var == varName {
			return i
		}
	}

	return -1
}

type DiscoveredKey struct {
	// Key holds the key found in source files of the project.
	Key Key

	// Err will hold the error message if the key was invalid for example.
	// Useful to report problems to the user.
	Err error

	// IsPlural holds the fact that the key was used as a plural key.
	IsPlural bool

	// IsCreatedByUser holds the fact that this key was created by the user in the reconciliation screen.
	IsCreatedByUser bool
}

type DiscoveredVar struct {
	// Var represent the variable name
	Var Var

	// Err represents the current status of this variable
	Err error
}

type Project struct {
	LocationPath string
	ExportDir    string
	// MainLanguage     Lang
	ManagedLanguages []Lang

	// VarsKeysLink holds every know variables and their linked keys.
	VarsKeysLink map[Var]Keys
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
