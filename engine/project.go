package engine

import (
	"encoding/json"
	"os"
)

func projectSave(p *Project, filePath string) error {
	content, err := json.Marshal(*p)

	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, content, 0777)

	if err != nil {
		return err
	}

	return nil
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

	return nil
}
