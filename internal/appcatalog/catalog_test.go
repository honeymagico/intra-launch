package appcatalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validCatalog = `{
  "version": "1",
  "apps": [
    {"id":"report-system","name":"報表系統","version":"v1150709","author":"IT Team","description":"報表工具","executable":"ReportSystem.exe"},
    {"id":"purchase-manager","name":"採購管理系統","version":"v1150708","author":"IT Team","description":"採購工具","executable":"PurchaseManager.exe"},
    {"id":"ops-toolkit","name":"維運支援工具","version":"v1141201","author":"IT Support","description":"維運工具","executable":"bin/OpsToolkit.exe"}
  ]
}`

func TestLoadParsesThreeApps(t *testing.T) {
	dir := t.TempDir()
	writeCatalog(t, dir, validCatalog)

	catalog, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Version != "1" || len(catalog.Apps) != 3 {
		t.Fatalf("Load() = version %q, %d apps", catalog.Version, len(catalog.Apps))
	}
	if catalog.Apps[2].Executable != "bin/OpsToolkit.exe" {
		t.Fatalf("third executable = %q", catalog.Apps[2].Executable)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name    string
		content *string
		want    string
	}{
		{name: "missing file", want: "讀取 catalog.json 失敗"},
		{name: "invalid json", content: stringPointer(`{"version":`), want: "catalog.json 格式錯誤"},
		{name: "missing field", content: stringPointer(`{"version":"1","apps":[{"id":"app"}]}`), want: "缺少必要欄位 name"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if test.content != nil {
				writeCatalog(t, dir, *test.content)
			}
			_, err := Load(dir)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load() error = %v, want containing %q", err, test.want)
			}
		})
	}
}

func writeCatalog(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "catalog.json"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func stringPointer(value string) *string {
	return &value
}
