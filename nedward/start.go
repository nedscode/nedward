package nedward

import (
	"github.com/pkg/errors"
	"github.com/nedscode/nedward/config"
)

func (c *Client) Start(names []string, skipBuild bool, tail bool, noWatch bool, exclude []string) error {
	if len(names) == 0 {
		return errors.New("At least one service or group must be specified")
	}

	sgs, err := config.GetServicesOrGroups(names)
	if err != nil {
		return errors.WithStack(err)
	}
	if c.ServiceChecks != nil {
		err = c.ServiceChecks(sgs)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = c.startAndTrack(sgs, skipBuild, tail, noWatch, exclude, c.NedwardExecutable)
	if err != nil {
		return errors.WithStack(err)
	}
	if tail {
		return errors.WithStack(c.tailFromFlag(names))
	}

	return nil
}
