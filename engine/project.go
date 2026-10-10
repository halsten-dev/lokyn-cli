package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func projectSave(p *Project, filePath string) error {
	content, err := json.Marshal(*p)

	if err != nil {
		return err
	}

	// Associations are committed before leaving reconciliation. Replace the
	// project atomically so a failed write cannot destroy the previous links.
	file, err := os.CreateTemp(filepath.Dir(filePath), ".lokyn-project-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	mode := os.FileMode(0644)
	if info, err := os.Stat(filePath); err == nil {
		mode = info.Mode().Perm()
	}
	if err := file.Chmod(mode); err != nil {
		return err
	}
	if _, err := file.Write(content); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), filePath)
}

func projectLoad(p *Project, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(content, p)
	if err != nil {
		return err
	}
	p.InitVariableLinks()

	return nil
}
