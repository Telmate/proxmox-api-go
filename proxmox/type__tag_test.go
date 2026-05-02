package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_tag"
	"github.com/stretchr/testify/require"
)

func test_TagMapToSDK_data() []struct {
	name   string
	input  string
	output Tags
} {
	return []struct {
		name   string
		input  string
		output Tags
	}{
		{name: `Whitespace`, // Handle Proxmox API bug: sometimes returns " " (whitespace) for VMs with no tags
			input: " "},
		{name: `Comma`,
			input:  "Test,a,BBB,cC",
			output: Tags{"Test", "a", "BBB", "cC"}},
		{name: `Semicolon`,
			input:  "Test;a;BBB;cC",
			output: Tags{"Test", "a", "BBB", "cC"}},
		{name: `Mixed`,
			input:  "Test;a,BBB,cC;x;bla;pve,pbs",
			output: Tags{"Test", "a", "BBB", "cC", "x", "bla", "pve", "pbs"}},
		{name: `Single a`,
			input:  "a",
			output: Tags{"a"}},
		{name: `Single aaa`,
			input:  "aaa",
			output: Tags{"aaa"}}}
}

func Benchmark_Tag_mapToSDK(b *testing.B) {
	tests := test_TagMapToSDK_data()
	for b.Loop() {
		for _, test := range tests {
			var tags Tags
			tags.mapToSDK(test.input)
		}
	}
}

func Test_Tag_mapToSDK(t *testing.T) {
	t.Parallel()
	tests := test_TagMapToSDK_data()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tags Tags
			tags.mapToSDK(test.input)
			require.Equal(t, test.output, tags)
		})
	}
}

func Test_Tag_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid Tag`,
			input:  test_data_tag.Tag_Legal(),
			output: nil,
		},
		{name: `Invalid Tag`,
			input:  test_data_tag.Tag_Character_Illegal(),
			output: errors.New(Tag_Error_Invalid),
		},
		{name: `Invalid Tag Empty`,
			input:  []string{test_data_tag.Tag_Empty()},
			output: errors.New(Tag_Error_Empty),
		},
		{name: `Invalid Tag Max Length`,
			input:  []string{test_data_tag.Tag_Max_Illegal()},
			output: errors.New(Tag_Error_MaxLength),
		},
	}
	for _, test := range tests {
		for _, e := range test.input {
			t.Run(test.name+": "+e, func(t *testing.T) {
				require.Equal(t, test.output, (Tag(e)).Validate())
			})
		}
	}
}

func testDataTagsMapToAPI() qemuTestsAPI {
	return qemuTestsAPI{category: `Tags`,
		create: []qemuTestCaseAPI{
			{name: `do nothing`,
				config: &ConfigQemu{Tags: new(Tags{})}}},
		createUpdate: []qemuTestCaseAPI{
			{name: `set`,
				currentUpdate: configQemuUpdate{raw: &rawConfigQemu{a: map[string]any{
					"tags": "tag5;tag6"}}},
				config: &ConfigQemu{Tags: new(Tags{"tag1", "tag2"})},
				body:   map[string]string{"tags": "tag1,tag2"}}},
		update: []qemuTestCaseAPI{
			{name: `create`,
				currentUpdate: configQemuUpdate{raw: &rawConfigQemu{a: map[string]any{}}},
				config:        &ConfigQemu{Tags: new(Tags{"tag1", "tag2"})},
				body:          map[string]string{"tags": "tag1,tag2"}},
			{name: `empty`,
				currentUpdate: configQemuUpdate{raw: &rawConfigQemu{a: map[string]any{
					"tags": "tag5;tag6"}}},
				config: &ConfigQemu{Tags: new(Tags{})},
				body:   map[string]string{"tags": ""}},
			{name: `do nothing`,
				currentUpdate: configQemuUpdate{raw: &rawConfigQemu{a: map[string]any{
					"tags": "tag5;tag6"}}},
				config: &ConfigQemu{Tags: new(Tags{"tag5", "tag6"})}},
			{name: `different order, do nothing`,
				currentUpdate: configQemuUpdate{raw: &rawConfigQemu{a: map[string]any{
					"tags": "tag5;tag623;tag2;tag8"}}},
				config: &ConfigQemu{Tags: new(Tags{"tag2", "tag8", "tag623", "tag5"})}}},
	}
}
