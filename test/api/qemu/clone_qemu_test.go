package qemu

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/pad"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	api_test "github.com/Telmate/proxmox-api-go/test/api"
	"github.com/stretchr/testify/require"
)

func Test_Qemu_Clone_With_ID(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Clone-With-ID"
	const guestID = 1013
	ctx := context.Background()
	cl, err := api_test.NewClient()
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	cloneVmr := pveSDK.NewVmRef(test.QemuTemplateID)
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Ensure guest does not exist`,
			test: func(t *testing.T) {
				existed, err := c.Guest.Delete(ctx, *pveSDK.NewVmRef(pveSDK.GuestID(guestID)))
				require.NoError(t, err)
				require.False(t, existed)
			}},
		{name: `Clone guest`,
			test: func(t *testing.T) {
				vmr, err := c.QemuGuest.Clone(ctx, *cloneVmr, pveSDK.CloneQemuTarget{
					Linked: &pveSDK.CloneLinked{
						ID:   new(pveSDK.GuestID(guestID)),
						Node: node,
						Name: new(pveSDK.GuestName(guestName)),
					}})
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest existence`,
			test: func(t *testing.T) {
				raws, err := c.Guest.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, raws)
				guestMap := raws.AsMap()
				_, ok := guestMap[guestID]
				require.True(t, ok)
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				c.Guest.Delete(ctx, *pveSDK.NewVmRef(guestID))
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}

// Create 5 guests at the same time. forcing a race condition on the API.
// This is to ensure that the API client can handle such a situation gracefully.
func Test_Qemu_Clone_Without_ID(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Clone-Without-ID"
	ctx := context.Background()
	cl, err := api_test.NewClient()
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	cloneVmr := pveSDK.NewVmRef(test.QemuTemplateID)
	var vmr *pveSDK.VmRef
	var previousVmrS []*pveSDK.VmRef
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Find previously created guests`,
			test: func(t *testing.T) {
				rawGuests, err := c.Guest.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, rawGuests)
				guests := rawGuests.AsArray()
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
					_, err := c.Guest.Delete(ctx, *previousVmrS[i])
					require.NoError(t, err)
				}
				previousVmrS = nil
			}},
		{name: `Clone guests`,
			test: func(t *testing.T) {
				vmr, err = c.QemuGuest.Clone(ctx, *cloneVmr, pveSDK.CloneQemuTarget{
					Linked: &pveSDK.CloneLinked{
						Node: node,
						Name: new(pveSDK.GuestName(guestName))}})
				require.NoError(t, err)
				require.NotNil(t, vmr)
			}},
		{name: `Check guest existence`,
			test: func(t *testing.T) {
				raws, err := c.Guest.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, raws)
				guestMap := raws.AsMap()
				_, ok := guestMap[vmr.VmId()]
				require.True(t, ok)
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

// Create 5 guests at the same time. forcing a race condition on the API.
// This is to ensure that the API client can handle such a situation gracefully.
func Test_Qemu_Clone_Client_Race(t *testing.T) {
	t.Parallel()
	const node = pveSDK.NodeName(test.FirstNode)
	const guestName = "Test-Qemu-Clone-Client-Race"
	guestsAmount := test.GuestRaceAmount
	ctx := context.Background()
	cl, err := api_test.NewClient()
	require.NoError(t, err)
	require.NoError(t, cl.Login(ctx, test.UserID, test.Password, ""))
	c := cl.New()
	vmrS := make([]*pveSDK.VmRef, guestsAmount)
	var previousVmrS []*pveSDK.VmRef
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{name: `Find previously created guests`,
			test: func(t *testing.T) {
				rawGuests, err := c.Guest.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, rawGuests)
				guests := rawGuests.AsArray()
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
					_, err := c.Guest.Delete(ctx, *previousVmrS[i])
					require.NoError(t, err)
				}
				previousVmrS = nil
			}},
		{name: `Clone guests`,
			test: func(t *testing.T) {

				var wg sync.WaitGroup
				errCh := make(chan error, guestsAmount)

				for i := range guestsAmount {
					wg.Add(1)
					go func(i int) {
						defer wg.Done()

						cl, err := api_test.NewClient()
						if err != nil {
							errCh <- err
							return
						}
						if err := cl.Login(ctx, test.UserID, test.Password, ""); err != nil {
							errCh <- err
							return
						}

						// TODO we should create the template during the test
						cloneVmr := pveSDK.NewVmRef(test.QemuTemplateID)
						cloneVmr.SetNode(string(node))
						vmr, err := c.QemuGuest.Clone(ctx, *cloneVmr, pveSDK.CloneQemuTarget{
							Linked: &pveSDK.CloneLinked{
								Node: node,
								Name: util.Pointer(pveSDK.GuestName(guestName)),
							}})
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
				time.Sleep(1 * time.Second) // wait for Proxmox to catch up
			}},
		{name: `Check guest existence`,
			test: func(t *testing.T) {
				raws, err := c.Guest.List(ctx)
				require.NoError(t, err)
				require.NotNil(t, raws)
				guestMap := raws.AsMap()
				for i := range vmrS {
					_, exists := guestMap[vmrS[i].VmId()]
					require.True(t, exists, "guest does not exist")
				}
			}},
		{name: `Delete guest`,
			test: func(t *testing.T) {
				for i := range vmrS {
					existed, err := c.Guest.Delete(ctx, *vmrS[i])
					require.NoError(t, err)
					require.True(t, existed)
				}
			}},
	}
	for i, test := range tests {
		t.Run(pad.Left(strconv.Itoa(i), 2, '0')+" "+test.name, test.test)
	}
}
