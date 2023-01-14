package machine

import (
	"encoding/json"
	"io"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/v3/host"
)

type UserInfo struct {
	Username string `json:"username"`
	UserID   string `json:"userId"`
	GroupID  string `json:"groupId"`
}

func (i *UserInfo) Explore() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	i.Username = u.Username
	i.UserID = u.Uid
	i.GroupID = u.Gid
	return nil
}

type OSInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	VersionID string `json:"versionId"`
	Kernel    string `json:"kernel"`
}

func (i *OSInfo) Explore() error {
	ctx, cancel := utils.DefaultTimeoutCtx()
	defer cancel()
	if runtime.GOOS != "linux" {
		i.Name = runtime.GOOS
	} else {
		if err := i.exploreNameVersion(); err != nil {
			return err
		}
	}
	kernel, err := host.KernelVersionWithContext(ctx)
	if err != nil {
		return err
	}
	i.Kernel = kernel
	return nil
}

var (
	redhatVersionRegexp = regexp.MustCompile(`\d+\.\d+`)
)

func (i *OSInfo) exploreNameVersion() error {
	const osReleasePath = "/etc/os-release"
	ret, err := os.ReadFile(osReleasePath)
	if err != nil {
		i.Name = runtime.GOOS
		i.Version = "Unknown"
		return nil
	}
	lines := utils.Lines(string(ret))
	for _, line := range lines {
		key, val := utils.SplitKeyVal(line, "=")
		if key == "NAME" {
			i.Name = strings.Trim(val, `"`)
		}
		if key == "VERSION_ID" {
			i.VersionID = strings.Trim(val, `"`)
		}
		if key == "VERSION" {
			i.Version = strings.Trim(val, `"`)
		}
	}
	if i.Name == "CentOS Linux" || i.Name == "Red Hat Enterprise Linux Server" {
		ret, err := os.ReadFile("/etc/redhat-release")
		if err == nil {
			version := redhatVersionRegexp.FindString(string(ret))
			if version != "" {
				i.VersionID = version
			}
		}
	}
	return nil
}

type MachineInfo struct {
	User     UserInfo     `json:"user"`
	Hardware HardwareInfo `json:"hardware"`
	OS       OSInfo       `json:"os"`
	Network  NetworkInfo  `json:"network"`
	System   SystemInfo   `json:"system"`
}

func (i *MachineInfo) Explore() error {
	return ExploreMulti(
		i.User.Explore,
		i.Hardware.Explore,
		i.OS.Explore,
		i.Network.Explore,
		i.System.Explore)
}

func (i *MachineInfo) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(i); err != nil {
		return errors.Wrap(err, "write machine info json")
	}
	return nil
}

func CollectMachineInfo() (*MachineInfo, error) {
	mi := &MachineInfo{}
	if err := mi.Explore(); err != nil {
		return nil, err
	}
	return mi, nil
}
