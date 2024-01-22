package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
)

func FindAllTargets(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	targets := make([]string, 0)
	if ssmPath, err := getModulePath(ssmModulePath, ssmModuleVersion); err == nil {
		targets = append(targets, ssmPath)
	}
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

func getModulePath(name, version string) (string, error) {
	// first we need GOMODCACHE
	cache, ok := os.LookupEnv("GOMODCACHE")
	if !ok {
		cache = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	// then we need to escape path
	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", err
	}

	if version == "(devel)" {
		versions := make([]module.Version, 0)
		potentials, _ := filepath.Glob(filepath.Join(cache, ssmModulePath+"*"))
		for _, path := range potentials {
			d, err := os.Stat(path)
			if err != nil {
				continue
			}
			if !d.IsDir() || !strings.Contains(path, escapedPath) {
				continue
			}
			if pieces := strings.Split(path, "@"); len(pieces) == 2 {
				if strings.Count(pieces[1], "/") > 0 {
					continue
				}
				versions = append(versions, module.Version{Path: name, Version: pieces[1]})
			}
		}
		if len(versions) > 0 {
			module.Sort(versions)
			version = versions[len(versions)-1].Version
		}
	}
	// version also
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", err
	}

	return filepath.Join(cache, escapedPath+"@"+escapedVersion), nil
}
