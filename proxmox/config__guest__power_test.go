package proxmox

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/stretchr/testify/require"
)

func Test_GuestInterface_Reboot(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		vmr      VmRef
		requests []mockServer.Request
		err      error
	}{
		{name: `VmRef minimal`,
			vmr: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"}}}),
				mockServer.RequestsPostResponse("/nodes/pve3/lxc/200/status/reboot", map[string]any{},
					[]byte(`{"data":"`+UPID("pve3", "qmreboot", GuestID(200))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmreboot", GuestID(200)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `VmRef maximal`,
			vmr: VmRef{node: "test", vmId: 100, vmType: GuestQemu},
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/test/qemu/100/status/reboot", map[string]any{},
					[]byte(`{"data":"`+UPID("test", "qmreboot", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmreboot", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Guest.Reboot(context.Background(), test.vmr)
			require.Equal(t, test.err, err)
			server.Clear(t)
			c.clearVersion()
		})
	}
}

func Test_GuestInterface_Shutdown(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		vmr      VmRef
		requests []mockServer.Request
		err      error
	}{
		{name: `VmRef minimal`,
			vmr: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"}}}),
				mockServer.RequestsPostResponse("/nodes/pve3/lxc/200/status/shutdown", map[string]any{},
					[]byte(`{"data":"`+UPID("pve3", "qmshutdown", GuestID(200))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmshutdown", GuestID(200)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `VmRef minimal stopped`,
			vmr: VmRef{vmId: 300},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu", "status": "stopped"}}}))},
		{name: `VmRef maximal`,
			vmr: VmRef{node: "test", vmId: 100, vmType: GuestQemu},
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/test/qemu/100/status/shutdown", map[string]any{},
					[]byte(`{"data":"`+UPID("test", "qmshutdown", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmshutdown", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Guest.Shutdown(context.Background(), test.vmr)
			require.Equal(t, test.err, err)
			server.Clear(t)
			c.clearVersion()
		})
	}
}

func Test_GuestInterface_Start(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		vmr      VmRef
		requests []mockServer.Request
		err      error
	}{
		{name: `VmRef minimal`,
			vmr: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"}}}),
				mockServer.RequestsPostResponse("/nodes/pve3/lxc/200/status/start", map[string]any{},
					[]byte(`{"data":"`+UPID("pve3", "qmstart", GuestID(200))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmstart", GuestID(200)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
		{name: `VmRef minimal running`,
			vmr: VmRef{vmId: 300},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu", "status": "running"}}}))},
		{name: `VmRef maximal`,
			vmr: VmRef{node: "test", vmId: 100, vmType: GuestQemu},
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/test/qemu/100/status/start", map[string]any{},
					[]byte(`{"data":"`+UPID("test", "qmstart", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstart", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}))},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Guest.Start(context.Background(), test.vmr)
			require.Equal(t, test.err, err)
			server.Clear(t)
			c.clearVersion()
		})
	}
}

func Test_GuestInterface_Stop(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		vmr      VmRef
		force    bool
		requests []mockServer.Request
		err      error
	}{
		{name: `VmRef minimal false`,
			vmr: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"},
				}}),
				mockServer.RequestsPostResponse("/nodes/pve3/lxc/200/status/stop", map[string]any{},
					[]byte(`{"data":"`+UPID("pve3", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmstop", GuestID(200)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `VmRef minimal true`,
			vmr:   VmRef{vmId: 200},
			force: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu"},
				}}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPostResponse("/nodes/pve3/lxc/200/status/stop", map[string]any{"overrule-shutdown": "1"},
					[]byte(`{"data":"`+UPID("pve3", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJson("/nodes/pve3/tasks/"+mockServer.Path(UPID("pve3", "qmstop", GuestID(200)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `VmRef minimal false stopped`,
			vmr: VmRef{vmId: 300},
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu", "status": "stopped"},
				}}))},
		{name: `VmRef minimal true stopped`,
			vmr:   VmRef{vmId: 300},
			force: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJson("/cluster/resources?type=vm", map[string]any{"data": []any{
					map[string]any{"vmid": float64(100), "node": "myNode", "type": "qemu"},
					map[string]any{"vmid": float64(200), "node": "pve3", "type": "lxc"},
					map[string]any{"vmid": float64(300), "node": "myNode", "type": "qemu", "status": "stopped"},
				}}))},
		{name: `VmRef maximal false`,
			vmr: VmRef{node: "test", vmId: 100, vmType: GuestQemu},
			requests: mockServer.Append(
				mockServer.RequestsPostResponse("/nodes/test/qemu/100/status/stop", map[string]any{},
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
		{name: `VmRef maximal true`,
			vmr:   VmRef{node: "test", vmId: 100, vmType: GuestQemu},
			force: true,
			requests: mockServer.Append(
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPostResponse("/nodes/test/qemu/100/status/stop", map[string]any{"overrule-shutdown": "1"},
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(100))+`"}`)),
				mockServer.RequestsGetJson("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(100)))+"/status",
					map[string]any{"data": map[string]any{"exitstatus": string("OK")}}),
			)},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			server.Set(test.requests, t)
			err := c.New().Guest.Stop(context.Background(), test.vmr, test.force)
			require.Equal(t, test.err, err)
			server.Clear(t)
			c.clearVersion()
		})
	}
}
