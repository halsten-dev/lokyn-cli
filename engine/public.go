package engine

import (
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
func ProjectNew(exportPath string, languages []Lang) Project {
	project := Project{}

	project.ExportDir = exportPath
	project.ManagedLanguages = languages

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
