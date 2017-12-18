package services

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/nedscode/nedward/home"
	"github.com/pkg/errors"
)

func LoadRunningServices() ([]ServiceOrGroup, error) {
	dir := home.NedwardConfig.StateDir
	stateFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var services []ServiceOrGroup
	for _, file := range stateFiles {
		// Skip directories (these contain instance state)
		if file.IsDir() {
			continue
		}

		command := &ServiceCommand{}
		raw, err := ioutil.ReadFile(path.Join(dir, file.Name()))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		json.Unmarshal(raw, command)
		command.Service.ConfigFile = command.ConfigFile

		// Check this service is actually running
		valid, err := command.validateState()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if valid {
			services = append(services, command.Service)
		}
	}
	return services, nil
}
