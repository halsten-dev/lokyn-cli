package engine

import (
	"errors"
	"os"
	"path"
)

func discoverProject(project *Project, directory string) error {
	var projectDirFound bool
	var projectFileFound bool

	projectDirFound = false
	projectFileFound = false

	elements, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	for _, e := range elements {
		if e.Name() == PROJECT_DIR_NAME {
			projectDirFound = true
			break
		}
	}

	if !projectDirFound {
		return errors.New(ERROR_NO_PROJECT_DIR_FOUND)
	}

	projectPath := path.Join(directory, PROJECT_DIR_NAME)

	elements, err = os.ReadDir(projectPath)

	for _, e := range elements {
		if e.Name() == PROJECT_FILE_NAME {
			err = projectLoad(project, path.Join(projectPath, PROJECT_FILE_NAME))

			if err != nil {
				return err
			}

			projectFileFound = true
			break
		}
	}

	if !projectFileFound {
		return errors.New(ERROR_NO_PROJECT_FOUND)
	}

	return nil
}
