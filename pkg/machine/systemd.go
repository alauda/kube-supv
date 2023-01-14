package machine

import (
	"strings"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

const (
	disabled = "disabled"
	enabled  = "enabled"
	active   = "active"
)

type ServiceState struct {
	Active  bool `json:"active"`
	Enabled bool `json:"enabled"`
}

type SystemdInfo struct {
	Version  string                   `json:"version"`
	Services map[string]*ServiceState `json:"services"`
}

func (i *SystemdInfo) Explore() error {
	return ExploreMulti(i.exploreVersion, i.exploreServices)
}

func (i *SystemdInfo) exploreVersion() error {
	cmd := "systemctl"
	if !utils.CommandExist(cmd) {
		return nil
	}
	args := []string{"show", "-p", "Version"}
	ret, err := utils.Exec(cmd, args...)
	if err != nil {
		return errors.Wrapf(err, `%s %s`, cmd, strings.Join(args, " "))
	}
	fields := strings.Split(ret, "=")
	i.Version = fields[len(fields)-1]
	return nil
}

func (i *SystemdInfo) exploreServices() error {
	cmd := "systemctl"
	if !utils.CommandExist(cmd) {
		return nil
	}
	const suffix = ".service"
	args := []string{"list-unit-files", "*.service", "--no-pager"}
	ret, err := utils.Exec(cmd, args...)
	if err != nil {
		return errors.Wrapf(err, `%s %s`, cmd, strings.Join(args, " "))
	}
	if i.Services == nil {
		i.Services = map[string]*ServiceState{}
	}
	for _, line := range utils.Lines(ret) {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimSuffix(fields[0], suffix)
		state := fields[1]
		if state != enabled && state != disabled {
			continue
		}
		if i.Services[name] == nil {
			i.Services[name] = &ServiceState{}
		}
		i.Services[name].Enabled = state == enabled
	}

	args2 := []string{"list-units", "*.service", "-t", "service", "--no-pager"}
	ret, err = utils.Exec(cmd, args2...)
	if err != nil {
		return errors.Wrapf(err, `%s %s`, cmd, strings.Join(args2, " "))
	}

	for _, line := range utils.Lines(ret) {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		name := strings.TrimSuffix(fields[0], suffix)
		state := fields[2]
		if i.Services[name] == nil {
			continue
		}
		i.Services[name].Active = state == active
	}

	return nil
}
