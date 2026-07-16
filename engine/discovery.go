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
func discoverDirectory(keys *DiscoveredKeys, vars *DiscoveredVars, directory string) error {
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
			err = discoverDirectory(keys, vars, path.Join(directory, e.Name()))

			if err != nil {
				return err
			}

			continue
		}

		// If it's not a source file.
		if filepath.Ext(e.Name()) != ".go" {
			continue
		}

		discoverSourceFile(keys, vars, path.Join(directory, e.Name()))
	}

	return nil
}

// discoverSourceFile is the function that read a source file and extract all found Lokyn keys.
func discoverSourceFile(keys *DiscoveredKeys, vars *DiscoveredVars, filePath string) {
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

	findCalls(keys, vars, content, alias)
}

// The optional `import\s+` prefix handles single-line imports (e.g. templ's
// generated files: `import "github.com/halsten-dev/lokyn"`); without it the
// `import` keyword is captured as the package alias, so calls in those files
// are searched as `import.L(` and silently missed.
var importPattern = regexp.MustCompile(`(?m)^\s*(?:import\s+)?(?:(\w+)\s+)?"github\.com/halsten-dev/lokyn"`)

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

func findCalls(keys *DiscoveredKeys, vars *DiscoveredVars, content []byte, prefix string) {
	var pCounter int
	var bKey strings.Builder

	pattern := regexp.MustCompile(prefix + `\.([LP])\(`)

	matches := pattern.FindAllSubmatch(content, -1)
	matchesIndex := pattern.FindAllIndex(content, -1)

	for i, match := range matches {
		pCounter = 1

		if len(match) < 2 {
			continue
		}

		callType := string(match[1]) // L or P

		index := matchesIndex[i][1]

		// To help extract the key and to avoid "in string" parenthesis
		isInString := false
		lastStringRune := rune(0)
		sCounter := 0

		bKey.Reset()

		for {
			r := rune(content[index])

			if lastStringRune != 0 {
				if r == lastStringRune && isInString {
					isInString = false
					lastStringRune = rune(0)
				}
			} else {
				if r == '"' || r == '`' {
					isInString = true
					lastStringRune = r
					sCounter++
				}
			}

			if r == ',' && !isInString && callType == "P" {
				break
			}

			if r == '(' && !isInString {
				pCounter++
			}

			if r == ')' && !isInString {
				pCounter--
			}

			if pCounter == 0 {
				break
			}

			bKey.WriteRune(r)

			index++
		}

		strKey := bKey.String()

		if (isSurroundedBy(strKey, `"`) || isSurroundedBy(strKey, "`")) && sCounter == 1 {

			key := DiscoveredKey{
				IsPlural: callType == "P",
			}

			key.Key = Key(strKey[1 : len(strKey)-1])
			key.Err = nil

			if !keys.ContainsKey(key.Key) {
				*keys = append(*keys, key)
			}

		} else {
			variable := DiscoveredVar{
				Var: Var(strKey),
				Err: nil,
			}

			if !vars.ContainsVar(variable.Var) {
				*vars = append(*vars, variable)
			}
		}
	}
}

func isSurroundedBy(value string, character string) bool {
	if strings.HasPrefix(value, character) && strings.HasSuffix(value, character) {
		return true
	}

	return false
}
