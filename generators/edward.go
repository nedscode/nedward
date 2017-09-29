package generators

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

// NedwardGenerator generates imports for all Nedward config files in the directory
// hierarchy.
type NedwardGenerator struct {
	generatorBase
	found []string
}

// Name returns 'nedward' to identify this generator.
func (v *NedwardGenerator) Name() string {
	return "nedward"
}

// VisitDir searches a directory for nedward.json files, and will store an import
// for any found. Returns true in the first return value if an import was found.
func (v *NedwardGenerator) VisitDir(path string) (bool, error) {
	if path == v.basePath {
		return false, nil
	}
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.Name() == "nedward.json" {
			relPath, err := filepath.Rel(v.basePath, filepath.Join(path, f.Name()))
			if err != nil {
				return false, errors.WithStack(err)
			}
			v.found = append(v.found, relPath)
			return true, SkipAll
		}
	}

	return false, nil
}

// Imports returns all imports found during previous walks.
func (v *NedwardGenerator) Imports() []string {
	return v.found
}
