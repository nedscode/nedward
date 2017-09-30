package nedward

import (
	"github.com/pkg/errors"
	"github.com/nedscode/nedward/services"
	"github.com/nedscode/nedward/tracker"
	"github.com/nedscode/nedward/worker"
)

func (c *Client) Stop(names []string, force bool, exclude []string, all bool) error {
	sgs, err := c.getServiceList(names, all)

	// Prompt user to confirm as needed
	if len(names) == 0 && !force && !c.askForConfirmation("Are you sure you want to stop all services?") {
		return nil
	}

	// Perform required checks and actions for services
	if c.ServiceChecks != nil {
		if err = c.ServiceChecks(sgs); err != nil {
			return errors.WithStack(err)
		}
	}

	cfg := services.OperationConfig{
		WorkingDir:       c.WorkingDir,
		NedwardExecutable: c.NedwardExecutable,
		Exclusions:       exclude,
	}

	task := tracker.NewTask(c.Follower.Handle)
	defer c.Follower.Done()

	poolSize := 3
	if c.DisableConcurrentPhases {
		poolSize = 0
	}

	p := worker.NewPool(poolSize)
	p.Start()
	defer func() {
		p.Stop()
		_ = <-p.Complete()
	}()
	for _, s := range sgs {
		_ = s.Stop(cfg, services.ContextOverride{}, task, p)
	}
	return nil
}
