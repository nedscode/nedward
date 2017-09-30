package nedward

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/nedscode/nedward/output"
	"github.com/nedscode/nedward/services"
	"github.com/nedscode/nedward/tracker"
	"github.com/nedscode/nedward/worker"
)

type Client struct {
	Logger *log.Logger

	Input  io.Reader
	Output io.Writer

	Config string

	ServiceChecks func([]services.ServiceOrGroup) error

	NedwardExecutable string

	Follower TaskFollower

	// Prevent build, launch and stop phases from running concurrently
	DisableConcurrentPhases bool

	WorkingDir string

	basePath   string
	groupMap   map[string]*services.ServiceGroupConfig
	serviceMap map[string]*services.ServiceConfig
}

type TaskFollower interface {
	Handle(update tracker.Task)
	Done()
}

// NewClient creates an nedward client an empty configuration
func NewClient() (*Client, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Client{
		Input:      os.Stdin,
		Output:     os.Stdout,
		Follower:   output.NewFollower(),
		Logger:     log.New(ioutil.Discard, "", 0), // Default to a logger that discards output
		WorkingDir: wd,
		groupMap:   make(map[string]*services.ServiceGroupConfig),
		serviceMap: make(map[string]*services.ServiceConfig),
	}, nil
}

// NewClientWithConfig creates an Nedward client and loads the config from the given path
func NewClientWithConfig(configPath, version string) (*Client, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := &Client{
		Input:      os.Stdin,
		Output:     os.Stdout,
		Follower:   output.NewFollower(),
		Logger:     log.New(ioutil.Discard, "", 0), // Default to a logger that discards output
		WorkingDir: wd,
		Config:     configPath,
		groupMap:   make(map[string]*services.ServiceGroupConfig),
		serviceMap: make(map[string]*services.ServiceConfig),
	}
	err = client.LoadConfig(version)
	return client, errors.WithStack(err)
}

func (c *Client) BasePath() string {
	return c.basePath
}

func (c *Client) ServiceMap() map[string]*services.ServiceConfig {
	return c.serviceMap
}

func (c *Client) startAndTrack(sgs []services.ServiceOrGroup, skipBuild bool, tail bool, noWatch bool, exclude []string, nedwardExecutable string) error {
	cfg := services.OperationConfig{
		WorkingDir:       c.WorkingDir,
		NedwardExecutable: nedwardExecutable,
		Exclusions:       exclude,
		SkipBuild:        skipBuild,
		NoWatch:          noWatch,
	}

	task := tracker.NewTask(c.Follower.Handle)
	defer c.Follower.Done()

	poolSize := 1
	if c.DisableConcurrentPhases {
		poolSize = 0
	}
	p := worker.NewPool(poolSize)
	p.Start()
	defer func() {
		p.Stop()
		_ = <-p.Complete()
	}()
	var err error
	for _, s := range sgs {
		if skipBuild {
			err = s.Launch(cfg, services.ContextOverride{}, task, p)
			if err != nil {
				return errors.WithMessage(err, "Error launching "+s.GetName())
			}
		} else {
			err = s.Start(cfg, services.ContextOverride{}, task, p)
			if err != nil {
				return errors.WithMessage(err, "Error starting "+s.GetName())
			}
		}
	}
	return nil
}

func (c *Client) tailFromFlag(names []string) error {
	fmt.Println("=== Logs ===")
	return errors.WithStack(c.Log(names))
}

func (c *Client) askForConfirmation(question string) bool {

	// Skip confirmations for children. For example, for enabling sudo.
	isChild := os.Getenv("ISCHILD")
	if isChild != "" {
		return true
	}

	reader := bufio.NewReader(c.Input)
	for {
		fmt.Fprintf(c.Output, "%s [y/n]? ", question)

		response, err := reader.ReadString('\n')
		if err != nil {
			return false
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

type serviceConfigByPID []*services.ServiceConfig

func (s serviceConfigByPID) Len() int {
	return len(s)
}
func (s serviceConfigByPID) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s serviceConfigByPID) Less(i, j int) bool {
	cmd1, _ := s[i].GetCommand(services.ContextOverride{})
	cmd2, _ := s[j].GetCommand(services.ContextOverride{})
	return cmd1.Pid < cmd2.Pid
}
