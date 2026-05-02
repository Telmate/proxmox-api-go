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

func Test_Lxc_Create_Minimal_Unprivileged(t *testing.T) {
	t.Parallel()
	const guest = pveSDK.GuestID(1000)
	const node = pveSDK.NodeName(test.FirstNode)
	const storage = pveSDK.StorageName(test.GuestStorage)
	const name = pveSDK.GuestName("Test-Lxc-Create-Minimal-Unprivileged")
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	set, expected := MinimumConfig(guest, node, storage, new(false), name)
	var vmr *pveSDK.VmRef
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
				vmr, err = set.Create(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guest, expected)
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

func Test_Lxc_Create_Minimal_Privileged(t *testing.T) {
	t.Parallel()
	const guest = pveSDK.GuestID(1001)
	const node = pveSDK.NodeName(test.FirstNode)
	const storage = pveSDK.StorageName(test.GuestStorage)
	const name = pveSDK.GuestName("Test-Lxc-Create-Minimal-Privileged")
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	set, expected := MinimumConfig(guest, node, storage, new(true), name)
	var vmr *pveSDK.VmRef
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
				vmr, err = set.Create(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guest, expected)
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

func Test_Lxc_Create_Minimal_Privileged_Unset(t *testing.T) {
	t.Parallel()
	const guest = pveSDK.GuestID(1002)
	const node = pveSDK.NodeName(test.FirstNode)
	const storage = pveSDK.StorageName(test.GuestStorage)
	const name = pveSDK.GuestName("Test-Lxc-Create-Minimal-Privileged-Unset")
	cl, err := pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	set, expected := MinimumConfig(guest, node, storage, nil, name)
	var vmr *pveSDK.VmRef
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
				vmr, err = set.Create(ctx, cl)
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest config`,
			test: func(t *testing.T) {
				CheckConfig(t, ctx, cl, guest, expected)
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

// Create 5 guests at the same time. forcing a race condition on the API.
// This is to ensure that the API client can handle such a situation gracefully.
func Test_Lxc_Create_Client_Race(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Lxc-Create-Client-Race"
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
				config := pveSDK.ConfigLXC{
					BootMount: &pveSDK.LxcBootMount{
						SizeInKibibytes: util.Pointer(pveSDK.LxcMountSize(128 * 1024)),
						Storage:         util.Pointer(string(test.GuestStorage))},
					CreateOptions: &pveSDK.LxcCreateOptions{
						OsTemplate: &pveSDK.LxcTemplate{
							Storage: test.TemplateStorage,
							File:    test.DownloadedLXCTemplate,
						}},
					Name: util.Pointer(pveSDK.GuestName(guestName)),
					Node: util.Pointer(node),
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
						vmr, err := config.Create(ctx, cl)
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
