package engine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

// This file only contains the public API of the engine package.

// DiscoverProject is the function that allows to explore the current project root,
// find the .lokyn folder if existing and load the project.
func DiscoverProject(path string) (Project, error) {
	var project Project

	err := discoverProject(&project, path)

	return project, err
}

// DiscoverKeys is the function that allows to explore the current project root,
// find every go files and analyse them to find all the Lokyn keys.
func DiscoverKeys(path string) (DiscoveredKeys, DiscoveredVars, error) {
	var keys DiscoveredKeys
	var vars DiscoveredVars

	err := discoverDirectory(&keys, &vars, path)

	return keys, vars, err
}

// ProjectNew creates a new project and returns it
func ProjectNew(path, exportPath string, languages []string) Project {
	langs := make([]Lang, len(languages))

	project := Project{}

	for i, l := range languages {
		langs[i] = Lang(l)
	}

	project.LocationPath = path
	project.ExportDir = exportPath
	project.ManagedLanguages = langs

	return project
}

func ProjectSave(project *Project) error {
	projectPath := path.Join(project.LocationPath, PROJECT_DIR_NAME)

	err := os.MkdirAll(projectPath, 0777)

	if err != nil {
		return err
	}

	err = projectSave(project, path.Join(projectPath, PROJECT_FILE_NAME))

	if err != nil {
		return err
	}

	return nil
}

func ImportTranslations(project *Project) (KeyLangMap, error) {
	langKeyMap, err := importAllTranslationFiles(*project)

	if err != nil {
		return nil, err
	}

	keyLangMap := ConvertLangKeyMap(langKeyMap)

	return keyLangMap, nil
}

func ExportAllTranslations(project *Project, data KeyLangMap) error {
	var err error
	var content bytes.Buffer
	var filePath string
	var count int

	exportData := ConvertKeyLangMap(data)

	for _, l := range project.ManagedLanguages {
		filePath = path.Join(project.LocationPath, project.ExportDir, fmt.Sprintf("%s.json", l))
		err = os.Remove(filePath)

		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		content.Reset()

		content.WriteString("{")

		count = 0

		for k, t := range exportData[l] {
			if t.OneValue == "" {
				continue
			}

			if count > 0 {
				content.WriteString(",")
			}

			// Marshal each key/value through encoding/json so quotes, backslashes,
			// newlines, etc. in a translation are escaped. Interpolating them raw
			// produced invalid JSON (e.g. a value containing " broke the file).
			// json.Marshal of a string never errors, so the errors are ignored.
			keyJSON, _ := json.Marshal(string(k))

			if t.IsPlural {
				oneJSON, _ := json.Marshal(t.OneValue)
				otherJSON, _ := json.Marshal(t.OtherValue)
				content.WriteString(fmt.Sprintf(`%s:{"One":%s,"Other":%s}`,
					keyJSON, oneJSON, otherJSON))
			} else {
				valueJSON, _ := json.Marshal(t.OneValue)
				content.WriteString(fmt.Sprintf(`%s:%s`, keyJSON, valueJSON))
			}

			count++
		}

		content.WriteString("}")

		err = os.WriteFile(filePath, content.Bytes(), 0777)

		if err != nil {
			return err
		}
	}

	return nil
}
