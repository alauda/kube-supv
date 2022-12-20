package machineinfo

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/alauda/kube-supv/pkg/utils"
)

type SELinuxStatus string

const (
	Enforcing  SELinuxStatus = "Enforcing"
	Permissive SELinuxStatus = "Permissive"
	Disabled   SELinuxStatus = "Disabled"
)

type PathExist string

const (
	PathExistFile PathExist = "ExistFile"
	PathExistDir  PathExist = "ExistDir"
	PathNotExist  PathExist = "NotExist"
)

type SystemInfo struct {
	SELinux  SELinuxStatus        `json:"SELinux"`
	Apparmor bool                 `json:"apparmor"`
	Swap     bool                 `json:"swap"`
	LongBit  int                  `json:"longBit"`
	Time     time.Time            `json:"time"`
	Timezone string               `json:"timezone"`
	Hostname string               `json:"hostname"`
	Tools    map[string]string    `json:"tools"`
	Pathes   map[string]PathExist `json:"pathes"`
	Systemd  SystemdInfo          `json:"systemd"`
}

func (i *SystemInfo) Explore() error {
	return ExploreMulti(
		i.exploreSELinux,
		i.exploreApparmor,
		i.exploreSwap,
		i.exploreLongBit,
		i.exploreTime,
		i.exploreTimezone,
		i.exploreHostname,
		i.exploreTools,
		i.explorePathes,
		i.Systemd.Explore)
}

func (i *SystemInfo) exploreSELinux() error {
	const cmd = "getenforce"
	if !utils.CommandExist(cmd) {
		i.SELinux = Disabled
	} else {
		ret, err := utils.Exec(cmd)
		if err != nil {
			return err
		}
		i.SELinux = SELinuxStatus(ret)
	}
	return nil
}

func (i *SystemInfo) exploreApparmor() error {
	const path = "/sys/module/apparmor/parameters/enabled"
	ret, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			i.Apparmor = false
			return nil
		}
		return err
	}
	i.Apparmor = strings.TrimSpace(string(ret)) == "Y"
	return nil
}

func (i *SystemInfo) exploreSwap() error {
	const path = "/proc/swaps"
	ret, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	i.Swap = len(utils.Lines(strings.TrimSpace(string(ret)))) > 1
	return nil
}

func (i *SystemInfo) exploreLongBit() error {
	const cmd = "getconf"
	if utils.CommandExist(cmd) {
		ret, err := utils.Exec(cmd)
		if err != nil {
			return err
		}
		bits, err := strconv.Atoi(ret)
		if err != nil {
			return err
		}
		i.LongBit = bits
	} else {
		var e int = 0
		i.LongBit = int(unsafe.Sizeof(e))
	}
	return nil
}

func (i *SystemInfo) exploreTime() error {
	i.Time = time.Now()
	return nil
}

func (i *SystemInfo) exploreTimezone() error {
	location, err := time.LoadLocation("Local")
	if err != nil {
		return err
	}
	i.Timezone = location.String()
	return nil
}

func (i *SystemInfo) exploreHostname() error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	i.Hostname = hostname
	return nil
}

var necessaryTools = []string{}

func (i *SystemInfo) exploreTools() error {
	if i.Tools == nil {
		i.Tools = map[string]string{}
	}
	for _, tool := range necessaryTools {
		path, err := exec.LookPath(tool)
		if err == nil {
			i.Tools[tool] = path
		}
	}
	return nil
}

var importantPathes = []string{}

func (i *SystemInfo) explorePathes() error {
	if i.Pathes == nil {
		i.Pathes = map[string]PathExist{}
	}
	for _, path := range importantPathes {
		info, err := os.Stat(path)
		if err != nil {
			i.Pathes[path] = PathNotExist
		}
		if info.IsDir() {
			i.Pathes[path] = PathExistDir
		} else {
			i.Pathes[path] = PathExistFile
		}
	}
	return nil
}
