package sysd

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNetwork(t *testing.T) {
	tcs := []struct {
		name string
		conf Network
		out  string
	}{
		{
			"empty",
			Network{},
			"[Network]\n\n",
		},
		{
			"basic dhcp",
			Network{
				Match: NetDevMatch{
					Name: "en*",
				},
				Net: NetSettings{
					DHCP: NetDHCPv4,
				},
			},
			"[Match]\nName=en*\n\n[Network]\nDHCP=ipv4\n\n",
		},
		{
			"basic static",
			Network{
				Match: NetDevMatch{
					Name: "en*",
				},
				Net: NetSettings{
					Address: net.IPNet{IP: net.ParseIP("192.168.1.12"), Mask: net.IPMask{255, 255, 255, 0}},
					Gateway: net.ParseIP("192.168.1.1"),
					DNS:     []string{"8.8.8.8"},
				},
			},
			"[Match]\nName=en*\n\n[Network]\nAddress=192.168.1.12/24\nGateway=192.168.1.1\nDNS=8.8.8.8\n\n",
		},
		{
			"unmanaged",
			Network{
				Match: NetDevMatch{
					Name: "bridge*",
				},
				Link: NetLink{
					Unmanaged: true,
				},
			},
			"[Match]\nName=bridge*\n\n[Link]\nUnmanaged=yes\n\n[Network]\n\n",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			out := tc.conf.String()
			if diff := cmp.Diff(out, tc.out); diff != "" {
				t.Errorf("unexpected output (+got, -want):\n%s", diff)
			}
		})
	}
}
