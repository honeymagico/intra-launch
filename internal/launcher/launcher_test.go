package launcher

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStartBuildsExecutablePathAndWorkingDirectory(t *testing.T) {
	base := t.TempDir()
	executable := filepath.Join("bin", "tool.exe")
	path := filepath.Join(base, "tool", executable)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotPath, gotDirectory string
	err := start(base, "tool", executable, func(path, directory string) error {
		gotPath, gotDirectory = path, directory
		return nil
	})
	if err != nil {
		t.Fatalf("start() error = %v", err)
	}
	if gotPath != path || gotDirectory != filepath.Dir(path) {
		t.Fatalf("starter got (%q, %q), want (%q, %q)", gotPath, gotDirectory, path, filepath.Dir(path))
	}
}

func TestStartRejectsMissingExecutable(t *testing.T) {
	err := start(t.TempDir(), "tool", "missing.exe", func(string, string) error {
		t.Fatal("starter should not be called")
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "找不到執行檔") {
		t.Fatalf("start() error = %v, want missing executable error", err)
	}
}

func TestLaunchCommandUsesCmdForBatchFile(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific command behavior")
	}
	path := filepath.Join(`C:\Apps`, "Start Portable.bat")
	command := launchCommand(path)
	if !strings.EqualFold(filepath.Base(command.Path), "cmd.exe") {
		t.Fatalf("command path = %q, want cmd.exe", command.Path)
	}
}

func TestLaunchCommandLeavesExecutableUnchanged(t *testing.T) {
	path := filepath.Join(`C:\Apps`, "Tool.exe")
	command := launchCommand(path)
	if len(command.Args) != 1 || command.Args[0] != path {
		t.Fatalf("command args = %#v, want [%q]", command.Args, path)
	}
}
