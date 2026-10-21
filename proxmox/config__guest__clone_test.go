package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CloneLxcTarget_mapToAPI(t *testing.T) {
	t.Parallel()
	type testOutput struct {
		id   GuestID
		node NodeName
		pool PoolName
		api  map[string]string
	}
	tests := []struct {
		name   string
		input  CloneLxcTarget
		output testOutput
	}{
		{name: `nil`},
		{name: `Full ID`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				ID: new(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]string{
					"full":   "1",
					"target": "",
					"newid":  "100"}}},
		{name: `Full Name`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Name: new(GuestName("test"))}},
			output: testOutput{
				api: map[string]string{
					"full":     "1",
					"target":   "",
					"hostname": "test"}}},
		{name: `Full Node`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]string{
					"full":   "1",
					"target": "test"}}},
		{name: `Full Pool`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Pool: new(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]string{
					"full":   "1",
					"target": "",
					"pool":   "test"}}},
		{name: `Full Storage`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Storage: new(StorageName("test"))}},
			output: testOutput{
				api: map[string]string{
					"full":    "1",
					"target":  "",
					"storage": "test"}}},
		{name: `Linked ID`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				ID: new(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]string{
					"target": "",
					"newid":  "100"}}},
		{name: `Linked Name`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Name: new(GuestName("test"))}},
			output: testOutput{
				api: map[string]string{
					"target":   "",
					"hostname": "test"}}},
		{name: `Linked Node`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]string{
					"target": "test"}}},
		{name: `Linked Pool`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Pool: new(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]string{
					"target": "",
					"pool":   "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, node, pool, api := test.input.mapToAPI()
			require.Equal(t, test.output.id, id)
			require.Equal(t, test.output.node, node)
			require.Equal(t, test.output.pool, pool)
			testParamsEqualRaw(t, test.output.api, api)
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
				ID: new(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Full Name errors.New(GuestName_Error_Empty)`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node: "test",
				Name: new(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Full Pool errors.New(PoolName_Error_Empty)`,
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Pool: new(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Full Storage`, // TODO this should be an error
			input: CloneLxcTarget{Full: &CloneLxcFull{
				Node:    "test",
				Storage: new(StorageName("test"))}},
			output: nil},
		{name: `Invalid Linked Node errors.New(NodeName_Error_Empty)`,
			input:  CloneLxcTarget{Linked: &CloneLinked{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Linked ID errors.New(GuestID_Error_Minimum)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				ID: new(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Linked Name errors.New(GuestName_Error_Empty)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Node: "test",
				Name: new(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Linked Pool errors.New(PoolName_Error_Empty)`,
			input: CloneLxcTarget{Linked: &CloneLinked{
				Pool: new(PoolName(""))}},
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
		api  map[string]string
	}
	tests := []struct {
		name   string
		input  CloneQemuTarget
		output testOutput
	}{
		{name: `nil`},
		{name: `Full ID`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				ID: new(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]string{
					"full":   "1",
					"target": "",
					"newid":  "100"}}},
		{name: `Full Name`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Name: new(GuestName("test"))}},
			output: testOutput{
				api: map[string]string{
					"full":   "1",
					"target": "",
					"name":   "test"}}},
		{name: `Full Node`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]string{
					"full":   "1",
					"target": "test"}}},
		{name: `Full Pool`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Pool: new(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]string{
					"full":   "1",
					"target": "",
					"pool":   "test"}}},
		{name: `Full Storage`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Storage: new(StorageName("test"))}},
			output: testOutput{
				api: map[string]string{
					"full":    "1",
					"target":  "",
					"storage": "test"}}},
		{name: `Full StorageFormat`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				StorageFormat: new(QemuDiskFormat("test"))}},
			output: testOutput{
				api: map[string]string{
					"full":   "1",
					"target": "",
					"format": "test"}}},
		{name: `Linked ID`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				ID: new(GuestID(100))}},
			output: testOutput{
				id: 100,
				api: map[string]string{
					"target": "",
					"newid":  "100"}}},
		{name: `Linked Name`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Name: new(GuestName("test"))}},
			output: testOutput{
				api: map[string]string{
					"target": "",
					"name":   "test"}}},
		{name: `Linked Node`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test"}},
			output: testOutput{
				node: "test",
				api: map[string]string{
					"target": "test"}}},
		{name: `Linked Pool`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Pool: new(PoolName("test"))}},
			output: testOutput{
				pool: "test",
				api: map[string]string{
					"target": "",
					"pool":   "test"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, node, pool, api := test.input.mapToAPI()
			require.Equal(t, test.output.id, id)
			require.Equal(t, test.output.node, node)
			require.Equal(t, test.output.pool, pool)
			testParamsEqualRaw(t, test.output.api, api)
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
				ID: new(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Full Name errors.New(GuestName_Error_Empty)`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node: "test",
				Name: new(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Full Pool errors.New(PoolName_Error_Empty)`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Pool: new(PoolName(""))}},
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Full Storage`, // TODO this should be an error
			input: CloneQemuTarget{Full: &CloneQemuFull{
				Node:    "test",
				Storage: new(StorageName(""))}},
			output: nil},
		{name: `Invalid Full StorageFormat`,
			input: CloneQemuTarget{Full: &CloneQemuFull{
				StorageFormat: new(QemuDiskFormat(""))}},
			output: QemuDiskFormat("").Error()},
		{name: `Invalid Linked Node errors.New(NodeName_Error_Empty)`,
			input:  CloneQemuTarget{Linked: &CloneLinked{}},
			output: errors.New(NodeName_Error_Empty)},
		{name: `Invalid Linked ID errors.New(GuestID_Error_Minimum)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				ID: new(GuestID(99))}},
			output: errors.New(GuestID_Error_Minimum)},
		{name: `Invalid Linked Name errors.New(GuestName_Error_Empty)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Node: "test",
				Name: new(GuestName(""))}},
			output: errors.New(GuestNameErrorEmpty)},
		{name: `Invalid Linked Pool errors.New(PoolName_Error_Empty)`,
			input: CloneQemuTarget{Linked: &CloneLinked{
				Pool: new(PoolName(""))}},
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
