//go:build dev

package frontend

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

var DistFS fs.FS = loadDevDistFS()

func loadDevDistFS() fs.FS {
	candidates := devDistCandidates()
	for _, dir := range candidates {
		if info, err := os.Stat(filepath.Join(dir, "dist")); err == nil && info.IsDir() {
			return os.DirFS(dir)
		}
	}
	return fs.FS(nil)
}

func devDistCandidates() []string {
	var dirs []string

	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "internal", "frontend"))
	}

	if exePath, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exePath), "internal", "frontend"))
	}

	if _, file, _, ok := runtime.Caller(0); ok {
		dirs = append(dirs, filepath.Join(filepath.Dir(file)))
	}

	return dirs
}
