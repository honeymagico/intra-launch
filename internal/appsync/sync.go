// Package appsync mirrors one portable application directory into local storage.
package appsync

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
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
func Directory(source, destination string, app appcatalog.App, workers int, reportProgress func(Progress)) error {
	return DirectoryContext(context.Background(), source, destination, app, workers, reportProgress)
}

// DirectoryContext behaves like Directory and stops safely when ctx is cancelled.
func DirectoryContext(ctx context.Context, source, destination string, app appcatalog.App, workers int, reportProgress func(Progress)) error {
	if workers < 1 {
		return fmt.Errorf("同步 worker 數量必須大於 0")
	}
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
	if err := copyTree(ctx, source, staging, workers, reportProgress); err != nil {
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

type copyJob struct {
	source   string
	target   string
	relative string
}

func copyTree(ctx context.Context, source, destination string, workers int, reportProgress func(Progress)) error {
	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan copyJob, workers*2)
	errors := make(chan error, 1)
	var workerGroup sync.WaitGroup
	var progressMu sync.Mutex
	completedFiles := 0

	reportError := func(err error) {
		select {
		case errors <- err:
			cancel()
		default:
		}
	}

	workerGroup.Add(workers)
	for range workers {
		go func() {
			defer workerGroup.Done()
			for job := range jobs {
				if err := copyFile(workerContext, job.source, job.target); err != nil {
					reportError(err)
					return
				}
				progressMu.Lock()
				completedFiles++
				if reportProgress != nil {
					reportProgress(Progress{CompletedFiles: completedFiles, CurrentFile: job.relative})
				}
				progressMu.Unlock()
			}
		}()
	}

	walkErr := filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := workerContext.Err(); err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		select {
		case jobs <- copyJob{source: path, target: target, relative: relative}:
			return nil
		case <-workerContext.Done():
			return workerContext.Err()
		}
	})
	if walkErr != nil {
		reportError(walkErr)
	}
	close(jobs)
	workerGroup.Wait()

	select {
	case err := <-errors:
		return err
	default:
		return nil
	}
}

func copyFile(ctx context.Context, source, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(source)
	if err != nil {
		return err
	}
	input, err := os.Open(source)
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
	return closeErr
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
