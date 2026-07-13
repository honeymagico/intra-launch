//go:build windows

package launcher

import (
	"os/exec"
	"strings"
	"syscall"
)

func batchCommand(path string) *exec.Cmd {
	command := exec.Command("cmd.exe")
	command.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: `cmd.exe /d /s /c start "" cmd.exe /d /c call "` + escapeCmdPath(path) + `"`,
	}
	return command
}

func escapeCmdPath(path string) string {
	return strings.ReplaceAll(path, `"`, `""`)
}
