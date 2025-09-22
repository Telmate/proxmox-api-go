package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/array"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/stretchr/testify/require"
)

// This tests quite a bit as it tests the following:
// Getting the list from the API.
// Getting the specific rule types from the generic HaRule type.
// Getting the fields from the generic HaRule type.
// Getting the actual config from the specific rule types.
// This way we test the full path from the API to the actual config that the user wants.
func Test_HaRule_Get(t *testing.T) {
	baseNodeRule := func(r HaNodeAffinityRule) HaNodeAffinityRule {
		if r.Comment == nil {
			r.Comment = util.Pointer("")
		}
		if r.Enabled == nil {
			r.Enabled = util.Pointer(true)
		}
		if r.Guests == nil {
			r.Guests = util.Pointer(array.Nil[VmRef]())
		}
		if r.Nodes == nil {
			r.Nodes = util.Pointer(array.Nil[HaNode]())
		}
		if r.Strict == nil {
			r.Strict = util.Pointer(false)
		}
		return r
	}
	baseResourceRule := func(r HaResourceAffinityRule) HaResourceAffinityRule {
		if r.Affinity == nil {
			r.Affinity = util.Pointer(HaAffinity(0))
		}
		if r.Comment == nil {
			r.Comment = util.Pointer("")
		}
		if r.Enabled == nil {
			r.Enabled = util.Pointer(true)
		}
		if r.Guests == nil {
			r.Guests = util.Pointer(array.Nil[VmRef]())
		}
		return r
	}
	type OutputNode struct {
		id   HaRuleID
		rule HaNodeAffinityRule
	}
	type OutputResource struct {
		id   HaRuleID
		rule HaResourceAffinityRule
	}
	type test struct {
		name           string
		input          []map[string]any
		outputNode     *OutputNode
		outputResource *OutputResource
		err            error
	}
	tests := []struct {
		category string
		tests    []test
	}{
		{category: `Affinity`,
			tests: []test{
				{name: `Positive`,
					input: []map[string]any{
						{"rule": string("ha-rule-1"),
							"type":     string("resource-affinity"),
							"affinity": string("positive")}},
					outputResource: &OutputResource{
						id: "ha-rule-1",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:       "ha-rule-1",
							Affinity: util.Pointer(HaAffinity(1))})}},
				{name: `Negative`,
					input: []map[string]any{
						{"rule": string("ha-rule-2"),
							"type":     string("resource-affinity"),
							"affinity": string("negative")}},
					outputResource: &OutputResource{
						id: "ha-rule-2",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:       "ha-rule-2",
							Affinity: util.Pointer(HaAffinity(-1))})}}}},
		{category: `Comment`,
			tests: []test{
				{
					input: []map[string]any{
						{"rule": string("ha-rule-1"),
							"type":    string("node-affinity"),
							"comment": string("This is a comment")},
						{"rule": string("ha-rule-2"),
							"type":    string("resource-affinity"),
							"comment": string("Test resource comment")}},
					outputNode: &OutputNode{
						id: "ha-rule-1",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:      "ha-rule-1",
							Comment: util.Pointer("This is a comment")})},
					outputResource: &OutputResource{
						id: "ha-rule-2",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:      "ha-rule-2",
							Comment: util.Pointer("Test resource comment"),
						})},
				}}},
		{category: `Digest`,
			tests: []test{
				{
					input: []map[string]any{
						{"rule": string("test"),
							"type":   string("node-affinity"),
							"digest": string("ebefeede3059417444308d2e58d2a5e504fe6151")},
						{"rule": string("test2"),
							"type":   string("resource-affinity"),
							"digest": string("676544e695863189f91f09b12107813b394347fd")}},
					outputNode: &OutputNode{
						id: "test",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:     "test",
							Digest: [20]byte{0xeb, 0xef, 0xee, 0xde, 0x30, 0x59, 0x41, 0x74, 0x44, 0x30, 0x8d, 0x2e, 0x58, 0xd2, 0xa5, 0xe5, 0x04, 0xfe, 0x61, 0x51}})},
					outputResource: &OutputResource{
						id: "test2",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:     "test2",
							Digest: [20]byte{0x67, 0x65, 0x44, 0xe6, 0x95, 0x86, 0x31, 0x89, 0xf9, 0x1f, 0x09, 0xb1, 0x21, 0x07, 0x81, 0x3b, 0x39, 0x43, 0x47, 0xfd}})}}}},
		{category: `Enabled`,
			tests: []test{
				{name: "false",
					input: []map[string]any{
						{"rule": string("enabled-false"),
							"type":    string("node-affinity"),
							"disable": string("1")},
						{"rule": string("enabled-false-2"),
							"type":    string("resource-affinity"),
							"disable": string("1")}},
					outputNode: &OutputNode{
						id: "enabled-false",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:      "enabled-false",
							Enabled: util.Pointer(false)})},
					outputResource: &OutputResource{
						id: "enabled-false-2",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:      "enabled-false-2",
							Enabled: util.Pointer(false)})}},
				{name: "true",
					input: []map[string]any{
						{"rule": string("enabled-true-2"),
							"type":    string("resource-affinity"),
							"disable": string("0")},
						{"rule": string("enabled-true"),
							"type":    string("node-affinity"),
							"disable": string("0")}},
					outputNode: &OutputNode{
						id: "enabled-true",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:      "enabled-true",
							Enabled: util.Pointer(true)})},
					outputResource: &OutputResource{
						id: "enabled-true-2",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:      "enabled-true-2",
							Enabled: util.Pointer(true)})}}}},
		{category: `Error`,
			tests: []test{
				{name: "api error",
					err: errors.New("api error")}}},
		{category: `Guests`,
			tests: []test{
				{name: `Multiple`,
					input: []map[string]any{
						{"rule": string("guest-multiple"),
							"type":      string("node-affinity"),
							"resources": string("vm:100,ct:200,vm:300")},
						{"rule": string("guest-multiple-resource"),
							"type":      string("resource-affinity"),
							"resources": string("vm:100,ct:200,vm:300,vm:101")}},
					outputNode: &OutputNode{
						id: "guest-multiple",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID: "guest-multiple",
							Guests: &[]VmRef{
								{vmType: GuestQemu, vmId: 100},
								{vmType: GuestLxc, vmId: 200},
								{vmType: GuestQemu, vmId: 300}}})},
					outputResource: &OutputResource{
						id: "guest-multiple-resource",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID: "guest-multiple-resource",
							Guests: &[]VmRef{
								{vmType: GuestQemu, vmId: 100},
								{vmType: GuestLxc, vmId: 200},
								{vmType: GuestQemu, vmId: 300},
								{vmType: GuestQemu, vmId: 101}}})}},
				{name: `Single`,
					input: []map[string]any{
						{"rule": string("guest-single"),
							"type":      string("node-affinity"),
							"resources": string("ct:342")},
						{"rule": string("guest-single-resource"),
							"type":      string("resource-affinity"),
							"resources": string("vm:8565")}},
					outputNode: &OutputNode{
						id: "guest-single",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:     "guest-single",
							Guests: &[]VmRef{{vmType: GuestLxc, vmId: 342}}})},
					outputResource: &OutputResource{
						id: "guest-single-resource",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID:     "guest-single-resource",
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 8565}}})}}}},
		{category: `ID`,
			tests: []test{
				{
					input: []map[string]any{
						{"rule": string("my-id"),
							"type": string("node-affinity")},
						{"rule": string("my-id-resource"),
							"type": string("resource-affinity")}},
					outputNode: &OutputNode{
						id: "my-id",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID: "my-id"})},
					outputResource: &OutputResource{
						id: "my-id-resource",
						rule: baseResourceRule(HaResourceAffinityRule{
							ID: "my-id-resource"})}}}},
		{category: `Nodes`,
			tests: []test{
				{name: `Multiple`,
					input: []map[string]any{
						{"rule": string("node-multiple"),
							"type":  string("node-affinity"),
							"nodes": string("node1:1,node2,node3:1000")}},
					outputNode: &OutputNode{
						id: "node-multiple",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID: "node-multiple",
							Nodes: &[]HaNode{
								{Node: "node1", Priority: 1},
								{Node: "node2"},
								{Node: "node3", Priority: 1000}}})}},
				{name: `Single`,
					input: []map[string]any{
						{"rule": string("node-single"),
							"type":  string("node-affinity"),
							"nodes": string("node42:99")}},
					outputNode: &OutputNode{
						id: "node-single",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:    "node-single",
							Nodes: &[]HaNode{{Node: "node42", Priority: 99}}})}}}},
		{category: `Strict`,
			tests: []test{
				{name: "false",
					input: []map[string]any{
						{"rule": string("strict-false"),
							"type":   string("node-affinity"),
							"strict": string("0")},
					},
					outputNode: &OutputNode{
						id: "strict-false",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:     "strict-false",
							Strict: util.Pointer(false)}),
					}},
				{name: "true",
					input: []map[string]any{
						{"rule": string("strict-true"),
							"type":   string("node-affinity"),
							"strict": string("1")},
					},
					outputNode: &OutputNode{
						id: "strict-true",
						rule: baseNodeRule(HaNodeAffinityRule{
							ID:     "strict-true",
							Strict: util.Pointer(true)}),
					}}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			rules, err := listHaRules(context.Background(), &mockClientAPI{
				listHaRulesFunc: func(ctx context.Context) ([]any, error) {
					outArray := make([]any, len(subTest.input))
					for i, e := range subTest.input {
						outArray[i] = e
					}
					return outArray, subTest.err
				}})

			var arrayRules []HaRule
			var mapRules map[HaRuleID]HaRule
			if subTest.outputNode != nil || subTest.outputResource != nil {
				arrayRules = rules.ConvertArray()
				mapRules = rules.ConvertMap()
			}

			t.Run(name+"/error", func(*testing.T) {
				require.Equal(t, subTest.err, err)
			})

			if subTest.outputNode != nil {
				t.Run(name+"/node/map", func(*testing.T) {
					id := subTest.outputNode.id
					rule := subTest.outputNode.rule
					raw, ok := mapRules[id].GetNodeAffinity()
					require.True(t, ok)
					require.Equal(t, *rule.Comment, mapRules[id].GetComment())
					require.Equal(t, rule.Digest, mapRules[id].GetDigest())
					require.Equal(t, *rule.Enabled, mapRules[id].GetEnabled())
					require.Equal(t, rule.ID, mapRules[id].GetID())
					require.Equal(t, rule, raw.Get())
				})
				t.Run(name+"/node/array", func(*testing.T) {
					rule := subTest.outputNode.rule
					var ok bool
					var comment string
					var digest [20]byte
					var enabled bool
					var id HaRuleID
					var config HaNodeAffinityRule
					for _, r := range arrayRules {
						if r.GetID() == subTest.outputNode.id {
							var raw RawHaNodeAffinityRule
							raw, ok = r.GetNodeAffinity()
							comment = r.GetComment()
							digest = r.GetDigest()
							enabled = r.GetEnabled()
							id = r.GetID()
							config = raw.Get()
						}
					}
					require.True(t, ok)
					require.Equal(t, *rule.Comment, comment)
					require.Equal(t, rule.Digest, digest)
					require.Equal(t, *rule.Enabled, enabled)
					require.Equal(t, rule.ID, id)
					require.Equal(t, rule, config)
				})
			}

			if subTest.outputResource != nil {
				t.Run(name+"/resource/map", func(*testing.T) {
					id := subTest.outputResource.id
					rule := subTest.outputResource.rule
					raw, ok := mapRules[id].GetResourceAffinity()
					require.True(t, ok)
					require.Equal(t, *rule.Comment, mapRules[id].GetComment())
					require.Equal(t, rule.Digest, mapRules[id].GetDigest())
					require.Equal(t, *rule.Enabled, mapRules[id].GetEnabled())
					require.Equal(t, rule.ID, mapRules[id].GetID())
					require.Equal(t, rule, raw.Get())
				})
				t.Run(name+"/resource/array", func(*testing.T) {
					rule := subTest.outputResource.rule
					var ok bool
					var comment string
					var digest [20]byte
					var enabled bool
					var id HaRuleID
					var config HaResourceAffinityRule
					for _, r := range arrayRules {
						if r.GetID() == subTest.outputResource.id {
							var raw RawHaResourceAffinityRule
							raw, ok = r.GetResourceAffinity()
							comment = r.GetComment()
							digest = r.GetDigest()
							enabled = r.GetEnabled()
							id = r.GetID()
							config = raw.Get()
						}
					}
					require.True(t, ok)
					require.Equal(t, *rule.Comment, comment)
					require.Equal(t, rule.Digest, digest)
					require.Equal(t, *rule.Enabled, enabled)
					require.Equal(t, rule.ID, id)
					require.Equal(t, rule, config)
				})
			}
		}
	}
}

func Test_HaNodeAffinityRule_Validate(t *testing.T) {
	type test struct {
		name  string
		input HaNodeAffinityRule
		err   error
	}
	type testType struct {
		create       []test
		createUpdate []test
		update       []test
	}
	tests := []struct {
		category string
		valid    testType
		invalid  testType
	}{
		{category: `all`,
			valid: testType{
				create: []test{
					{name: `minimum`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 100}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{{Node: "node1"}}}}},
				update: []test{
					{name: `minimum`,
						input: HaNodeAffinityRule{ID: "ha-rule-1"}}}}},
		{category: `Guests`,
			invalid: testType{
				create: []test{
					{name: `errors.New(HaNodeAffinityRule_Error_GuestsRequired)`,
						err: errors.New(HaNodeAffinityRule_Error_GuestsRequired)}},
				createUpdate: []test{
					{name: `errors.New(HaNodeAffinityRule_Error_GuestsEmpty)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{}},
						err: errors.New(HaNodeAffinityRule_Error_GuestsEmpty)},
					{name: `errors.New(GuestID_Error_Minimum)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 99}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{}},
						err: errors.New(GuestID_Error_Minimum)}}}},
		{category: `ID`,
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(HaRuleID_Error_MinLength)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{},
							Nodes:  &[]HaNode{}},
						err: errors.New(HaRuleID_Error_MinLength)}}}},
		{category: `Nodes`,
			invalid: testType{
				create: []test{
					{name: `errors.New(HaNodeAffinityRule_Error_NodesRequired)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{}},
						err: errors.New(HaNodeAffinityRule_Error_NodesRequired)}},
				createUpdate: []test{
					{name: `errors.New(HaNodeAffinityRule_Error_NodesEmpty)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 100}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{}},
						err: errors.New(HaNodeAffinityRule_Error_NodesEmpty)},
					{name: `errors.New(NodeName_Error_Empty)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 100}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{{Node: ""}}},
						err: errors.New(NodeName_Error_Empty)},
					{name: `errors.New(HaPriority_Error_Invalid)`,
						input: HaNodeAffinityRule{
							Guests: &[]VmRef{{vmType: GuestQemu, vmId: 100}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{{Node: "node1", Priority: 1001}}},
						err: errors.New(HaPriority_Error_Invalid)}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.valid.create, test.valid.createUpdate...) {
			name := test.category + "/Valid/Create"
			if len(test.valid.create)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.valid.update, test.valid.createUpdate...) {
			name := test.category + "/Valid/Update"
			if len(test.valid.update)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(&HaNodeAffinityRule{}), name)
			})
		}
		for _, subTest := range append(test.invalid.create, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Create"
			if len(test.invalid.create)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.invalid.update, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Update"
			if len(test.invalid.update)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(&HaNodeAffinityRule{}), name)
			})
		}
	}
}

func Test_HaResourceAffinityRule_Validate(t *testing.T) {
	type test struct {
		name  string
		input HaResourceAffinityRule
		err   error
	}
	type testType struct {
		create       []test
		createUpdate []test
		update       []test
	}
	tests := []struct {
		category string
		valid    testType
		invalid  testType
	}{
		{category: `all`,
			valid: testType{
				create: []test{
					{name: `minimum`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityPositive),
							Guests:   &[]VmRef{{vmType: GuestQemu, vmId: 100}},
							ID:       "ha-rule-1"}}},
				createUpdate: []test{
					{name: `normal`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityNegative),
							Comment:  util.Pointer("This is a comment"),
							Enabled:  util.Pointer(true),
							Guests: &[]VmRef{
								{vmType: GuestQemu, vmId: 100},
								{vmType: GuestLxc, vmId: 200},
								{vmType: GuestQemu, vmId: 300}},
							ID: "ha-rule-1"}}},
				update: []test{
					{name: `minimum`,
						input: HaResourceAffinityRule{ID: "ha-rule-1"}}}}},
		{category: `Affinity`,
			invalid: testType{
				create: []test{
					{name: `errors.New(HaResourceAffinityRule_Error_AffinityRequired)`,
						err: errors.New(HaResourceAffinityRule_Error_AffinityRequired)},
					{name: `errors.New(HaAffinity_Error_Invalid)`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinity(25)),
							Guests:   &[]VmRef{},
							ID:       "valid-id"},
						err: errors.New(HaAffinity_Error_Invalid)}}}},
		{category: `Guests`,
			invalid: testType{
				create: []test{
					{name: `errors.New(HaResourceAffinityRule_Error_GuestsRequired)`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityUnknown)},
						err: errors.New(HaResourceAffinityRule_Error_GuestsRequired)}},
				createUpdate: []test{
					{name: `errors.New(HaResourceAffinityRule_Error_GuestsEmpty)`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityPositive),
							Guests:   &[]VmRef{},
							ID:       "ha-rule-1"},
						err: errors.New(HaResourceAffinityRule_Error_GuestsEmpty)},
					{name: `errors.New(GuestID_Error_Minimum)`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityNegative),
							Guests:   &[]VmRef{{vmType: GuestQemu, vmId: 99}},
							ID:       "ha-rule-1"},
						err: errors.New(GuestID_Error_Minimum)}}}},
		{category: `ID`,
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(HaRuleID_Error_MinLength)`,
						input: HaResourceAffinityRule{
							Affinity: util.Pointer(HaAffinityPositive),
							Guests:   &[]VmRef{}},
						err: errors.New(HaRuleID_Error_MinLength)}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.valid.create, test.valid.createUpdate...) {
			name := test.category + "/Valid/Create"
			if len(test.valid.create)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.valid.update, test.valid.createUpdate...) {
			name := test.category + "/Valid/Update"
			if len(test.valid.update)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(&HaResourceAffinityRule{}), name)
			})
		}
		for _, subTest := range append(test.invalid.create, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Create"
			if len(test.invalid.create)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil), name)
			})
		}
		for _, subTest := range append(test.invalid.update, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Update"
			if len(test.invalid.update)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(&HaResourceAffinityRule{}), name)
			})
		}
	}
}

func Test_HaRuleID_Validate(t *testing.T) {
	tests := []struct {
		name   string
		input  []string
		output error
	}{
		{name: `Valid HaRuleID`,
			input: test_data_ha.HaRuleID_Legal()},
		{name: `Invalid HaRuleID Empty`,
			input:  test_data_ha.HaRuleID_MinLength(),
			output: errors.New(HaRuleID_Error_MinLength)},
		{name: `Invalid HaRuleID Invalid`,
			input:  test_data_ha.HaRuleID_CharacterIllegal(),
			output: errors.New(HaRuleID_Error_Invalid)},
		{name: `Invalid HaRuleID Max Length`,
			input:  []string{test_data_ha.HaRuleID_MaxIllegal()},
			output: errors.New(HaRuleID_Error_MaxLength)},
		{name: `Invalid HaRuleID begin with illegal start character`,
			input:  test_data_ha.HaRuleID_StartIllegal(),
			output: errors.New(HaRuleID_Error_Start)},
	}
	for _, test := range tests {
		for _, e := range test.input {
			t.Run(test.name+"/"+e, func(t *testing.T) {
				require.Equal(t, test.output, (HaRuleID(e)).Validate())
			})
		}
	}
}
