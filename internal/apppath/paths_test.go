package apppath

import (
	"path/filepath"
	"testing"
)

func TestPathComposition(t *testing.T) {
	smbBase := filepath.Join("server", "apps")
	localBase := filepath.Join("users", "tester", "intra-launch")

	smb, err := SMBAppDir(smbBase, "report-system")
	if err != nil || smb != filepath.Join(smbBase, "report-system") {
		t.Fatalf("SMBAppDir() = %q, %v", smb, err)
	}
	local, err := LocalAppDir(localBase, "report-system")
	if err != nil || local != filepath.Join(localBase, "report-system") {
		t.Fatalf("LocalAppDir() = %q, %v", local, err)
	}
	executable, err := ExecutablePath(localBase, "report-system", filepath.Join("bin", "Report.exe"))
	if err != nil || executable != filepath.Join(localBase, "report-system", "bin", "Report.exe") {
		t.Fatalf("ExecutablePath() = %q, %v", executable, err)
	}
}

func TestPathCompositionRejectsTraversal(t *testing.T) {
	for _, test := range []struct {
		name string
		call func() error
	}{
		{"app id", func() error { _, err := LocalAppDir("base", `..\other`); return err }},
		{"nested app id", func() error { _, err := LocalAppDir("base", `group\app`); return err }},
		{"executable", func() error { _, err := ExecutablePath("base", "app", `..\other.exe`); return err }},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); err == nil {
				t.Fatal("error = nil, want traversal error")
			}
		})
	}
}
