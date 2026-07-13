// Package appsync mirrors one portable application directory into local storage.
package appsync

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"intra-launch/internal/appcatalog"
	"intra-launch/internal/appstate"
)

// Progress describes completed file copies during one application sync.
type Progress struct {
	CompletedFiles int    `json:"completedFiles"`
	CurrentFile    string `json:"currentFile"`
}

// Directory copies source into a staging directory and replaces destination only
// after the application executable and metadata have been written successfully.
func Directory(source, destination string, app appcatalog.App, reportProgress func(Progress)) error {
	return DirectoryContext(context.Background(), source, destination, app, reportProgress)
}

// DirectoryContext behaves like Directory and stops safely when ctx is cancelled.
func DirectoryContext(ctx context.Context, source, destination string, app appcatalog.App, reportProgress func(Progress)) error {
	info, err := os.Stat(source)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("找不到內網應用目錄 %q", source)
	}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("建立本機目錄失敗: %w", err)
	}
	staging, err := os.MkdirTemp(parent, ".intra-launch-sync-*")
	if err != nil {
		return fmt.Errorf("建立同步暫存目錄失敗: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := copyTree(ctx, source, staging, reportProgress); err != nil {
		return fmt.Errorf("同步應用檔案失敗: %w", err)
	}
	executable := filepath.Join(staging, app.Executable)
	if info, err := os.Stat(executable); err != nil || info.IsDir() {
		return fmt.Errorf("同步完成但找不到執行檔 %q", app.Executable)
	}
	if err := appstate.SaveMetadata(staging, app); err != nil {
		return err
	}
	backup := destination + fmt.Sprintf(".backup-%d", time.Now().UnixNano())
	hadDestination := false
	if _, err := os.Stat(destination); err == nil {
		if err := os.Rename(destination, backup); err != nil {
			return fmt.Errorf("準備替換舊版本失敗: %w", err)
		}
		hadDestination = true
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("檢查本機應用目錄失敗: %w", err)
	}
	if err := os.Rename(staging, destination); err != nil {
		if hadDestination {
			_ = os.Rename(backup, destination)
		}
		return fmt.Errorf("套用同步結果失敗: %w", err)
	}
	if hadDestination {
		_ = os.RemoveAll(backup)
	}
	return nil
}

func copyTree(ctx context.Context, source, destination string, reportProgress func(Progress)) error {
	completedFiles := 0
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, &contextReader{ctx: ctx, reader: input})
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		completedFiles++
		if reportProgress != nil {
			reportProgress(Progress{CompletedFiles: completedFiles, CurrentFile: relative})
		}
		return nil
	})
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(buffer)
}
