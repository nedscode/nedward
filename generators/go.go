package generators

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/nedscode/nedward/services"
)

// GoGenerator generates go services from main packages
type GoGenerator struct {
	generatorBase
	foundServices []*services.ServiceConfig
}

// Name returns 'go' to identify this generator
func (v *GoGenerator) Name() string {
	return "go"
}

// StartWalk lets a generator know that a directory walk has been started, with the
// given starting path
func (v *GoGenerator) StartWalk(path string) {
	v.generatorBase.StartWalk(path)
}

// VisitDir searches a directory for .go files, and will store a service if a main
// package is detected. Returns true in the first return value if a service was found.
func (v *GoGenerator) VisitDir(path string) (bool, error) {
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		fPath := filepath.Join(path, f.Name())
		if filepath.Ext(fPath) != ".go" {
			continue
		}

		input, err := ioutil.ReadFile(fPath)
		if err != nil {
			return false, errors.WithStack(err)
		}

		packageExpr := regexp.MustCompile(`package main\n`)
		if packageExpr.Match(input) {
			packageName := filepath.Base(path)
			packagePath, err := filepath.Rel(v.basePath, path)
			if err != nil {
				return false, errors.WithStack(err)
			}
			service, err := v.goService(packageName, packagePath)
			if err != nil {
				return false, errors.WithStack(err)
			}
			v.foundServices = append(v.foundServices, service)
			return true, nil
		}

	}

	return false, nil
}

// Services returns the services generated during the last walk
func (v *GoGenerator) Services() []*services.ServiceConfig {
	return v.foundServices
}

func (v *GoGenerator) goService(name, packagePath string) (*services.ServiceConfig, error) {
	service := &services.ServiceConfig{
		Name: name,
		Path: &packagePath,
		Env:  []string{},
		Commands: services.ServiceConfigCommands{
			Build:  "go install",
			Launch: name,
		},
	}

	return service, nil
}

func (v *GoGenerator) createWatch(service *services.ServiceConfig) (services.ServiceWatch, error) {
	return services.ServiceWatch{
		Service:       service,
		IncludedPaths: v.getImportList(service),
	}, nil
}

func (v *GoGenerator) getImportList(service *services.ServiceConfig) []string {
	if service.Path == nil {
		return nil
	}

	// Get a list of imports using 'go list'
	var imports = []string{}
	cmd := exec.Command("go", "list", "-f", "{{ join .Imports \":\" }}")
	cmd.Dir = *service.Path
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(errBuf.String())
		return []string{*service.Path}
	}
	imports = append(imports, strings.Split(out.String(), ":")...)

	// Verify the import paths exist
	var checkedImports = []string{*service.Path}
	for _, i := range imports {
		path := os.ExpandEnv(fmt.Sprintf("$GOPATH/src/%v", i))
		if _, err := os.Stat(path); err == nil {
			rel, err := filepath.Rel(v.basePath, path)
			if err != nil {
				// TODO: Handle this error more effectively
				fmt.Println(err)
				continue
			}
			checkedImports = append(checkedImports, rel)
		}
	}
	// Remove subpaths
	sort.Strings(checkedImports)
	var outImports []string
	for i, path := range checkedImports {
		include := true
		for j, earlier := range checkedImports {
			if i > j && strings.HasPrefix(path, earlier) {
				include = false
			}
		}
		if include {
			outImports = append(outImports, path)
		}
	}
	return outImports
}
