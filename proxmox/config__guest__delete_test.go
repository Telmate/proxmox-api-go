package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/stretchr/testify/require"
)

func Test_GuestInterface_Delete(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		guest    VmRef
		deleted  bool
		requests []mockServer.Request
		err      error
	}{
		{name: `error no GuestID`,
			err: errors.New(VmRef_Error_IDnotSet)},
		{name: `guest does not exists`,
			guest: VmRef{vmId: 100},
			requests: mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
				map[string]any{"vmid": float64(200)},
				map[string]any{"vmid": float64(300)},
				map[string]any{"vmid": float64(400)}})},
		{name: `failed to list guests`,
			guest:    VmRef{vmId: 100},
			requests: mockServer.RequestsError("/cluster/resources?type=vm", mockServer.GET, 500, 3),
			err:      errors.New(mockServer.InternalServerError)},
		{name: `get config lxc failed`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "pve2", "type": "lxc"}}),
				mockServer.RequestsError("/nodes/pve2/lxc/200/config", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `get config qemu failed`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "qemu"}}),
				mockServer.RequestsError("/nodes/test/qemu/200/config", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `get config lxc protectted`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "lxc"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{
					"protection": float64(1)})),
			err: &errorWrapper[GuestID]{
				err: Error.GuestIsProtectedCantDelete(),
				id:  200}},
		{name: `get config qemu protectted`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "qemu"}}),
				mockServer.RequestsGetJson("/nodes/test/qemu/200/config", map[string]any{"data": map[string]any{
					"protection": float64(1)}})),
			err: &errorWrapper[GuestID]{
				err: Error.GuestIsProtectedCantDelete(),
				id:  200}},
		{name: `guest lxc deleted by external after list`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "lxc"}}),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/config", mockServer.GET, mockServer.JsonError(500, map[string]any{
					"message": "Configuration file 'nodes/test/lxc-server/200.conf' does not exist"})))},
		{name: `guest qemu deleted by external after list`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "qemu"}}),
				mockServer.RequestsErrorHandled("/nodes/test/qemu/200/config", mockServer.GET, mockServer.JsonError(500, map[string]any{
					"message": "Configuration file 'nodes/test/qemu-server/200.conf' does not exist"})))},
		{name: `success qemu | running | no HA | overrule`,
			guest:   VmRef{vmId: 200},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "qemu"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsVersion("8.0.0"),
				mockServer.RequestsPostResponse("/nodes/test/qemu/200/status/stop", map[string]any{"overrule-shutdown": "1"},
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}),
				mockServer.RequestsDeleteResponse("/nodes/test/qemu/200", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
		{name: `success qemu | stopped | no HA | overrule`,
			guest:   VmRef{vmId: 200},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "qemu",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsDeleteResponse("/nodes/test/qemu/200", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
		{name: `success qemu | stopped | HA | overrule`,
			guest:   VmRef{vmId: 200},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":    float64(200),
						"node":    "test",
						"type":    "qemu",
						"status":  "stopped",
						"hastate": ""}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsDeleteResponse("/nodes/test/qemu/200?purge=1", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
		{name: `success qemu | running | HA`,
			guest:   VmRef{vmId: 200},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":    float64(200),
						"node":    "test",
						"type":    "qemu",
						"status":  "running",
						"hastate": ""}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsDelete("/cluster/ha/resources/200", nil),
				mockServer.RequestsVersion("7.255.255"),
				mockServer.RequestsPostResponse("/nodes/test/qemu/200/status/stop", nil,
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}),
				mockServer.RequestsDeleteResponse("/nodes/test/qemu/200", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
		{name: `fail qemu | running | HA | delete HA`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":    float64(200),
						"node":    "test",
						"type":    "qemu",
						"status":  "running",
						"hastate": ""}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsError("/cluster/ha/resources/200", mockServer.DELETE, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `success lxc all issues`,
			guest:   VmRef{vmId: 200},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "lxc",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{}),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsDeleteResponse("/nodes/test/lxc/200", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
		{name: `fail lxc loop exhausetd | stopped`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "lxc",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{}),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"}))),
			err: errors.New("unable to delete guest in 3 attempts")},
		{name: `fail lxc loop exhausetd | running`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "lxc",
						"status": "running"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{}),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "CT 200 not running"})),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"}))),
			err: errors.New("unable to delete guest in 3 attempts")},
		{name: `error qemu | stopped | no HA`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "qemu",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsError("/nodes/test/qemu/200", mockServer.DELETE, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `error api qemu | stopped | no HA`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "qemu",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsErrorHandled("/nodes/test/qemu/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "random api error"}))),
			err: &ApiError{
				Code:    "500",
				Message: "random api error"}},
		{name: `error version | qemu | running | no HA`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{"vmid": float64(200), "node": "test", "type": "qemu"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsError("/version", mockServer.GET, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `error | lxc all issues | stop error`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "lxc",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{}),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsError("/nodes/test/lxc/200/status/stop", mockServer.POST, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `error api | lxc all issues | stop error`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "lxc",
						"status": "stopped"}}),
				mockServer.RequestsGetJsonData("/nodes/test/lxc/200/config", map[string]any{}),
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "unable to destroy CT 200 - container is running"})),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsErrorHandled("/nodes/test/lxc/200/status/stop", mockServer.POST, mockServer.JsonError(500, map[string]any{
					"message": "random api error"}))),
			err: &ApiError{
				Code:    "500",
				Message: "random api error"}},
		{name: `error | qemu all issues | delete error`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "qemu",
						"status": "running"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsPostResponse("/nodes/test/qemu/200/status/stop", nil,
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}),
				mockServer.RequestsError("/nodes/test/qemu/200", mockServer.DELETE, 500, 3)),
			err: errors.New(mockServer.InternalServerError)},
		{name: `error api | qemu all issues | delete error`,
			guest: VmRef{vmId: 200},
			requests: mockServer.Append(
				mockServer.RequestsGetJsonData("/cluster/resources?type=vm", []any{
					map[string]any{
						"vmid":   float64(200),
						"node":   "test",
						"type":   "qemu",
						"status": "running"}}),
				mockServer.RequestsGetJsonData("/nodes/test/qemu/200/config", map[string]any{}),
				mockServer.RequestsVersion("7.255.255"),
				// in loop
				mockServer.RequestsPostResponse("/nodes/test/qemu/200/status/stop", nil,
					[]byte(`{"data":"`+UPID("test", "qmstop", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmstop", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}),
				mockServer.RequestsErrorHandled("/nodes/test/qemu/200", mockServer.DELETE, mockServer.JsonError(500, map[string]any{
					"message": "random api error"}))),
			err: &ApiError{
				Code:    "500",
				Message: "random api error"}},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server.Set(test.requests, t)
			deleted, err := c.New().Guest.Delete(t.Context(), test.guest)
			require.Equal(t, test.err, err)
			require.Equal(t, test.deleted, deleted)
			server.Clear(t)
			c.clearVersion()
		})
	}
}

func Test_GuestInterface_DeleteNoCheck(t *testing.T) {
	t.Parallel()
	UPID := func(node NodeName, task string, guest GuestID) string {
		return generateUPID(node, task, guest, UserID{Name: "root", Realm: "pam"})
	}
	tests := []struct {
		name     string
		guest    VmRef
		deleted  bool
		requests []mockServer.Request
		err      error
	}{
		{name: `success`,
			guest:   VmRef{vmId: 200, node: "test", vmType: GuestQemu},
			deleted: true,
			requests: mockServer.Append(
				mockServer.RequestsDeleteResponse("/nodes/test/qemu/200", nil,
					[]byte(`{"data":"`+UPID("test", "qmdestroy", GuestID(200))+`"}`)),
				mockServer.RequestsGetJsonData("/nodes/test/tasks/"+mockServer.Path(UPID("test", "qmdestroy", GuestID(200)))+"/status",
					map[string]any{"exitstatus": string("OK")}))},
	}
	server, c := testMockServerInit(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server.Set(test.requests, t)
			deleted, err := c.New().Guest.DeleteNoCheck(t.Context(), test.guest)
			require.Equal(t, test.err, err)
			require.Equal(t, test.deleted, deleted)
			server.Clear(t)
			c.clearVersion()
		})
	}
}
