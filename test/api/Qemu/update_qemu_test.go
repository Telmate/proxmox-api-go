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
	c := cl.New()
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
				vmr, err = c.QemuGuest.Create(ctx, config)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, &pveSDK.ConfigQemu{
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
					Tags:            new(pveSDK.Tags),
				})
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
				err = c.QemuGuest.Update(ctx, *vmr, true, false, config)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, &pveSDK.ConfigQemu{
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
					Tags:            new(pveSDK.Tags),
				})
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				DeleteGuest(t, ctx, cl, guestID)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Qemu_Upate_Reduced_To_Max(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Update-Reduced-To-Max"
	const guestID = 1021
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var vmr *pveSDK.VmRef
	setReduced, expectedReduced := ReducedConfig(guestID, node, guestName)
	setMax, expectedMax := MaximumConfig(guestID, node, guestName)
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
				vmr, err = c.QemuGuest.Create(ctx, setReduced)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedReduced)
			}},
		{name: `Update guest`,
			test: func(t *testing.T) {
				err = c.QemuGuest.Update(ctx, *vmr, true, false, setMax)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedMax)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				DeleteGuest(t, ctx, cl, guestID)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Qemu_Upate_Max_To_Reduced(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Update-Max-To-Reduced"
	const guestID = 1022
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var vmr *pveSDK.VmRef
	setMax, expectedMax := MaximumConfig(guestID, node, guestName)
	setReduced, expectedReduced := ReducedConfig(guestID, node, guestName)
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
				vmr, err = c.QemuGuest.Create(ctx, setMax)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedMax)
			}},
		{name: `Update guest`,
			test: func(t *testing.T) {
				err = c.QemuGuest.Update(ctx, *vmr, true, false, setReduced)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedReduced)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				DeleteGuest(t, ctx, cl, guestID)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Qemu_Upate_Min_To_Reduced(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Update-Min-To-Reduced"
	const guestID = 1023
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var vmr *pveSDK.VmRef
	setMin, expectedMin := MinimumConfig(guestID, node, guestName)
	setReduced, expectedReduced := ReducedConfig(guestID, node, guestName)
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
				vmr, err = c.QemuGuest.Create(ctx, setMin)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedMin)
			}},
		{name: `Update guest`,
			test: func(t *testing.T) {
				err = c.QemuGuest.Update(ctx, *vmr, true, false, setReduced)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedReduced)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				DeleteGuest(t, ctx, cl, guestID)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Qemu_Upate_Min_To_Max(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Update-Min-To-Max"
	const guestID = 1024
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	var vmr *pveSDK.VmRef
	setMin, expectedMin := MinimumConfig(guestID, node, guestName)
	setMax, expectedMax := MaximumConfig(guestID, node, guestName)
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
				vmr, err = c.QemuGuest.Create(ctx, setMin)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedMin)
			}},
		{name: `Update guest`,
			test: func(t *testing.T) {
				err = c.QemuGuest.Update(ctx, *vmr, true, false, setMax)
				require.NoError(t, err)
			}},
		{name: `Check guest config after update`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, expectedMax)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				DeleteGuest(t, ctx, cl, guestID)
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
