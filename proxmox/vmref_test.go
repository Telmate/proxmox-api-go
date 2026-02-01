package proxmox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_CloneLxcTarget_mapToAPI(t *testing.T) {
	t.Parallel()
	type testOutput struct {
		id   GuestID
		node NodeName
		pool PoolName
		api  map[string]interface{}
	}
	tests := []struct {
		name   string
		input  CloneLxcTarget
		output testOutput
	}{
		{name: `nil`},
		{name: `Full ID`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				ID: util.Pointer(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"newid":  100}}},
		{name: `Full Name`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Name: util.Pointer(GuestName("test"))}},
			output: testOutput{
				api: map[string]interface{}{
					"full":     true,
					"target":   "",
					"hostname": "test"}}},
		{name: `Full Node`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]interface{}{
					"full":   true,
					"target": "test"}}},
		{name: `Full Pool`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Pool: util.Pointer(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"pool":   "test"}}},
		{name: `Full Storage`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Storage: util.Pointer("test")}},
			output: testOutput{
				api: map[string]interface{}{
					"full":    true,
					"target":  "",
					"storage": "test"}}},
		{name: `Linked ID`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				ID: util.Pointer(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]interface{}{
					"full":   false,
					"target": "",
					"newid":  100}}},
		{name: `Linked Name`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Name: util.Pointer(GuestName("test"))}},
			output: testOutput{
				api: map[string]interface{}{
					"full":     false,
					"target":   "",
					"hostname": "test"}}},
		{name: `Linked Node`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]interface{}{
					"full":   false,
					"target": "test"}}},
		{name: `Linked Pool`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Pool: util.Pointer(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]interface{}{
					"full":   false,
					"target": "",
					"pool":   "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, node, pool, api := test.input.mapToAPI()
			require.Equal(t, test.output.id, id)
			require.Equal(t, test.output.node, node)
			require.Equal(t, test.output.pool, pool)
			require.Equal(t, test.output.api, api)
		})
	}
}

func Test_CloneLxcTarget_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloneLxcTarget
		output error
	}{
		{name: `Invalid errors.New(CloneLxcTarget_Error_MutualExclusivity)`,
			input: CloneLxcTarget{
				Full:   &CloneLxcFull{},
				Linked: &CloneLinked{}},
			output: errors.New(CloneLxcTarget_Error_MutualExclusivity)},
		{name: `Invalid errors.New(CloneLxcTarget_Error_NoneSet)`,
			input:  CloneLxcTarget{},
			output: errors.New(CloneLxcTarget_Error_NoneSet)},
		{name: `Invalid Full Node errors.New(NodeName_Error_Empty)`,
			input:  CloneLxcTarget{Full: &CloneLxcFull{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Full ID errors.New(GuestID_Error_Minimum)`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				ID: util.Pointer(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Full Name errors.New(GuestName_Error_Empty)`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test",
				Name: util.Pointer(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Full Pool errors.New(PoolName_Error_Empty)`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Pool: util.Pointer(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Full Storage`, // TODO this should be an error
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node:    "test",
				Storage: util.Pointer("")}},
			output: nil},
		{name: `Invalid Linked Node errors.New(NodeName_Error_Empty)`,
			input:  CloneLxcTarget{Linked: &CloneLinked{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Linked ID errors.New(GuestID_Error_Minimum)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				ID: util.Pointer(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Linked Name errors.New(GuestName_Error_Empty)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test",
				Name: util.Pointer(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Linked Pool errors.New(PoolName_Error_Empty)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Pool: util.Pointer(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Valid Full`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test"}}},
		{name: `Valid Linked`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_CloneQemuTarget_mapToAPI(t *testing.T) {
	t.Parallel()
	type testOutput struct {
		id   GuestID
		node NodeName
		pool PoolName
		api  map[string]interface{}
	}
	tests := []struct {
		name   string
		input  CloneQemuTarget
		output testOutput
	}{
		{name: `nil`},
		{name: `Full ID`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				ID: util.Pointer(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"newid":  100}}},
		{name: `Full Name`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Name: util.Pointer(GuestName("test"))}},
			output: testOutput{
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"name":   "test"}}},
		{name: `Full Node`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]interface{}{
					"full":   true,
					"target": "test"}}},
		{name: `Full Pool`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Pool: util.Pointer(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"pool":   "test"}}},
		{name: `Full Storage`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Storage: util.Pointer("test")}},
			output: testOutput{
				api: map[string]interface{}{
					"full":    true,
					"target":  "",
					"storage": "test"}}},
		{name: `Full StorageFormat`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				StorageFormat: util.Pointer(QemuDiskFormat("test"))}},
			output: testOutput{
				api: map[string]interface{}{
					"full":   true,
					"target": "",
					"format": "test"}}},
		{name: `Linked ID`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				ID: util.Pointer(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]interface{}{
					"full":   false,
					"target": "",
					"newid":  100}}},
		{name: `Linked Name`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Name: util.Pointer(GuestName("test"))}},
			output: testOutput{
				api: map[string]interface{}{
					"full":   false,
					"target": "",
					"name":   "test"}}},
		{name: `Linked Node`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]interface{}{
					"full":   false,
					"target": "test"}}},
		{name: `Linked Pool`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Pool: util.Pointer(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]interface{}{
					"full":   false,
					"target": "",
					"pool":   "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, node, pool, api := test.input.mapToAPI()
			require.Equal(t, test.output.id, id)
			require.Equal(t, test.output.node, node)
			require.Equal(t, test.output.pool, pool)
			require.Equal(t, test.output.api, api)
		})
	}
}

func Test_CloneQemuTarget_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  CloneQemuTarget
		output error
	}{
		{name: `Invalid errors.New(CloneQemuTarget_Error_MutualExclusivity)`,
			input: CloneQemuTarget{
				Full:   &CloneQemuFull{},
				Linked: &CloneLinked{}},
			output: errors.New(CloneQemuTarget_Error_MutualExclusivity)},
		{name: `Invalid errors.New(CloneQemuTarget_Error_NoneSet)`,
			input:  CloneQemuTarget{},
			output: errors.New(CloneQemuTarget_Error_NoneSet)},
		{name: `Invalid Full Node errors.New(NodeName_Error_Empty)`,
			input:  CloneQemuTarget{Full: &CloneQemuFull{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Full ID errors.New(GuestID_Error_Minimum)`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				ID: util.Pointer(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Full Name errors.New(GuestName_Error_Empty)`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test",
				Name: util.Pointer(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Full Pool errors.New(PoolName_Error_Empty)`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Pool: util.Pointer(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Full Storage`, // TODO this should be an error
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node:    "test",
				Storage: util.Pointer("")}},
			output: nil},
		{name: `Invalid Full StorageFormat`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				StorageFormat: util.Pointer(QemuDiskFormat(""))}},
			output: QemuDiskFormat("").Error()},
		{name: `Invalid Linked Node errors.New(NodeName_Error_Empty)`,
			input:  CloneQemuTarget{Linked: &CloneLinked{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Linked ID errors.New(GuestID_Error_Minimum)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				ID: util.Pointer(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Linked Name errors.New(GuestName_Error_Empty)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test",
				Name: util.Pointer(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Linked Pool errors.New(PoolName_Error_Empty)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Pool: util.Pointer(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Valid Full`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test"}}},
		{name: `Valid Linked`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_VmRef_Migrate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		c   *Client
		ctx context.Context
		vmr *VmRef
	}
	tests := []struct {
		name  string
		input testInput
	}{
		{name: `Client nil`,
			input: testInput{
				ctx: context.Background(),
				vmr: &VmRef{}}},
		{name: `Context nil`,
			input: testInput{
				c:   fakeClient(),
				vmr: &VmRef{}}},
		{name: `VmRef nil`,
			input: testInput{
				c:   fakeClient(),
				ctx: context.Background()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { test.input.vmr.Migrate(test.input.ctx, test.input.c, "valid", false) })
			require.Error(t, test.input.vmr.Migrate(test.input.ctx, test.input.c, "valid", false))
		})
	}
}

func Test_VmRef_MigrateNoCheck(t *testing.T) {
	t.Parallel()
	type testInput struct {
		c   *Client
		ctx context.Context
		vmr *VmRef
	}
	tests := []struct {
		name  string
		input testInput
	}{
		{name: `Client nil`,
			input: testInput{
				ctx: context.Background(),
				vmr: &VmRef{}}},
		{name: `Context nil`,
			input: testInput{
				c:   fakeClient(),
				vmr: &VmRef{}}},
		{name: `VmRef nil`,
			input: testInput{
				c:   fakeClient(),
				ctx: context.Background()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { test.input.vmr.MigrateNoCheck(test.input.ctx, test.input.c, "valid", false) })
			require.Error(t, test.input.vmr.MigrateNoCheck(test.input.ctx, test.input.c, "valid", false))
		})
	}
}

func Test_VmRef_pendingChanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []any
		output bool
		err    error
	}{
		{name: `No pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cpu"), "value": float64(2)},
				map[string]any{"key": string("disks"), "value": string("sata0")}}},
		{name: `Pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2), "pending": float64(3)},
				map[string]any{"key": string("disks"), "value": string("sata0")}},
			output: true},
		{name: `Delete`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2)},
				map[string]any{"key": string("tpmstate0"), "value": string("local-zfs:vm-1001-disk-2,size=4M,version=v1.2"), "delete": float64(1)}},
			output: true},
		{name: `Missing Value Pending (Real-World Case)`,
			input: []any{
				map[string]any{"key": string("bios"), "pending": string("ovmf")},
				map[string]any{"key": string("cores"), "value": float64(2)}},
			output: true},
		{name: `Error`,
			err: errors.New("test")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pending, err := (&VmRef{}).pendingChanges(context.Background(), &mockClientAPI{
				getGuestPendingChangesFunc: func(ctx context.Context, vmr *VmRef) ([]any, error) {
					return test.input, test.err
				}})
			require.Equal(t, test.err, err)
			require.Equal(t, test.output, pending)
		})
	}
}

func Test_VmRef_pendingConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   []any
		pending bool
		output  map[string]any
		err     error
	}{
		{name: `No pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cpu"), "value": float64(2)},
				map[string]any{"key": string("disks"), "value": string("sata0")}},
			output: map[string]any{
				"cpu":   float64(2),
				"disks": string("sata0"),
				"test":  string("value")}},
		{name: `Pending`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2), "pending": float64(3)},
				map[string]any{"key": string("disks"), "value": string("sata0")}},
			output: map[string]any{
				"cores": float64(2),
				"disks": string("sata0"),
				"test":  string("value")},
			pending: true},
		{name: `Delete`,
			input: []any{
				map[string]any{"key": string("test"), "value": string("value")},
				map[string]any{"key": string("cores"), "value": float64(2)},
				map[string]any{"key": string("tpmstate0"), "value": string("local-zfs:vm-1001-disk-2,size=4M,version=v1.2"), "delete": float64(1)}},
			output: map[string]any{
				"cores":     float64(2),
				"test":      string("value"),
				"tpmstate0": string("local-zfs:vm-1001-disk-2,size=4M,version=v1.2")},
			pending: true},
		{name: `Missing Value Pending (Real-World Case)`,
			input: []any{
				map[string]any{"key": string("bios"), "pending": string("ovmf")},
				map[string]any{"key": string("cores"), "value": float64(2)}},
			output:  map[string]any{"cores": float64(2)},
			pending: true},
		{name: `Error`,
			err: errors.New("test")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, pending, err := (&VmRef{}).pendingConfig(context.Background(), &mockClientAPI{
				getGuestPendingChangesFunc: func(ctx context.Context, vmr *VmRef) ([]any, error) {
					return test.input, test.err
				}})
			require.Equal(t, test.err, err)
			require.Equal(t, test.pending, pending)
			require.Equal(t, test.output, output)
		})
	}
}

func Test_RawGuestStatus_Get(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  RawGuestStatus
		output GuestStatus
	}{
		{name: `Name`,
			input:  RawGuestStatus{"name": "test"},
			output: GuestStatus{Name: "test"}},
		{name: `State`,
			input:  RawGuestStatus{"status": "running"},
			output: GuestStatus{State: PowerStateRunning}},
		{name: `Uptime`,
			input:  RawGuestStatus{"uptime": float64(12345)},
			output: GuestStatus{Uptime: time.Duration(12345) * time.Second}},
		{name: `All`,
			input: RawGuestStatus{
				"name":   "guest100",
				"status": "stopped",
				"uptime": float64(95673)},
			output: GuestStatus{
				Name:   "guest100",
				State:  PowerStateStopped,
				Uptime: time.Duration(95673) * time.Second}},
		{name: `Empty`,
			input:  RawGuestStatus{},
			output: GuestStatus{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			require.Equal(t, test.output, test.input.Get(), test.name)
		})
	}
}
