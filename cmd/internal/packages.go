package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FindAllTargets(args ...string) []string {
	if len(args) == 0 {
		return nil
	}
	targets := make([]string, 0)
	for _, arg := range args {
		if abs, err := filepath.Abs(filepath.Clean(arg)); err == nil {
			arg = abs
		}
		fi, err := os.Stat(arg)
		if err != nil {
			continue
		}
		if fi.IsDir() {
			_ = filepath.WalkDir(arg, func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					return nil
				}
				if strings.Contains(path, ".git") {
					return filepath.SkipDir
				}
				targets = append(targets, path)
				return nil
			})
		} else {
			targets = append(targets, arg)
		}
	}
	return targets
}

func validGoFile(info fs.FileInfo) bool {
	n := info.Name()
	return !strings.Contains(n, "_test") && filepath.Ext(n) == ".go"
}
