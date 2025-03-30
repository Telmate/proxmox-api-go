package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

func Test_CloneLxcTarget_mapToAPI(t *testing.T) {
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
				Name: util.Pointer("test")}},
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
				Name: util.Pointer("test")}},
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
		{name: `Invalid Full Name`, // TODO this should be an error
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test",
				Name: util.Pointer("")}},
			output: nil},
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
		{name: `Invalid Linked Name`, // TODO this should be an error
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test",
				Name: util.Pointer("")}},
			output: nil},
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
				Name: util.Pointer("test")}},
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
				Name: util.Pointer("test")}},
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
		{name: `Invalid Full Name`, // TODO this should be an error
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test",
				Name: util.Pointer("")}},
			output: nil},
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
		{name: `Invalid Linked Name`, // TODO this should be an error
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test",
				Name: util.Pointer("")}},
			output: nil},
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
			_, err := test.input.vmr.Migrate(test.input.ctx, test.input.c, "valid", false)
			require.Error(t, err)
		})
	}
}

func Test_VmRef_MigrateNoCheck(t *testing.T) {
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
			_, err := test.input.vmr.MigrateNoCheck(test.input.ctx, test.input.c, "valid", false)
			require.Error(t, err)
		})
	}
}
