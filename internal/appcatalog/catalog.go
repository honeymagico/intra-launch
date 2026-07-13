// Package appcatalog reads and validates the application catalog published on SMB.
package appcatalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileName = "catalog.json"

// Catalog is the single application list stored at the SMB root.
type Catalog struct {
	Version string `json:"version"`
	Apps    []App  `json:"apps"`
}

// App describes one portable application in the catalog.
type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Executable  string `json:"executable"`
}

// Load reads catalog.json from the given SMB base directory.
func Load(smbBasePath string) (Catalog, error) {
	if strings.TrimSpace(smbBasePath) == "" {
		return Catalog{}, fmt.Errorf("SMB base path 不可為空")
	}
	return LoadFile(filepath.Join(smbBasePath, fileName))
}

// LoadFile reads and validates one catalog file.
func LoadFile(path string) (Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Catalog{}, fmt.Errorf("讀取 catalog.json 失敗 %q: %w", path, err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("catalog.json 格式錯誤 %q: %w", path, err)
	}
	if err := catalog.Validate(); err != nil {
		return Catalog{}, fmt.Errorf("catalog.json 內容錯誤 %q: %w", path, err)
	}
	return catalog, nil
}

// Validate checks fields required by the launcher UI and application operations.
func (c Catalog) Validate() error {
	if strings.TrimSpace(c.Version) == "" {
		return fmt.Errorf("缺少 catalog version")
	}
	for i, app := range c.Apps {
		if err := app.Validate(); err != nil {
			return fmt.Errorf("第 %d 個應用: %w", i+1, err)
		}
	}
	return nil
}

// Validate checks the fields required for one application.
func (a App) Validate() error {
	required := []struct {
		name  string
		value string
	}{
		{"id", a.ID},
		{"name", a.Name},
		{"version", a.Version},
		{"author", a.Author},
		{"description", a.Description},
		{"executable", a.Executable},
	}
	for _, field := range required {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("缺少必要欄位 %s", field.name)
		}
	}
	return nil
}
