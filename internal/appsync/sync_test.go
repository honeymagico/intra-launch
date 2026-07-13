package appsync

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"intra-launch/internal/appcatalog"
)

func TestDirectoryContextRemovesStagingAfterCancellation(t *testing.T) {
	source, local := t.TempDir(), t.TempDir()
	app := appcatalog.App{ID: "tool", Name: "Tool", Version: "v1", Author: "IT", Description: "Test", Executable: "tool.exe"}
	if err := os.WriteFile(filepath.Join(source, app.Executable), []byte("app"), 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := DirectoryContext(ctx, source, filepath.Join(local, app.ID), app, nil); err == nil {
		t.Fatal("expected cancellation error")
	}
	entries, err := os.ReadDir(local)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("staging directory was not removed: %#v", entries)
	}
}

func TestDirectoryCopiesAndMirrorsApplication(t *testing.T) {
	source, local := t.TempDir(), t.TempDir()
	app := appcatalog.App{ID: "tool", Name: "Tool", Version: "v2", Author: "IT", Description: "Test", Executable: "bin/tool.exe"}
	if err := os.MkdirAll(filepath.Join(source, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "bin", "tool.exe"), []byte("v2"), 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(local, app.ID)
	if err := os.MkdirAll(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "obsolete"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := Directory(source, destination, app, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(destination, "obsolete")); !os.IsNotExist(err) {
		t.Fatalf("obsolete file was not removed")
	}
	if _, err := os.Stat(filepath.Join(destination, ".intra-launch.json")); err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryKeepsExistingApplicationWhenExecutableIsMissing(t *testing.T) {
	source, local := t.TempDir(), t.TempDir()
	destination := filepath.Join(local, "tool")
	if err := os.Mkdir(destination, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destination, "keep"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	app := appcatalog.App{ID: "tool", Name: "Tool", Version: "v2", Author: "IT", Description: "Test", Executable: "missing.exe"}
	if err := Directory(source, destination, app, nil); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(filepath.Join(destination, "keep")); err != nil {
		t.Fatal("existing application was changed")
	}
}

func TestDirectoryReportsCompletedFiles(t *testing.T) {
	source, local := t.TempDir(), t.TempDir()
	app := appcatalog.App{ID: "tool", Name: "Tool", Version: "v1", Author: "IT", Description: "Test", Executable: "tool.exe"}
	if err := os.WriteFile(filepath.Join(source, "tool.exe"), []byte("app"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "data.txt"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	var progress []Progress
	err := Directory(source, filepath.Join(local, app.ID), app, func(update Progress) {
		progress = append(progress, update)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(progress) != 2 || progress[len(progress)-1].CompletedFiles != 2 {
		t.Fatalf("progress = %#v, want 2 completed files", progress)
	}
}
