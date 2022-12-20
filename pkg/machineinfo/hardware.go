package machineinfo

import (
	"runtime"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type CPUInfo struct {
	Family string   `json:"family"`
	Mode   string   `json:"mode"`
	Vendor string   `json:"vendor"`
	Arch   string   `json:"arch"`
	Flags  []string `json:"flags"`
	CPUs   int      `json:"CPUs"`
	Mhz    float64  `json:"mhz"`
}

func (i *CPUInfo) Explore() error {
	if err := i.exploreArch(); err != nil {
		return err
	}
	ctx, cancel := utils.DefaultTimeoutCtx()
	defer cancel()
	info, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return err
	}
	i.CPUs, err = cpu.Counts(true)
	if err != nil {
		return err
	}
	if len(info) > 0 {
		i.Flags = info[0].Flags
		i.Family = info[0].Family
		i.Mode = info[0].Model
		for _, it := range info {
			i.Mhz += it.Mhz
		}
		i.Mhz = i.Mhz / float64(len(info))
	}
	return nil
}

func (i *CPUInfo) exploreArch() error {
	const cmd = "arch"
	if utils.CommandExist(cmd) {
		ret, err := utils.Exec(cmd)
		if err != nil {
			return err
		}
		i.Arch = ret
	} else {
		switch runtime.GOARCH {
		case "arm64":
			i.Arch = "aarch64"
		case "amd64":
			i.Arch = "x86_64"
		default:
			i.Arch = runtime.GOARCH
		}
	}
	return nil
}

type MemoryInfo struct {
	Total          uint64 `json:"total"`
	Available      uint64 `json:"available"`
	SwapTotal      uint64 `json:"swapTotal"`
	SwapFree       uint64 `json:"swapFree"`
	HugePageSize   uint64 `json:"hugePageSize"`
	HugePagesTotal uint64 `json:"hugePagesTotal"`
	HugePagesFree  uint64 `json:"hugePagesFree"`
	HugePagesRsvd  uint64 `json:"hugePagesRsvd"`
	HugePagesSurp  uint64 `json:"hugePagesSurp"`
}

func (i *MemoryInfo) Explore() error {
	ctx, cancel := utils.DefaultTimeoutCtx()
	defer cancel()
	info, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}
	i.Total = info.Total
	i.Available = info.Available
	i.SwapTotal = info.SwapTotal
	i.SwapFree = info.SwapFree
	i.HugePageSize = info.HugePageSize
	i.HugePagesTotal = info.HugePagesTotal
	i.HugePagesFree = info.HugePagesFree
	i.HugePagesRsvd = info.HugePagesRsvd
	i.HugePagesSurp = info.HugePagesSurp
	return nil
}

type HardwareInfo struct {
	CPU    CPUInfo    `json:"CPU"`
	Memory MemoryInfo `json:"memory"`
	PCIs   []string   `json:"PCIs"`
}

func (i *HardwareInfo) Explore() error {
	return ExploreMulti(
		i.CPU.Explore,
		i.Memory.Explore,
		i.explorePCIs)
}

func (i *HardwareInfo) explorePCIs() error {
	const cmd = "lspci"
	if utils.CommandExist(cmd) {
		ret, err := utils.Exec(cmd, "-nn")
		if err != nil {
			return err
		}
		i.PCIs = utils.Lines(ret)
	}
	return nil
}
