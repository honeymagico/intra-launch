package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPathUsesUserProfile(t *testing.T) {
	profile := t.TempDir()
	t.Setenv("USERPROFILE", profile)

	got, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(profile, "intra-launch", "config.json")
	if got != want {
		t.Fatalf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"smbBasePath":"\\\\server\\apps","smbUsername":"user","smbPassword":"secret"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.SMBBasePath != `\\server\apps` || got.SMBUsername != "user" || got.SMBPassword != "secret" {
		t.Fatalf("Load() = %#v", got)
	}
}

func TestLoadRejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load() error = nil, want invalid JSON error")
	}
}

func TestLoadOrCreateCreatesDefaultConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")

	got, err := LoadOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != defaultConfig {
		t.Fatalf("LoadOrCreate() = %#v, want %#v", got, defaultConfig)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded != defaultConfig {
		t.Fatalf("saved config = %#v, want %#v", loaded, defaultConfig)
	}
}

func TestLoadOrCreatePreservesExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	want := Config{SMBBasePath: `\\server\custom`, SMBUsername: "custom", SMBPassword: "secret"}
	data := []byte(`{"smbBasePath":"\\\\server\\custom","smbUsername":"custom","smbPassword":"secret"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := LoadOrCreate(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("LoadOrCreate() = %#v, want %#v", got, want)
	}
}

func TestLoadOrCreateDoesNotReplaceInvalidConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadOrCreate(path); err == nil {
		t.Fatal("LoadOrCreate() error = nil, want invalid JSON error")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{" {
		t.Fatalf("invalid config was replaced: %q", data)
	}
}
