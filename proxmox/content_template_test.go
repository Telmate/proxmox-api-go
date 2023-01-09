package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ConfigContent_Template_error(t *testing.T) {
	require.Equal(t, errors.New("the value of (Node) may not be empty"), ConfigContent_Template{}.error("Node"))
}

func Test_ConfigContent_Template_mapToApiValues(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Template
		output map[string]interface{}
	}{
		{
			input: ConfigContent_Template{
				Storage:  "a",
				Template: "b",
			},
			output: map[string]interface{}{
				"storage":  "a",
				"template": "b",
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.mapToApiValues())
	}
}

func Test_ConfigContent_Template_Validate(t *testing.T) {
	testData := []struct {
		input  ConfigContent_Template
		output error
	}{
		{
			input:  ConfigContent_Template{},
			output: ConfigContent_Template{}.error("Node"),
		},
		{
			input:  ConfigContent_Template{Node: "notEmpty"},
			output: ConfigContent_Template{}.error("Storage"),
		},
		{
			input: ConfigContent_Template{
				Node:    "notEmpty",
				Storage: "notEmpty",
			},
			output: ConfigContent_Template{}.error("Template"),
		},
		{
			input: ConfigContent_Template{
				Node:     "notEmpty",
				Storage:  "notEmpty",
				Template: "notEmpty",
			},
			output: nil,
		},
	}
	for _, e := range testData {
		require.Equal(t, e.output, e.input.Validate())
	}
}

func Test_createTemplateList(t *testing.T) {
	testData := []struct {
		Input  []interface{}
		Output *[]TemplateItem
	}{
		{
			Input: []interface{}{
				map[string]interface{}{
					"location":    "http://mirror.turnkeylinux.org/turnkeylinux/images/proxmox/debian-11-turnkey-drupal7_17.1-1_amd64.tar.gz",
					"manageurl":   "http://__IPADDRESS__/",
					"description": "TurnKey Drupal 7 - Content Management Framework",
					"version":     "17.1-1",
					"template":    "debian-11-turnkey-drupal7_17.1-1_amd64.tar.gz",
					"headline":    "TurnKey Drupal 7",
					"source":      "https://releases.turnkeylinux.org/pve",
					"package":     "turnkey-drupal7",
					"section":     "turnkeylinux",
				},
				map[string]interface{}{
					"infopage":     "http://www.turnkeylinux.org/mumble",
					"section":      "turnkeylinux",
					"headline":     "TurnKey Mumble",
					"manageurl":    "http://__IPADDRESS__/",
					"sha512sum":    "30a81eec66968817b9e06502141a476d0707643e3649034f0edc69692fcbff6ccde49e43ec9ef6ee09e2655faab1202e915428cc30e77af30749dabbac944d28",
					"type":         "lxc",
					"architecture": "amd64",
					"os":           "debian-11",
					"template":     "debian-11-turnkey-mumble_17.1-1_amd64.tar.gz",
				},
				map[string]interface{}{
					"location":     "http://mirror.turnkeylinux.org/turnkeylinux/images/proxmox/debian-11-turnkey-mattermost_17.1-1_amd64.tar.gz",
					"type":         "lxc",
					"os":           "debian-11",
					"package":      "turnkey-mattermost",
					"sha512sum":    "5b6d3cdd8fac06494cd094dac36a9b2820889c4b3e0ef033a17b3f4d82d12390fccbcb0ab1425472595d163c6ce28c047bda5bd8e4e9c551a7db2744ada3c130",
					"architecture": "amd64",
					"description":  "TurnKey Mattermost - Open Source, self-hosted Slack-alternative",
					"version":      "17.1-1",
					"infopage":     "http://www.turnkeylinux.org/mattermost",
					"source":       "https://releases.turnkeylinux.org/pve",
				},
			},
			Output: &[]TemplateItem{
				{
					Architecture:   "",
					Description:    "TurnKey Drupal 7 - Content Management Framework",
					Headline:       "TurnKey Drupal 7",
					InfoPage:       "",
					Location:       "http://mirror.turnkeylinux.org/turnkeylinux/images/proxmox/debian-11-turnkey-drupal7_17.1-1_amd64.tar.gz",
					ManageURL:      "http://__IPADDRESS__/",
					OS:             "",
					Package:        "turnkey-drupal7",
					Section:        "turnkeylinux",
					SHA512Checksum: "",
					Source:         "https://releases.turnkeylinux.org/pve",
					Template:       "debian-11-turnkey-drupal7_17.1-1_amd64.tar.gz",
					Type:           "",
					Version:        "17.1-1",
				},
				{
					Architecture:   "amd64",
					Description:    "",
					Headline:       "TurnKey Mumble",
					InfoPage:       "http://www.turnkeylinux.org/mumble",
					Location:       "",
					ManageURL:      "http://__IPADDRESS__/",
					OS:             "debian-11",
					Package:        "",
					Section:        "turnkeylinux",
					SHA512Checksum: "30a81eec66968817b9e06502141a476d0707643e3649034f0edc69692fcbff6ccde49e43ec9ef6ee09e2655faab1202e915428cc30e77af30749dabbac944d28",
					Source:         "",
					Template:       "debian-11-turnkey-mumble_17.1-1_amd64.tar.gz",
					Type:           "lxc",
					Version:        "",
				},
				{
					Architecture:   "amd64",
					Description:    "TurnKey Mattermost - Open Source, self-hosted Slack-alternative",
					Headline:       "",
					InfoPage:       "http://www.turnkeylinux.org/mattermost",
					Location:       "http://mirror.turnkeylinux.org/turnkeylinux/images/proxmox/debian-11-turnkey-mattermost_17.1-1_amd64.tar.gz",
					ManageURL:      "",
					OS:             "debian-11",
					Package:        "turnkey-mattermost",
					Section:        "",
					SHA512Checksum: "5b6d3cdd8fac06494cd094dac36a9b2820889c4b3e0ef033a17b3f4d82d12390fccbcb0ab1425472595d163c6ce28c047bda5bd8e4e9c551a7db2744ada3c130",
					Source:         "https://releases.turnkeylinux.org/pve",
					Template:       "",
					Type:           "lxc",
					Version:        "17.1-1",
				},
			},
		},
	}
	for _, e := range testData {
		require.Equal(t, e.Output, createTemplateList(e.Input))
	}
}
