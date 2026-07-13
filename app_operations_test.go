//go:build windows

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"intra-launch/internal/appcatalog"
	"intra-launch/internal/settings"
)

func TestMVPApplicationLifecycle(t *testing.T) {
	profile, configRoot, smb := t.TempDir(), t.TempDir(), t.TempDir()
	t.Setenv("USERPROFILE", profile)
	t.Setenv("APPDATA", configRoot)
	app := appcatalog.App{ID: "stage9-test-tool", Name: "Stage 9 Tool", Version: "v1", Author: "IT", Description: "lifecycle", Executable: "Stage9TestTool.exe"}
	writeTestCatalog(t, smb, app)
	sourceDirectory := filepath.Join(smb, app.ID)
	if err := os.MkdirAll(sourceDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDirectory, app.Executable), []byte("portable-v1"), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath, err := settings.DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	config, _ := json.Marshal(settings.Config{SMBBasePath: smb})
	if err := os.WriteFile(configPath, config, 0o600); err != nil {
		t.Fatal(err)
	}

	backend := NewApp()
	if err := backend.DownloadApplication(app.ID); err != nil {
		t.Fatal(err)
	}
	lists, err := backend.ApplicationLists()
	if err != nil || len(lists.Installed) != 1 || len(lists.Available) != 0 {
		t.Fatalf("after download: %#v, %v", lists, err)
	}

	localExecutable := filepath.Join(profile, "intra-launch", app.ID, app.Executable)
	if err := os.Remove(localExecutable); err != nil {
		t.Fatal(err)
	}
	if err := backend.RepairApplication(app.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(localExecutable); err != nil {
		t.Fatalf("repair did not restore executable: %v", err)
	}

	app.Version = "v2"
	writeTestCatalog(t, smb, app)
	lists, err = backend.ApplicationLists()
	if err != nil || !lists.Installed[0].UpdateAvailable {
		t.Fatalf("update not detected: %#v, %v", lists, err)
	}
	if err := backend.UpdateApplication(app.ID); err != nil {
		t.Fatal(err)
	}
	lists, _ = backend.ApplicationLists()
	if lists.Installed[0].UpdateAvailable {
		t.Fatal("update state remained after synchronization")
	}

	if err := backend.UninstallApplication(app.ID); err != nil {
		t.Fatal(err)
	}
	lists, _ = backend.ApplicationLists()
	if len(lists.Installed) != 0 || len(lists.Available) != 1 {
		t.Fatalf("after uninstall: %#v", lists)
	}
}

func writeTestCatalog(t *testing.T, root string, app appcatalog.App) {
	t.Helper()
	data, err := json.Marshal(appcatalog.Catalog{Version: "1", Apps: []appcatalog.App{app}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}
