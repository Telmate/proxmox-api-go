package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_pool"
	"github.com/stretchr/testify/require"
)

func Test_ConfigPool_mapToApi(t *testing.T) {
	type testInput struct {
		new     ConfigPool
		current *ConfigPool
	}
	tests := []struct {
		name   string
		input  testInput
		output map[string]interface{}
	}{
		{name: `Create Full`,
			input: testInput{
				new: ConfigPool{
					Name:    "test",
					Comment: util.Pointer("test-comment"),
					Guests:  &[]uint{100, 300, 200}}},
			output: map[string]interface{}{
				"poolid":  "test",
				"comment": "test-comment"}},
		{name: `Create poolid`,
			input: testInput{
				new: ConfigPool{Name: "test"}},
			output: map[string]interface{}{"poolid": "test"}},
		{name: `Create comment`,
			input: testInput{
				new: ConfigPool{Comment: util.Pointer("test-comment")}},
			output: map[string]interface{}{
				"poolid":  "",
				"comment": "test-comment"}},
		{name: `Create members`,
			input: testInput{
				new: ConfigPool{Guests: &[]uint{100, 300, 200}}},
			output: map[string]interface{}{"poolid": ""}},
		{name: `Update Full`,
			input: testInput{
				new: ConfigPool{
					Name:    "test",
					Comment: util.Pointer("test-comment"),
					Guests:  &[]uint{100, 300, 200}},
				current: &ConfigPool{
					Name:    "test",
					Comment: util.Pointer("old-comment"),
					Guests:  &[]uint{100, 300}}},
			output: map[string]interface{}{
				"comment": "test-comment"}},
		{name: `Update poolid`,
			input: testInput{
				new:     ConfigPool{Name: "test"},
				current: &ConfigPool{Name: "old"}},
			output: map[string]interface{}{}},
		{name: `Update comment`,
			input: testInput{
				new:     ConfigPool{Comment: util.Pointer("test-comment")},
				current: &ConfigPool{Comment: util.Pointer("old-comment")}},
			output: map[string]interface{}{
				"comment": "test-comment"}},
		{name: `Update members`,
			input: testInput{
				new:     ConfigPool{Guests: &[]uint{100, 300, 200}},
				current: &ConfigPool{Guests: &[]uint{100, 300}}},
			output: map[string]interface{}{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.new.mapToApi(test.input.current))
		})
	}
}

func Test_ConfigPool_mapToSDK(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string]interface{}
		output ConfigPool
	}{
		{name: "All",
			input: map[string]interface{}{
				"poolid":  "test",
				"comment": "test",
				"members": []interface{}{
					map[string]interface{}{"vmid": float64(100)},
					map[string]interface{}{"vmid": float64(300)},
					map[string]interface{}{"vmid": float64(200)}}},
			output: ConfigPool{
				Name:    "test",
				Comment: util.Pointer("test"),
				Guests:  &[]uint{100, 300, 200}}},
		{name: "poolid",
			input:  map[string]interface{}{"poolid": "test"},
			output: ConfigPool{Name: "test"}},
		{name: "comment",
			input:  map[string]interface{}{"comment": "test"},
			output: ConfigPool{Comment: util.Pointer("test")}},
		{name: "members",
			input: map[string]interface{}{
				"members": []interface{}{
					map[string]interface{}{"vmid": float64(100)},
					map[string]interface{}{"vmid": float64(300)},
					map[string]interface{}{"vmid": float64(200)}}},
			output: ConfigPool{Guests: &[]uint{100, 300, 200}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, ConfigPool{}.mapToSDK(test.input))
		})
	}
}

func Test_ConfigPool_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  ConfigPool
		output error
	}{
		{name: "Valid PoolName",
			input: ConfigPool{Name: PoolName(test_data_pool.PoolName_Legal())}},
		{name: "Invalid PoolName Empty",
			input:  ConfigPool{Name: ""},
			output: errors.New(PoolName_Error_Empty)},
		{name: "Invalid PoolName Length",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Max_Illegal())},
			output: errors.New(PoolName_Error_Length)},
		{name: "Invalid PoolName Characters",
			input:  ConfigPool{Name: PoolName(test_data_pool.PoolName_Error_Characters()[0])},
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_PoolName_guestsToRemoveFromPools(t *testing.T) {
	type testInput struct {
		guests      []GuestResource
		guestsToAdd []uint
	}
	tests := []struct {
		name   string
		input  testInput
		output map[PoolName][]uint
	}{
		{name: `'guestsToAdd' Not in 'guests'`,
			input: testInput{
				guests: []GuestResource{
					{Id: 100, Pool: "test"},
					{Id: 200, Pool: "poolA"},
					{Id: 300, Pool: "test"}},
				guestsToAdd: []uint{700, 800, 900}},
			output: map[PoolName][]uint{}},
		{name: `Empty`,
			output: map[PoolName][]uint{}},
		{name: `Empty 'guests'`,
			input: testInput{
				guestsToAdd: []uint{100, 300, 200}},
			output: map[PoolName][]uint{}},
		{name: `Empty 'guestsToAdd'`,
			input: testInput{
				guests: []GuestResource{
					{Id: 100, Pool: "test"},
					{Id: 200, Pool: "poolA"},
					{Id: 300, Pool: "test"}}},
			output: map[PoolName][]uint{}},
		{name: `Full`,
			input: testInput{
				guests: []GuestResource{
					{Id: 100, Pool: "test"},
					{Id: 200, Pool: "poolA"},
					{Id: 300, Pool: "test"}},
				guestsToAdd: []uint{100, 300, 200}},
			output: map[PoolName][]uint{
				"test":  {100, 300},
				"poolA": {200}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, PoolName("").guestsToRemoveFromPools(test.input.guests, test.input.guestsToAdd))
		})
	}
}

func Test_PoolName_mapToString(t *testing.T) {
	tests := []struct {
		name   string
		input  []uint
		output string
	}{
		{name: `empty`,
			input: []uint{}},
		{name: `full`,
			input:  []uint{100, 300, 200},
			output: "100,300,200"},
		{name: `nil`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, PoolName("").mapToString(test.input))
		})
	}
}

func Test_PoolName_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid PoolName`,
			input: test_data_pool.PoolName_Legals()},
		{name: `Invalid Empty`,
			output: errors.New(PoolName_Error_Empty)},
		{name: `Invalid Length`,
			input:  []string{test_data_pool.PoolName_Max_Illegal()},
			output: errors.New(PoolName_Error_Length)},
		{name: `Invalid Characters`,
			input:  test_data_pool.PoolName_Error_Characters(),
			output: errors.New(PoolName_Error_Characters)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range test.input {
				require.Equal(t, test.output, PoolName(input).Validate())
			}
		})
	}
}
