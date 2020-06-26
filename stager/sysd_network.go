package stager

import (
	"fmt"
	"net"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/conf/sysd"
	"github.com/twitchylinux/builder/units"
)

// systemdNetwork describes configuration of a network.
type systemdNetwork struct {
	If    *StepCondition   `toml:"if"`
	Match sysd.NetDevMatch `toml:"match"`
	Conf  systemdNetConf   `toml:"config"`
}

// systemdNetConf describes what configuration to apply to a network.
type systemdNetConf struct {
	DHCP           sysd.NetDHCP `toml:"dhcp"`
	IsDefaultRoute bool         `toml:"default_route"`
	Address        string       `toml:"address"`
	Gateway        string       `toml:"gateway"`
	DNS            []string     `toml:"dns_servers"`
}

func systemdNetConfig(opts Options, tree *toml.Tree) (*units.InstallFiles, error) {
	conf := map[string]systemdNetwork{}
	t := tree.Get(keySysdNetworks)
	if t == nil {
		return nil, nil
	}
	ge, ok := t.(*toml.Tree)
	if !ok {
		return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyUdevRules, t)
	}
	if err := ge.Unmarshal(&conf); err != nil {
		return nil, err
	}
	if len(conf) == 0 {
		return nil, nil
	}

	var outFiles []units.FileInfo
	for name, ruleSet := range conf {
		skip, err := ruleSet.If.ShouldSkip(tree, opts)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		c := sysd.Network{
			Match: ruleSet.Match,
			Net: sysd.NetSettings{
				DHCP:           ruleSet.Conf.DHCP,
				IsDefaultRoute: ruleSet.Conf.IsDefaultRoute,
				DNS:            ruleSet.Conf.DNS,
			},
		}
		if ruleSet.Conf.Address != "" {
			ip, netm, err := net.ParseCIDR(ruleSet.Conf.Address)
			if err != nil {
				return nil, fmt.Errorf("parsing network %q.address: %v", name, err)
			}
			c.Net.Address = net.IPNet{IP: ip, Mask: netm.Mask}
		}
		if ruleSet.Conf.Gateway != "" {
			c.Net.Gateway = net.ParseIP(ruleSet.Conf.Gateway)
		}
		outFiles = append(outFiles, units.FileInfo{
			Path: fmt.Sprintf("/etc/systemd/network/%s.network", name),
			Data: []byte(c.String()),
		})
	}

	if len(outFiles) == 0 {
		return nil, nil
	}
	return &units.InstallFiles{
		UnitName: "systemd-network",
		Mkdir:    "/etc/systemd/network",
		Files:    outFiles,
	}, nil
}
