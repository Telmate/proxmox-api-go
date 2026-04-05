package api_test

import (
	"context"
	"crypto/tls"
	"strconv"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

func Test_Qemu_Update_Max_Transform(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Update-Max"
	const guestID = 1020
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	var vmr *pveSDK.VmRef
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				vmr = pveSDK.NewVmRef(pveSDK.GuestID(guestID))
				require.Error(t, vmr.Delete(ctx, cl))
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				config := pveSDK.ConfigQemu{
					ID:     util.Pointer(pveSDK.GuestID(guestID)),
					CPU:    &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:   util.Pointer(pveSDK.GuestName(guestName)),
					Node:   util.Pointer(node),
					EfiDisk: &pveSDK.EfiDisk{
						Format:  util.Pointer(pveSDK.QemuDiskFormat("raw")),
						Storage: util.Pointer(pveSDK.StorageName(test.GuestStorage)),
					},
				}
				vmr, err := config.Create(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				vmr = pveSDK.NewVmRef(pveSDK.GuestID(guestID))
				raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, cl)
				require.NoError(t, err)
				require.NotNil(t, raw)
				config, err := raw.Get(*vmr)
				require.NoError(t, err)
				require.NotNil(t, config)
				config.Smbios1 = "" // This field is unique and cannot be predicted, so we ignore it for the comparison.
				require.EqualExportedValues(t, &pveSDK.ConfigQemu{
					Bios:        "seabios",
					Boot:        " ",
					CPU:         &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Description: util.Pointer(""),
					EfiDisk: &pveSDK.EfiDisk{
						Size:            1024,
						Format:          util.Pointer(pveSDK.QemuDiskFormat("raw")),
						PreEnrolledKeys: util.Pointer(false),
						Storage:         util.Pointer(pveSDK.StorageName(test.GuestStorage)),
					},
					Hotplug:         "network,disk,usb",
					ID:              util.Pointer(pveSDK.GuestID(guestID)),
					Memory:          &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:            util.Pointer(pveSDK.GuestName(guestName)),
					Node:            util.Pointer(node),
					Protection:      util.Pointer(false),
					QemuDisks:       pveSDK.QemuDevices{},
					QemuKVM:         util.Pointer(true),
					QemuOs:          "other",
					QemuUnusedDisks: pveSDK.QemuDevices{},
					QemuVga:         pveSDK.QemuDevice{},
					Scsihw:          "lsi",
					StartAtNodeBoot: util.Pointer(false),
					Tablet:          util.Pointer(true),
				}, config)
			}},
		{name: `Update guest`,
			test: func(t *testing.T) {
				config := pveSDK.ConfigQemu{
					ID:     util.Pointer(pveSDK.GuestID(guestID)),
					CPU:    &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:   util.Pointer(pveSDK.GuestName(guestName)),
					Node:   util.Pointer(node),
					EfiDisk: &pveSDK.EfiDisk{
						PreEnrolledKeys: util.Pointer(true),
						Format:          util.Pointer(pveSDK.QemuDiskFormat("raw")),
						Storage:         util.Pointer(pveSDK.StorageName(test.GuestStorage)),
					},
				}
				_, err = config.Update(ctx, true, vmr, cl)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
				raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, cl)
				require.NoError(t, err)
				require.NotNil(t, raw)
				config, err := raw.Get(*vmr)
				require.NoError(t, err)
				require.NotNil(t, config)
				config.Smbios1 = ""                           // This field is unique and cannot be predicted, so we ignore it for the comparison.
				config.QemuUnusedDisks = pveSDK.QemuDevices{} // TODO include this field in test once reworked
				require.EqualExportedValues(t, &pveSDK.ConfigQemu{
					Bios:        "seabios",
					Boot:        " ",
					CPU:         &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Description: util.Pointer(""),
					EfiDisk: &pveSDK.EfiDisk{
						Size:            1024,
						Format:          util.Pointer(pveSDK.QemuDiskFormat("raw")),
						PreEnrolledKeys: util.Pointer(true),
						Storage:         util.Pointer(pveSDK.StorageName(test.GuestStorage)),
					},
					Hotplug:         "network,disk,usb",
					ID:              util.Pointer(pveSDK.GuestID(guestID)),
					Memory:          &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:            util.Pointer(pveSDK.GuestName(guestName)),
					Node:            util.Pointer(node),
					Protection:      util.Pointer(false),
					QemuDisks:       pveSDK.QemuDevices{},
					QemuKVM:         util.Pointer(true),
					QemuOs:          "other",
					QemuUnusedDisks: pveSDK.QemuDevices{},
					QemuVga:         pveSDK.QemuDevice{},
					Scsihw:          "lsi",
					StartAtNodeBoot: util.Pointer(false),
					Tablet:          util.Pointer(true),
				}, config)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
				raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, cl)
				require.NoError(t, err)
				require.NotNil(t, raw)
				require.NoError(t, vmr.Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
