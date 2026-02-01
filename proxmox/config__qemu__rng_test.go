package proxmox

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

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
