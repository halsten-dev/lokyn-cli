package engine

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
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

// discoverDirectory is a recursive function that go in the whole hierarchy.
func discoverDirectory(keys *Keys, directory string) error {
	// Get all files of the folder.
	// If it's a folder > call this function again
	// -> Else, if it's a go file -> discoverSourceFile
	// -> -> Else do nothing

	elements, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	for _, e := range elements {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		if e.IsDir() {
			err = discoverDirectory(keys, path.Join(directory, e.Name()))

			if err != nil {
				return err
			}

			continue
		}

		// If it's not a source file.
		if filepath.Ext(e.Name()) != ".go" {
			continue
		}

		discoverSourceFile(keys, path.Join(directory, e.Name()))
	}

	return nil
}

// discoverSourceFile is the function that read a source file and extract all found Lokyn keys.
func discoverSourceFile(keys *Keys, filePath string) {
	// First, determine if the lokyn package uses an alias.
	// Read line by line and fetch : lokyn.L / lokyn.P
	// Get the key between double quotes, if there is no double quotes. Key are invalid.
	f, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	content, err := os.ReadFile(filePath)

	found, alias := findImport(content)

	// No Lokyn import, no need to analyse the source file
	if !found {
		return
	}

	if len(alias) == 0 {
		alias = "lokyn"
	}

	findCalls(keys, content, alias)
}

var importPattern = regexp.MustCompile(`(?m)^\s*(?:(\w+)\s+)?"github\.com/halsten-dev/lokyn"`)

func findImport(content []byte) (bool, string) {
	matches := importPattern.FindSubmatch(content)
	if matches == nil {
		return false, ""
	}

	if len(matches) > 1 && matches[1] != nil {
		return true, string(matches[1])
	}

	return true, ""
}

func findCalls(keys *Keys, content []byte, prefix string) {
	pattern := regexp.MustCompile(prefix + `\.(L|P)\(\s*([^,)]+)`)

	matches := pattern.FindAllSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		callType := string(match[1]) // L or P
		callKey := string(match[2])
		callKey = strings.TrimSpace(callKey)

		key := DiscoveredKey{
			Key:      Key(callKey),
			IsPlural: callType == "P",
		}

		if isSurroundedBy(callKey, `"`) || isSurroundedBy(callKey, "`") {
			key.Key = Key(callKey[1 : len(callKey)-1])
			key.Err = nil
		} else {
			key.Key = Key(callKey)
			key.Err = errors.New("Key is variable, need manual matching")
		}

		*keys = append(*keys, key)
	}
}

func isSurroundedBy(value string, character string) bool {
	if strings.HasPrefix(value, character) && strings.HasSuffix(value, character) {
		return true
	}

	return false
}
