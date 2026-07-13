package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"intra-launch/internal/appcatalog"
	"intra-launch/internal/apppath"
	"intra-launch/internal/appstate"
	"intra-launch/internal/appsync"
	"intra-launch/internal/launcher"
	"intra-launch/internal/settings"
	"intra-launch/internal/smbconn"
)

// App is the backend entry point exposed to the Wails frontend.
type App struct {
	ctx        context.Context
	startupErr error
	mu         sync.RWMutex
	lastSync   time.Time
	syncMu     sync.Mutex
	syncCancel context.CancelFunc
	syncDone   chan struct{}
}

// LaunchApplication starts an installed portable application.
func (a *App) LaunchApplication(id, executable string) error {
	basePath, err := apppath.EnsureLocalAppsBase()
	if err != nil {
		return err
	}
	return launcher.Start(basePath, id, executable)
}

// NewApp creates the application backend.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	basePath, err := apppath.EnsureLocalAppsBase()
	a.startupErr = err
	if a.startupErr == nil {
		a.startupErr = cleanupSyncArtifacts(basePath)
	}
	if a.startupErr == nil {
		if _, err := a.LoadSettings(); err != nil {
			log.Printf("載入啟動設定失敗: %v", err)
		}
	}
}

// Health gives the frontend a small callable endpoint for startup verification.
// Application discovery and SMB access are implemented in later MVP stages.
func (a *App) Health() string {
	if a.startupErr != nil {
		return "error: " + a.startupErr.Error()
	}
	return "ready"
}

// LoadSettings reads the current user's SMB connection settings.
func (a *App) LoadSettings() (settings.Config, error) {
	path, err := settings.DefaultPath()
	if err != nil {
		return settings.Config{}, err
	}
	return settings.LoadOrCreate(path)
}

// ConfigPath returns the settings file location shown by the UI.
func (a *App) ConfigPath() (string, error) {
	return settings.DefaultPath()
}

// LocalAppsBasePath ensures and returns the fixed local application directory.
func (a *App) LocalAppsBasePath() (string, error) {
	return apppath.EnsureLocalAppsBase()
}

// LoadCatalog reads catalog.json from the SMB root configured for this user.
func (a *App) LoadCatalog() (appcatalog.Catalog, error) {
	config, err := a.LoadSettings()
	if err != nil {
		return appcatalog.Catalog{}, err
	}
	return loadCatalog(config)
}

// ApplicationLists scans local applications first, then enriches them with SMB state.
// Catalog failures are returned as status while valid local applications remain usable.
func (a *App) ApplicationLists() (appstate.Lists, error) {
	basePath, err := apppath.EnsureLocalAppsBase()
	if err != nil {
		return appstate.Lists{}, err
	}
	local, warnings, err := appstate.ScanLocal(basePath)
	if err != nil {
		return appstate.Lists{}, err
	}

	catalog, catalogErr := a.LoadCatalog()
	result := appstate.Merge(local, catalog)
	result.Warnings = warnings
	if catalogErr != nil {
		result.CatalogError = catalogErr.Error()
	} else {
		a.mu.Lock()
		a.lastSync = time.Now()
		a.mu.Unlock()
	}
	return result, nil
}

// LastSync returns the last successful catalog read in RFC3339 format.
func (a *App) LastSync() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.lastSync.IsZero() {
		return ""
	}
	return a.lastSync.Format(time.RFC3339)
}

// DownloadApplication installs an application that exists in the current catalog.
func (a *App) DownloadApplication(id string) error { return a.syncApplication(id, false) }

// UpdateApplication replaces an installed application after checking it is not running.
func (a *App) UpdateApplication(id string) error { return a.syncApplication(id, true) }

// RepairApplication mirrors the catalog version again.
func (a *App) RepairApplication(id string) error { return a.syncApplication(id, true) }

func (a *App) syncApplication(id string, checkRunning bool) error {
	syncContext, finishSync, err := a.beginSync()
	if err != nil {
		return err
	}
	defer finishSync()
	config, err := a.LoadSettings()
	if err != nil {
		return err
	}
	catalog, err := loadCatalog(config)
	if err != nil {
		return err
	}
	if err := syncContext.Err(); err != nil {
		return err
	}
	app, found := catalogApp(catalog, id)
	if !found {
		return fmt.Errorf("catalog 中找不到應用 %q", id)
	}
	if checkRunning {
		running, err := launcher.Running(app.Executable)
		if err != nil {
			return err
		}
		if running {
			return fmt.Errorf("應用正在執行，請先關閉 %s", app.Name)
		}
	}
	base, err := apppath.EnsureLocalAppsBase()
	if err != nil {
		return err
	}
	source, err := apppath.SMBAppDir(config.SMBBasePath, id)
	if err != nil {
		return err
	}
	destination, err := apppath.LocalAppDir(base, id)
	if err != nil {
		return err
	}
	log.Printf("同步應用 %s（%d workers）: %s -> %s", id, config.SyncWorkers, source, destination)
	lastProgress := time.Time{}
	return appsync.DirectoryContext(syncContext, source, destination, app, config.SyncWorkers, func(progress appsync.Progress) {
		if a.ctx == nil || time.Since(lastProgress) < 250*time.Millisecond {
			return
		}
		lastProgress = time.Now()
		runtime.EventsEmit(a.ctx, "sync:progress", map[string]interface{}{
			"appId":          id,
			"completedFiles": progress.CompletedFiles,
		})
	})
}

// CancelSync requests cancellation of the active download, update, or repair.
func (a *App) CancelSync() bool {
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.syncCancel == nil {
		return false
	}
	a.syncCancel()
	return true
}

func (a *App) beginSync() (context.Context, func(), error) {
	a.syncMu.Lock()
	defer a.syncMu.Unlock()
	if a.syncCancel != nil {
		return nil, nil, fmt.Errorf("另一個同步作業正在進行")
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	a.syncCancel = cancel
	a.syncDone = done
	return ctx, func() {
		cancel()
		a.syncMu.Lock()
		a.syncCancel = nil
		a.syncDone = nil
		close(done)
		a.syncMu.Unlock()
	}, nil
}

func (a *App) beforeClose(ctx context.Context) bool {
	a.syncMu.Lock()
	cancel, done := a.syncCancel, a.syncDone
	a.syncMu.Unlock()
	if cancel == nil {
		return false
	}
	choice, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "同步作業尚未完成",
		Message:       "下載、更新或修復仍在進行。\n\n是否中止作業並關閉啟動器？",
		DefaultButton: "No",
	})
	if err != nil || choice != "Yes" {
		return true
	}
	cancel()
	<-done
	return false
}

func cleanupSyncArtifacts(basePath string) error {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("掃描同步殘留目錄失敗: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() || (!strings.HasPrefix(entry.Name(), ".intra-launch-sync-") && !strings.Contains(entry.Name(), ".backup-")) {
			continue
		}
		path := filepath.Join(basePath, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("清除同步殘留目錄 %q 失敗: %w", path, err)
		}
		log.Printf("已清除同步殘留目錄 %s", path)
	}
	return nil
}

func loadCatalog(config settings.Config) (appcatalog.Catalog, error) {
	connectionErr := smbconn.Connect(config.SMBBasePath, config.SMBUsername, config.SMBPassword)
	catalog, err := appcatalog.Load(config.SMBBasePath)
	if err == nil {
		return catalog, nil
	}
	if connectionErr != nil {
		return appcatalog.Catalog{}, fmt.Errorf("%w；使用設定帳密連線 SMB 亦失敗: %v", err, connectionErr)
	}
	return appcatalog.Catalog{}, err
}

// UninstallApplication removes only the managed directory for the selected app.
func (a *App) UninstallApplication(id string) error {
	base, err := apppath.EnsureLocalAppsBase()
	if err != nil {
		return err
	}
	directory, err := apppath.LocalAppDir(base, id)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(directory, ".intra-launch.json")); err != nil {
		return fmt.Errorf("無法解除安裝：找不到有效的本機應用狀態")
	}
	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("解除安裝失敗: %w", err)
	}
	log.Printf("已解除安裝應用 %s", id)
	return nil
}

func catalogApp(catalog appcatalog.Catalog, id string) (appcatalog.App, bool) {
	for _, app := range catalog.Apps {
		if app.ID == id {
			return app, true
		}
	}
	return appcatalog.App{}, false
}
