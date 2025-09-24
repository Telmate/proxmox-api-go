package proxmox

import (
	"context"
	"errors"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/array"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_ha"
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

func Test_HaNodeAffinityRule_create(t *testing.T) {
	type test struct {
		name   string
		config HaNodeAffinityRule
		output map[string]any
	}
	tests := []struct {
		category string
		create   []test
	}{
		{category: `Comment`,
			create: []test{
				{name: `set`,
					config: HaNodeAffinityRule{Comment: util.Pointer("This is a comment")},
					output: map[string]any{
						"comment": string("This is a comment"),
						"rule":    string(""),
						"strict":  string("0"),
						"type":    string("node-affinity")}},
				{name: `empty`,
					config: HaNodeAffinityRule{Comment: util.Pointer("")},
					output: map[string]any{
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}}}},
		{category: `Digest`,
			create: []test{
				{name: `set no effect`,
					config: HaNodeAffinityRule{Digest: [20]byte{0xeb, 0xef, 0xee, 0xde, 0x30, 0x59, 0x41, 0x74, 0x44, 0x30, 0x8d, 0x2e, 0x58, 0xd2, 0xa5, 0xe5, 0x04, 0xfe, 0x61, 0x51}},
					output: map[string]any{
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}}}},
		{category: `Enabled`,
			create: []test{
				{name: "false",
					config: HaNodeAffinityRule{Enabled: util.Pointer(false)},
					output: map[string]any{
						"disable": string("1"),
						"rule":    string(""),
						"strict":  string("0"),
						"type":    string("node-affinity")}},
				{name: "true",
					config: HaNodeAffinityRule{Enabled: util.Pointer(true)},
					output: map[string]any{
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}}}},
		{category: `Guests`,
			create: []test{
				{name: `Multiple`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					output: map[string]any{
						"resources": []string{"vm:100", "ct:200", "vm:300"},
						"rule":      string(""),
						"strict":    string("0"),
						"type":      string("node-affinity")}},
				{name: `Single`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestLxc, vmId: 342}}},
					output: map[string]any{
						"resources": []string{"ct:342"},
						"rule":      string(""),
						"strict":    string("0"),
						"type":      string("node-affinity")}}}},
		{category: `ID`,
			create: []test{
				{name: `set`,
					config: HaNodeAffinityRule{ID: "my-id"},
					output: map[string]any{
						"rule":   string("my-id"),
						"strict": string("0"),
						"type":   string("node-affinity")}}}},
		{category: `Nodes`,
			create: []test{
				{name: `Multiple`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{
						{Node: "node1", Priority: 1},
						{Node: "node2"},
						{Node: "node3", Priority: 1000}}},
					output: map[string]any{
						"nodes":  string("node1:1,node2,node3:1000"),
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}},
				{name: `Single`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{
						{Node: "node42", Priority: 99}}},
					output: map[string]any{
						"nodes":  string("node42:99"),
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}}}},
		{category: `Strict`,
			create: []test{
				{name: "false",
					config: HaNodeAffinityRule{Strict: util.Pointer(false)},
					output: map[string]any{
						"rule":   string(""),
						"strict": string("0"),
						"type":   string("node-affinity")}},
				{name: "true",
					config: HaNodeAffinityRule{Strict: util.Pointer(true)},
					output: map[string]any{
						"rule":   string(""),
						"strict": string("1"),
						"type":   string("node-affinity")}}}},
	}
	for _, test := range tests {
		for _, subTest := range test.create {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				var tmpParams map[string]any
				subTest.config.create(context.Background(), &mockClientAPI{
					createHaRuleFunc: func(ctx context.Context, params map[string]any) error {
						tmpParams = params
						return nil
					}})
				require.Equal(t, subTest.output, tmpParams, name)
			})
		}
	}
}

func Test_HaNodeAffinityRule_Get(t *testing.T) {
	baseRule := func(r HaNodeAffinityRule) HaNodeAffinityRule {
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
	type test struct {
		name          string
		input         map[string]any
		id            HaRuleID
		outputPublic  HaNodeAffinityRule
		outputPrivate *HaNodeAffinityRule
		err           error
	}
	tests := []struct {
		category string
		tests    []test
	}{
		{category: `Comment`,
			tests: []test{
				{name: `set`,
					input: map[string]any{
						"type":    string("node-affinity"),
						"comment": string("This is a comment")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Comment: util.Pointer("This is a comment")}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Comment: util.Pointer("This is a comment")}))}}},
		{category: `Digest`,
			tests: []test{
				{name: `set`,
					input: map[string]any{
						"type":   string("node-affinity"),
						"digest": string("ebefeede3059417444308d2e58d2a5e504fe6151")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Digest: [20]byte{0xeb, 0xef, 0xee, 0xde, 0x30, 0x59, 0x41, 0x74, 0x44, 0x30, 0x8d, 0x2e, 0x58, 0xd2, 0xa5, 0xe5, 0x04, 0xfe, 0x61, 0x51}}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						rawDigest: "ebefeede3059417444308d2e58d2a5e504fe6151"}))}}},
		{category: `Enabled`,
			tests: []test{
				{name: "false",
					input: map[string]any{
						"type":    string("node-affinity"),
						"disable": string("1")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Enabled: util.Pointer(false)}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Enabled: util.Pointer(false)}))},
				{name: "true",
					input: map[string]any{
						"type":    string("node-affinity"),
						"disable": string("0")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Enabled: util.Pointer(true)}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Enabled: util.Pointer(true)}))}}},
		{category: `Guests`,
			tests: []test{
				{name: `Multiple`,
					input: map[string]any{
						"type":      string("node-affinity"),
						"resources": string("vm:100,ct:200,vm:300")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestQemu, vmId: 100},
							{vmType: GuestLxc, vmId: 200},
							{vmType: GuestQemu, vmId: 300}}}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestQemu, vmId: 100},
							{vmType: GuestLxc, vmId: 200},
							{vmType: GuestQemu, vmId: 300}}}))},
				{name: `Single`,
					input: map[string]any{
						"type":      string("node-affinity"),
						"resources": string("ct:342")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestLxc, vmId: 342}}}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestLxc, vmId: 342}}}))}}},
		{category: `ID`,
			tests: []test{
				{name: `set`,
					id: "my-id",
					input: map[string]any{
						"rule": string("my-id"),
						"type": string("node-affinity")},
					outputPublic: baseRule(HaNodeAffinityRule{
						ID: "my-id"}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						ID: "my-id"}))}}},
		{category: `Nodes`,
			tests: []test{
				{name: `Multiple`,
					input: map[string]any{
						"type":  string("node-affinity"),
						"nodes": string("node1:1,node2,node3:1000")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Nodes: &[]HaNode{
							{Node: "node1", Priority: 1},
							{Node: "node2"},
							{Node: "node3", Priority: 1000}}}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Nodes: &[]HaNode{
							{Node: "node1", Priority: 1},
							{Node: "node2"},
							{Node: "node3", Priority: 1000}}}))},
				{name: `Single`,
					input: map[string]any{
						"type":  string("node-affinity"),
						"nodes": string("node42:99")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Nodes: &[]HaNode{
							{Node: "node42", Priority: 99}}}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Nodes: &[]HaNode{
							{Node: "node42", Priority: 99}}}))}}},
		{category: `Strict`,
			tests: []test{
				{name: "false",
					input: map[string]any{
						"type":   string("node-affinity"),
						"strict": string("0")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Strict: util.Pointer(false)}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Strict: util.Pointer(false)}))},
				{name: "true",
					input: map[string]any{
						"type":   string("node-affinity"),
						"strict": string("1")},
					outputPublic: baseRule(HaNodeAffinityRule{
						Strict: util.Pointer(true)}),
					outputPrivate: util.Pointer(baseRule(HaNodeAffinityRule{
						Strict: util.Pointer(true)}))}}},
		{category: `error`,
			tests: []test{
				{name: "api error",
					err: errors.New("api error")}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				var tmpID HaRuleID
				rawRule, err := subTest.id.get(context.Background(), &mockClientAPI{
					getHaRuleFunc: func(ctx context.Context, id HaRuleID) (map[string]any, error) {
						tmpID = id
						return subTest.input, subTest.err
					}})
				require.Equal(t, subTest.err, err)
				require.Equal(t, subTest.id, tmpID)
				if subTest.err == nil {
					raw, ok := rawRule.getNodeAffinity()
					require.Equal(t, true, ok)
					_, notOk := rawRule.getResourceAffinity()
					require.Equal(t, false, notOk)
					require.Equal(t, subTest.outputPrivate, raw.get())
					require.Equal(t, subTest.outputPublic, raw.Get())
				}
			})
		}
	}
}

func Test_HaNodeAffinityRule_update(t *testing.T) {
	baseConfig := func(current *rawHaNodeAffinityRule) *rawHaNodeAffinityRule {
		if current == nil {
			current = &rawHaNodeAffinityRule{a: map[string]any{}}
		}
		if current.a[haRuleApiKeyComment] == nil {
			current.a[haRuleApiKeyComment] = string("This is a comment")
		}
		if current.a[haRuleApiKeyDisabled] == nil {
			current.a[haRuleApiKeyDisabled] = string("0")
		}
		if current.a[haRuleApiKeyNodes] == nil {
			current.a[haRuleApiKeyNodes] = string("node1:1,node2,node3:1000")
		}
		if current.a[haRuleApiKeyStrict] == nil {
			current.a[haRuleApiKeyStrict] = string("0")
		}
		if current.a[haRuleApiKeyResources] == nil {
			current.a[haRuleApiKeyResources] = string("vm:100,ct:200,vm:300")
		}
		current.a[haRuleApiKeyDigest] = string("da39a3ee5e6b4b0d3255bfef95601890afd80709")
		return current
	}
	type test struct {
		name          string
		config        HaNodeAffinityRule
		currentConfig *rawHaNodeAffinityRule
		id            HaRuleID
		output        map[string]any
	}
	tests := []struct {
		category string
		update   []test
	}{
		{category: `Comment`,
			update: []test{
				{name: `replace`,
					config:        HaNodeAffinityRule{Comment: util.Pointer("New comment")},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"comment": string("New comment"),
						"type":    string("node-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove`,
					config:        HaNodeAffinityRule{Comment: util.Pointer("")},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"comment": string(""),
						"type":    string("node-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config:        HaNodeAffinityRule{Comment: util.Pointer("This is a comment")},
					currentConfig: baseConfig(nil)}}},
		{category: `Enabled`,
			update: []test{
				{name: `replace true`,
					config: HaNodeAffinityRule{Enabled: util.Pointer(true)},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						haRuleApiKeyDisabled: string("1")}}),
					output: map[string]any{
						"delete": string("disable"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace false`,
					config: HaNodeAffinityRule{Enabled: util.Pointer(false)},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						haRuleApiKeyDisabled: string("0")}}),
					output: map[string]any{
						"disable": string("1"),
						"type":    string("node-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config:        HaNodeAffinityRule{Enabled: util.Pointer(true)},
					currentConfig: baseConfig(nil)}}},
		{category: `Guests`,
			update: []test{
				{name: `replace amount different`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 8565},
						{vmType: GuestLxc, vmId: 342}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"vm:8565", "ct:342"},
						"type":      string("node-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace amount same`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 8565},
						{vmType: GuestLxc, vmId: 342},
						{vmType: GuestQemu, vmId: 777}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"ct:342", "vm:777", "vm:8565"},
						"type":      string("node-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace single`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 342}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"vm:342"},
						"type":      string("node-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove (not supported in API)`,
					config:        HaNodeAffinityRule{Guests: &[]VmRef{}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{},
						"type":      string("node-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change different order`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"resources": string("ct:200,vm:300,vm:100")}})},
				{name: `no change same order`,
					config: HaNodeAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"resources": string("vm:100,ct:200,vm:300")}})}}},
		{category: `ID`,
			update: []test{
				{name: `no change`,
					config:        HaNodeAffinityRule{ID: "my-id"},
					currentConfig: baseConfig(nil)},
				{name: `with change`,
					config: HaNodeAffinityRule{
						ID:      "my-id",
						Enabled: util.Pointer(true)},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"disable": string("1")}}),
					output: map[string]any{
						"delete": string("disable"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")},
					id: "my-id"}}},
		{category: `Nodes`,
			update: []test{
				{name: `replace Multiple`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{
						{Node: "node42", Priority: 1000},
						{Node: "nod342"},
						{Node: "node7", Priority: 1}}},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"nodes": string("node3:1000,node1:1,node2")}}),
					output: map[string]any{
						"nodes":  string("nod342,node42:1000,node7:1"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace Single`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{
						{Node: "node7", Priority: 1}}},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"nodes": string("node1;node3:1000")}}),
					output: map[string]any{
						"nodes":  string("node7:1"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove (not supported in API)`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{}},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"nodes": string("node1:1,node3:1000")}}),
					output: map[string]any{
						"nodes":  string(""),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change different order`,
					config: HaNodeAffinityRule{Nodes: &[]HaNode{
						{Node: "node3", Priority: 1000},
						{Node: "node1", Priority: 1},
						{Node: "node2"}}},
					currentConfig: baseConfig(nil)},
				{name: `no change same order`,
					config: HaNodeAffinityRule{},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"nodes": string("node1:1,node2,node3:1000")}})}}},
		{category: `Strict`,
			update: []test{
				{name: `replace false`,
					config: HaNodeAffinityRule{Strict: util.Pointer(false)},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"strict": string("1")}}),
					output: map[string]any{
						"strict": string("0"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace true`,
					config:        HaNodeAffinityRule{Strict: util.Pointer(true)},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"strict": string("1"),
						"type":   string("node-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config: HaNodeAffinityRule{Strict: util.Pointer(true)},
					currentConfig: baseConfig(&rawHaNodeAffinityRule{a: map[string]any{
						"strict": string("1")}})}}},
	}
	for _, test := range tests {
		for _, subTest := range test.update {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				var tmpID HaRuleID
				var tmpParams map[string]any
				subTest.config.update(context.Background(), subTest.currentConfig, &mockClientAPI{
					updateHaRuleFunc: func(ctx context.Context, id HaRuleID, params map[string]any) error {
						tmpID = id
						tmpParams = params
						return nil
					}})
				require.Equal(t, subTest.id, tmpID)
				require.Equal(t, subTest.output, tmpParams)
			})
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
				createUpdate: []test{
					{name: `normal`,
						input: HaNodeAffinityRule{
							Comment: util.Pointer("This is a comment"),
							Enabled: util.Pointer(true),
							Guests: &[]VmRef{
								{vmType: GuestQemu, vmId: 100},
								{vmType: GuestLxc, vmId: 200},
								{vmType: GuestQemu, vmId: 300}},
							ID:     "ha-rule-1",
							Nodes:  &[]HaNode{{Node: "node1", Priority: 1}, {Node: "node2"}, {Node: "node3", Priority: 1000}},
							Strict: util.Pointer(true)}}},
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

func Test_HaResourceAffinityRule_create(t *testing.T) {
	type test struct {
		name   string
		config HaResourceAffinityRule
		output map[string]any
	}
	tests := []struct {
		category string
		create   []test
	}{
		{category: `Affinity`,
			create: []test{
				{name: `empty (not supported in API)`,
					config: HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityUnknown)},
					output: map[string]any{
						"affinity": string(""),
						"rule":     string(""),
						"type":     string("resource-affinity")}},
				{name: `positive`,
					config: HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityPositive)},
					output: map[string]any{
						"affinity": string("positive"),
						"rule":     string(""),
						"type":     string("resource-affinity")}},
				{name: `negative`,
					config: HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityNegative)},
					output: map[string]any{
						"affinity": string("negative"),
						"rule":     string(""),
						"type":     string("resource-affinity")}}}},
		{category: `Comment`,
			create: []test{
				{name: `set`,
					config: HaResourceAffinityRule{Comment: util.Pointer("This is a comment")},
					output: map[string]any{
						"comment": string("This is a comment"),
						"rule":    string(""),
						"type":    string("resource-affinity")}},
				{name: `empty`,
					config: HaResourceAffinityRule{Comment: util.Pointer("")},
					output: map[string]any{
						"rule": string(""),
						"type": string("resource-affinity")}}}},
		{category: `Digest`,
			create: []test{
				{name: `set no effect`,
					config: HaResourceAffinityRule{Digest: [20]byte{0xeb, 0xef, 0xee, 0xde, 0x30, 0x59, 0x41, 0x74, 0x44, 0x30, 0x8d, 0x2e, 0x58, 0xd2, 0xa5, 0xe5, 0x04, 0xfe, 0x61, 0x51}},
					output: map[string]any{
						"rule": string(""),
						"type": string("resource-affinity")}}}},
		{category: `Enabled`,
			create: []test{
				{name: "false",
					config: HaResourceAffinityRule{Enabled: util.Pointer(false)},
					output: map[string]any{
						"disable": string("1"),
						"rule":    string(""),
						"type":    string("resource-affinity")}},
				{name: "true",
					config: HaResourceAffinityRule{Enabled: util.Pointer(true)},
					output: map[string]any{
						"rule": string(""),
						"type": string("resource-affinity")}}}},
		{category: `Guests`,
			create: []test{
				{name: `Multiple`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					output: map[string]any{
						"resources": []string{"vm:100", "ct:200", "vm:300"},
						"rule":      string(""),
						"type":      string("resource-affinity")}},
				{name: `Single`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestLxc, vmId: 342}}},
					output: map[string]any{
						"resources": []string{"ct:342"},
						"rule":      string(""),
						"type":      string("resource-affinity")}}}},
		{category: `ID`,
			create: []test{
				{name: `set`,
					config: HaResourceAffinityRule{ID: "my-id"},
					output: map[string]any{
						"rule": string("my-id"),
						"type": string("resource-affinity")}}}},
	}
	for _, test := range tests {
		for _, subTest := range test.create {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				var tmpParams map[string]any
				subTest.config.create(context.Background(), &mockClientAPI{
					createHaRuleFunc: func(ctx context.Context, params map[string]any) error {
						tmpParams = params
						return nil
					}})
				require.Equal(t, subTest.output, tmpParams, name)
			})
		}
	}
}

func Test_HaResourceAffinityRule_Get(t *testing.T) {
	baseRule := func(r HaResourceAffinityRule) HaResourceAffinityRule {
		if r.Affinity == nil {
			r.Affinity = util.Pointer(HaAffinityUnknown)
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
	type test struct {
		name          string
		input         map[string]any
		id            HaRuleID
		outputPublic  HaResourceAffinityRule
		outputPrivate *HaResourceAffinityRule
		err           error
	}
	tests := []struct {
		category string
		tests    []test
	}{
		{category: `Affinity`,
			tests: []test{
				{name: `positive`,
					input: map[string]any{
						"type":     string("resource-affinity"),
						"affinity": string("positive")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Affinity: util.Pointer(HaAffinityPositive)}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Affinity: util.Pointer(HaAffinityPositive)}))},
				{name: `negative`,
					input: map[string]any{
						"type":     string("resource-affinity"),
						"affinity": string("negative")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Affinity: util.Pointer(HaAffinityNegative)}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Affinity: util.Pointer(HaAffinityNegative)}))}}},
		{category: `Comment`,
			tests: []test{
				{name: `set`,
					input: map[string]any{
						"type":    string("resource-affinity"),
						"comment": string("This is a comment")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Comment: util.Pointer("This is a comment")}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Comment: util.Pointer("This is a comment")}))}}},
		{category: `Digest`,
			tests: []test{
				{name: `set`,
					input: map[string]any{
						"type":   string("resource-affinity"),
						"digest": string("ebefeede3059417444308d2e58d2a5e504fe6151")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Digest: [20]byte{0xeb, 0xef, 0xee, 0xde, 0x30, 0x59, 0x41, 0x74, 0x44, 0x30, 0x8d, 0x2e, 0x58, 0xd2, 0xa5, 0xe5, 0x04, 0xfe, 0x61, 0x51}}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						rawDigest: "ebefeede3059417444308d2e58d2a5e504fe6151"}))}}},
		{category: `Enabled`,
			tests: []test{
				{name: "false",
					input: map[string]any{
						"type":    string("resource-affinity"),
						"disable": string("1")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Enabled: util.Pointer(false)}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Enabled: util.Pointer(false)}))},
				{name: "true",
					input: map[string]any{
						"type":    string("resource-affinity"),
						"disable": string("0")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Enabled: util.Pointer(true)}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Enabled: util.Pointer(true)}))}}},
		{category: `Guests`,
			tests: []test{
				{name: `Multiple`,
					input: map[string]any{
						"type":      string("resource-affinity"),
						"resources": string("vm:100,ct:200,vm:300")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestQemu, vmId: 100},
							{vmType: GuestLxc, vmId: 200},
							{vmType: GuestQemu, vmId: 300}}}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestQemu, vmId: 100},
							{vmType: GuestLxc, vmId: 200},
							{vmType: GuestQemu, vmId: 300}}}))},
				{name: `Single`,
					input: map[string]any{
						"type":      string("resource-affinity"),
						"resources": string("ct:342")},
					outputPublic: baseRule(HaResourceAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestLxc, vmId: 342}}}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						Guests: &[]VmRef{
							{vmType: GuestLxc, vmId: 342}}}))}}},
		{category: `ID`,
			tests: []test{
				{name: `set`,
					id: "my-id",
					input: map[string]any{
						"rule": string("my-id"),
						"type": string("resource-affinity")},
					outputPublic: baseRule(HaResourceAffinityRule{
						ID: "my-id"}),
					outputPrivate: util.Pointer(baseRule(HaResourceAffinityRule{
						ID: "my-id"}))}}},
		{category: `error`,
			tests: []test{
				{name: "api error",
					err: errors.New("api error")}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				var tmpID HaRuleID
				rawRule, err := subTest.id.get(context.Background(), &mockClientAPI{
					getHaRuleFunc: func(ctx context.Context, id HaRuleID) (map[string]any, error) {
						tmpID = id
						return subTest.input, subTest.err
					}})
				require.Equal(t, subTest.err, err)
				require.Equal(t, subTest.id, tmpID)
				if subTest.err == nil {
					raw, ok := rawRule.getResourceAffinity()
					require.Equal(t, true, ok)
					_, notOk := rawRule.getNodeAffinity()
					require.Equal(t, false, notOk)
					require.Equal(t, subTest.outputPrivate, raw.get())
					require.Equal(t, subTest.outputPublic, raw.Get())
				}
			})
		}
	}
}

func Test_HaResourceAffinityRule_update(t *testing.T) {
	baseConfig := func(current *rawHaResourceAffinityRule) *rawHaResourceAffinityRule {
		if current == nil {
			current = &rawHaResourceAffinityRule{a: map[string]any{}}
		}
		if current.a[haRuleApiKeyAffinity] == nil {
			current.a[haRuleApiKeyAffinity] = string("positive")
		}
		if current.a[haRuleApiKeyComment] == nil {
			current.a[haRuleApiKeyComment] = string("This is a comment")
		}
		if current.a[haRuleApiKeyResources] == nil {
			current.a[haRuleApiKeyResources] = string("vm:100,ct:200,vm:300")
		}
		if current.a[haRuleApiKeyDigest] == nil {
			current.a[haRuleApiKeyDigest] = string("da39a3ee5e6b4b0d3255bfef95601890afd80709")
		}
		return current
	}
	type test struct {
		name          string
		config        HaResourceAffinityRule
		currentConfig *rawHaResourceAffinityRule
		id            HaRuleID
		output        map[string]any
	}
	tests := []struct {
		category string
		update   []test
	}{
		{category: `Affinity`,
			update: []test{
				{name: `replace positive`,
					config: HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityPositive)},
					currentConfig: baseConfig(&rawHaResourceAffinityRule{a: map[string]any{
						haRuleApiKeyAffinity: string("negative")}}),
					output: map[string]any{
						"affinity": string("positive"),
						"type":     string("resource-affinity"),
						"digest":   string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace negative`,
					config:        HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityNegative)},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"affinity": string("negative"),
						"type":     string("resource-affinity"),
						"digest":   string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove (not supported in API)`,
					config:        HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityUnknown)},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"affinity": string(""),
						"type":     string("resource-affinity"),
						"digest":   string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config:        HaResourceAffinityRule{Affinity: util.Pointer(HaAffinityPositive)},
					currentConfig: baseConfig(nil)}}},
		{category: `Comment`,
			update: []test{
				{name: `replace`,
					config:        HaResourceAffinityRule{Comment: util.Pointer("New comment")},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"comment": string("New comment"),
						"type":    string("resource-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove`,
					config:        HaResourceAffinityRule{Comment: util.Pointer("")},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"comment": string(""),
						"type":    string("resource-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config:        HaResourceAffinityRule{Comment: util.Pointer("This is a comment")},
					currentConfig: baseConfig(nil)}}},
		{category: `Enabled`,
			update: []test{
				{name: `replace true`,
					config: HaResourceAffinityRule{Enabled: util.Pointer(true)},
					currentConfig: baseConfig(&rawHaResourceAffinityRule{a: map[string]any{
						"disable": string("1")}}),
					output: map[string]any{
						"delete": string("disable"),
						"type":   string("resource-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace false`,
					config:        HaResourceAffinityRule{Enabled: util.Pointer(false)},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"disable": string("1"),
						"type":    string("resource-affinity"),
						"digest":  string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change`,
					config:        HaResourceAffinityRule{Enabled: util.Pointer(true)},
					currentConfig: baseConfig(nil)}}},
		{category: `Guests`,
			update: []test{
				{name: `replace amount different`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 8565},
						{vmType: GuestLxc, vmId: 342}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"vm:8565", "ct:342"},
						"type":      string("resource-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace amount same`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 8565},
						{vmType: GuestLxc, vmId: 342},
						{vmType: GuestQemu, vmId: 777}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"ct:342", "vm:777", "vm:8565"},
						"type":      string("resource-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `replace single`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 342}}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{"vm:342"},
						"type":      string("resource-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `remove (not supported in API)`,
					config:        HaResourceAffinityRule{Guests: &[]VmRef{}},
					currentConfig: baseConfig(nil),
					output: map[string]any{
						"resources": []string{},
						"type":      string("resource-affinity"),
						"digest":    string("da39a3ee5e6b4b0d3255bfef95601890afd80709")}},
				{name: `no change different order`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					currentConfig: baseConfig(&rawHaResourceAffinityRule{a: map[string]any{
						"resources": string("ct:200,vm:300,vm:100")}})},
				{name: `no change same order`,
					config: HaResourceAffinityRule{Guests: &[]VmRef{
						{vmType: GuestQemu, vmId: 100},
						{vmType: GuestLxc, vmId: 200},
						{vmType: GuestQemu, vmId: 300}}},
					currentConfig: baseConfig(&rawHaResourceAffinityRule{a: map[string]any{
						"resources": string("vm:100,ct:200,vm:300")}})}}},
		{category: `ID`,
			update: []test{
				{name: `no change`,
					config:        HaResourceAffinityRule{ID: "my-id"},
					currentConfig: baseConfig(nil)},
				{name: `with change`,
					config: HaResourceAffinityRule{
						ID:      "my-id",
						Enabled: util.Pointer(true)},
					currentConfig: baseConfig(&rawHaResourceAffinityRule{a: map[string]any{
						"disable": string("1")}}),
					output: map[string]any{
						"delete": string("disable"),
						"type":   string("resource-affinity"),
						"digest": string("da39a3ee5e6b4b0d3255bfef95601890afd80709")},
					id: "my-id"}}},
	}
	for _, test := range tests {
		for _, subTest := range test.update {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				var tmpID HaRuleID
				var tmpParams map[string]any
				subTest.config.update(context.Background(), subTest.currentConfig, &mockClientAPI{
					updateHaRuleFunc: func(ctx context.Context, id HaRuleID, params map[string]any) error {
						tmpID = id
						tmpParams = params
						return nil
					}})
				require.Equal(t, subTest.id, tmpID)
				require.Equal(t, subTest.output, tmpParams)
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
