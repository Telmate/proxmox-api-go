package proxmox

import (
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_tag"
	"github.com/stretchr/testify/require"
)

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
