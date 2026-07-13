package apppath

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalAppsBase returns the fixed per-user application directory.
func LocalAppsBase() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("取得使用者目錄失敗: %w", err)
	}
	return filepath.Join(home, "intra-launch"), nil
}

// EnsureLocalAppsBase creates the local application directory when absent.
func EnsureLocalAppsBase() (string, error) {
	base, err := LocalAppsBase()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return "", fmt.Errorf("建立本機應用目錄失敗 %q: %w", base, err)
	}
	return base, nil
}

// SMBAppDir returns the source directory for an application.
func SMBAppDir(base, id string) (string, error) {
	if err := validateAppID(id); err != nil {
		return "", err
	}
	return filepath.Join(base, id), nil
}

// LocalAppDir returns the managed local directory for an application.
func LocalAppDir(base, id string) (string, error) {
	if err := validateAppID(id); err != nil {
		return "", err
	}
	return filepath.Join(base, id), nil
}

func validateAppID(id string) error {
	if err := validateRelativePath("應用 ID", id); err != nil {
		return err
	}
	if filepath.Base(id) != id {
		return fmt.Errorf("應用 ID 必須是單一目錄名稱: %q", id)
	}
	return nil
}

// ExecutablePath returns the full local executable path for an application.
func ExecutablePath(base, id, executable string) (string, error) {
	appDir, err := LocalAppDir(base, id)
	if err != nil {
		return "", err
	}
	if err := validateRelativePath("執行檔路徑", executable); err != nil {
		return "", err
	}
	return filepath.Join(appDir, executable), nil
}

func validateRelativePath(name, value string) error {
	if value == "" || filepath.IsAbs(value) || filepath.Clean(value) == "." {
		return fmt.Errorf("%s無效: %q", name, value)
	}
	clean := filepath.Clean(value)
	if clean == ".." || len(clean) > 3 && clean[:3] == ".."+string(filepath.Separator) {
		return fmt.Errorf("%s不可超出應用目錄: %q", name, value)
	}
	return nil
}
