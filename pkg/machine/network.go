package machine

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
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
	Flags []string `json:"flags"`
	IP    []net.IP `json:"IP"`
}

func NewNetworkInterface(iface net.Interface) NetworkInterface {
	ni := NetworkInterface{
		Name: iface.Name,
		MTU:  iface.MTU,
		MAC:  iface.HardwareAddr.String(),
	}

	if addrs, err := iface.Addrs(); err == nil {
		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err == nil {
				ni.IP = append(ni.IP, ip)
			}
		}
	}
	if flags := iface.Flags.String(); flags != "0" {
		ni.Flags = strings.Split(flags, "|")
	}
	return ni
}

type NetworkInfo struct {
	Listening  map[int][]string   `json:"listening"`
	Interfaces []NetworkInterface `json:"interfaces"`
	Routes     []Route            `json:"routes"`
}

type Route struct {
	Dest    string `json:"dest"`
	Src     string `json:"src"`
	Gateway string `json:"gateway"`
	Device  string `json:"device"`
}

func NewRoute(r netlink.Route) Route {
	route := Route{}
	if r.MPLSDst != nil {
		route.Dest = strconv.Itoa(*r.MPLSDst)
	} else if r.Dst != nil {
		route.Dest = r.Dst.String()
	}
	if r.Src != nil {
		route.Src = r.Src.String()
	}
	if len(r.MultiPath) > 0 {
		route.Gateway = fmt.Sprintf("%s", r.MultiPath)
	} else if r.Gw != nil {
		route.Gateway = r.Gw.String()
	}

	link, err := netlink.LinkByIndex(r.LinkIndex)
	if err == nil {
		if attrs := link.Attrs(); attrs != nil {
			route.Device = attrs.Name
		}
	}

	return route
}

func (i *NetworkInfo) Explore() error {
	return ExploreMulti(i.exploreListening, i.exploreInterfaces, i.exploreRoutes)
}

func (i *NetworkInfo) exploreListening() error {
	if runtime.GOOS != "linux" {
		return nil
	}

	const ListenState = "0A"
	if i.Listening == nil {
		i.Listening = map[int][]string{}
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
		return errors.Wrap(err, `netlink.LinkList`)
	}
	for index, link := range links {
		routes, err := netlink.RouteList(link, 0)
		if err != nil {
			return errors.Wrapf(err, `netlink.RouteList link: %d`, index)
		}
		for _, route := range routes {
			i.Routes = append(i.Routes, NewRoute(route))
		}
	}

	return nil
}

func (i *NetworkInfo) exploreInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return errors.Wrap(err, `net.Interfaces`)
	}
	for _, iface := range ifaces {
		i.Interfaces = append(i.Interfaces, NewNetworkInterface(iface))
	}
	return nil
}
