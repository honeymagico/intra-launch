package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanupSyncArtifactsOnlyRemovesManagedDirectories(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{".intra-launch-sync-123", "tool.backup-456", "tool"} {
		if err := os.Mkdir(filepath.Join(base, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := cleanupSyncArtifacts(base); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(base, "tool")); err != nil {
		t.Fatal("regular application directory was removed")
	}
	for _, name := range []string{".intra-launch-sync-123", "tool.backup-456"} {
		if _, err := os.Stat(filepath.Join(base, name)); !os.IsNotExist(err) {
			t.Fatalf("managed directory %q was not removed", name)
		}
	}
}
