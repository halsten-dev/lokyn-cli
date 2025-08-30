package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
)

// This file only contains the public API of the engine package.

// DiscoverProject is the function that allows to explore the current project root,
// find the .lokyn folder if existing and load the project.
func DiscoverProject() (Project, error) {
	var project Project

	err := discoverProject(&project, "./")

	return project, err
}

// DiscoverKeys is the function that allows to explore the current project root,
// find every go files and analyse them to find all the Lokyn keys.
func DiscoverKeys() (Keys, error) {
	var keys Keys

	err := discoverDirectory(&keys, "./")

	return keys, err
}

// ProjectNew creates a new project and returns it
func ProjectNew(exportPath string, languages []string) Project {
	langs := make([]Lang, len(languages))

	project := Project{}

	for i, l := range languages {
		langs[i] = Lang(l)
	}

	project.ExportDir = exportPath
	project.ManagedLanguages = langs

	return project
}

func ProjectSave(project *Project) error {
	projectPath := path.Join(".", PROJECT_DIR_NAME)

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
		filePath = path.Join(project.ExportDir, fmt.Sprintf("%s.json", l))
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

			if t.IsPlural {
				content.WriteString(fmt.Sprintf(`"%s":{"One":"%s", "Other":"%s"}`,
					k, t.OneValue, t.OtherValue))
			} else {
				content.WriteString(fmt.Sprintf(`"%s":"%s"`, k, t.OneValue))
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
