package lxc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Lxc_Terminal(t *testing.T) {
	t.Parallel()
	const guest = pveSDK.GuestID(1100)
	const node = pveSDK.NodeName(test.FirstNode)
	const storage = pveSDK.StorageName(test.GuestStorage)
	const name = pveSDK.GuestName("Test-Lxc-Terminal")
	const password = "Enter123!"
	cl, err := api_test.NewClient()
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	set, _ := MinimumConfig(guest, node, storage, new(false), name, new(pveSDK.PamPassword(password)))
	var vmr *pveSDK.VmRef
	var conn pveSDK.TerminalSession
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
				vmr, err = c.LxcGuest.Create(ctx, set)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Power on guest`,
			test: func(t *testing.T) {
				err := c.Guest.Start(ctx, *vmr)
				require.NoError(t, err)
			}},
		{name: `Establish Terminal connection`,
			test: func(t *testing.T) {
				newConn, err := c.LxcGuest.ConsoleXTerminal(ctx, *vmr)
				require.NoError(t, err)
				require.NotNil(t, newConn)
				conn = newConn
			}},
		{name: `Use Terminal connection`,
			test: func(t *testing.T) {
				err := conn.Write([]byte("root\n"))
				require.NoError(t, err)
				time.Sleep(time.Millisecond * 100)
				err = conn.Write([]byte("\n"))
				require.NoError(t, err)
				time.Sleep(time.Millisecond * 100)
				err = conn.Write([]byte(password))
				require.NoError(t, err)
				time.Sleep(time.Millisecond * 100)
				err = conn.Write([]byte("\n"))
				require.NoError(t, err)
			}},
		{name: `Close Terminal connection`,
			test: func(t *testing.T) {
				err := conn.Close()
				require.NoError(t, err)
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
