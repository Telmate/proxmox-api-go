package lxc

import (
	"context"
	"net"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func CheckConfig(t *testing.T, ctx context.Context, c *pveSDK.Client, guestID pveSDK.GuestID, expected pveSDK.ConfigLXC) {
	raw, err := c.New().LxcGuest.Read(ctx, *pveSDK.NewVmRef(pveSDK.GuestID(guestID)))
	require.NoError(t, err)
	require.NotNil(t, raw)
	config := raw.Get(nil, pveSDK.PowerStateUnknown)
	config.Digest = [20]byte{} // Ignore digest for comparison as it is always different
	require.EqualExportedValues(t, expected, *config)
}

func ParseMAC(rawMAC string) net.HardwareAddr {
	mac, err := net.ParseMAC(rawMAC)
	if err != nil {
		panic(err)
	}
	return mac
}

func ParseCIDR(rawCIDR string) (net.IP, *net.IPNet) {
	ip, cidr, err := net.ParseCIDR(rawCIDR)
	if err != nil {
		panic(err)
	}
	return ip, cidr
}
