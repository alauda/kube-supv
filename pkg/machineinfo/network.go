package machineinfo

import (
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/vishvananda/netlink"
)

type ListeningPort struct {
	Port      int      `json:"port"`
	Protocols []string `json:"protocols"`
}

type NetworkInterface struct {
	Name  string   `json:"name"`
	MTU   int      `json:"MTU"`
	MAC   string   `json:"MAC"`
	Flags string   `json:"flags"`
	IPs   []net.IP `json:"IPs"`
}

type NetworkInfo struct {
	Listening  map[int][]string   `json:"listening"`
	Interfaces []NetworkInterface `json:"interfaces"`
	Routes     []string           `json:"routes"`
}

func (i *NetworkInfo) Explore() error {
	return ExploreMulti(i.exploreListening, i.exploreInterfaces, i.exploreRoutes)
}

func (i *NetworkInfo) exploreListening() error {
	const ListenState = "0A"
	if runtime.GOOS != "linux" {
		return nil
	}
	for path, protocol := range map[string]string{
		"/proc/net/tcp":  "tcp",
		"/proc/net/tcp6": "tcp6",
	} {
		ret, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		for _, line := range utils.Lines(string(ret)) {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			if fields[3] == ListenState {
				_, val := utils.SplitKeyVal(fields[1], ":")
				port, err := strconv.ParseInt(val, 16, 32)
				if err == nil {
					i.Listening[int(port)] = append(i.Listening[int(port)], protocol)
				}
			}
		}
	}
	return nil
}

func (i *NetworkInfo) exploreRoutes() error {
	if runtime.GOOS != "linux" {
		return nil
	}
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}
	for _, link := range links {
		routes, err := netlink.RouteList(link, 0)
		if err != nil {
			return err
		}
		for _, route := range routes {
			i.Routes = append(i.Routes, route.String())
		}
	}

	return nil
}

func (i *NetworkInfo) exploreInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, iface := range ifaces {
		var ips []net.IP
		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ip, _, err := net.ParseCIDR(addr.String())
				if err != nil {
					ips = append(ips, ip)
				}
			}
		}
		i.Interfaces = append(i.Interfaces, NetworkInterface{
			Name:  iface.Name,
			MTU:   iface.MTU,
			MAC:   iface.HardwareAddr.String(),
			Flags: iface.Flags.String(),
			IPs:   ips,
		})
	}
	return nil
}
