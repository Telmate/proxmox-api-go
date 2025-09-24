package proxmox

import (
	"context"
	"crypto/sha1"
	"errors"
	"regexp"
	"sort"
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

func (id HaRuleID) get(ctx context.Context, c clientApiInterface) (*haRule, error) {
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
	Kind() HaRuleKind
}

type haRule struct {
	a map[string]any
}

func (r *haRule) GetComment() string { return haGetComment(r.a) }

func (r *haRule) GetDigest() [sha1.Size]byte { return haGetDigest(r.a).sha1() }

func (r *haRule) GetEnabled() bool { return haGetEnabled(r.a) }

func (r *haRule) GetID() HaRuleID { return haGetID(r.a) }

func (r *haRule) GetNodeAffinity() (RawHaNodeAffinityRule, bool) {
	return r.getNodeAffinity()
}

func (r *haRule) getNodeAffinity() (*rawHaNodeAffinityRule, bool) {
	if r.Kind() == HaRuleKindNodeAffinity {
		return &rawHaNodeAffinityRule{a: r.a}, true
	}
	return nil, false
}

func (r *haRule) GetResourceAffinity() (RawHaResourceAffinityRule, bool) {
	return r.getResourceAffinity()
}

func (r *haRule) getResourceAffinity() (*rawHaResourceAffinityRule, bool) {
	if r.Kind() == HaRuleKindResourceAffinity {
		return &rawHaResourceAffinityRule{a: r.a}, true
	}
	return nil, false
}

func (r *haRule) Kind() HaRuleKind {
	if v, ok := r.a[haRuleApiKeyType]; ok {
		switch v.(string) {
		case haTypeNodeAffinity:
			return HaRuleKindNodeAffinity
		case haTypeResourceAffinity:
			return HaRuleKindResourceAffinity
		}
	}
	return HaRuleKindUnknown
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
}

type rawHaNodeAffinityRule struct {
	a map[string]any
}

func (r *rawHaNodeAffinityRule) Get() HaNodeAffinityRule {
	return HaNodeAffinityRule{
		Comment: util.Pointer(r.GetComment()),
		Digest:  r.GetDigest(),
		Enabled: util.Pointer(r.GetEnabled()),
		Guests:  util.Pointer(r.GetGuests()),
		ID:      r.GetID(),
		Nodes:   util.Pointer(r.GetNodes()),
		Strict:  util.Pointer(r.GetStrict())}
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

func (r *rawHaNodeAffinityRule) get() *HaNodeAffinityRule {
	return &HaNodeAffinityRule{
		Comment:   util.Pointer(r.GetComment()),
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
	haStrictFalse      string = "0"
)

const (
	HaNodeAffinityRule_Error_GuestsEmpty    = "guests must not be empty"
	HaNodeAffinityRule_Error_GuestsRequired = "guests must be specified during creation"
	HaNodeAffinityRule_Error_Kind           = "rule is not a node affinity rule"
	HaNodeAffinityRule_Error_NodesEmpty     = "nodes must not be empty"
	HaNodeAffinityRule_Error_NodesRequired  = "nodes must be specified during creation"
)

func (config HaNodeAffinityRule) Create(ctx context.Context, c *Client) error {
	return c.new().haCreateNodeAffinityRule(ctx, config)
}

func (c *clientNew) haCreateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error {
	if err := haVersionCheck(ctx, c); err != nil {
		return err
	}
	if err := ha.validateCreate(); err != nil {
		return err
	}
	for i := range *ha.Guests {
		if err := c.guestCheckVmRef_Unsafe(ctx, &(*ha.Guests)[i]); err != nil {
			return err
		}
	}
	return ha.create(ctx, c.api)
}

func (config HaNodeAffinityRule) CreateNoCheck(ctx context.Context, c *Client) error {
	return c.new().haCreateNodeAffinityRuleNoCheck(ctx, config)
}

func (c *clientNew) haCreateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error {
	return ha.create(ctx, c.api)
}

func (config HaNodeAffinityRule) create(ctx context.Context, c clientApiInterface) error {
	return c.createHaRule(ctx, config.mapToApiCreate())
}

func (config HaNodeAffinityRule) Update(ctx context.Context, c *Client) error {
	return c.new().haUpdateNodeAffinityRule(ctx, config)
}

func (c *clientNew) haUpdateNodeAffinityRule(ctx context.Context, ha HaNodeAffinityRule) error {
	if err := haVersionCheck(ctx, c); err != nil {
		return err
	}
	if err := ha.validateUpdate(); err != nil {
		return err
	}
	rawRule, err := ha.ID.get(ctx, c.api)
	if err != nil {
		return err
	}
	raw, ok := rawRule.getNodeAffinity()
	if !ok {
		return errors.New(HaNodeAffinityRule_Error_Kind)
	}
	return ha.update(ctx, raw, c.api)
}

func (config HaNodeAffinityRule) UpdateNoCheck(ctx context.Context, c *Client) error {
	return c.new().haUpdateNodeAffinityRuleNoCheck(ctx, config)
}

func (c *clientNew) haUpdateNodeAffinityRuleNoCheck(ctx context.Context, ha HaNodeAffinityRule) error {
	rawRule, err := ha.ID.get(ctx, c.api)
	if err != nil {
		return err
	}
	raw, ok := rawRule.getNodeAffinity()
	if !ok {
		return errors.New(HaNodeAffinityRule_Error_Kind)
	}
	return ha.update(ctx, raw, c.api)
}

func (config HaNodeAffinityRule) update(ctx context.Context, current *rawHaNodeAffinityRule, c clientApiInterface) error {
	params := config.mapToApiUpdate_Unsafe(current)
	if len(params) <= haRuleMinimumParams {
		return nil
	}
	return c.updateHaRule(ctx, config.ID, params)
}

func (config HaNodeAffinityRule) mapToApiCreate() map[string]any {
	params := map[string]any{
		haRuleApiKeyRuleID: config.ID.String(),
		haRuleApiKeyType:   haTypeNodeAffinity}
	if config.Comment != nil && *config.Comment != "" {
		params[haRuleApiKeyComment] = *config.Comment
	}
	if config.Enabled != nil && !*config.Enabled {
		params[haRuleApiKeyDisabled] = "1"
	}
	if config.Guests != nil && len(*config.Guests) > 0 {
		haMapToApiGuests(*config.Guests, params)
	}
	if config.Nodes != nil && len(*config.Nodes) > 0 {
		haMapToApiNodes(*config.Nodes, params)
	}
	if config.Strict != nil && *config.Strict {
		params[haRuleApiKeyStrict] = haStrictTrue
	} else {
		params[haRuleApiKeyStrict] = haStrictFalse
	}
	return params
}

func (config HaNodeAffinityRule) mapToApiUpdate_Unsafe(current *rawHaNodeAffinityRule) map[string]any {
	params := map[string]any{
		haRuleApiKeyDigest: current.getDigest().String(),
		haRuleApiKeyType:   haTypeNodeAffinity}
	if config.Comment != nil && *config.Comment != current.GetComment() {
		params[haRuleApiKeyComment] = *config.Comment
	}
	var delete string
	if config.Enabled != nil && *config.Enabled != current.GetEnabled() {
		if *config.Enabled {
			delete = haRuleApiKeyDisabled
		} else {
			params[haRuleApiKeyDisabled] = "1"
		}
	}
	if config.Guests != nil && haDifferenceGuests(*config.Guests, current.GetGuests()) {
		haMapToApiGuests(*config.Guests, params)
	}
	if config.Nodes != nil && haDifferenceNodes(*config.Nodes, current.GetNodes()) {
		haMapToApiNodes(*config.Nodes, params)
	}
	if config.Strict != nil && *config.Strict != current.GetStrict() {
		if *config.Strict {
			params[haRuleApiKeyStrict] = haStrictTrue
		} else {
			params[haRuleApiKeyStrict] = haStrictFalse
		}
	}
	if delete != "" {
		params[haRuleApiKeyDelete] = delete
	}
	return params
}

func (config HaNodeAffinityRule) Validate(current *HaNodeAffinityRule) error {
	if current != nil {
		return config.validateUpdate()
	}
	return config.validateCreate()
}

func (config HaNodeAffinityRule) validateCreate() error {
	if config.Guests == nil {
		return errors.New(HaNodeAffinityRule_Error_GuestsRequired)
	}
	if config.Nodes == nil {
		return errors.New(HaNodeAffinityRule_Error_NodesRequired)
	}
	return config.validateUpdate()
}

func (config HaNodeAffinityRule) validateUpdate() error {
	if err := config.ID.Validate(); err != nil {
		return err
	}
	if config.Guests != nil {
		if len(*config.Guests) == 0 {
			return errors.New(HaNodeAffinityRule_Error_GuestsEmpty)
		}
		for i := range *config.Guests {
			if err := (*config.Guests)[i].vmId.Validate(); err != nil {
				return err
			}
		}
	}
	if config.Nodes != nil {
		if len(*config.Nodes) == 0 {
			return errors.New(HaNodeAffinityRule_Error_NodesEmpty)
		}
		for i := range *config.Nodes {
			if err := (*config.Nodes)[i].Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

type RawHaResourceAffinityRule interface {
	Get() HaResourceAffinityRule
	GetAffinity() HaAffinity
	GetComment() string
	GetDigest() [sha1.Size]byte
	GetEnabled() bool
	GetGuests() []VmRef
	GetID() HaRuleID
}

type rawHaResourceAffinityRule struct {
	a map[string]any
}

func (r *rawHaResourceAffinityRule) Get() HaResourceAffinityRule {
	return HaResourceAffinityRule{
		Affinity: util.Pointer(r.GetAffinity()),
		Comment:  util.Pointer(r.GetComment()),
		Digest:   r.GetDigest(),
		Enabled:  util.Pointer(r.GetEnabled()),
		Guests:   util.Pointer(r.GetGuests()),
		ID:       r.GetID(),
	}
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

func (r *rawHaResourceAffinityRule) get() *HaResourceAffinityRule {
	return &HaResourceAffinityRule{
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
	HaResourceAffinityRule_Error_AffinityRequired = "affinity must be specified during creation"
	HaResourceAffinityRule_Error_GuestsEmpty      = "guests must not be empty"
	HaResourceAffinityRule_Error_GuestsRequired   = "guests must be specified during creation"
	HaResourceAffinityRule_Error_Kind             = "rule is not a resource affinity rule"
)

func (config HaResourceAffinityRule) Create(ctx context.Context, c *Client) error {
	return c.new().haCreateResourceAffinityRule(ctx, config)
}

func (c *clientNew) haCreateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error {
	if err := haVersionCheck(ctx, c); err != nil {
		return err
	}
	return ha.create(ctx, c.api)
}

func (config HaResourceAffinityRule) CreateNoCheck(ctx context.Context, c *Client) error {
	return c.new().haCreateResourceAffinityRuleNoCheck(ctx, config)
}

func (c *clientNew) haCreateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error {
	return ha.create(ctx, c.api)
}

func (config HaResourceAffinityRule) create(ctx context.Context, c clientApiInterface) error {
	return c.createHaRule(ctx, config.mapToApiCreate())
}

func (config HaResourceAffinityRule) Update(ctx context.Context, c *Client) error {
	return c.new().haUpdateResourceAffinityRule(ctx, config)
}

func (c *clientNew) haUpdateResourceAffinityRule(ctx context.Context, ha HaResourceAffinityRule) error {
	if err := haVersionCheck(ctx, c); err != nil {
		return err
	}
	if err := ha.validateUpdate(); err != nil {
		return err
	}
	rawRule, err := ha.ID.get(ctx, c.api)
	if err != nil {
		return err
	}
	raw, ok := rawRule.getResourceAffinity()
	if !ok {
		return errors.New(HaResourceAffinityRule_Error_Kind)
	}
	return ha.update(ctx, raw, c.api)
}

func (config HaResourceAffinityRule) UpdateNoCheck(ctx context.Context, c *Client) error {
	return c.new().haUpdateResourceAffinityRuleNoCheck(ctx, config)
}

func (c *clientNew) haUpdateResourceAffinityRuleNoCheck(ctx context.Context, ha HaResourceAffinityRule) error {
	rawRule, err := ha.ID.get(ctx, c.api)
	if err != nil {
		return err
	}
	raw, ok := rawRule.getResourceAffinity()
	if !ok {
		return errors.New(HaResourceAffinityRule_Error_Kind)
	}
	return ha.update(ctx, raw, c.api)
}

func (config HaResourceAffinityRule) update(ctx context.Context, current *rawHaResourceAffinityRule, c clientApiInterface) error {
	params := config.mapToApiUpdate_Unsafe(current)
	if len(params) <= haRuleMinimumParams {
		return nil
	}
	return c.updateHaRule(ctx, config.ID, params)
}

func (config HaResourceAffinityRule) mapToApiCreate() map[string]any {
	params := map[string]any{
		haRuleApiKeyRuleID: config.ID.String(),
		haRuleApiKeyType:   haTypeResourceAffinity}
	if config.Affinity != nil {
		params[haRuleApiKeyAffinity] = config.Affinity.String()
	}
	if config.Comment != nil && *config.Comment != "" {
		params[haRuleApiKeyComment] = *config.Comment
	}
	if config.Enabled != nil && !*config.Enabled {
		params[haRuleApiKeyDisabled] = "1"
	}
	if config.Guests != nil && len(*config.Guests) > 0 {
		haMapToApiGuests(*config.Guests, params)
	}
	return params
}

func (config HaResourceAffinityRule) mapToApiUpdate_Unsafe(current *rawHaResourceAffinityRule) map[string]any {
	params := map[string]any{
		haRuleApiKeyDigest: current.getDigest().String(),
		haRuleApiKeyType:   haTypeResourceAffinity}
	if config.Affinity != nil && *config.Affinity != current.GetAffinity() {
		params[haRuleApiKeyAffinity] = config.Affinity.String()
	}
	if config.Comment != nil && *config.Comment != current.GetComment() {
		params[haRuleApiKeyComment] = *config.Comment
	}
	var delete string
	if config.Enabled != nil && *config.Enabled != current.GetEnabled() {
		if *config.Enabled {
			delete = haRuleApiKeyDisabled
		} else {
			params[haRuleApiKeyDisabled] = "1"
		}
	}
	if config.Guests != nil && haDifferenceGuests(*config.Guests, current.GetGuests()) {
		haMapToApiGuests(*config.Guests, params)
	}
	if delete != "" {
		params[haRuleApiKeyDelete] = delete
	}
	return params
}

func (config HaResourceAffinityRule) Validate(current *HaResourceAffinityRule) error {
	if current != nil {
		return config.validateUpdate()
	}
	return config.validateCreate()
}

func (config HaResourceAffinityRule) validateCreate() error {
	if config.Affinity == nil {
		return errors.New(HaResourceAffinityRule_Error_AffinityRequired)
	}
	if config.Guests == nil {
		return errors.New(HaResourceAffinityRule_Error_GuestsRequired)
	}
	return config.validateUpdate()
}

func (config HaResourceAffinityRule) validateUpdate() error {
	if err := config.ID.Validate(); err != nil {
		return err
	}
	if config.Affinity != nil {
		if err := config.Affinity.Validate(); err != nil {
			return err
		}
	}
	if config.Guests != nil {
		if len(*config.Guests) == 0 {
			return errors.New(HaResourceAffinityRule_Error_GuestsEmpty)
		}
		for i := range *config.Guests {
			if err := (*config.Guests)[i].vmId.Validate(); err != nil {
				return err
			}
		}
	}
	return nil
}

const (
	haRuleApiKeyAffinity  string = "affinity"
	haRuleApiKeyComment   string = "comment"
	haRuleApiKeyDelete    string = "delete"
	haRuleApiKeyDigest    string = "digest"
	haRuleApiKeyDisabled  string = "disable"
	haRuleApiKeyNodes     string = "nodes"
	haRuleApiKeyResources string = "resources"
	haRuleApiKeyRuleID    string = "rule"
	haRuleApiKeyStrict    string = "strict"
	haRuleApiKeyType      string = "type"
)

const (
	haRuleMinimumParams = 2 // type and digest
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

func haDifferenceGuests(a, b []VmRef) bool {
	if len(a) != len(b) {
		return true
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i].vmId < a[j].vmId
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].vmId < b[j].vmId
	})
	for i := range a {
		if a[i].vmId != b[i].vmId || a[i].vmType != b[i].vmType {
			return true
		}
	}
	return false
}

func haDifferenceNodes(a, b []HaNode) bool {
	if len(a) != len(b) {
		return true
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i].Node < a[j].Node
	})
	sort.Slice(b, func(i, j int) bool {
		return b[i].Node < b[j].Node
	})
	for i := range a {
		if a[i].Node != b[i].Node || a[i].Priority != b[i].Priority {
			return true
		}
	}
	return false
}

func haMapToApiGuests(guests []VmRef, params map[string]any) {
	resources := make([]string, len(guests))
	for i := range guests {
		switch (guests)[i].vmType {
		case GuestQemu:
			resources[i] = haGuestPrefixVm + (guests)[i].vmId.String()
		case GuestLxc:
			resources[i] = haGuestPrefixCt + (guests)[i].vmId.String()
		}
	}
	params[haRuleApiKeyResources] = resources
}

func haMapToApiNodes(nodes []HaNode, params map[string]any) {
	builder := strings.Builder{}
	for i := range nodes {
		builder.WriteString(",")
		if nodes[i].Priority == 0 {
			builder.WriteString(string(nodes[i].Node.String()))
		} else {
			builder.WriteString(string(nodes[i].Node.String()))
			builder.WriteString(":")
			builder.WriteString(nodes[i].Priority.String())
		}
	}
	if builder.Len() > 0 {
		params[haRuleApiKeyNodes] = builder.String()[1:] // cut off first comma
	} else {
		params[haRuleApiKeyNodes] = builder.String()
	}
}

func haGetID(params map[string]any) HaRuleID {
	if v, ok := params[haRuleApiKeyRuleID]; ok {
		return HaRuleID(v.(string))
	}
	return ""
}

// HaAffinity is an enum.
type HaAffinity int8

const (
	HaAffinityPositive HaAffinity = 1
	HaAffinityUnknown  HaAffinity = 0
	HaAffinityNegative HaAffinity = -1
)

const HaAffinity_Error_Invalid = "affinity must be either positive or negative"

func (a HaAffinity) String() string {
	switch a {
	case HaAffinityPositive:
		return "positive"
	case HaAffinityNegative:
		return "negative"
	default:
		return ""
	}
}

func (a HaAffinity) Validate() error {
	if a != HaAffinityPositive && a != HaAffinityNegative {
		return errors.New(HaAffinity_Error_Invalid)
	}
	return nil
}

// HaRuleID has a minimim of 2 characters and max of 128 characters.
type HaRuleID string

var haRuleIdRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)

const (
	HaRuleID_Error_MinLength = `ha rule ID must at least be 2 characters`
	HaRuleID_Error_MaxLength = `ha rule ID has a maximum of 128 characters`
	HaRuleID_Error_Invalid   = `ha rule ID did not match the following regex '^[a-zA-Z][a-zA-Z0-9\-_]{2,127}$'`
	HaRuleID_Error_Start     = `ha rule ID can only start with a lower or upper case letter`
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

// HaRuleKind is an enum.
type HaRuleKind int8

const (
	HaRuleKindUnknown          HaRuleKind = 0
	HaRuleKindNodeAffinity     HaRuleKind = 1
	HaRuleKindResourceAffinity HaRuleKind = 2
)

// Max 1000
type HaPriority uint16

const (
	HaPriority_Error_Invalid = "priority must be between 0 and 1000"
	HaPriorityMax            = 1000
)

func (p HaPriority) String() string { return strconv.Itoa(int(p)) } // for fmt.Stringer interface

func (p HaPriority) Validate() error {
	if p > HaPriorityMax {
		return errors.New(HaPriority_Error_Invalid)
	}
	return nil
}

type HaNode struct {
	Node     NodeName
	Priority HaPriority
}

func (n HaNode) Validate() error {
	if err := n.Node.Validate(); err != nil {
		return err
	}
	return n.Priority.Validate()
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
