package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/nedscode/nedward/home"
)

func GetConfigPathFromWorkingDirectory() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return GetConfigPath(wd), nil
}

// GetConfigPath identifies the location of nedward.json, if any exists
func GetConfigPath(wd string) string {
	var pathOptions []string

	// Config file in Nedward Config dir
	pathOptions = append(pathOptions, filepath.Join(home.NedwardConfig.Dir, "nedward.json"))

	// Config file in current working directory
	pathOptions = append(pathOptions, filepath.Join(wd, "nedward.json"))
	for path.Dir(wd) != wd {
		wd = path.Dir(wd)
		pathOptions = append(pathOptions, filepath.Join(wd, "nedward.json"))
	}

	for _, path := range pathOptions {
		_, err := os.Stat(path)
		if err != nil {
			continue
		}
		absfp, absErr := filepath.Abs(path)
		if absErr != nil {
			fmt.Println("Error getting config file: ", absErr)
			return ""
		}
		return absfp
	}

	return ""
}
