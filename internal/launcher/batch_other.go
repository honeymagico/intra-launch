//go:build !windows

package launcher

import "os/exec"

func batchCommand(path string) *exec.Cmd {
	return exec.Command(path)
}
