package launcher

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Running reports whether Windows currently lists the executable image.
func Running(executable string) (bool, error) {
	name := filepath.Base(executable)
	command := exec.Command("tasklist", "/FI", "IMAGENAME eq "+name, "/FO", "CSV", "/NH")
	output, err := command.Output()
	if err != nil {
		return false, fmt.Errorf("檢查應用執行狀態失敗: %w", err)
	}
	return bytes.Contains(bytes.ToLower(output), []byte(strings.ToLower(`"`+name+`"`))), nil
}
