package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/logrusorgru/aurora"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
)

func mustReplaceDataDir(dataDir, in string) string {
	d, err := filepath.Abs(dataDir)
	if err != nil {
		panic(fmt.Errorf("file path abs: %w", err))
	}

	in = strings.Replace(in, "{fh-data-dir}", d, -1)
	return in
}

func mkdirStorePathIfLocal(storeURL string) (err error) {
	userLog.Debug("creating directory and its parent(s)", zap.String("directory", storeURL))
	if dirs := getDirsToMake(storeURL); len(dirs) > 0 {
		err = makeDirs(dirs)
	}
	return
}

func getDirsToMake(storeURL string) []string {
	parts := strings.Split(storeURL, "://")
	if len(parts) > 1 {
		if parts[0] != "file" {
			// Not a local store, nothing to do
			return nil
		}
		storeURL = parts[1]
	}

	// Some of the store URL are actually a file directly, let's try our best to cope for that case
	filename := filepath.Base(storeURL)
	if strings.Contains(filename, ".") {
		storeURL = filepath.Dir(storeURL)
	}

	// If we reach here, it's a local store path
	return []string{storeURL}
}

func makeDirs(directories []string) error {
	for _, directory := range directories {
		err := os.MkdirAll(directory, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %q: %w", directory, err)
		}
	}

	return nil
}

func copyFile(ctx context.Context, in, out string) error {
	reader, _, _, err := dstore.OpenObject(ctx, in)
	if err != nil {
		return fmt.Errorf("unable : %w", err)
	}

	writer, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	if _, err := io.Copy(writer, reader); err != nil {
		_ = os.Remove(out)
		return fmt.Errorf("copy content: %w", err)
	}

	return nil
}

func cliErrorAndExit(message string) {
	fmt.Println(aurora.Red(message).String())
	os.Exit(1)
}

func dedentf(format string, args ...interface{}) string {
	return fmt.Sprintf(dedent.Dedent(strings.TrimPrefix(format, "\n")), args...)
}

func expandDir(dir string) (string, error) {
	if !strings.HasPrefix(dir, "~") {
		return filepath.Abs(dir)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return strings.Replace(dir, "~", homeDir, 1), nil
}

func dirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
