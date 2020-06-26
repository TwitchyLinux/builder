package sysd

import (
	"fmt"
	"net"
	"strings"
)

// NetDHCP describes valid modes for configuring a network.
type NetDHCP string

// Valid NetDHCP values.
const (
	NetDHCPDual NetDHCP = "yes"
	NetDHCPv4   NetDHCP = "ipv4"
	NetDHCPv6   NetDHCP = "ipv6"
)

// Network describes configuration for a systemd-networkd.
type Network struct {
	Match NetDevMatch
	Link  NetLink
	Net   NetSettings
}

// String returns the structure serialized as configuration.
func (n *Network) String() string {
	var out strings.Builder
	if !n.Match.IsEmpty() {
		out.WriteString(n.Match.String())
	}
	if !n.Link.IsEmpty() {
		out.WriteString(n.Link.String())
	}

	out.WriteString(n.Net.String())
	return out.String()
}

// NetSettings describes configuration to apply to a network.
type NetSettings struct {
	DHCP           NetDHCP
	IsDefaultRoute bool
	Address        net.IPNet
	Gateway        net.IP
	DNS            []string
}

// String returns the structure serialized as configuration.
func (n *NetSettings) String() string {
	var out strings.Builder
	out.WriteString("[Network]\n")
	if n.DHCP != "" {
		out.WriteString(fmt.Sprintf("DHCP=%s\n", n.DHCP))
	}
	if n.IsDefaultRoute {
		out.WriteString("DefaultRouteOnDevice=true\n")
	}
	if len(n.Address.IP) > 0 {
		out.WriteString(fmt.Sprintf("Address=%s\n", n.Address.String()))
	}
	if len(n.Gateway) > 0 {
		out.WriteString(fmt.Sprintf("Gateway=%s\n", n.Gateway.String()))
	}
	for _, dns := range n.DNS {
		out.WriteString(fmt.Sprintf("DNS=%s\n", dns))
	}

	out.WriteString("\n")
	return out.String()
}

// NetLink describes configuration to apply to a network link
// when the configuration matches a network device.
type NetLink struct {
	MAC               string
	MTU               uint
	Unmanaged         bool
	RequiredForOnline bool
}

func (n *NetLink) IsEmpty() bool {
	return *n == NetLink{}
}

// String returns the structure serialized as configuration.
func (n *NetLink) String() string {
	var out strings.Builder
	out.WriteString("[Link]\n")
	if n.MAC != "" {
		out.WriteString(fmt.Sprintf("MACAddress=%s\n", n.MAC))
	}
	if n.MTU != 0 {
		out.WriteString(fmt.Sprintf("MTUBytes=%d\n", n.MTU))
	}
	if n.Unmanaged {
		out.WriteString("Unmanaged=yes\n")
	}
	if n.RequiredForOnline {
		out.WriteString("RequiredForOnline=yes\n")
	}

	out.WriteString("\n")
	return out.String()
}

// NetDevMatch enumerates constraints which determine if the configuration in
// the file should be applied to a given device. If multiple configuration
// files would match, only the first (lexically) is applied.
type NetDevMatch struct {
	Name string `toml:"name"`
	MAC  string `toml:"mac_address"`
	Path string `toml:"path"`
	Type string `toml:"type"`
	SSID string `toml:"ssid"`
	Host string `toml:"host"`
}

func (m *NetDevMatch) IsEmpty() bool {
	return *m == NetDevMatch{}
}

// String returns the structure serialized as configuration.
func (m *NetDevMatch) String() string {
	var out strings.Builder
	out.WriteString("[Match]\n")
	if m.Name != "" {
		out.WriteString(fmt.Sprintf("Name=%s\n", m.Name))
	}
	if m.MAC != "" {
		out.WriteString(fmt.Sprintf("MACAddress=%s\n", m.MAC))
	}
	if m.Path != "" {
		out.WriteString(fmt.Sprintf("Path=%s\n", m.Path))
	}
	if m.Type != "" {
		out.WriteString(fmt.Sprintf("Type=%s\n", m.Type))
	}
	if m.SSID != "" {
		out.WriteString(fmt.Sprintf("SSID=%s\n", m.SSID))
	}
	if m.Host != "" {
		out.WriteString(fmt.Sprintf("Host=%s\n", m.Host))
	}

	out.WriteString("\n")
	return out.String()
}
