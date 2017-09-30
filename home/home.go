package home

import (
	"os"
	"os/user"
	"path"

	"github.com/pkg/errors"
)

// NedwardConfiguration defines the application config for Nedward
type NedwardConfiguration struct {
	Dir          string
	NedwardLogDir string
	LogDir       string
	PidDir       string
	StateDir     string
	ScriptDir    string
}

// NedwardConfig stores a shared instance of NedwardConfiguration for use across the app
var NedwardConfig = NedwardConfiguration{}

func createDirIfNeeded(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0777)
	}
}

// Initialize sets up NedwardConfig based on the location of .nedward in the home dir
func (e *NedwardConfiguration) Initialize() error {
	user, err := user.Current()
	if err != nil {
		return errors.WithStack(err)
	}
	e.Dir = path.Join(user.HomeDir, ".nedward")
	createDirIfNeeded(e.Dir)
	e.NedwardLogDir = path.Join(e.Dir, "nedward_logs")
	createDirIfNeeded(e.NedwardLogDir)
	e.LogDir = path.Join(e.Dir, "logs")
	createDirIfNeeded(e.LogDir)
	e.PidDir = path.Join(e.Dir, "pidFiles")
	createDirIfNeeded(e.PidDir)
	e.StateDir = path.Join(e.Dir, "stateFiles")
	createDirIfNeeded(e.StateDir)
	e.ScriptDir = path.Join(e.Dir, "scriptFiles")
	createDirIfNeeded(e.ScriptDir)
	return nil
}
