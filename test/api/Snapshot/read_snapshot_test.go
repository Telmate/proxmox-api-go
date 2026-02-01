package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Snapshot_ReadQemu(t *testing.T) {
	t.Parallel()
	const snap1 = pveSDK.SnapshotName("snap1")
	const guest = pveSDK.GuestID(804)
	const node = pveSDK.NodeName(test.FirstNode)
	snapshot := pveSDK.SnapshotInfo{
		Name:        snap1,
		Description: "Test snapshot" + body.Symbols,
		VmState:     util.Pointer(false)}
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var currentTime time.Time
	currentTimePtr := &currentTime
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				require.Error(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				guestCreate(t, ctx, cl, guest, node, "Test-Snapshot-ReadQemu")
			}},
		{name: `Create snapshot`,
			test: func(t *testing.T) {
				*currentTimePtr = time.Now()
				require.NoError(t, c.Snapshot.CreateQemu(ctx, *pveSDK.NewVmRef(guest), snapshot.Name, snapshot.Description, false))
			}},
		{name: `Read snapshot`,
			test: func(t *testing.T) {
				raw, err := c.Snapshot.ReadQemu(ctx, *pveSDK.NewVmRef(guest), snapshot.Name)
				require.NoError(t, err)
				require.NotNil(t, raw)
				config, _ := raw.Get(nil)
				expected := util.Pointer(pveSDK.ConfigQemu{
					Name: util.Pointer(pveSDK.GuestName("Test-Snapshot-ReadQemu")),
					Bios: "seabios",
					Boot: " ",
					CPU: &pveSDK.QemuCPU{
						Cores: util.Pointer(pveSDK.QemuCpuCores(1)),
					},
					Description: util.Pointer("Test snapshot" + body.Symbols),
					EFIDisk:     pveSDK.QemuDevice{},
					Hotplug:     "network,disk,usb",
					Memory: &pveSDK.QemuMemory{
						CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16)),
					},
					Protection:      util.Pointer(false),
					QemuDisks:       pveSDK.QemuDevices{},
					QemuKVM:         util.Pointer(true),
					QemuOs:          "other",
					QemuUnusedDisks: pveSDK.QemuDevices{},
					QemuVga:         pveSDK.QemuDevice{},
					Tablet:          util.Pointer(true),
					StartAtNodeBoot: util.Pointer(false),
					Scsihw:          "lsi",
				})
				config.Smbios1 = "" // ignore smbios differences
				require.Equal(t, expected, config)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				require.NoError(t, pveSDK.NewVmRef(guest).Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
