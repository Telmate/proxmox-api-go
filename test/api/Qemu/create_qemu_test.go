package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	"github.com/stretchr/testify/require"
)

// Create 5 guests at the same time. forcing a race condition on the API.
// This is to ensure that the API client can handle such a situation gracefully.
func Test_Qemu_Create_Client_Race(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Create-Client-Race"
	const guestsAmount = 5
	ctx := context.Background()
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	vmrS := make([]*pveSDK.VmRef, guestsAmount)
	var previousVmrS []*pveSDK.VmRef
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Find previously created guests`,
			test: func(t *testing.T) {
				guests, err := pveSDK.ListGuests(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, guests)
				for i := range guests {
					if guests[i].GetName() == guestName {
						previousVmrS = append(previousVmrS, pveSDK.NewVmRef(guests[i].GetID()))
					}
				}
				require.Len(t, previousVmrS, 0)
			}},
		{name: `Delete previously created guests`,
			test: func(t *testing.T) {
				for i := range previousVmrS {
					require.NoError(t, previousVmrS[i].Delete(ctx, cl))
				}
				previousVmrS = nil
			}},
		{name: `Create guests`,
			test: func(t *testing.T) {
				config := pveSDK.ConfigQemu{
					CPU:    &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:   util.Pointer(pveSDK.GuestName(guestName)),
					Node:   util.Pointer(node),
				}

				var wg sync.WaitGroup
				errCh := make(chan error, guestsAmount)

				for i := range guestsAmount {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()

						cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
						if err != nil {
							errCh <- err
							return
						}
						if err := cl.Login(ctx, test.UserID, test.Password, ""); err != nil {
							errCh <- err
							return
						}
						vmr, err := cl.New().QemuGuest.Create(ctx, config)
						if err != nil {
							errCh <- err
							return
						}
						if vmr == nil {
							errCh <- fmt.Errorf("nil vmr for index %d", i)
							return
						}
						vmrS[i] = vmr
					}(i)
				}

				wg.Wait()
				close(errCh)

				for err := range errCh {
					require.NoError(t, err)
				}
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				for i := range vmrS {
					require.NoError(t, vmrS[i].Delete(ctx, cl))
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

// TODO add tests for the efi disks
func Test_Qemu_Create_Max(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Create-Max"
	const guestID = 1010
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
				vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
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
				}
				vmr, err = c.QemuGuest.Create(ctx, config)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guestID, &pveSDK.ConfigQemu{
					CPU:             &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					Description:     util.Pointer(""),
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
					Bios:            "seabios",
					Boot:            " ",
					Tags:            new(pveSDK.Tags),
				})
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				require.NoError(t, vmr.Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

func Test_Qemu_Create_Disk_Minimal_Size(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Create-Disk-Minimal-Size"
	const guestID = 1011
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
				vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
				require.Error(t, vmr.Delete(ctx, cl))
			}},
		{name: `Create guest`,
			test: func(t *testing.T) {
				config := pveSDK.ConfigQemu{
					ID:  util.Pointer(pveSDK.GuestID(guestID)),
					CPU: &pveSDK.QemuCPU{Cores: util.Pointer(pveSDK.QemuCpuCores(1))},
					EfiDisk: &pveSDK.EfiDisk{
						Type:    util.Pointer(pveSDK.EfiDiskType4M),
						Storage: util.Pointer(pveSDK.StorageName(test.GuestStorage))},
					Memory: &pveSDK.QemuMemory{CapacityMiB: util.Pointer(pveSDK.QemuMemoryCapacity(16))},
					Name:   util.Pointer(pveSDK.GuestName(guestName)),
					Node:   util.Pointer(node),
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
						PreEnrolledKeys: util.Pointer(false),
						Type:            util.Pointer(pveSDK.EfiDiskType4M),
						Format:          util.Pointer(pveSDK.QemuDiskFormat_Raw),
						Size:            pveSDK.EfiDiskSize(1024),
						Storage:         util.Pointer(pveSDK.StorageName(test.GuestStorage))},
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
				require.NoError(t, vmr.Delete(ctx, cl))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
