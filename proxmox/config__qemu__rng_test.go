package proxmox

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_RandomnessDevice_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `Source /dev/random`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "source=/dev/random"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(EntropySourceRandom)}})},
			{name: `Source /dev/urandom`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "source=/dev/urandom"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(EntropySourceURandom)}})},
			{name: `Source /dev/hwrng`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "source=/dev/hwrng"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(EntropySourceHwRNG)}})},
			{name: `Limit`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "max_bytes=1000"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(entropySourceInvalid),
					Limit:  new(uint(1000))}})},
			{name: `Period`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "period=1000"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(entropySourceInvalid),
					Period: new(time.Duration(1000) * time.Millisecond)}})},
			{name: `all`,
				input: rawConfigQemu{a: map[string]any{
					"rng0": "period=1024,max_bytes=1000,source=/dev/urandom"}},
				output: testQemuBaseConfig_get(ConfigQemu{RandomnessDevice: &VirtIoRNG{
					Source: new(EntropySourceURandom),
					Limit:  new(uint(1000)),
					Period: new(time.Duration(1024) * time.Millisecond)}})}}
	}
}

func Test_EntropySource_MarshalJSON(t *testing.T) {
	t.Parallel()
	type testData struct {
		Source EntropySource `json:"source"`
	}
	tests := []struct {
		name   string
		input  testData
		output []byte
		err    error
	}{
		{name: `Random`,
			input:  testData{Source: EntropySourceRandom},
			output: []byte(`{"source":"/dev/random"}`)},
		{name: `URandom`,
			input:  testData{Source: EntropySourceURandom},
			output: []byte(`{"source":"/dev/urandom"}`)},
		{name: `HwRNG`,
			input:  testData{Source: EntropySourceHwRNG},
			output: []byte(`{"source":"/dev/hwrng"}`)},
		{name: `Invalid`,
			input: testData{Source: entropySourceInvalid},
			err:   errors.New(EntropySourceErrorInvalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := json.Marshal(test.input)
			require.Equal(t, test.output, output)
			if test.err == nil {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, test.err.Error())
			}
		})
	}
}

func Test_EntropySource_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	type testData struct {
		EntropySource EntropySource `json:"entropySource"`
	}
	tests := []struct {
		name   string
		input  string
		output testData
		err    error
	}{
		{name: `Random`,
			input:  `{"entropySource":"/dev/random"}`,
			output: testData{EntropySource: EntropySourceRandom}},
		{name: `URandom`,
			input:  `{"entropySource":"/dev/urandom"}`,
			output: testData{EntropySource: EntropySourceURandom}},
		{name: `HwRNG`,
			input:  `{"entropySource":"/dev/hwrng"}`,
			output: testData{EntropySource: EntropySourceHwRNG}},
		{name: `Invalid`,
			input: `{"entropySource":"invalid"}`,
			err:   errors.New(EntropySourceErrorInvalid)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output testData
			err := json.Unmarshal([]byte(test.input), &output)
			require.Equal(t, test.output, output)
			require.Equal(t, test.err, err)
		})
	}
}

func Test_RawConfigQemu_GetRandomnessDevice(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_RandomnessDevice_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.RandomnessDevice, raw.GetRandomnessDevice())
		})
}

func Test_RawConfigQemu_RandomnessDevice_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_RandomnessDevice_Get().Test(t)
}
