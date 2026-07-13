// Package appstate scans locally installed applications and merges them with a catalog.
package appstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"intra-launch/internal/appcatalog"
)

const metadataFileName = ".intra-launch.json"

// InstalledApp is a local application with its catalog comparison state.
type InstalledApp struct {
	Application      appcatalog.App `json:"application"`
	UpdateAvailable  bool           `json:"updateAvailable"`
	AvailableVersion string         `json:"availableVersion,omitempty"`
}

// Lists contains the two application sections required by the UI.
type Lists struct {
	Installed    []InstalledApp   `json:"installed"`
	Available    []appcatalog.App `json:"available"`
	Warnings     []string         `json:"warnings,omitempty"`
	CatalogError string           `json:"catalogError,omitempty"`
}

// ScanLocal reads metadata from each immediate application directory.
// A damaged application is reported as a warning without hiding other applications.
func ScanLocal(basePath string) ([]appcatalog.App, []string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, nil, fmt.Errorf("掃描本機應用目錄失敗 %q: %w", basePath, err)
	}

	apps := make([]appcatalog.App, 0, len(entries))
	var warnings []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".intra-launch-sync-") || strings.Contains(entry.Name(), ".backup-") {
			continue
		}
		path := filepath.Join(basePath, entry.Name(), metadataFileName)
		app, err := loadMetadata(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("略過本機應用 %q: %v", entry.Name(), err))
			continue
		}
		if app.ID != entry.Name() {
			warnings = append(warnings, fmt.Sprintf("略過本機應用 %q: 狀態檔 ID 為 %q", entry.Name(), app.ID))
			continue
		}
		apps = append(apps, app)
	}
	return apps, warnings, nil
}

// SaveMetadata records the catalog data needed for later offline use.
func SaveMetadata(appDirectory string, app appcatalog.App) error {
	if err := app.Validate(); err != nil {
		return fmt.Errorf("本機應用狀態無效: %w", err)
	}
	data, err := json.MarshalIndent(app, "", "  ")
	if err != nil {
		return fmt.Errorf("建立本機應用狀態失敗: %w", err)
	}
	path := filepath.Join(appDirectory, metadataFileName)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("寫入本機應用狀態失敗 %q: %w", path, err)
	}
	return nil
}

// Merge separates installed and downloadable applications and compares versions.
func Merge(local []appcatalog.App, remote appcatalog.Catalog) Lists {
	remoteByID := make(map[string]appcatalog.App, len(remote.Apps))
	for _, app := range remote.Apps {
		remoteByID[app.ID] = app
	}

	result := Lists{
		Installed: make([]InstalledApp, 0, len(local)),
		Available: make([]appcatalog.App, 0, len(remote.Apps)),
	}
	installedIDs := make(map[string]struct{}, len(local))
	for _, app := range local {
		installedIDs[app.ID] = struct{}{}
		installed := InstalledApp{Application: app}
		if available, ok := remoteByID[app.ID]; ok {
			installed.AvailableVersion = available.Version
			installed.UpdateAvailable = app.Version != available.Version
		}
		result.Installed = append(result.Installed, installed)
	}

	for _, app := range remote.Apps {
		if _, installed := installedIDs[app.ID]; !installed {
			result.Available = append(result.Available, app)
		}
	}
	return result
}

func loadMetadata(path string) (appcatalog.App, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return appcatalog.App{}, fmt.Errorf("讀取狀態檔失敗: %w", err)
	}
	var app appcatalog.App
	if err := json.Unmarshal(data, &app); err != nil {
		return appcatalog.App{}, fmt.Errorf("狀態檔格式錯誤: %w", err)
	}
	if err := app.Validate(); err != nil {
		return appcatalog.App{}, fmt.Errorf("狀態檔內容錯誤: %w", err)
	}
	return app, nil
}
