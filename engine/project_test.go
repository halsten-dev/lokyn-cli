package engine

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverProjectUsesSelectedDirectory(t *testing.T) {
	oldRoot := t.TempDir()
	for _, savedRoot := range []string{oldRoot, filepath.Join(oldRoot, "missing"), ""} {
		t.Run(savedRoot, func(t *testing.T) {
			root := t.TempDir()
			original := ProjectNew(root, "translations", []string{"en", "fr"})
			if err := ProjectSave(&original); err != nil {
				t.Fatal(err)
			}
			original.LocationPath = savedRoot
			configPath := filepath.Join(root, PROJECT_DIR_NAME, PROJECT_FILE_NAME)
			if err := projectSave(&original, configPath); err != nil {
				t.Fatal(err)
			}

			project, err := DiscoverProject(root)
			if err != nil {
				t.Fatal(err)
			}
			if project.LocationPath != root {
				t.Fatalf("LocationPath = %q, want %q", project.LocationPath, root)
			}
			if project.ExportDir != original.ExportDir || !reflect.DeepEqual(project.ManagedLanguages, original.ManagedLanguages) {
				t.Fatalf("project settings changed: %+v", project)
			}

			var saved Project
			if err := projectLoad(&saved, configPath); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(saved, project) {
				t.Fatalf("saved project = %+v, want %+v", saved, project)
			}
		})
	}
}

func TestDiscoverMovedProjectFromCurrentDirectory(t *testing.T) {
	parent := t.TempDir()
	oldRoot := filepath.Join(parent, "old")
	newRoot := filepath.Join(parent, "moved")
	project := ProjectNew(oldRoot, "translations", []string{"en"})
	if err := ProjectSave(&project); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(oldRoot, "translations"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldRoot, "translations", "en.json"), []byte(`{"Hello":"Hello there"}`), 0644); err != nil {
		t.Fatal(err)
	}
	source := "package example\nimport (\n\"github.com/halsten-dev/lokyn\"\n)\nvar greeting = lokyn.L(\"Hello\")\n"
	if err := os.WriteFile(filepath.Join(oldRoot, "example.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(oldRoot, newRoot); err != nil {
		t.Fatal(err)
	}
	t.Chdir(newRoot)

	project, err := DiscoverProject(".")
	if err != nil {
		t.Fatal(err)
	}
	if project.LocationPath != newRoot {
		t.Fatalf("LocationPath = %q, want %q", project.LocationPath, newRoot)
	}
	keys, err := DiscoverKeys(project.LocationPath)
	if err != nil {
		t.Fatal(err)
	}
	if !keys.ContainsKey("Hello") {
		t.Fatal("key discovery did not use the moved project")
	}
	data, err := ImportTranslations(&project)
	if err != nil {
		t.Fatal(err)
	}
	if got := data["Hello"]["en"].OneValue; got != "Hello there" {
		t.Fatalf("imported translation = %q, want %q", got, "Hello there")
	}
	// Export using the persisted project, not just the in-memory correction.
	if err := projectLoad(&project, filepath.Join(newRoot, PROJECT_DIR_NAME, PROJECT_FILE_NAME)); err != nil {
		t.Fatal(err)
	}
	if err := ExportAllTranslations(&project, data); err != nil {
		t.Fatalf("export from moved project: %v", err)
	}
}
