package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GuestHA_mapToApi(t *testing.T) {
	PointerHaState := func(i HaState) *HaState { return &i }             // TODO remove when we have a generic pointer function
	PointerHaGroupName := func(i HaGroupName) *HaGroupName { return &i } // TODO remove when we have a generic pointer function
	PointerHaRelocate := func(i HaRelocate) *HaRelocate { return &i }    // TODO remove when we have a generic pointer function
	PointerHaRestart := func(i HaRestart) *HaRestart { return &i }       // TODO remove when we have a generic pointer function
	PointerString := func(i string) *string { return &i }                // TODO remove when we have a generic pointer function
	type testInput struct {
		vmID   int
		config GuestHA
	}
	tests := []struct {
		name   string
		input  testInput
		output map[string]interface{}
	}{
		{name: `Test Comment Create`,
			input:  testInput{config: GuestHA{Comment: PointerString("abc")}},
			output: map[string]interface{}{"comment": string("abc")}},
		{name: `Test Comment Update`,
			input: testInput{vmID: 100, config: GuestHA{Comment: PointerString("abc")}},
			output: map[string]interface{}{
				"comment": string("abc"),
				"sid":     int(100)}},
		{name: `Test Full Create`,
			input: testInput{
				config: GuestHA{
					Comment:     PointerString("test"),
					Delete:      true,
					Group:       PointerHaGroupName("test-group"),
					Reallocates: PointerHaRelocate(1),
					Restarts:    PointerHaRestart(10),
					State:       PointerHaState(HaState_Started)}},
			output: map[string]interface{}{
				"comment":      string("test"),
				"group":        string("test-group"),
				"max_relocate": int(1),
				"max_restart":  int(10),
				"state":        string("started")}},
		{name: `Test Full Create nil`,
			input:  testInput{config: GuestHA{}},
			output: map[string]interface{}{}},
		{name: `Test Full Update`,
			input: testInput{vmID: 100, config: GuestHA{
				Comment:     PointerString("test"),
				Delete:      true,
				Group:       PointerHaGroupName("test-group"),
				Reallocates: PointerHaRelocate(10),
				Restarts:    PointerHaRestart(1),
				State:       PointerHaState(HaState_Stopped)}},
			output: map[string]interface{}{
				"comment":      string("test"),
				"group":        string("test-group"),
				"max_relocate": int(10),
				"max_restart":  int(1),
				"sid":          int(100),
				"state":        string("stopped")}},
		{name: `Test Full Update nil`,
			input:  testInput{vmID: 100, config: GuestHA{}},
			output: map[string]interface{}{"sid": int(100)}},
		{name: `Test Group Create`,
			input:  testInput{config: GuestHA{Group: PointerHaGroupName("test-group")}},
			output: map[string]interface{}{"group": string("test-group")}},
		{name: `Test Group Create nil`,
			input:  testInput{config: GuestHA{Group: nil}},
			output: map[string]interface{}{}},
		{name: `Test Group Create ""`,
			input:  testInput{config: GuestHA{Group: PointerHaGroupName("")}},
			output: map[string]interface{}{}},
		{name: `Test Group Update`,
			input: testInput{vmID: 100, config: GuestHA{Group: PointerHaGroupName("test-group")}},
			output: map[string]interface{}{
				"group": string("test-group"),
				"sid":   int(100)}},
		{name: `Test Group Update nil`,
			input: testInput{vmID: 100, config: GuestHA{Group: nil}},
			output: map[string]interface{}{
				"sid": int(100)}},
		{name: `Test Group Update ""`,
			input: testInput{vmID: 100, config: GuestHA{Group: PointerHaGroupName("")}},
			output: map[string]interface{}{
				"delete": string("group"),
				"sid":    int(100)}},
		{name: `Test Reallocates Create`,
			input:  testInput{config: GuestHA{Reallocates: PointerHaRelocate(10)}},
			output: map[string]interface{}{"max_relocate": int(10)}},
		{name: `Test Reallocates Update`,
			input: testInput{vmID: 100, config: GuestHA{Reallocates: PointerHaRelocate(10)}},
			output: map[string]interface{}{
				"max_relocate": int(10),
				"sid":          int(100)}},
		{name: `Test Restarts Create`,
			input:  testInput{config: GuestHA{Restarts: PointerHaRestart(10)}},
			output: map[string]interface{}{"max_restart": int(10)}},
		{name: `Test Restarts Update`,
			input: testInput{vmID: 100, config: GuestHA{Restarts: PointerHaRestart(10)}},
			output: map[string]interface{}{
				"max_restart": int(10),
				"sid":         int(100)}},
		{name: `Test State Create`,
			input:  testInput{config: GuestHA{State: PointerHaState(HaState_Started)}},
			output: map[string]interface{}{"state": string("started")}},
		{name: `Test State Update`,
			input: testInput{vmID: 100, config: GuestHA{State: PointerHaState(HaState_Started)}},
			output: map[string]interface{}{
				"sid":   int(100),
				"state": string("started")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.config.mapToApi(test.input.vmID), test.name)
		})
	}
}

func Test_GuestHA_mapToSDK(t *testing.T) {
	PointerHaState := func(i HaState) *HaState { return &i }             // TODO remove when we have a generic pointer function
	PointerHaGroupName := func(i HaGroupName) *HaGroupName { return &i } // TODO remove when we have a generic pointer function
	PointerHaRelocate := func(i HaRelocate) *HaRelocate { return &i }    // TODO remove when we have a generic pointer function
	PointerHaRestart := func(i HaRestart) *HaRestart { return &i }       // TODO remove when we have a generic pointer function
	PointerString := func(i string) *string { return &i }                // TODO remove when we have a generic pointer function
	tests := []struct {
		name   string
		input  map[string]interface{}
		output GuestHA
	}{
		{name: "Test Comment",
			input:  map[string]interface{}{"comment": "abc"},
			output: GuestHA{Comment: PointerString("abc")}},
		{name: "Test Full",
			input: map[string]interface{}{
				"comment":      "test",
				"group":        "test-group",
				"max_relocate": float64(10),
				"max_restart":  float64(1),
				"state":        "stopped"},
			output: GuestHA{
				Comment:     PointerString("test"),
				Group:       PointerHaGroupName("test-group"),
				Reallocates: PointerHaRelocate(10),
				Restarts:    PointerHaRestart(1),
				State:       PointerHaState(HaState_Stopped)}},
		{name: "Test Group",
			input:  map[string]interface{}{"group": "test-group"},
			output: GuestHA{Group: PointerHaGroupName("test-group")}},
		{name: "Test Reallocates",
			input:  map[string]interface{}{"max_relocate": float64(10)},
			output: GuestHA{Reallocates: PointerHaRelocate(10)}},
		{name: "Test Restarts",
			input:  map[string]interface{}{"max_restart": float64(10)},
			output: GuestHA{Restarts: PointerHaRestart(10)}},
		{name: "Test State",
			input:  map[string]interface{}{"state": "started"},
			output: GuestHA{State: PointerHaState(HaState_Started)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, GuestHA{}.mapToSDK(test.input), test.name)
		})
	}
}

func Test_GuestHA_Set(t *testing.T) {
	type testInput struct {
		vmr *VmRef
		c   *Client
	}
	tests := []struct {
		name  string
		input testInput
	}{
		{name: "* nil", input: testInput{vmr: &VmRef{}}},
		{name: "nil *", input: testInput{c: &Client{}}},
		{name: "nil nil", input: testInput{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { GuestHA{}.Set(test.input.vmr, test.input.c) }, test.name)
		})
	}
}

func Test_GuestHA_Validate(t *testing.T) {
	PointerHaState := func(i HaState) *HaState { return &i }             // TODO remove when we have a generic pointer function
	PointerHaGroupName := func(i HaGroupName) *HaGroupName { return &i } // TODO remove when we have a generic pointer function
	PointerHaRelocate := func(i HaRelocate) *HaRelocate { return &i }    // TODO remove when we have a generic pointer function
	PointerHaRestart := func(i HaRestart) *HaRestart { return &i }       // TODO remove when we have a generic pointer function
	tests := []struct {
		name   string
		input  GuestHA
		output error
	}{
		{name: "Invalid Group", input: GuestHA{Group: PointerHaGroupName("inv&lid")}, output: errors.New(HaGroupName_Error_Illegal)},
		{name: "Invalid Reallocates", input: GuestHA{Reallocates: PointerHaRelocate(11)}, output: errors.New(HaRelocate_Error_UpperBound)},
		{name: "Invalid Restarts", input: GuestHA{Restarts: PointerHaRestart(11)}, output: errors.New(HaRestart_Error_UpperBound)},
		{name: "Invalid State", input: GuestHA{State: PointerHaState("invalid")}, output: errors.New(HaState_Error_Invalid)},
		{name: "Valid", input: GuestHA{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(), test.name)
		})
	}
}

func Test_HaGroupName_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  HaGroupName
		output error
	}{
		{name: "Invalid Length", input: "i", output: errors.New(HaGroupName_Error_Length)},
		{name: "Invalid Illegal", input: "inv&^%lid", output: errors.New(HaGroupName_Error_Illegal)},
		{name: "Invalid Illegal_End", input: "invalid&", output: errors.New(HaGroupName_Error_Illegal_End)},
		{name: "Invalid Illegal_Start", input: "0invalid", output: errors.New(HaGroupName_Error_Illegal_Start)},
		{name: "Valid", input: "valid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(), test.name)
		})
	}
}

func Test_HaRelocate_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  HaRelocate
		output error
	}{
		{name: "Invalid Upper Bound", input: 11, output: errors.New(HaRelocate_Error_UpperBound)},
		{name: "Valid Lower Bound", input: 0},
		{name: "Valid Upper Bound", input: 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(), test.name)
		})
	}
}

func Test_HaRestart_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  HaRestart
		output error
	}{
		{name: "Invalid Upper Bound", input: 11, output: errors.New(HaRestart_Error_UpperBound)},
		{name: "Valid Lower Bound", input: 0},
		{name: "Valid Upper Bound", input: 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(), test.name)
		})
	}
}

func Test_HaState_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  HaState
		output error
	}{
		{name: "Invalid", input: "invalid", output: errors.New(HaState_Error_Invalid)},
		{name: "Valid Disabled", input: HaState_Disabled},
		{name: "Valid Ignored", input: HaState_Ignored},
		{name: "Valid Started", input: HaState_Started},
		{name: "Valid Stopped", input: HaState_Stopped},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate(), test.name)
		})
	}
}

func Test_NewGuestHAFromApi(t *testing.T) {
	type testInput struct {
		vmr *VmRef
		c   *Client
	}
	tests := []struct {
		name   string
		input  testInput
		output GuestHA
	}{
		{name: "* nil", input: testInput{vmr: &VmRef{}}, output: GuestHA{}},
		{name: "nil *", input: testInput{c: &Client{}}, output: GuestHA{}},
		{name: "nil nil", input: testInput{}, output: GuestHA{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { NewGuestHAFromApi(test.input.vmr, test.input.c) }, test.name)
		})
	}
}
