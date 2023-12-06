package proxmox

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/test/data/test_data_snapshot"
	"github.com/stretchr/testify/require"
)

func Test_ConfigSnapshot_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input ConfigSnapshot
		err   bool
	}{
		// Valid
		{name: "Valid ConfigSnapshot",
			input: ConfigSnapshot{Name: SnapshotName(test_data_snapshot.SnapshotName_Max_Legal())},
		},
		// Invalid
		{name: "Invalid ConfigSnapshot",
			input: ConfigSnapshot{Name: SnapshotName(test_data_snapshot.SnapshotName_Max_Illegal())},
			err:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			if test.err {
				require.Error(t, test.input.Validate(), test.name)
			} else {
				require.NoError(t, test.input.Validate(), test.name)
			}
		})
	}
}

// TODO rename this test
// Test the formatting logic to build the tree of snapshots
func Test_FormatSnapshotsTree(t *testing.T) {
	input := test_FormatSnapshots_Input()
	output := test_FormatSnapshotsTree_Output()
	for i, e := range input {
		result, _ := json.Marshal(e.FormatSnapshotsTree())
		require.JSONEq(t, output[i], string(result))
	}
}

// TODO rename this test
// Test the formatting logic to build the list of snapshots
func Test_FormatSnapshotsList(t *testing.T) {
	input := test_FormatSnapshots_Input()
	output := test_FormatSnapshotsList_Output()
	for i, e := range input {
		result, _ := json.Marshal(e.FormatSnapshotsList())
		require.JSONEq(t, output[i], string(result))
	}
}

func test_FormatSnapshots_Input() []rawSnapshots {
	return []rawSnapshots{{map[string]interface{}{
		"name":        "aa",
		"snaptime":    float64(1666361849),
		"description": "",
		"parent":      "",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaa",
		"snaptime":    float64(1666361866),
		"description": "",
		"parent":      "aa",
		"vmstate":     float64(1),
	}, map[string]interface{}{
		"name":        "aaaa",
		"snaptime":    float64(1666362071),
		"description": "123456",
		"parent":      "aaa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaab",
		"snaptime":    float64(1666362062),
		"description": "",
		"parent":      "aaa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaac",
		"snaptime":    float64(1666361873),
		"description": "",
		"parent":      "aaa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaad",
		"snaptime":    float64(1666361937),
		"description": "abcdefg",
		"parent":      "aaa",
		"vmstate":     float64(1),
	}, map[string]interface{}{
		"name":        "aaae",
		"snaptime":    float64(1666362084),
		"description": "",
		"parent":      "aaa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "current",
		"description": "You are here!",
		"parent":      "aaae",
	}, map[string]interface{}{
		"name":        "aab",
		"snaptime":    float64(1666361920),
		"description": "",
		"parent":      "aa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaba",
		"snaptime":    float64(1666361952),
		"description": "",
		"parent":      "aab",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aabaa",
		"snaptime":    float64(1666361960),
		"description": "",
		"parent":      "aaba",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aac",
		"snaptime":    float64(1666361896),
		"description": "",
		"parent":      "aa",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aaca",
		"snaptime":    float64(1666361988),
		"description": "!@#()&",
		"parent":      "aac",
		"vmstate":     float64(1),
	}, map[string]interface{}{
		"name":        "aacaa",
		"snaptime":    float64(1666362006),
		"description": "",
		"parent":      "aaca",
		"vmstate":     float64(1),
	}, map[string]interface{}{
		"name":        "aacb",
		"snaptime":    float64(1666361977),
		"description": "",
		"parent":      "aac",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aacba",
		"snaptime":    float64(1666362021),
		"description": "QWERTY",
		"parent":      "aacb",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aacc",
		"snaptime":    float64(1666361904),
		"description": "",
		"parent":      "aac",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "aacca",
		"snaptime":    float64(1666361910),
		"description": "",
		"parent":      "aacc",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "bb",
		"snaptime":    float64(1666361866),
		"description": "aA1!",
		"parent":      "",
		"vmstate":     float64(1),
	}, map[string]interface{}{
		"name":        "bba",
		"snaptime":    float64(1666362071),
		"description": "",
		"parent":      "bb",
		"vmstate":     float64(0),
	}, map[string]interface{}{
		"name":        "bbb",
		"snaptime":    float64(1666362062),
		"description": "",
		"parent":      "bb",
		"vmstate":     float64(0),
	}}}
}

func test_FormatSnapshotsTree_Output() []string {
	return []string{`[{
		"name":"aa","time":1666361849,"children":[{
			"name":"aaa","time":1666361866,"ram":true,"children":[{
				"name":"aaaa","time":1666362071,"description":"123456"},{
				"name":"aaab","time":1666362062},{
				"name":"aaac","time":1666361873},{
				"name":"aaad","time":1666361937,"description":"abcdefg","ram":true},{
				"name":"aaae","time":1666362084,"children":[{
					"name":"current","description":"You are here!"}]}]},{
			"name":"aab","time":1666361920,"children":[{
				"name":"aaba","time":1666361952,"children":[{
					"name":"aabaa","time":1666361960}]}]},{
			"name":"aac","time":1666361896,"children":[{
				"name":"aaca","time":1666361988,"description":"!@#()\u0026","ram":true,"children":[{
					"name":"aacaa","time":1666362006,"ram":true}]},{
				"name":"aacb","time":1666361977,"children":[{
					"name":"aacba","time":1666362021,"description":"QWERTY"}]},{
				"name":"aacc","time":1666361904,"children":[{
					"name":"aacca","time":1666361910}]}]}]},{
		"name":"bb","time":1666361866,"description":"aA1!","ram":true,"children":[{
			"name":"bba","time":1666362071},{
			"name":"bbb","time":1666362062}]}]`}
}

func test_FormatSnapshotsList_Output() []string {
	return []string{`[{
		"name":"aa","time":1666361849},{
		"name":"aaa","time":1666361866,"ram":true,"parent":"aa"},{
		"name":"aaaa","time":1666362071,"description":"123456","parent":"aaa"},{
		"name":"aaab","time":1666362062,"parent":"aaa"},{
		"name":"aaac","time":1666361873,"parent":"aaa"},{
		"name":"aaad","time":1666361937,"description":"abcdefg","ram":true,"parent":"aaa"},{
		"name":"aaae","time":1666362084,"parent":"aaa"},{
		"name":"current","description":"You are here!","parent":"aaae"},{
		"name":"aab","time":1666361920,"parent":"aa"},{
		"name":"aaba","time":1666361952,"parent":"aab"},{
		"name":"aabaa","time":1666361960,"parent":"aaba"},{
		"name":"aac","time":1666361896,"parent":"aa"},{
		"name":"aaca","time":1666361988,"description":"!@#()\u0026","ram":true,"parent":"aac"},{
		"name":"aacaa","time":1666362006,"ram":true,"parent":"aaca"},{
		"name":"aacb","time":1666361977,"parent":"aac"},{
		"name":"aacba","time":1666362021,"description":"QWERTY","parent":"aacb"},{
		"name":"aacc","time":1666361904,"parent":"aac"},{
		"name":"aacca","time":1666361910,"parent":"aacc"},{
		"name":"bb","time":1666361866,"description":"aA1!","ram":true},{
		"name":"bba","time":1666362071,"parent":"bb"},{
		"name":"bbb","time":1666362062,"parent":"bb"}]`}
}

func Test_SnapshotName_Validate(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		err   error
	}{
		// Valid
		{name: "Valid", input: test_data_snapshot.SnapshotName_Legal()},
		// Invalid
		{name: "Invalid SnapshotName_Error_MinLength",
			input: []string{"", test_data_snapshot.SnapshotName_Min_Illegal()},
			err:   errors.New(SnapshotName_Error_MinLength),
		},
		{name: "Invalid SnapshotName_Error_MaxLength",
			input: []string{test_data_snapshot.SnapshotName_Max_Illegal()},
			err:   errors.New(SnapshotName_Error_MaxLength),
		},
		{name: "Invalid SnapshotName_Error_StartNoLetter",
			input: test_data_snapshot.SnapshotName_Start_Illegal(),
			err:   errors.New(SnapshotName_Error_StartNoLetter),
		},
		{name: "Invalid SnapshotName_Error_StartNoLetter",
			input: test_data_snapshot.SnapshotName_Character_Illegal(),
			err:   errors.New(SnapshotName_Error_IllegalCharacters),
		},
	}
	for _, test := range tests {
		for _, snapshot := range test.input {
			t.Run(test.name+" :"+snapshot, func(*testing.T) {
				require.Equal(t, SnapshotName(snapshot).Validate(), test.err, test.name+" :"+snapshot)
			})
		}
	}
}
