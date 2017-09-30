package nedward_test

import (
	"errors"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/theothertomelliott/must"
	"github.com/nedscode/nedward/common"
	"github.com/nedscode/nedward/nedward"
)

func TestRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var tests = []struct {
		name             string
		path             string
		config           string
		servicesStart    []string
		servicesRestart  []string
		skipBuild        bool
		tail             bool
		noWatch          bool
		exclude          []string
		expectedStates   map[string]string
		expectedMessages []string
		expectedServices int
		err              error
	}{
		{
			name:            "single service",
			path:            "testdata/single",
			config:          "nedward.json",
			servicesStart:   []string{"service"},
			servicesRestart: []string{"service"},
			expectedStates: map[string]string{
				"service":         "Pending", // This isn't technically right
				"service > Stop":  "Success",
				"service > Build": "Success",
				"service > Start": "Success",
			},
			expectedServices: 1,
		},
		{
			name:             "service not found",
			path:             "testdata/single",
			config:           "nedward.json",
			servicesStart:    []string{"service"},
			servicesRestart:  []string{"missing"},
			err:              errors.New("Service or group not found"),
			expectedServices: 1,
		},
		{
			name:          "group, restart all",
			path:          "testdata/group",
			config:        "nedward.json",
			servicesStart: []string{"group"},
			expectedStates: map[string]string{
				"service1":         "Pending",
				"service1 > Stop":  "Success",
				"service1 > Start": "Success",
				"service1 > Build": "Success",
				"service2":         "Pending",
				"service2 > Stop":  "Success",
				"service2 > Start": "Success",
				"service2 > Build": "Success",
				"service3":         "Pending",
				"service3 > Stop":  "Success",
				"service3 > Start": "Success",
				"service3 > Build": "Success",
			},
			expectedServices: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var err error

			// Copy test content into a temp dir on the GOPATH & defer deletion
			wd, cleanup := createWorkingDir(t, test.name, test.path)
			defer cleanup()

			client, err := nedward.NewClientWithConfig(path.Join(wd, test.config), common.NedwardVersion)
			if err != nil {
				t.Fatal(err)
			}
			client.WorkingDir = wd
			tf := newTestFollower()
			client.Follower = tf
			client.NedwardExecutable = nedwardExecutable
			client.DisableConcurrentPhases = true

			err = client.Start(test.servicesStart, test.skipBuild, false, test.noWatch, test.exclude)
			if err != nil {
				t.Fatal(err)
			}

			childProcesses := getRunnerAndServiceProcesses(t)

			// Reset the follower
			tf = newTestFollower()
			client.Follower = tf

			err = client.Restart(test.servicesRestart, true, test.skipBuild, false, test.noWatch, test.exclude)
			must.BeEqualErrors(t, test.err, err)
			must.BeEqual(t, test.expectedStates, tf.states)
			must.BeEqual(t, test.expectedMessages, tf.messages)

			if err == nil {
				for _, p := range childProcesses {
					process, err := os.FindProcess(int(p.Pid))
					if err != nil {
						t.Fatal(err)
					}
					if err == nil {
						if process.Signal(syscall.Signal(0)) == nil {
							t.Errorf("process should not still be running: %v", p.Pid)
						}
					}
				}
			}

			// Verify that the process actually started
			verifyAndStopRunners(t, test.expectedServices)
		})
	}
}
