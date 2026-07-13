// Package launcher starts locally installed portable applications.
package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"intra-launch/internal/apppath"
)

// Start validates and starts an application without waiting for it to exit.
func Start(basePath, id, executable string) error {
	return start(basePath, id, executable, startProcess)
}

func start(basePath, id, executable string, starter func(string, string) error) error {
	path, err := apppath.ExecutablePath(basePath, id, executable)
	if err != nil {
		return err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("找不到執行檔 %q", path)
		}
		return fmt.Errorf("檢查執行檔失敗 %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("執行檔路徑是目錄 %q", path)
	}
	if err := starter(path, filepath.Dir(path)); err != nil {
		return fmt.Errorf("啟動應用失敗 %q: %w", path, err)
	}
	return nil
}

func startProcess(path, workingDirectory string) error {
	command := launchCommand(path)
	command.Dir = workingDirectory
	return command.Start()
}

func launchCommand(path string) *exec.Cmd {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".bat", ".cmd":
		return batchCommand(path)
	default:
		return exec.Command(path)
	}
}
