package lxc

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Lxc_Read_Interface_Info(t *testing.T) {
	t.Parallel()
	const guest = pveSDK.GuestID(1003)
	const node = pveSDK.NodeName(test.FirstNode)
	const storage = pveSDK.StorageName(test.GuestStorage)
	const name = pveSDK.GuestName("Test-Lxc-Read-Interface-Info")
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, nil)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, expected := MinimumConfig(guest, node, storage, new(true), name)

	mac := "28:4D:C2:C8:7B:8A"
	networkName := pveSDK.LxcNetworkName("eth0")
	ipv4 := "10.10.10.2/24"
	ipv6 := "6951:f53f:a891:f5c:cca0:ad8:38b2:ea66/48"

	networks := pveSDK.LxcNetworks{pveSDK.LxcNetworkID0: pveSDK.LxcNetwork{
		Bridge:      new("vmbr0"),
		Connected:   new(true),
		Firewall:    new(false),
		HostManaged: new(true),
		MAC:         new(ParseMAC(mac)),
		Name:        &networkName,
		IPv4:        &pveSDK.LxcIPv4{Address: new(pveSDK.IPv4CIDR(ipv4))},
		IPv6:        &pveSDK.LxcIPv6{Address: new(pveSDK.IPv6CIDR(ipv6))}}}

	set.Networks = networks
	expected.Networks = networks
	ipv4address, _ := ParseCIDR(ipv4)
	ipv6address, _ := ParseCIDR(ipv6)
	var vmr *pveSDK.VmRef
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(guest))
				require.NoError(t, err)
				require.False(t, existed)
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				vmr, err = set.Create(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guest, expected)
			}},
		{name: `Start guest`,
			test: func(t *testing.T) {
				err = c.Guest.Start(ctx, *vmr)
				require.NoError(t, err)
			}},
		{name: `Get Interface info`,
			test: func(t *testing.T) {
				raw, err := c.LxcGuest.ReadNetworkInterfaceInfo(ctx, *vmr)
				require.NoError(t, err)
				require.NotNil(t, raw)
				rawNetwork, ok := raw.SelectMacAddress(ParseMAC(mac))
				require.True(t, ok)
				require.NotNil(t, rawNetwork)
				network := rawNetwork.Get()

				require.Equal(t, networkName, network.Name)
				require.Equal(t, ParseMAC(mac), network.MacAddress)
				ipAddresses := []net.IP{ipv4address, ipv6address}
				require.GreaterOrEqual(t, len(network.IpAddresses), len(ipAddresses))
				for _, e := range ipAddresses {
					var exists bool
					for _, ee := range network.IpAddresses {
						if e.String() == ee.String() {
							exists = true
						}
					}
					require.True(t, exists)
				}

			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				existed, err := c.Guest.Delete(ctx, *vmr)
				require.NoError(t, err)
				require.True(t, existed)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
