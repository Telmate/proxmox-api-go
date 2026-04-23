package proxmox

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_QemuGuestInterface_Create(t *testing.T) {
	baseConfig := func(config ConfigQemu) ConfigQemu {
		if config.CPU == nil {
			config.CPU = &QemuCPU{Cores: util.Pointer(QemuCpuCores(1))}
		}
		if config.Memory == nil {
			config.Memory = &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(512))}
		}
		return config
	}
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		config   ConfigQemu
		vmr      VmRef
		requests []mockServer.Request
		err      error
	}{
		{name: `no power state`,
			vmr: VmRef{node: "pve3", vmId: 100, vmType: GuestQemu},
			config: baseConfig(ConfigQemu{
				ID:   util.Pointer(GuestID(100)),
				Node: util.Pointer(NodeName("pve3"))}),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsPostResponse("/nodes/pve3/qemu", map[string]any{
					"cores":  "1",
					"memory": "512",
					"vmid":   "100"},
					[]byte(`{"data":"`+UPID("pve3", "qmcreate", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmcreate", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsGetJson("/access/permissions", map[string]any{"data": map[string]any{}}))},
		{name: `started`,
			vmr: VmRef{node: "test", vmId: 2345, vmType: GuestQemu},
			config: baseConfig(ConfigQemu{
				ID:    util.Pointer(GuestID(2345)),
				Node:  util.Pointer(NodeName("test")),
				State: util.Pointer(PowerStateRunning)}),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsPostResponse("/nodes/test/qemu", map[string]any{
					"cores":  "1",
					"memory": "512",
					"start":  "1",
					"vmid":   "2345"},
					[]byte(`{"data":"`+UPID("test", "qmstart", GuestID(2345))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstart", GuestID(2345)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsPostResponse("/nodes/test/qemu/2345/status/start", map[string]any{},
					[]byte(`{"data":"`+UPID("test", "qmstart", GuestID(2345))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstart", GuestID(2345)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsGetJson("/access/permissions", map[string]any{"data": map[string]any{}}))},
		{name: `stopped`,
			vmr: VmRef{node: "pve-9l", vmId: 9000, vmType: GuestQemu},
			config: baseConfig(ConfigQemu{
				ID:    util.Pointer(GuestID(9000)),
				Node:  util.Pointer(NodeName("pve-9l")),
				State: util.Pointer(PowerStateStopped)}),
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsPostResponse("/nodes/pve-9l/qemu", map[string]any{
					"cores":  "1",
					"memory": "512",
					"vmid":   "9000"},
					[]byte(`{"data":"`+UPID("pve-9l", "qmcreate", GuestID(9000))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve-9l/tasks/"+mockServer.Path(UPID("pve-9l", "qmcreate", GuestID(9000)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsGetJson("/access/permissions", map[string]any{"data": map[string]any{}}))},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			vmr, err := c.New().QemuGuest.Create(context.Background(), test.config)
			require.Equal(t, test.err, err)
			require.Equal(t, test.vmr, *vmr)
			server.Clear(t)
			c.clearVersion()
		})
	}
}

func Test_QemuGuestInterface_Update(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name           string
		config         ConfigQemu
		vmr            VmRef
		allowRestart   bool
		allowForceStop bool
		requests       []mockServer.Request
		err            error
	}{
		{name: `stopped to running, allowRestart false`,
			config: ConfigQemu{State: util.Pointer(PowerStateRunning)},
			vmr:    VmRef{node: "pve", vmId: 100},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/config", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": 100, "node": "pve", "type": "qemu"}}}),
				mockServer.RequestsGetJson("/cluster/ha/resources/100", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/status/current", map[string]any{"data": map[string]any{"status": "stopped"}}),
				mockServer.RequestsPostResponse("/nodes/pve/qemu/100/status/start", nil, []byte(`{"data":"`+UPID("pve", "qmstart", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve/tasks/"+mockServer.Path(UPID("pve", "qmstart", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `running to stopped, allowRestart false`,
			config: ConfigQemu{State: util.Pointer(PowerStateStopped)},
			vmr:    VmRef{node: "pve2", vmId: 100},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/pve2/qemu/100/config", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": 100, "node": "pve2", "type": "qemu"}}}),
				mockServer.RequestsGetJson("/cluster/ha/resources/100", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsGetJson("/nodes/pve2/qemu/100/status/current", map[string]any{"data": map[string]any{"status": "running"}}),
				mockServer.RequestsPostResponse("/nodes/pve2/qemu/100/status/shutdown", nil, []byte(`{"data":"`+UPID("pve2", "qmshutdown", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve2/tasks/"+mockServer.Path(UPID("pve2", "qmshutdown", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `migrate efi disk, running, allowRestart true`,
			config: ConfigQemu{
				EfiDisk: &EfiDisk{
					Storage: util.Pointer(StorageName("local-lvm"))}},
			vmr:          VmRef{node: "pve", vmId: 100},
			allowRestart: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/config", map[string]any{"data": map[string]any{
					"efidisk0": "local-dir:100/vm-100-disk-0.qcow2,size=528K,efitype=4m,pre-enrolled-keys=0"}}),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": 100, "node": "pve", "type": "qemu"}}}),
				mockServer.RequestsGetJson("/cluster/ha/resources/100", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/status/current", map[string]any{"data": map[string]any{"status": "running"}}),
				mockServer.RequestsPostResponse("/nodes/pve/qemu/100/status/shutdown", nil, []byte(`{"data":"`+UPID("pve", "qmshutdown", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve/tasks/"+mockServer.Path(UPID("pve", "qmshutdown", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsPostResponse("/nodes/pve/qemu/100/move_disk", map[string]any{
					"delete":  "1",
					"disk":    "efidisk0",
					"storage": "local-lvm"}, []byte(`{"data":"`+UPID("pve", "qmmovedisk", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve/tasks/"+mockServer.Path(UPID("pve", "qmmovedisk", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
				mockServer.RequestsPostResponse("/nodes/pve/qemu/100/status/start", nil, []byte(`{"data":"`+UPID("pve", "qmstart", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve/tasks/"+mockServer.Path(UPID("pve", "qmstart", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `migrate efi disk, stopped, allowRestart true`,
			config: ConfigQemu{
				EfiDisk: &EfiDisk{
					Storage: util.Pointer(StorageName("local-lvm"))}},
			vmr:          VmRef{node: "pve", vmId: 100},
			allowRestart: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/config", map[string]any{"data": map[string]any{
					"efidisk0": "local-dir:100/vm-100-disk-0.qcow2,size=528K,efitype=4m,pre-enrolled-keys=0"}}),
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": 100, "node": "pve", "type": "qemu"}}}),
				mockServer.RequestsGetJson("/cluster/ha/resources/100", map[string]any{"data": map[string]any{}}), // TODO we don't need this info here
				mockServer.RequestsGetJson("/version", map[string]any{"data": map[string]any{"version": "8.0.1"}}),
				mockServer.RequestsGetJson("/nodes/pve/qemu/100/status/current", map[string]any{"data": map[string]any{"status": "stopped"}}),
				mockServer.RequestsPostResponse("/nodes/pve/qemu/100/move_disk", map[string]any{
					"delete":  "1",
					"disk":    "efidisk0",
					"storage": "local-lvm"}, []byte(`{"data":"`+UPID("pve", "qmmovedisk", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve/tasks/"+mockServer.Path(UPID("pve", "qmmovedisk", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().QemuGuest.Update(context.Background(), test.vmr, test.allowRestart, test.allowForceStop, test.config)
			require.Equal(t, test.err, err)
			server.Clear(t)
			c.clearVersion()
		})
	}
}
