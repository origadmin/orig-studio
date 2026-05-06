//go:build dev

package web

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
		dirs = append(dirs, filepath.Join(cwd, "web"))
	}

	if exePath, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exePath), "web"))
	}

	if _, file, _, ok := runtime.Caller(0); ok {
		dirs = append(dirs, filepath.Dir(file))
	}

	return dirs
}
