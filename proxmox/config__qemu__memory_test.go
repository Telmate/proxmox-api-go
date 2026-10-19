package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_Memory_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `All float64`,
				input: rawConfigQemu{a: map[string]any{
					"memory":  float64(1024),
					"balloon": float64(512),
					"shares":  float64(50)}},
				output: testQemuBaseConfig_get(ConfigQemu{Memory: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1024)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(512)),
					Shares:             new(QemuMemoryShares(50))}})},
			{name: `All string`,
				input: rawConfigQemu{a: map[string]any{
					"memory":  "1024",
					"balloon": "512",
					"shares":  "50"}},
				output: testQemuBaseConfig_get(ConfigQemu{Memory: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1024)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(512)),
					Shares:             new(QemuMemoryShares(50))}})},
			{name: `memory`,
				input: rawConfigQemu{a: map[string]any{
					"memory": float64(2000)}},
				output: testQemuBaseConfig_get(ConfigQemu{Memory: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(2000))}})},
			{name: `balloon`,
				input: rawConfigQemu{a: map[string]any{
					"balloon": float64(1000)}},
				output: testQemuBaseConfig_get(ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))}})},
			{name: `shares`,
				input: rawConfigQemu{a: map[string]any{
					"shares": float64(100)}},
				output: testQemuBaseConfig_get(ConfigQemu{Memory: &QemuMemory{Shares: new(QemuMemoryShares(100))}})}}
	}
}

func Test_QemuMemory_Validate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		new     QemuMemory
		current *QemuMemory
	}
	tests := []struct {
		name   string
		input  testInput
		output error
	}{
		// there could still be some edge cases that are not covered
		{name: `Valid Create new.CapacityMiB`,
			input: testInput{new: QemuMemory{CapacityMiB: new(QemuMemoryCapacity(qemuMemoryCapacity_Max))}}},
		{name: `Valid Update new.CapacityMiB smaller then current.MinimumCapacityMiB`,
			input: testInput{
				new:     QemuMemory{CapacityMiB: new(QemuMemoryCapacity(1000))},
				current: &QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(2000))}}},
		{name: `Valid Update new.CapacityMiB smaller then current.MinimumCapacityMiB and MinimumCapacityMiB lowered`,
			input: testInput{
				new:     QemuMemory{CapacityMiB: new(QemuMemoryCapacity(1000)), MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))},
				current: &QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(2000))}}},
		{name: `Valid Create new.MinimumCapacityMiB`,
			input: testInput{new: QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))}}},
		{name: `Valid Update new.CapacityMiB == new.MinimumCapacityMiB && new.CapacityMiB > current.CapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1500)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1500))},
				current: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(1000))}}},
		{name: `Valid Update new.MinimumCapacityMiB > current.MinimumCapacityMiB && new.MinimumCapacityMiB < new.CapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(3000)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(2000))},
				current: &QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1500))}}},
		{name: `Valid Create new.Shares(qemuMemoryShares_Max) new.CapacityMiB & new.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1001)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000)),
					Shares:             new(QemuMemoryShares(qemuMemoryShares_Max))}}},
		{name: `Valid Create new.Shares new.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000)),
					Shares:             new(QemuMemoryShares(0))}}},
		{name: `Valid Create new.Shares(0) new.CapacityMiB & new.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1001)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000)),
					Shares:             new(QemuMemoryShares(0))}}},
		{name: `Valid Update new.Shares(0) current.CapacityMiB == current.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{Shares: new(QemuMemoryShares(0))},
				current: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1000)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))}}},
		{name: `Invalid Create new.CapacityMiB(0)`,
			input:  testInput{new: QemuMemory{CapacityMiB: new(QemuMemoryCapacity(0))}},
			output: errors.New(QemuMemoryCapacity_Error_Minimum)},
		{name: `Invalid Update new.CapacityMiB(0)`,
			input: testInput{
				new:     QemuMemory{CapacityMiB: new(QemuMemoryCapacity(0))},
				current: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(1000))}},
			output: errors.New(QemuMemoryCapacity_Error_Minimum)},
		{name: `Invalid Create new.CapacityMiB > qemuMemoryCapacity_Max`,
			input:  testInput{new: QemuMemory{CapacityMiB: new(QemuMemoryCapacity(qemuMemoryCapacity_Max + 1))}},
			output: errors.New(QemuMemoryCapacity_Error_Maximum)},
		{name: `Invalid Update new.CapacityMiB > qemuMemoryCapacity_Max`,
			input: testInput{
				new:     QemuMemory{CapacityMiB: new(QemuMemoryCapacity(qemuMemoryCapacity_Max + 1))},
				current: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(qemuMemoryCapacity_Max))}},
			output: errors.New(QemuMemoryCapacity_Error_Maximum)},
		{name: `Invalid Update new.CapacityMiB(0)`,
			input: testInput{
				new:     QemuMemory{CapacityMiB: new(QemuMemoryCapacity(0))},
				current: &QemuMemory{CapacityMiB: new(QemuMemoryCapacity(1000))}},
			output: errors.New(QemuMemoryCapacity_Error_Minimum)},
		{name: `Invalid Create new.MinimumCapacityMiB to big`,
			input:  testInput{new: QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(qemuMemoryBalloonCapacity_Max + 1))}},
			output: errors.New(QemuMemoryBalloonCapacity_Error_Invalid)},
		{name: `Invalid Update new.MinimumCapacityMiB to big`,
			input: testInput{
				new:     QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(qemuMemoryBalloonCapacity_Max + 1))},
				current: &QemuMemory{MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))}},
			output: errors.New(QemuMemoryBalloonCapacity_Error_Invalid)},
		{name: `Invalid Create new.MinimumCapacityMiB > new.CapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1000)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(2000))}},
			output: errors.New(QemuMemory_Error_MinimumCapacityMiB_GreaterThan_CapacityMiB)},
		{name: `Invalid Create new.Shares(1)`,
			input:  testInput{new: QemuMemory{Shares: new(QemuMemoryShares(1))}},
			output: errors.New(QemuMemory_Error_NoMemoryCapacity)},
		{name: `Invalid Create new.Shares() too big and new.CapacityMiB & new.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1001)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000)),
					Shares:             new(QemuMemoryShares(qemuMemoryShares_Max + 1))}},
			output: errors.New(QemuMemoryShares_Error_Invalid)},
		{name: `Invalid Update new.Shares() too big and new.CapacityMiB & new.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1001)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000)),
					Shares:             new(QemuMemoryShares(qemuMemoryShares_Max + 1))},
				current: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(512)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(256)),
					Shares:             new(QemuMemoryShares(1))}},
			output: errors.New(QemuMemoryShares_Error_Invalid)},
		{name: `Invalid Create new.Shares(1) when new.CapacityMiB == new.MinimumCapacityMiB`,
			input:  testInput{new: QemuMemory{Shares: new(QemuMemoryShares(1))}},
			output: errors.New(QemuMemory_Error_NoMemoryCapacity)},
		{name: `Invalid Update new.Shares(1) when current.CapacityMiB == current.MinimumCapacityMiB`,
			input: testInput{
				new: QemuMemory{Shares: new(QemuMemoryShares(1))},
				current: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(1000)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1000))}},
			output: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)},
		{name: `Invalid Update new.Shares(1) new.CapacityMiB == current.MinimumCapacityMiB & MinimumCapacityMiB not updated`,
			input: testInput{
				new: QemuMemory{
					CapacityMiB: new(QemuMemoryCapacity(1024)),
					Shares:      new(QemuMemoryShares(1))},
				current: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(2048)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1024))}},
			output: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)},
		{name: `Invalid Update new.Shares(1) new.MinimumCapacityMiB == current.CapacityMiB & CapacityMiB not updated`,
			input: testInput{
				new: QemuMemory{
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(2048)),
					Shares:             new(QemuMemoryShares(1))},
				current: &QemuMemory{
					CapacityMiB:        new(QemuMemoryCapacity(2048)),
					MinimumCapacityMiB: new(QemuMemoryBalloonCapacity(1024))}},
			output: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.new.Validate(test.input.current))
		})
	}
}

func Test_QemuMemoryBalloonCapacity_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuMemoryBalloonCapacity
		output error
	}{
		{name: `Valid`,
			input: qemuMemoryBalloonCapacity_Max},
		{name: `Invalid Max`,
			input:  qemuMemoryBalloonCapacity_Max + 1,
			output: errors.New(QemuMemoryBalloonCapacity_Error_Invalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuMemoryCapacity_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuMemoryCapacity
		output error
	}{
		{name: `Valid`,
			input: qemuMemoryCapacity_Max},
		{name: `Invalid Max`,
			input:  qemuMemoryCapacity_Max + 1,
			output: errors.New(QemuMemoryCapacity_Error_Maximum)},
		{name: `Invalid Min`,
			input:  0,
			output: errors.New(QemuMemoryCapacity_Error_Minimum)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_QemuMemoryShares_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  QemuMemoryShares
		output error
	}{
		{name: `Valid`,
			input: qemuMemoryShares_Max,
		},
		{name: `Invalid`,
			input:  qemuMemoryShares_Max + 1,
			output: errors.New(QemuMemoryShares_Error_Invalid),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.output, test.input.Validate())
		})
	}
}

func Test_RawConfigQemu_GetMemory(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Memory_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.Memory, raw.GetMemory())
		})
}

func Test_RawConfigQemu_Memory_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Memory_Get().Test(t)
}
