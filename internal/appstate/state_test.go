package appstate

import (
	"os"
	"path/filepath"
	"testing"

	"intra-launch/internal/appcatalog"
)

func TestScanLocalKeepsValidAppsAndWarnsAboutInvalidApps(t *testing.T) {
	base := t.TempDir()
	local := testApp("report", "v2")
	appDir := filepath.Join(base, local.ID)
	if err := os.Mkdir(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SaveMetadata(appDir, local); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(base, "damaged"), 0o755); err != nil {
		t.Fatal(err)
	}

	apps, warnings, err := ScanLocal(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 1 || apps[0].Version != "v2" || len(warnings) != 1 {
		t.Fatalf("ScanLocal() = %#v, warnings %#v", apps, warnings)
	}
}

func TestScanLocalIgnoresSyncAndBackupDirectories(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{".intra-launch-sync-123", "tool.backup-456"} {
		if err := os.Mkdir(filepath.Join(base, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	apps, warnings, err := ScanLocal(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps) != 0 || len(warnings) != 0 {
		t.Fatalf("ScanLocal() = %#v, warnings %#v", apps, warnings)
	}
}

func TestMergeSeparatesListsAndDetectsAnyVersionDifference(t *testing.T) {
	local := []appcatalog.App{testApp("report", "v2"), testApp("offline-only", "v1")}
	remote := appcatalog.Catalog{Version: "1", Apps: []appcatalog.App{
		testApp("report", "v1"), // A lower remote version is still an available update.
		testApp("download", "v1"),
	}}

	result := Merge(local, remote)
	if len(result.Installed) != 2 || len(result.Available) != 1 {
		t.Fatalf("Merge() = %#v", result)
	}
	if !result.Installed[0].UpdateAvailable || result.Installed[0].AvailableVersion != "v1" {
		t.Fatalf("installed report = %#v", result.Installed[0])
	}
	if result.Available[0].ID != "download" {
		t.Fatalf("available = %#v", result.Available)
	}
}

func TestMergeWithoutCatalogStillReturnsLocalApps(t *testing.T) {
	result := Merge([]appcatalog.App{testApp("offline", "v1")}, appcatalog.Catalog{})
	if len(result.Installed) != 1 || len(result.Available) != 0 {
		t.Fatalf("Merge() = %#v", result)
	}
}

func testApp(id, version string) appcatalog.App {
	return appcatalog.App{
		ID: id, Name: id, Version: version, Author: "team",
		Description: "description", Executable: id + ".exe",
	}
}
