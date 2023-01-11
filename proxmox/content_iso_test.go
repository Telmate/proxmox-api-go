package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigContent_Iso_error(t *testing.T) {
	require.Equal(t, errors.New("the value of (Node) may not be empty"), ConfigContent_Iso{}.error("Node"))
}

func Test_ConfigContent_Iso_mapToApiValues(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Iso
		output map[string]interface{}
	}{
		{
			input: ConfigContent_Iso{
				Checksum:          "xxx",
				ChecksumAlgorithm: "sha512",
				DownloadUrl:       "https://eample.com/distro.iso",
				Filename:          "distro.iso",
				Storage:           "local",
			},
			output: map[string]interface{}{
				"checksum-algorithm": "sha512",
				"checksum":           "xxx",
				"content":            "iso",
				"filename":           "distro.iso",
				"storage":            "local",
				"url":                "https://eample.com/distro.iso",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.mapToApiValues())
	}
}

func Test_ConfigContent_Iso_Validate(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Iso
		output error
	}{
		{
			input:  ConfigContent_Iso{},
			output: ConfigContent_Iso{}.error("Node"),
		},
		{
			input:  ConfigContent_Iso{Node: "notEmpty"},
			output: ConfigContent_Iso{}.error("Storage"),
		},
		{
			input: ConfigContent_Iso{
				Node:    "notEmpty",
				Storage: "notEmpty",
			},
			output: ConfigContent_Iso{}.error("URL"),
		},
		{
			input: ConfigContent_Iso{
				Node:        "notEmpty",
				Storage:     "notEmpty",
				DownloadUrl: "notEmpty",
			},
			output: ConfigContent_Iso{}.error("Filename"),
		},
		{
			input: ConfigContent_Iso{
				Node:        "notEmpty",
				Storage:     "notEmpty",
				DownloadUrl: "notEmpty",
				Filename:    "notEmpty",
			},
			output: nil,
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.Validate())
	}
}
