package internal

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/module"
)

func init() {
	if ssmPath, err := getModulePath(ssmModulePath, ssmModuleVersion); err == nil {
		SSMStates, _ = LoadStates(ssmPath)
	}
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

	if version == "(devel)" || version == "" {
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
