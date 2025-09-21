package proxmox

import (
	"context"
	"crypto/sha1"
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

const HaRule_Error_VersionTooLow = "HA rules require Proxmox VE 9.0 or higher"

func ListHaRules(ctx context.Context, c *Client) (HaRules, error) { return c.new().haListRules(ctx) }

func (c *clientNew) haListRules(ctx context.Context) (HaRules, error) {
	if err := haVersionCheck(ctx, c); err != nil {
		return nil, err
	}
	return listHaRules(ctx, c.api)
}

func ListHaRulesNoCheck(ctx context.Context, c *Client) (HaRules, error) {
	return c.new().haListRulesNoCheck(ctx)
}

func (c *clientNew) haListRulesNoCheck(ctx context.Context) (HaRules, error) {
	return listHaRules(ctx, c.api)
}

func listHaRules(ctx context.Context, c clientApiInterface) (HaRules, error) {
	rawRules, err := c.listHaRules(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]map[string]any, len(rawRules))
	for i := range rawRules {
		rules[i] = rawRules[i].(map[string]any)
	}
	return &haRules{a: rules}, nil
}

func NewHaRuleFromApi(ctx context.Context, id HaRuleID, c *Client) (HaRule, error) {
	return c.new().haGetRule(ctx, id)
}

func (c *clientNew) haGetRule(ctx context.Context, id HaRuleID) (HaRule, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	if err := haVersionCheck(ctx, c); err != nil {
		return nil, err
	}
	return id.get(ctx, c.api)
}

func (id HaRuleID) get(ctx context.Context, c clientApiInterface) (HaRule, error) {
	raw, err := c.getHaRule(ctx, id)
	if err != nil {
		return nil, err
	}
	return &haRule{a: raw}, nil
}

type HaRules interface {
	ConvertArray() []HaRule
	ConvertMap() map[HaRuleID]HaRule
}

type haRules struct {
	a []map[string]any
}

func (r *haRules) ConvertArray() []HaRule {
	rules := make([]HaRule, len(r.a))
	for i := range r.a {
		rules[i] = &haRule{a: r.a[i]}
	}
	return rules
}

func (r *haRules) ConvertMap() map[HaRuleID]HaRule {
	rules := make(map[HaRuleID]HaRule, len(r.a))
	for i := range r.a {
		rules[HaRuleID(r.a[i][haRuleApiKeyRuleID].(string))] = &haRule{a: r.a[i]}
	}
	return rules
}

type HaRule interface {
	GetComment() string
	GetDigest() [sha1.Size]byte
	GetEnabled() bool
	GetID() HaRuleID
	GetNodeAffinity() (RawHaNodeAffinityRule, bool)
	GetResourceAffinity() (RawHaResourceAffinityRule, bool)
	IsNodeAffinity() bool
	IsResourceAffinity() bool
}

type haRule struct {
	a map[string]any
}

func (r *haRule) GetComment() string { return haGetComment(r.a) }

func (r *haRule) GetDigest() [sha1.Size]byte { return haGetDigest(r.a).sha1() }

func (r *haRule) GetEnabled() bool { return haGetEnabled(r.a) }

func (r *haRule) GetID() HaRuleID { return haGetID(r.a) }

func (r *haRule) GetNodeAffinity() (RawHaNodeAffinityRule, bool) {
	if r.IsNodeAffinity() {
		return &rawHaNodeAffinityRule{a: r.a}, true
	}
	return nil, false
}

func (r *haRule) GetResourceAffinity() (RawHaResourceAffinityRule, bool) {
	if r.IsResourceAffinity() {
		return &rawHaResourceAffinityRule{a: r.a}, true
	}
	return nil, false
}

func (r *haRule) IsNodeAffinity() bool {
	if v, ok := r.a[haRuleApiKeyType]; ok && v == haTypeNodeAffinity {
		return true
	}
	return false
}

func (r *haRule) IsResourceAffinity() bool {
	if v, ok := r.a[haRuleApiKeyType]; ok && v != haTypeNodeAffinity {
		return true
	}
	return false
}

type RawHaNodeAffinityRule interface {
	Get() HaNodeAffinityRule
	GetComment() string
	GetDigest() [sha1.Size]byte
	GetEnabled() bool
	GetGuests() []VmRef
	GetID() HaRuleID
	GetNodes() []HaNode
	GetStrict() bool
	get() HaNodeAffinityRule
	getDigest() digest
}

type rawHaNodeAffinityRule struct {
	a map[string]any
}

func (r *rawHaNodeAffinityRule) Get() HaNodeAffinityRule {
	rule := r.get()
	rule.Digest = rule.rawDigest.sha1()
	rule.rawDigest = ""
	return rule
}

func (r *rawHaNodeAffinityRule) GetComment() string { return haGetComment(r.a) }

func (r *rawHaNodeAffinityRule) GetDigest() [sha1.Size]byte { return r.getDigest().sha1() }

func (r *rawHaNodeAffinityRule) GetEnabled() bool { return haGetEnabled(r.a) }

func (r *rawHaNodeAffinityRule) GetGuests() []VmRef { return haGetGuests(r.a) }

func (r *rawHaNodeAffinityRule) GetID() HaRuleID { return haGetID(r.a) }

func (r *rawHaNodeAffinityRule) GetNodes() []HaNode {
	v, ok := r.a[haRuleApiKeyNodes]
	if !ok {
		return nil
	}
	rawNodes := strings.Split(v.(string), ",")
	nodes := make([]HaNode, len(rawNodes))
	for i := range rawNodes {
		if index := strings.IndexRune(rawNodes[i], ':'); index > 0 {
			nodes[i].Node = NodeName(rawNodes[i][:index])
			p, _ := strconv.Atoi(rawNodes[i][index+1:])
			nodes[i].Priority = HaPriority(p)
		} else {
			nodes[i].Node = NodeName(rawNodes[i])
		}
	}
	return nodes
}

func (r *rawHaNodeAffinityRule) GetStrict() bool {
	if v, ok := r.a[haRuleApiKeyStrict]; ok && v == haStrictTrue {
		return true
	}
	return false
}

func (r *rawHaNodeAffinityRule) get() HaNodeAffinityRule {
	return HaNodeAffinityRule{
		Comment:   util.Pointer(r.GetComment()),
		Digest:    r.GetDigest(),
		Enabled:   util.Pointer(r.GetEnabled()),
		Guests:    util.Pointer(r.GetGuests()),
		ID:        r.GetID(),
		Nodes:     util.Pointer(r.GetNodes()),
		Strict:    util.Pointer(r.GetStrict()),
		rawDigest: r.getDigest()}
}

func (r *rawHaNodeAffinityRule) getDigest() digest { return haGetDigest(r.a) }

type HaNodeAffinityRule struct {
	Comment   *string         `json:"comment,omitempty"` // Never nil when returned
	Digest    [sha1.Size]byte `json:"digest,omitempty"`  // only returned.
	Enabled   *bool           `json:"enabled,omitempty"` // Never nil when returned
	Guests    *[]VmRef        `json:"guests,omitempty"`  // Never nil when returned
	ID        HaRuleID        `json:"id"`
	Nodes     *[]HaNode       `json:"nodes,omitempty"`  // Never nil when returned
	Strict    *bool           `json:"strict,omitempty"` // Never nil when returned
	rawDigest digest          `json:"-"`
}

const (
	haTypeNodeAffinity string = "node-affinity"
	haGuestPrefixVm    string = "vm:"
	haGuestPrefixCt    string = "ct:"
	haStrictTrue       string = "1"
)

type RawHaResourceAffinityRule interface {
	Get() HaResourceAffinityRule
	GetAffinity() HaAffinity
	GetComment() string
	GetDigest() [sha1.Size]byte
	GetEnabled() bool
	GetGuests() []VmRef
	GetID() HaRuleID
	get() HaResourceAffinityRule
	getDigest() digest
}

type rawHaResourceAffinityRule struct {
	a map[string]any
}

func (r *rawHaResourceAffinityRule) Get() HaResourceAffinityRule {
	rule := r.get()
	rule.Digest = r.GetDigest()
	rule.rawDigest = ""
	return rule
}

func (r *rawHaResourceAffinityRule) GetAffinity() HaAffinity {
	if v, ok := r.a[haRuleApiKeyAffinity]; ok {
		switch v {
		case "positive":
			return HaAffinityPositive
		case "negative":
			return HaAffinityNegative
		}
	}
	return HaAffinityUnknown
}

func (r *rawHaResourceAffinityRule) GetComment() string { return haGetComment(r.a) }

func (r *rawHaResourceAffinityRule) GetDigest() [sha1.Size]byte { return r.getDigest().sha1() }

func (r *rawHaResourceAffinityRule) GetEnabled() bool { return haGetEnabled(r.a) }

func (r *rawHaResourceAffinityRule) GetGuests() []VmRef { return haGetGuests(r.a) }

func (r *rawHaResourceAffinityRule) GetID() HaRuleID { return haGetID(r.a) }

func (r *rawHaResourceAffinityRule) get() HaResourceAffinityRule {
	return HaResourceAffinityRule{
		Affinity:  util.Pointer(r.GetAffinity()),
		Comment:   util.Pointer(r.GetComment()),
		Enabled:   util.Pointer(r.GetEnabled()),
		Guests:    util.Pointer(r.GetGuests()),
		ID:        r.GetID(),
		rawDigest: r.getDigest()}
}

func (r *rawHaResourceAffinityRule) getDigest() digest { return haGetDigest(r.a) }

const (
	haTypeResourceAffinity string = "resource-affinity"
)

type HaResourceAffinityRule struct {
	Affinity  *HaAffinity     `json:"affinity,omitempty"` // Never nil when returned
	Comment   *string         `json:"comment,omitempty"`  // Never nil when returned
	Digest    [sha1.Size]byte `json:"digest,omitempty"`   // only returned.
	Enabled   *bool           `json:"enabled,omitempty"`  // Never nil when returned
	Guests    *[]VmRef        `json:"guests,omitempty"`   // Never nil when returned
	ID        HaRuleID        `json:"id"`
	rawDigest digest          `json:"-"`
}

const (
	haRuleApiKeyAffinity  string = "affinity"
	haRuleApiKeyDisabled  string = "disable"
	haRuleApiKeyComment   string = "comment"
	haRuleApiKeyDigest    string = "digest"
	haRuleApiKeyNodes     string = "nodes"
	haRuleApiKeyResources string = "resources"
	haRuleApiKeyRuleID    string = "rule"
	haRuleApiKeyStrict    string = "strict"
	haRuleApiKeyType      string = "type"
)

func haGetComment(params map[string]any) string {
	if v, ok := params[haRuleApiKeyComment]; ok {
		return v.(string)
	}
	return ""
}

func haGetDigest(params map[string]any) digest {
	if v, ok := params[haRuleApiKeyDigest]; ok {
		return digest(v.(string))
	}
	return ""
}

func haGetEnabled(params map[string]any) bool {
	if v, ok := params[haRuleApiKeyDisabled]; ok && v == "1" {
		return false
	}
	return true
}

func haGetGuests(params map[string]any) []VmRef {
	v, ok := params[haRuleApiKeyResources]
	if !ok {
		return nil
	}
	rawGuests := strings.Split(v.(string), ",")
	guests := make([]VmRef, len(rawGuests))
	for i := range rawGuests {
		switch rawGuests[i][:3] {
		case haGuestPrefixVm:
			guests[i].vmType = GuestQemu
		case haGuestPrefixCt:
			guests[i].vmType = GuestLxc
		}
		id, _ := strconv.Atoi(rawGuests[i][3:])
		guests[i].vmId = GuestID(id)
	}
	return guests
}

func haGetID(params map[string]any) HaRuleID { return HaRuleID(params[haRuleApiKeyRuleID].(string)) }

// HaAffinity is an enum.
type HaAffinity int8

const (
	HaAffinityPositive HaAffinity = 1
	HaAffinityUnknown  HaAffinity = 0
	HaAffinityNegative HaAffinity = -1
)

// HaRuleID has a minimim of 2 characters and max of 128 characters.
type HaRuleID string

var haRuleIdRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

const (
	HaRuleID_Error_MinLength = `ha rule ID must atleast be 2 characters`
	HaRuleID_Error_MaxLength = `ha rule ID has a maximum of 128 characters`
	HaRuleID_Error_Invalid   = `ha rule ID did not match the following regex '^[a-zA-Z][a-zA-Z0-9\-_]{2,127}$'`
	HaRuleID_Error_Start     = `ha rule ID can only with a lower or upper case letter`
	HaRuleIDMin              = 2
	HaRuleIDMax              = 128
)

func (id HaRuleID) Delete(ctx context.Context, c *Client) error {
	return c.new().haDeleteRule(ctx, id)
}

func (c *clientNew) haDeleteRule(ctx context.Context, id HaRuleID) error {
	if err := id.Validate(); err != nil {
		return err
	}
	return id.delete(ctx, c.api)
}

func (id HaRuleID) DeleteNoCheck(ctx context.Context, c *Client) error {
	return c.new().haDeleteRuleNoCheck(ctx, id)
}

func (c *clientNew) haDeleteRuleNoCheck(ctx context.Context, id HaRuleID) error {
	return id.delete(ctx, c.api)
}

func (id HaRuleID) delete(ctx context.Context, c clientApiInterface) error {
	return c.deleteHaRule(ctx, id)
}

func (id HaRuleID) String() string { return string(id) } // for fmt.Stringer interface

func (id HaRuleID) Validate() error {
	if len(id) < HaRuleIDMin {
		return errors.New(HaRuleID_Error_MinLength)
	}
	if len(id) > HaRuleIDMax {
		return errors.New(HaRuleID_Error_MaxLength)
	}
	switch id[0] {
	case '-', '_', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return errors.New(HaRuleID_Error_Start)
	}
	if !haRuleIdRegex.MatchString(string(id)) {
		return errors.New(HaRuleID_Error_Invalid)
	}
	return nil
}

// Max 1000
type HaPriority uint16

const (
	HaPriority_Error_Invalid = "priority must be between 0 and 1000"
	HaPriorityMax            = 1000
)

func (p HaPriority) String() string { return strconv.Itoa(int(p)) } // for fmt.Stringer interface

type HaNode struct {
	Node     NodeName
	Priority HaPriority
}

func haVersionCheck(ctx context.Context, c *clientNew) error {
	version, err := c.oldClient.Version(ctx)
	if err != nil {
		return err
	}
	if version.Encode() < version_9_0_0 {
		return errors.New(HaRule_Error_VersionTooLow)
	}
	return nil
}
