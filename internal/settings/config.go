package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName     = "config.json"
	DefaultSyncWorkers = 4
	MaximumSyncWorkers = 8
)

var defaultConfig = Config{
	SMBBasePath: `\\server\apps`,
	SMBUsername: "readonly-user",
	SMBPassword: "password",
	SyncWorkers: DefaultSyncWorkers,
}

// Config contains the connection settings for the internal SMB source.
type Config struct {
	SMBBasePath string `json:"smbBasePath"`
	SMBUsername string `json:"smbUsername"`
	SMBPassword string `json:"smbPassword"`
	SyncWorkers int    `json:"syncWorkers,omitempty"`
}

// DefaultPath returns the per-user settings file path.
func DefaultPath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("取得使用者 Profile 目錄失敗: %w", err)
	}
	return filepath.Join(dir, "intra-launch", configFileName), nil
}

// Load reads a JSON settings file from path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("讀取設定檔失敗 %q: %w", path, err)
	}

	// Start from defaults so settings created by older versions remain valid.
	config := defaultConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("設定檔格式錯誤 %q: %w", path, err)
	}
	if config.SyncWorkers < 1 || config.SyncWorkers > MaximumSyncWorkers {
		return Config{}, fmt.Errorf("設定檔格式錯誤 %q: syncWorkers 必須介於 1 到 %d", path, MaximumSyncWorkers)
	}
	return config, nil
}

// LoadOrCreate reads path, or creates it with the built-in defaults when absent.
func LoadOrCreate(path string) (Config, error) {
	config, err := Load(path)
	if err == nil {
		return config, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return Config{}, fmt.Errorf("建立設定目錄失敗 %q: %w", filepath.Dir(path), err)
	}
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return Config{}, fmt.Errorf("產生預設設定失敗: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return Config{}, fmt.Errorf("建立預設設定檔失敗 %q: %w", path, err)
	}
	return defaultConfig, nil
}
