package nedward

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/nedscode/nedward/services"
)

func (c *Client) Status(names []string, all bool) (string, error) {
	sgs, err := c.getServiceList(names, all)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if len(sgs) == 0 {
		return "No services found\n", nil
	}

	buf := new(bytes.Buffer)

	table := tablewriter.NewWriter(buf)
	headings := []string{
		"Name",
		"Status",
		"PID",
		"Ports",
		"Stdout",
		"Stderr",
		"Start Time",
	}
	if all {
		headings = append(headings, "Config")
	}
	table.SetHeader(headings)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	for _, s := range sgs {
		statuses, err := s.Status()
		if err != nil {
			return "", errors.WithStack(err)
		}
		for _, status := range statuses {
			row := []string{
				status.Service.Name,
				status.Status,
				strconv.Itoa(status.Pid),
				strings.Join(status.Ports, ", "),
				strconv.Itoa(status.StdoutCount) + " lines",
				strconv.Itoa(status.StderrCount) + " lines",
				status.StartTime.Format("2006-01-02 15:04:05"),
			}
			if all {
				configPath := status.Service.ConfigFile
				wd, err := os.Getwd()
				if err == nil {
					relativePath, err := filepath.Rel(wd, configPath)
					if err == nil && len(configPath) > len(relativePath) {
						configPath = relativePath
					}
				}
				row = append(row, configPath)
			}
			table.Append(row)
		}
	}
	table.Render()
	return buf.String(), nil
}

func (c *Client) getServiceList(names []string, all bool) ([]services.ServiceOrGroup, error) {
	var sgs []services.ServiceOrGroup
	var err error

	if all {
		runningServices, err := services.LoadRunningServices()
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return runningServices, nil
		}
		for _, service := range runningServices {
			for _, name := range names {
				if name == service.GetName() {
					sgs = append(sgs, service)
				}
			}
		}
		return sgs, nil
	}

	if len(names) == 0 {
		for _, service := range c.getAllServicesSorted() {
			var s []services.ServiceStatus
			s, err = service.Status()
			if err != nil {
				return nil, errors.WithStack(err)
			}
			for _, status := range s {
				if status.Status != services.StatusStopped {
					sgs = append(sgs, service)
				}
			}
		}
		return sgs, nil
	}

	sgs, err = c.getServicesOrGroups(names)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return sgs, nil
}
