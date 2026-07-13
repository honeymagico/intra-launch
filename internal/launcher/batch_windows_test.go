//go:build windows

package launcher

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBatchCommandStartsVisibleCommandWindow(t *testing.T) {
	path := filepath.Join(`C:\Apps`, "Start Portable.bat")
	command := batchCommand(path)
	if command.SysProcAttr == nil {
		t.Fatal("batch command SysProcAttr = nil")
	}
	for _, want := range []string{" start ", " cmd.exe ", " call ", `"` + path + `"`} {
		if !strings.Contains(command.SysProcAttr.CmdLine, want) {
			t.Fatalf("command line %q does not contain %q", command.SysProcAttr.CmdLine, want)
		}
	}
}
