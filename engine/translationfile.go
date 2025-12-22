package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
)

func importTranslationFile(exportDir string, lang Lang) (map[Key]Translation, error) {
	var translations map[Key]Translation
	var translation Translation
	var oneValue string
	var otherValue string
	var importValue map[string]any
	var ok bool
	var importedTranslations map[string]any

	content, err := os.ReadFile(path.Join(exportDir,
		fmt.Sprintf("%s.json", lang)))

	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}

		return nil, nil
	}

	if len(content) == 0 {
		return nil, nil
	}

	err = json.Unmarshal(content, &importedTranslations)

	if err != nil {
		return nil, err
	}

	translations = make(map[Key]Translation)

	for k, v := range importedTranslations {
		translation = Translation{
			Key:  Key(k),
			Lang: lang,
		}

		importValue, ok = v.(map[string]any)

		if ok {
			translation.IsPlural = true

			oneValue, ok = importValue["One"].(string)

			if !ok {
				return nil, fmt.Errorf(ERROR_MALFORMED_PLURAL_JSON, k)
			}

			otherValue, ok = importValue["Other"].(string)

			if !ok {
				return nil, fmt.Errorf(ERROR_MALFORMED_PLURAL_JSON, k)
			}

			translation.OneValue = oneValue
			translation.OtherValue = otherValue
		} else {
			translation.IsPlural = false
			translation.OneValue = v.(string)
			translation.OtherValue = ""
		}

		translations[Key(k)] = translation
	}

	return translations, nil

}

func importAllTranslationFiles(project Project) (LangKeyMap, error) {
	var langKeyMap LangKeyMap

	langKeyMap = make(LangKeyMap)

	for _, lang := range project.ManagedLanguages {
		translations, err := importTranslationFile(path.Join(project.LocationPath, project.ExportDir), lang)

		if err != nil {
			return nil, err
		}

		langKeyMap[lang] = translations

	}

	return langKeyMap, nil

}
