package proxmox

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_snapshot"
	"github.com/stretchr/testify/require"
)

func Test_snapshotClient_CreateLxc(t *testing.T) {
	t.Parallel()
	const UPID = "UPID:testNode:0006E4CB:17C8E729:6972A08C:qmsnapshot:198:root@pam:"
	tests := []struct {
		name         string
		guest        VmRef
		snapshotName SnapshotName
		description  string
		requests     []mockServer.Request
		err          error
	}{
		{name: `Create minimal`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/testNode/lxc/100/snapshot",
					map[string]any{
						"snapname": "snap1"},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Create maximal`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			description:  "This is a test snapshot" + body.Symbols,
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/testNode/lxc/100/snapshot",
					map[string]any{
						"snapname":    "snap1",
						"description": "This is a test snapshot" + body.Symbols},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Incomplete VmRef`,
			guest:        VmRef{vmId: 100},
			snapshotName: SnapshotName("mySnap"),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsPostResponse("/nodes/myNode/lxc/100/snapshot",
					map[string]any{
						"snapname": "mySnap"},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Validate error`,
			err: errors.New(SnapshotName_Error_MinLength)},
		{name: `Error Listing VMs`,
			guest:        VmRef{vmId: 100},
			snapshotName: SnapshotName("mySnap"),
			requests:     mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
		{name: `Error Creating snapshot`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			requests:     mockServer.RequestsError("/nodes/testNode/lxc/100/snapshot", mockServer.POST, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Snapshot.CreateLxc(context.Background(), test.guest, test.snapshotName, test.description)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_CreateQemu(t *testing.T) {
	t.Parallel()
	const UPID = "UPID:testNode:0006E4CB:17C8E729:6972A08C:qmsnapshot:198:root@pam:"
	tests := []struct {
		name         string
		guest        VmRef
		snapshotName SnapshotName
		description  string
		vmState      bool
		requests     []mockServer.Request
		err          error
	}{
		{name: `Create minimal`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/testNode/qemu/100/snapshot",
					map[string]any{
						"snapname": "snap1"},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Create maximal`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			description:  "This is a test snapshot" + body.Symbols,
			vmState:      true,
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/testNode/qemu/100/snapshot",
					map[string]any{
						"snapname":    "snap1",
						"description": "This is a test snapshot" + body.Symbols,
						"vmstate":     "1"},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Incomplete VmRef`,
			guest:        VmRef{vmId: 100},
			snapshotName: SnapshotName("mySnap"),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"},
				}}),
				mockServer.RequestsPostResponse("/nodes/myNode/qemu/100/snapshot",
					map[string]any{
						"snapname": "mySnap"},
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Validate error`,
			err: errors.New(SnapshotName_Error_MinLength)},
		{name: `Error Listing VMs`,
			guest:        VmRef{vmId: 100},
			snapshotName: SnapshotName("mySnap"),
			requests:     mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
		{name: `Error Creating snapshot`,
			guest:        VmRef{vmId: 100, node: "testNode"},
			snapshotName: SnapshotName("snap1"),
			requests:     mockServer.RequestsError("/nodes/testNode/qemu/100/snapshot", mockServer.POST, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Snapshot.CreateQemu(context.Background(), test.guest, test.snapshotName, test.description, test.vmState)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_Delete(t *testing.T) {
	t.Parallel()
	const UPID = "UPID:testNode:0006E4CB:17C8E729:6972A08C:qmsnapshot:198:root@pam:"
	tests := []struct {
		name         string
		guest        VmRef
		exists       bool
		snapshotName SnapshotName
		requests     []mockServer.Request
		err          error
	}{
		{name: `Delete existing`,
			guest:        VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapshotName: SnapshotName("snap1"),
			exists:       true,
			requests: mockServer.Append(
				mockServer.RequestsDeleteResponse("/nodes/testNode/qemu/100/snapshot/snap1", nil,
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Delete existing without node and vmType in VmRef`,
			guest:        VmRef{vmId: 200},
			snapshotName: SnapshotName("snap1"),
			exists:       true,
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"},
				}}),
				mockServer.RequestsDeleteResponse("/nodes/testNode/lxc/200/snapshot/snap1", nil,
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `Delete non-existing`,
			guest:        VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapshotName: SnapshotName("snap1"),
			exists:       false,
			requests: mockServer.Append(
				mockServer.RequestsDeleteResponse("/nodes/testNode/qemu/100/snapshot/snap1", nil,
					[]byte(`{"data":"`+UPID+`"}`)),
				mockServer.RequestsGetJson("/nodes/testNode/tasks/"+UPID+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string(`snapshot 'snap1' does not exist`)}}),
			)},
		{name: `Validate error`,
			err: errors.New(SnapshotName_Error_MinLength)},
		{name: `Error Listing VMs`,
			guest:        VmRef{vmId: 100},
			snapshotName: SnapshotName("mySnap"),
			requests:     mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
		{name: `Error Deleteing snapshot`,
			guest:        VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapshotName: SnapshotName("snap2"),
			requests:     mockServer.RequestsError("/nodes/testNode/qemu/100/snapshot/snap2", mockServer.DELETE, 500, 3),
			err:          errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			deleted, err := c.New().Snapshot.Delete(context.Background(), test.guest, test.snapshotName)
			require.Equal(t, test.err, err)
			require.Equal(t, test.exists, deleted)
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_List(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		guest     VmRef
		snapshots []SnapshotInfo
		requests  []mockServer.Request
		err       error
	}{
		{name: `List empty`,
			guest: VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapshots: []SnapshotInfo{
				{Name: "current",
					Description: "You are here!"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/testNode/qemu/100/snapshot", map[string]any{
					"data": []any{
						map[string]any{"name": "current",
							"description": "You are here!"},
					}}),
			)},
		{name: `List minimal VmRef`,
			guest: VmRef{vmId: 200},
			snapshots: []SnapshotInfo{
				{Name: "current",
					Description: "You are here!"}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsGetJson("/nodes/testNode/qemu/200/snapshot", map[string]any{
					"data": []any{
						map[string]any{"name": "current",
							"description": "You are here!"},
					}}),
			)},
		{name: `List multiple Lxc`,
			guest: VmRef{vmId: 100, node: "testNode", vmType: GuestLxc},
			snapshots: []SnapshotInfo{
				{Name: "snap1",
					Time: util.Pointer(time.Unix(1700000000, 0))},
				{Name: "snap2",
					Description: "This is snap2",
					Parent:      util.Pointer(SnapshotName("snap1")),
					Time:        util.Pointer(time.Unix(1700000100, 0))},
				{Name: "current",
					Description: "You are here!",
					Parent:      util.Pointer(SnapshotName("snap2"))}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/testNode/lxc/100/snapshot", map[string]any{
					"data": []any{
						map[string]any{"name": "snap1",
							"snaptime": float64(1700000000)},
						map[string]any{"name": "snap2",
							"description": "This is snap2",
							"parent":      "snap1",
							"snaptime":    float64(1700000100)},
						map[string]any{"name": "current",
							"description": "You are here!",
							"parent":      "snap2"},
					}}),
			)},
		{name: `List multiple Qemu`,
			guest: VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapshots: []SnapshotInfo{
				{Name: "snap1",
					Time:    util.Pointer(time.Unix(1700000000, 0)),
					VmState: util.Pointer(true)},
				{Name: "snap2",
					Description: "This is snap2",
					Parent:      util.Pointer(SnapshotName("snap1")),
					Time:        util.Pointer(time.Unix(1700000100, 0)),
					VmState:     util.Pointer(false)},
				{Name: "current",
					Description: "You are here!",
					Parent:      util.Pointer(SnapshotName("snap2"))}},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/testNode/qemu/100/snapshot", map[string]any{
					"data": []any{
						map[string]any{"name": "snap1",
							"snaptime": float64(1700000000),
							"vmstate":  float64(1)},
						map[string]any{"name": "snap2",
							"description": "This is snap2",
							"parent":      "snap1",
							"snaptime":    float64(1700000100),
							"vmstate":     float64(0)},
						map[string]any{"name": "current",
							"description": "You are here!",
							"parent":      "snap2"},
					}}),
			)},
		{name: `Error Listing VMs`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestLxc},
			requests: mockServer.RequestsError("/nodes/testNode/lxc/100/snapshot", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Error getting VmRef`,
			guest: VmRef{vmId: 200},
			snapshots: []SnapshotInfo{
				{Name: "current",
					Description: "You are here!"}},
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			rawSnapshots, err := c.New().Snapshot.List(context.Background(), test.guest)
			require.Equal(t, test.err, err)
			if test.err != nil {
				require.Nil(t, rawSnapshots)
			} else {
				snapshotMap := rawSnapshots.AsMap()
				for _, snapshot := range test.snapshots {
					raw, exists := snapshotMap[snapshot.Name]
					require.True(t, exists, "snapshot %q not found in list", snapshot.Name)
					require.NotNil(t, raw)
					require.Equal(t, snapshot, raw.Get(), "snapshot %q does not match", snapshot.Name)
				}
			}
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_ReadLxc(t *testing.T) {
	t.Parallel()
	baseConfig := func(c ConfigLXC) *ConfigLXC {
		if c.Memory == nil {
			c.Memory = util.Pointer(LxcMemory(0))
		}
		if c.Networks == nil {
			c.Networks = LxcNetworks{}
		}
		if c.Privileged == nil {
			c.Privileged = util.Pointer(true)
		}
		if c.Protection == nil {
			c.Protection = util.Pointer(false)
		}
		if c.StartAtNodeBoot == nil {
			c.StartAtNodeBoot = util.Pointer(false)
		}
		if c.Swap == nil {
			c.Swap = util.Pointer(LxcSwap(0))
		}
		if c.State == nil {
			c.State = util.Pointer(PowerStateRunning)
		}
		return &c
	}
	tests := []struct {
		name     string
		guest    VmRef
		snapName SnapshotName
		config   *ConfigLXC
		requests []mockServer.Request
		err      error
	}{
		{name: `Read LXC snapshot`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestLxc},
			snapName: SnapshotName("snap1"),
			config: baseConfig(ConfigLXC{
				ID:   util.Pointer(GuestID(100)),
				Name: util.Pointer(GuestName("lxc-name")),
				Node: util.Pointer(NodeName("testNode")),
			}),
			requests: mockServer.RequestsGetJson("/nodes/testNode/lxc/100/snapshot/snap1/config", map[string]any{
				"data": map[string]any{
					"hostname": "lxc-name"},
			})},
		{name: `Read LXC snapshot, minimal VmRef`,
			guest:    VmRef{vmId: 300},
			snapName: SnapshotName("snap1"),
			config: baseConfig(ConfigLXC{
				ID:   util.Pointer(GuestID(300)),
				Name: util.Pointer(GuestName("lxc-name")),
				Node: util.Pointer(NodeName("")),
			}),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsGetJson("/nodes/myNode/lxc/300/snapshot/snap1/config", map[string]any{
					"data": map[string]any{
						"hostname": "lxc-name",
					},
				}))},
		{name: `Validate error`,
			snapName: SnapshotName(""),
			err:      errors.New(SnapshotName_Error_MinLength)},
		{name: `Error during getting VmRef`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Error during config read`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestLxc},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/nodes/testNode/lxc/100/snapshot/snap1/config", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Snapshot.ReadLxc(context.Background(), test.guest, test.snapName)
			require.Equal(t, test.err, err)
			if err == nil {
				require.Equal(t, test.config, raw.Get(test.guest, PowerStateRunning))
			}
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_ReadQemu(t *testing.T) {
	t.Parallel()
	baseConfig := func(c ConfigQemu) *ConfigQemu {
		if c.Bios == "" {
			c.Bios = "seabios"
		}
		if c.Boot == "" {
			c.Boot = "cdn"
		}
		if c.CPU == nil {
			c.CPU = &QemuCPU{}
		}
		if c.Description == nil {
			c.Description = util.Pointer("")
		}
		if c.EFIDisk == nil {
			c.EFIDisk = QemuDevice{}
		}
		if c.Hotplug == "" {
			c.Hotplug = "network,disk,usb"
		}
		if c.Memory == nil {
			c.Memory = &QemuMemory{}
		}
		if c.Protection == nil {
			c.Protection = util.Pointer(false)
		}
		if c.QemuDisks == nil {
			c.QemuDisks = QemuDevices{}
		}
		if c.QemuKVM == nil {
			c.QemuKVM = util.Pointer(true)
		}
		if c.QemuOs == "" {
			c.QemuOs = "other"
		}
		if c.QemuUnusedDisks == nil {
			c.QemuUnusedDisks = QemuDevices{}
		}
		if c.QemuVga == nil {
			c.QemuVga = QemuDevice{}
		}
		if c.Scsihw == "" {
			c.Scsihw = "lsi"
		}
		if c.StartAtNodeBoot == nil {
			c.StartAtNodeBoot = util.Pointer(false)
		}
		if c.Tablet == nil {
			c.Tablet = util.Pointer(true)
		}
		return &c
	}
	tests := []struct {
		name     string
		guest    VmRef
		snapName SnapshotName
		config   *ConfigQemu
		requests []mockServer.Request
		err      error
	}{
		{name: `Read Qemu snapshot`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			config: baseConfig(ConfigQemu{
				ID:   util.Pointer(GuestID(100)),
				Name: util.Pointer(GuestName("test-qemu")),
				Node: util.Pointer(NodeName("testNode")),
			}),
			requests: mockServer.RequestsGetJson("/nodes/testNode/qemu/100/snapshot/snap1/config", map[string]any{
				"data": map[string]any{
					"name": "test-qemu"},
			})},
		{name: `Read Qemu snapshot, minimal VmRef`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			config: baseConfig(ConfigQemu{
				ID:   util.Pointer(GuestID(200)),
				Name: util.Pointer(GuestName("test-qemu")),
			}),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsGetJson("/nodes/testNode/qemu/200/snapshot/snap1/config", map[string]any{
					"data": map[string]any{
						"name": "test-qemu",
					},
				}))},
		{name: `Validate error`,
			snapName: SnapshotName(""),
			err:      errors.New(SnapshotName_Error_MinLength)},
		{name: `Error during getting VmRef`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Error during config read`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/nodes/testNode/qemu/100/snapshot/snap1/config", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			raw, err := c.New().Snapshot.ReadQemu(context.Background(), test.guest, test.snapName)
			require.Equal(t, test.err, err)
			if err == nil {
				config, err := raw.Get(&test.guest)
				require.NoError(t, err)
				require.Equal(t, test.config, config)
			}
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_Rollback(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		guest    VmRef
		snapName SnapshotName
		start    bool
		requests []mockServer.Request
		err      error
	}{
		{name: `Rolback false`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsPost("/nodes/testNode/qemu/100/snapshot/snap1/rollback", nil)},
		{name: `Rolback true`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			start:    true,
			requests: mockServer.RequestsPost("/nodes/testNode/qemu/100/snapshot/snap1/rollback", map[string]any{"start": "1"})},
		{name: `No VmRef node and vmType`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsPost("/nodes/testNode/qemu/200/snapshot/snap1/rollback", nil),
			)},
		{name: `Validate error`,
			snapName: SnapshotName(""),
			err:      errors.New(SnapshotName_Error_MinLength)},
		{name: `Error during getting VmRef`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Error during Rollback`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/nodes/testNode/qemu/100/snapshot/snap1/rollback", mockServer.POST, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Snapshot.Rollback(context.Background(), test.guest, test.snapName, test.start)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_snapshotClient_Update(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		guest       VmRef
		snapName    SnapshotName
		description string
		requests    []mockServer.Request
		err         error
	}{
		{name: `Update description`,
			guest:       VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName:    SnapshotName("snap1"),
			description: "upate description" + body.Symbols,
			requests: mockServer.RequestsPut("/nodes/testNode/qemu/100/snapshot/snap1/config",
				map[string]any{"description": "upate description" + body.Symbols})},
		{name: `Update empty description`,
			guest:    VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsPut("/nodes/testNode/qemu/100/snapshot/snap1/config",
				map[string]any{"description": ""})},
		{name: `Minimal VmRef`,
			guest:       VmRef{vmId: 200},
			snapName:    SnapshotName("snap1"),
			description: "upate description",
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "lxc"},
					map[string]any{"vmid": float64(200), "node": "testNode", "type": "qemu"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "lxc"},
				}}),
				mockServer.RequestsPut("/nodes/testNode/qemu/200/snapshot/snap1/config",
					map[string]any{"description": "upate description"}),
			)},
		{name: `Validate error`,
			snapName: SnapshotName(""),
			err:      errors.New(SnapshotName_Error_MinLength)},
		{name: `Error during getting VmRef`,
			guest:    VmRef{vmId: 200},
			snapName: SnapshotName("snap1"),
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `Error during Update`,
			guest:       VmRef{vmId: 100, node: "testNode", vmType: GuestQemu},
			snapName:    SnapshotName("snap1"),
			description: "upate description",
			requests:    mockServer.RequestsError("/nodes/testNode/qemu/100/snapshot/snap1/config", mockServer.PUT, 500, 3),
			err:         errors.New(mockServer.InternalServerError)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Snapshot.Update(context.Background(), test.guest, test.snapName, test.description)
			require.Equal(t, test.err, err)
			server.Clear(t)
		})
	}
}

func Test_ConfigSnapshot_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input ConfigSnapshot
		err   bool
	}{
		// Valid
		{name: "Valid ConfigSnapshot",
			input: ConfigSnapshot{Name: SnapshotName(test_data_snapshot.SnapshotName_Max_Legal())}},
		// Invalid
		{name: "Invalid ConfigSnapshot",
			input: ConfigSnapshot{Name: SnapshotName(test_data_snapshot.SnapshotName_Max_Illegal())},
			err:   true},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			if test.err {
				require.Error(t, test.input.Validate(), test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

func Test_RawSnapshots_AsArray(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  rawSnapshots
		output []SnapshotInfo
	}{
		{name: `missing name, should never happen`,
			input: rawSnapshots{
				a: []any{map[string]any{
					"description": "You are here!",
				}}},
			output: []SnapshotInfo{{Description: "You are here!"}}},
		{name: `only current snapshot`,
			input: rawSnapshots{
				a: []any{
					map[string]any{
						"name":        "current",
						"description": "You are here!"}}},
			output: []SnapshotInfo{
				{Name: "current", Description: "You are here!"}}},
		{name: `multiple snapshots`,
			input: rawSnapshots{
				a: []any{
					map[string]any{
						"name":        "snap1",
						"snaptime":    float64(1700000000),
						"description": "First snapshot",
						"vmstate":     float64(1)},
					map[string]any{
						"name":     "snap2",
						"snaptime": float64(1700000100),
						"parent":   "snap1",
						"vmstate":  float64(0)},
					map[string]any{
						"name":        "current",
						"description": "You are here!",
						"parent":      "snap2"}}},
			output: []SnapshotInfo{
				{Name: "snap1",
					Time:        util.Pointer(time.Unix(1700000000, 0)),
					Description: "First snapshot", VmState: util.Pointer(true)},
				{Name: "snap2",
					Time:    util.Pointer(time.Unix(1700000100, 0)),
					VmState: util.Pointer(false),
					Parent:  util.Pointer(SnapshotName("snap1"))},
				{Name: "current",
					Description: "You are here!",
					Parent:      util.Pointer(SnapshotName("snap2"))}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			array := RawSnapshots(&test.input).AsArray()
			require.NotNil(t, array)
			for i, snap := range test.output {
				require.Equal(t, snap, array[i].Get())
			}
		})
	}
}

func test_RawSnapshotTree_Current_and_Root_data() []struct {
	name    string
	input   rawSnapshots
	current *Snapshot
	root    []*Snapshot
} {
	// Helpers
	root := func(roots []string, snapshots map[string]*Snapshot) []*Snapshot {
		rootSnapshots := make([]*Snapshot, len(roots))
		for i := range roots {
			rootSnapshots[i] = snapshots[roots[i]]
		}
		return rootSnapshots
	}
	current := func(snapshots map[string]*Snapshot) *Snapshot { return snapshots["current"] }

	// Datasets
	single_chain_of_5_snapshots := func() map[string]*Snapshot {
		// root1 → a → b → c → current
		snaps := map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"a":       {Name: "a", Time: util.Pointer(time.Unix(1700000100, 0)), Description: "This is snap2"},
			"b":       {Name: "b", Time: util.Pointer(time.Unix(1700000200, 0))},
			"c":       {Name: "c", Time: util.Pointer(time.Unix(1700000300, 0))},
			"current": {Name: "current", Description: "You are here!"}}
		snaps["root1"].Children = []*Snapshot{snaps["a"]}
		snaps["a"].Parent = snaps["root1"]
		snaps["a"].Children = []*Snapshot{snaps["b"]}
		snaps["b"].Parent = snaps["a"]
		snaps["b"].Children = []*Snapshot{snaps["c"]}
		snaps["c"].Parent = snaps["b"]
		snaps["c"].Children = []*Snapshot{snaps["current"]}
		snaps["current"].Parent = snaps["c"]
		return snaps
	}

	single_chain_of_25_snapshots := func() map[string]*Snapshot {
		// root1 → snap1 → snap2 → ... → snap23 → current
		snaps := map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"current": {Name: "current", Description: "You are here!"}}
		for i := 1; i <= 23; i++ {
			snapName := fmt.Sprintf("snap%d", i)
			snaps[snapName] = &Snapshot{Name: SnapshotName(snapName), Time: util.Pointer(time.Unix(1700000000+int64(i*100), 0))}
		}

		// Build relationships
		snaps["root1"].Children = []*Snapshot{snaps["snap1"]}
		snaps["snap1"].Parent = snaps["root1"]
		for i := 1; i <= len(snaps)-3; i++ {
			snapName := fmt.Sprintf("snap%d", i)
			nextSnapName := fmt.Sprintf("snap%d", i+1)
			snaps[snapName].Children = []*Snapshot{snaps[nextSnapName]}
			snaps[nextSnapName].Parent = snaps[snapName]
		}
		snaps["snap23"].Children = []*Snapshot{snaps["current"]}
		snaps["current"].Parent = snaps["snap23"]

		return snaps
	}

	wide_25_snapshots := func() map[string]*Snapshot {
		// root1
		// ├─ snap1
		// ├─ snap2
		// ├─ snap3
		// ...
		// ├─ snap24
		// └─ current
		snaps := map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"current": {Name: "current", Description: "You are here!"}}
		for i := 1; i <= 3; i++ {
			snapName := fmt.Sprintf("snap%d", i)
			snaps[snapName] = &Snapshot{Name: SnapshotName(snapName), Time: util.Pointer(time.Unix(1700000000+int64(i*100), 0))}
		}

		// Build relationships
		children := make([]*Snapshot, 0, len(snaps)-2)
		for i := 1; i <= len(snaps)-2; i++ {
			snapName := fmt.Sprintf("snap%d", i)
			children = append(children, snaps[snapName])
			snaps[snapName].Parent = snaps["root1"]
		}
		children = append(children, snaps["current"])
		snaps["current"].Parent = snaps["root1"]
		snaps["root1"].Children = children

		return snaps
	}

	multiple_root_snapshots_trees := func() map[string]*Snapshot {
		// root1
		// ├─ 1a
		// │  ├─ 1aa
		// │  │  ├─ 1aaa
		// │  │  ├─ 1aab
		// │  │  │  └─ 1aaba
		// │  │  └─ 1aac
		// │  │     └─ current
		// │  ├─ 1ab
		// │  │  └─ 1aba
		// │  └─ 1ac
		// ├─ 1b
		// │  ├─ 1ba
		// │  │  └─ 1bab
		// │  ├─ 1bb
		// │  │  └─ 1bba
		// │  └─ 1bc
		// ├─ 1b
		// │  └─ 1ba
		// └─ 1c
		// root2
		// ├─ 2a
		// │  └─ 2aa
		// └─ 2c
		// root3
		// └─ 3a

		snaps := map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"1a":      {Name: "1a", Time: util.Pointer(time.Unix(1700000100, 0))},
			"1aa":     {Name: "1aa", Time: util.Pointer(time.Unix(1700000200, 0))},
			"1aaa":    {Name: "1aaa", Time: util.Pointer(time.Unix(1700000300, 0))},
			"1aab":    {Name: "1aab", Time: util.Pointer(time.Unix(1700000400, 0))},
			"1aaba":   {Name: "1aaba", Time: util.Pointer(time.Unix(1700000500, 0))},
			"1aac":    {Name: "1aac", Time: util.Pointer(time.Unix(1700000600, 0))},
			"1ab":     {Name: "1ab", Time: util.Pointer(time.Unix(1700000700, 0)), VmState: util.Pointer(true)},
			"1aba":    {Name: "1aba", Time: util.Pointer(time.Unix(1700000800, 0))},
			"1ac":     {Name: "1ac", Time: util.Pointer(time.Unix(1700000900, 0))},
			"1b":      {Name: "1b", Time: util.Pointer(time.Unix(1700001000, 0))},
			"1ba":     {Name: "1ba", Time: util.Pointer(time.Unix(1700001100, 0))},
			"1bab":    {Name: "1bab", Time: util.Pointer(time.Unix(1700001200, 0))},
			"1bb":     {Name: "1bb", Time: util.Pointer(time.Unix(1700001300, 0))},
			"1bba":    {Name: "1bba", Time: util.Pointer(time.Unix(1700001400, 0))},
			"1bc":     {Name: "1bc", Time: util.Pointer(time.Unix(1700001500, 0))},
			"1c":      {Name: "1c", Time: util.Pointer(time.Unix(1700001600, 0))},
			"root2":   {Name: "root2", Time: util.Pointer(time.Unix(1700002000, 0)), VmState: util.Pointer(true)},
			"2a":      {Name: "2a", Time: util.Pointer(time.Unix(1700002100, 0))},
			"2aa":     {Name: "2aa", Time: util.Pointer(time.Unix(1700002200, 0))},
			"2c":      {Name: "2c", Time: util.Pointer(time.Unix(1700002300, 0))},
			"root3":   {Name: "root3", Time: util.Pointer(time.Unix(1700003000, 0))},
			"3a":      {Name: "3a", Time: util.Pointer(time.Unix(1700003100, 0))},
			"current": {Name: "current", Description: "You are here!"}}
		// Build relationships
		snaps["root1"].Children = []*Snapshot{snaps["1a"], snaps["1b"], snaps["1c"]}
		snaps["1a"].Parent = snaps["root1"]
		snaps["1a"].Children = []*Snapshot{snaps["1aa"], snaps["1ab"], snaps["1ac"]}
		snaps["1aa"].Parent = snaps["1a"]
		snaps["1aa"].Children = []*Snapshot{snaps["1aaa"], snaps["1aab"], snaps["1aac"]}
		snaps["1aaa"].Parent = snaps["1aa"]
		snaps["1aab"].Parent = snaps["1aa"]
		snaps["1aab"].Children = []*Snapshot{snaps["1aaba"]}
		snaps["1aaba"].Parent = snaps["1aab"]
		snaps["1aac"].Parent = snaps["1aa"]
		snaps["1aac"].Children = []*Snapshot{snaps["current"]}
		snaps["current"].Parent = snaps["1aac"]
		snaps["1ab"].Parent = snaps["1a"]
		snaps["1ab"].Children = []*Snapshot{snaps["1aba"]}
		snaps["1aba"].Parent = snaps["1ab"]
		snaps["1ac"].Parent = snaps["1a"]
		snaps["1b"].Parent = snaps["root1"]
		snaps["1b"].Children = []*Snapshot{snaps["1ba"], snaps["1bb"], snaps["1bc"]}
		snaps["1ba"].Parent = snaps["1b"]
		snaps["1ba"].Children = []*Snapshot{snaps["1bab"]}
		snaps["1bab"].Parent = snaps["1ba"]
		snaps["1bb"].Parent = snaps["1b"]
		snaps["1bb"].Children = []*Snapshot{snaps["1bba"]}
		snaps["1bba"].Parent = snaps["1bb"]
		snaps["1bc"].Parent = snaps["1b"]
		snaps["1c"].Parent = snaps["root1"]

		snaps["root2"].Children = []*Snapshot{snaps["2a"], snaps["2c"]}
		snaps["2a"].Parent = snaps["root2"]
		snaps["2a"].Children = []*Snapshot{snaps["2aa"]}
		snaps["2aa"].Parent = snaps["2a"]
		snaps["2c"].Parent = snaps["root2"]

		snaps["root3"].Children = []*Snapshot{snaps["3a"]}
		snaps["3a"].Parent = snaps["root3"]

		return snaps
	}

	roots_with_same_time := func() map[string]*Snapshot {
		return map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"root2":   {Name: "root2", Time: util.Pointer(time.Unix(1700000000, 0))},
			"rOOt2":   {Name: "rOOt2", Time: util.Pointer(time.Unix(1700000000, 0))},
			"root222": {Name: "root222", Time: util.Pointer(time.Unix(1700000000, 0))},
			"rOot222": {Name: "rOot222", Time: util.Pointer(time.Unix(1700000000, 0))},
			"current": {Name: "current", Description: "You are here!"}}
	}

	children_with_same_time := func() map[string]*Snapshot {
		snaps := map[string]*Snapshot{
			"root1":   {Name: "root1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"snap1":   {Name: "snap1", Time: util.Pointer(time.Unix(1700000000, 0))},
			"snap2":   {Name: "snap2", Time: util.Pointer(time.Unix(1700000000, 0))},
			"sNAp2":   {Name: "sNAp2", Time: util.Pointer(time.Unix(1700000000, 0))},
			"snap222": {Name: "snap222", Time: util.Pointer(time.Unix(1700000000, 0))},
			"sNap222": {Name: "sNap222", Time: util.Pointer(time.Unix(1700000000, 0))},
			"current": {Name: "current", Description: "You are here!"}}
		snaps["current"].Parent = snaps["root1"]
		snaps["snap1"].Parent = snaps["root1"]
		snaps["snap2"].Parent = snaps["root1"]
		snaps["sNAp2"].Parent = snaps["root1"]
		snaps["snap222"].Parent = snaps["root1"]
		snaps["sNap222"].Parent = snaps["root1"]
		snaps["root1"].Children = []*Snapshot{
			snaps["sNAp2"],
			snaps["sNap222"],
			snaps["snap1"],
			snaps["snap2"],
			snaps["snap222"],
			snaps["current"],
		}

		return snaps
	}

	return []struct {
		name    string
		input   rawSnapshots
		current *Snapshot
		root    []*Snapshot
	}{
		{name: "Single current snapshot",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "current", "description": "You are here!"}}},
			current: &Snapshot{Name: "current", Description: "You are here!"},
			root: []*Snapshot{
				util.Pointer(Snapshot{
					Name:        "current",
					Description: "You are here!",
				})}},
		{name: "single chain of 5 snapshots",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "a", "snaptime": float64(1700000100), "parent": "root1", "description": "This is snap2"},
					map[string]any{"name": "b", "snaptime": float64(1700000200), "parent": "a"},
					map[string]any{"name": "c", "snaptime": float64(1700000300), "parent": "b"},
					map[string]any{"name": "current", "parent": "c", "description": "You are here!"},
				}},
			root:    root([]string{"root1"}, single_chain_of_5_snapshots()),
			current: current(single_chain_of_5_snapshots())},
		{name: "single chain of 25 snapshots",
			input: rawSnapshots{
				a: func() []any {
					snaps := []any{
						map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					}
					for i := 1; i <= 23; i++ {
						snapName := fmt.Sprintf("snap%d", i)
						snaps = append(snaps, map[string]any{
							"name":     snapName,
							"snaptime": float64(1700000000 + i*100),
							"parent": func() string {
								if i == 1 {
									return "root1"
								}
								return fmt.Sprintf("snap%d", i-1)
							}(),
						})
					}
					snaps = append(snaps, map[string]any{
						"name":        "current",
						"parent":      "snap23",
						"description": "You are here!",
					})
					return snaps
				}(),
			},
			root:    root([]string{"root1"}, single_chain_of_25_snapshots()),
			current: current(single_chain_of_25_snapshots())},
		{name: "wide 25 snapshots",
			input: rawSnapshots{
				a: func() []any {
					snaps := []any{
						map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					}
					for i := 1; i <= 3; i++ {
						snapName := fmt.Sprintf("snap%d", i)
						snaps = append(snaps, map[string]any{
							"name":     snapName,
							"snaptime": float64(1700000000 + i*100),
							"parent":   "root1",
						})
					}
					return append(snaps, map[string]any{
						"name":        "current",
						"parent":      "root1",
						"description": "You are here!",
					})
				}(),
			},
			root:    root([]string{"root1"}, wide_25_snapshots()),
			current: current(wide_25_snapshots())},
		{name: "Multiple root snapshots trees",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "1aaba", "snaptime": float64(1700000500), "parent": "1aab"},
					map[string]any{"name": "1aa", "snaptime": float64(1700000200), "parent": "1a"},
					map[string]any{"name": "1aab", "snaptime": float64(1700000400), "parent": "1aa"},
					map[string]any{"name": "1aaa", "snaptime": float64(1700000300), "parent": "1aa"},
					map[string]any{"name": "1bb", "snaptime": float64(1700001300), "parent": "1b"},
					map[string]any{"name": "1aac", "snaptime": float64(1700000600), "parent": "1aa"},
					map[string]any{"name": "1ac", "snaptime": float64(1700000900), "parent": "1a"},
					map[string]any{"name": "1ab", "snaptime": float64(1700000700), "parent": "1a", "vmstate": float64(1)},
					map[string]any{"name": "1aba", "snaptime": float64(1700000800), "parent": "1ab"},
					map[string]any{"name": "1bba", "snaptime": float64(1700001400), "parent": "1bb"},
					map[string]any{"name": "1ba", "snaptime": float64(1700001100), "parent": "1b"},
					map[string]any{"name": "1bab", "snaptime": float64(1700001200), "parent": "1ba"},
					map[string]any{"name": "root3", "snaptime": float64(1700003000)},
					map[string]any{"name": "1bc", "snaptime": float64(1700001500), "parent": "1b"},
					map[string]any{"name": "1b", "snaptime": float64(1700001000), "parent": "root1"},
					map[string]any{"name": "1a", "snaptime": float64(1700000100), "parent": "root1"},
					map[string]any{"name": "1c", "snaptime": float64(1700001600), "parent": "root1"},
					map[string]any{"name": "current", "parent": "1aac", "description": "You are here!"},
					map[string]any{"name": "root2", "snaptime": float64(1700002000), "vmstate": float64(1)},
					map[string]any{"name": "2a", "snaptime": float64(1700002100), "parent": "root2"},
					map[string]any{"name": "2aa", "snaptime": float64(1700002200), "parent": "2a"},
					map[string]any{"name": "2c", "snaptime": float64(1700002300), "parent": "root2"},
					map[string]any{"name": "3a", "snaptime": float64(1700003100), "parent": "root3"},
				}},
			root:    root([]string{"root1", "root2", "root3"}, multiple_root_snapshots_trees()),
			current: current(multiple_root_snapshots_trees())},
		{name: `Edgecase roots with same snaptime`,
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "current", "description": "You are here!"},
					map[string]any{"name": "root2", "snaptime": float64(1700000000)},
					map[string]any{"name": "root222", "snaptime": float64(1700000000)},
					map[string]any{"name": "rOot222", "snaptime": float64(1700000000)},
					map[string]any{"name": "rOOt2", "snaptime": float64(1700000000)},
				}},
			root:    root([]string{"rOOt2", "rOot222", "root1", "root2", "root222", "current"}, roots_with_same_time()),
			current: current(roots_with_same_time())},
		{name: `Edgecase children with same snaptime`,
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "sNAp2", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "current", "parent": "root1", "description": "You are here!"},
					map[string]any{"name": "sNap222", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap1", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap222", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap2", "parent": "root1", "snaptime": float64(1700000000)},
				}},
			root:    root([]string{"root1"}, children_with_same_time()),
			current: current(children_with_same_time())},
	}
}

func Test_RawSnapshotTree_Current_and_Root(t *testing.T) {
	t.Parallel()
	tests := test_RawSnapshotTree_Current_and_Root_data()
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			tree := RawSnapshots(&test.input).Tree()
			require.NotNil(t, tree)
			require.Equal(t, test.current, tree.Current())
			require.Equal(t, test.root, tree.Root())
		})
	}
}

func Benchmark_RawSnapshotTree_Current_and_Root(b *testing.B) {
	tests := test_RawSnapshotTree_Current_and_Root_data()
	b.ResetTimer()
	var result RawSnapshotTree
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			result = RawSnapshots(&test.input).Tree()
		}
	}
	if result == nil {
		b.Fatal("unexpected nil")
	}
}

func Test_RawSnapshotTree_Walk(t *testing.T) {
	t.Parallel()

	// Helper to collect snapshot names in order
	collect := func(tree RawSnapshotTree) []SnapshotName {
		var names []SnapshotName
		tree.Walk(func(s *Snapshot) bool {
			names = append(names, s.Name)
			return true
		})
		return names
	}

	// Helper to collect names until condition met
	collectUntil := func(tree RawSnapshotTree, stopAt SnapshotName) []SnapshotName {
		var names []SnapshotName
		tree.Walk(func(s *Snapshot) bool {
			names = append(names, s.Name)
			return s.Name != stopAt
		})
		return names
	}

	tests := []struct {
		name            string
		input           rawSnapshots
		expectedOrder   []SnapshotName
		stopAt          SnapshotName
		expectedStopped []SnapshotName
	}{
		{name: "Single current snapshot",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "current", "description": "You are here!"}}},
			expectedOrder: []SnapshotName{"current"}},
		{name: "Single chain of 5 snapshots",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "a", "snaptime": float64(1700000100), "parent": "root1"},
					map[string]any{"name": "b", "snaptime": float64(1700000200), "parent": "a"},
					map[string]any{"name": "c", "snaptime": float64(1700000300), "parent": "b"},
					map[string]any{"name": "current", "parent": "c", "description": "You are here!"},
				}},
			expectedOrder:   []SnapshotName{"root1", "a", "b", "c", "current"},
			stopAt:          "b",
			expectedStopped: []SnapshotName{"root1", "a", "b"}},
		{name: "Wide tree with multiple children",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap1", "snaptime": float64(1700000100), "parent": "root1"},
					map[string]any{"name": "snap2", "snaptime": float64(1700000200), "parent": "root1"},
					map[string]any{"name": "snap3", "snaptime": float64(1700000300), "parent": "root1"},
					map[string]any{"name": "current", "parent": "root1", "description": "You are here!"},
				}},
			expectedOrder:   []SnapshotName{"root1", "snap1", "snap2", "snap3", "current"},
			stopAt:          "snap2",
			expectedStopped: []SnapshotName{"root1", "snap1", "snap2"}},
		{name: "Multiple root snapshots trees",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "1a", "snaptime": float64(1700000100), "parent": "root1"},
					map[string]any{"name": "1aa", "snaptime": float64(1700000200), "parent": "1a"},
					map[string]any{"name": "1aaa", "snaptime": float64(1700000300), "parent": "1aa"},
					map[string]any{"name": "1aab", "snaptime": float64(1700000400), "parent": "1aa"},
					map[string]any{"name": "1aaba", "snaptime": float64(1700000500), "parent": "1aab"},
					map[string]any{"name": "1aac", "snaptime": float64(1700000600), "parent": "1aa"},
					map[string]any{"name": "1ab", "snaptime": float64(1700000700), "parent": "1a"},
					map[string]any{"name": "1aba", "snaptime": float64(1700000800), "parent": "1ab"},
					map[string]any{"name": "1ac", "snaptime": float64(1700000900), "parent": "1a"},
					map[string]any{"name": "1b", "snaptime": float64(1700001000), "parent": "root1"},
					map[string]any{"name": "1ba", "snaptime": float64(1700001100), "parent": "1b"},
					map[string]any{"name": "1bab", "snaptime": float64(1700001200), "parent": "1ba"},
					map[string]any{"name": "1bb", "snaptime": float64(1700001300), "parent": "1b"},
					map[string]any{"name": "1bba", "snaptime": float64(1700001400), "parent": "1bb"},
					map[string]any{"name": "1bc", "snaptime": float64(1700001500), "parent": "1b"},
					map[string]any{"name": "1c", "snaptime": float64(1700001600), "parent": "root1"},
					map[string]any{"name": "root2", "snaptime": float64(1700002000)},
					map[string]any{"name": "2a", "snaptime": float64(1700002100), "parent": "root2"},
					map[string]any{"name": "2aa", "snaptime": float64(1700002200), "parent": "2a"},
					map[string]any{"name": "2c", "snaptime": float64(1700002300), "parent": "root2"},
					map[string]any{"name": "root3", "snaptime": float64(1700003000)},
					map[string]any{"name": "3a", "snaptime": float64(1700003100), "parent": "root3"},
					map[string]any{"name": "current", "parent": "1aac", "description": "You are here!"},
				}},
			expectedOrder: []SnapshotName{
				// First root tree (depth-first, oldest to newest children)
				"root1", "1a", "1aa", "1aaa", "1aab", "1aaba", "1aac", "current", "1ab", "1aba", "1ac", "1b", "1ba", "1bab", "1bb", "1bba", "1bc", "1c",
				// Second root tree
				"root2", "2a", "2aa", "2c",
				// Third root tree
				"root3", "3a"},
			stopAt: "1ab",
			expectedStopped: []SnapshotName{
				"root1", "1a", "1aa", "1aaa", "1aab", "1aaba", "1aac", "current", "1ab"}},
		{name: "Stop at first snapshot",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "a", "snaptime": float64(1700000100), "parent": "root1"},
					map[string]any{"name": "b", "snaptime": float64(1700000200), "parent": "a"},
					map[string]any{"name": "current", "description": "You are here!", "parent": "a"},
				}},
			expectedOrder:   []SnapshotName{"root1", "a", "b", "current"},
			stopAt:          "root1",
			expectedStopped: []SnapshotName{"root1"}},
		{name: "Children with same snaptime ordered by name",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap1", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "snap2", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "sNAp2", "parent": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "current", "parent": "root1", "description": "You are here!"},
				}},
			expectedOrder: []SnapshotName{"root1", "sNAp2", "snap1", "snap2", "current"}},
		{name: "Multiple roots with same snaptime ordered by name",
			input: rawSnapshots{
				a: []any{
					map[string]any{"name": "root1", "snaptime": float64(1700000000)},
					map[string]any{"name": "root2", "snaptime": float64(1700000000)},
					map[string]any{"name": "rOOt2", "snaptime": float64(1700000000)},
					map[string]any{"name": "current", "description": "You are here!"},
				}},
			expectedOrder: []SnapshotName{"rOOt2", "root1", "root2", "current"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := RawSnapshots(&test.input).Tree()
			require.NotNil(t, tree)

			// Test full walk
			names := collect(tree)
			require.Equal(t, test.expectedOrder, names, "Walk should visit snapshots in depth-first order, oldest to newest")

			// Test early termination if stopAt is specified
			if test.stopAt != "" {
				stoppedNames := collectUntil(tree, test.stopAt)
				require.Equal(t, test.expectedStopped, stoppedNames, "Walk should stop early when function returns false")
			}
		})
	}

	// Test edge case: empty tree
	t.Run("Empty tree", func(t *testing.T) {
		emptyTree := &rawSnapshotTree{root: []*Snapshot{}, current: nil}
		var visited bool
		emptyTree.Walk(func(s *Snapshot) bool {
			visited = true
			return true
		})
		require.False(t, visited, "Walk should not visit any snapshots in an empty tree")
	})
}

func Test_SnapshotName_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input []string
		err   error
	}{
		// Valid
		{name: "Valid", input: test_data_snapshot.SnapshotName_Legal()},
		// Invalid
		{name: "Invalid SnapshotName_Error_MinLength",
			input: []string{"", test_data_snapshot.SnapshotName_Min_Illegal()},
			err:   errors.New(SnapshotName_Error_MinLength),
		},
		{name: "Invalid SnapshotName_Error_MaxLength",
			input: []string{test_data_snapshot.SnapshotName_Max_Illegal()},
			err:   errors.New(SnapshotName_Error_MaxLength),
		},
		{name: "Invalid SnapshotName_Error_StartNoLetter",
			input: test_data_snapshot.SnapshotName_Start_Illegal(),
			err:   errors.New(SnapshotName_Error_StartNoLetter),
		},
		{name: "Invalid SnapshotName_Error_StartNoLetter",
			input: test_data_snapshot.SnapshotName_Character_Illegal(),
			err:   errors.New(SnapshotName_Error_IllegalCharacters),
		},
	}
	for _, test := range tests {
		for _, snapshot := range test.input {
			t.Run(test.name+" :"+snapshot, func(*testing.T) {
				require.Equal(t, SnapshotName(snapshot).Validate(), test.err, test.name+" :"+snapshot)
			})
		}
	}
}
