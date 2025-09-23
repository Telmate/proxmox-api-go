package proxmox

import "crypto/sha1"

type HaRulesMock struct {
	ConvertArrayFunc func() []HaRule
	ConvertMapFunc   func() map[HaRuleID]HaRule
}

func (m *HaRulesMock) panic(field string) { panic(field + " not set in HaRulesMock") }

func (m *HaRulesMock) ConvertArray() []HaRule {
	if m.ConvertArrayFunc == nil {
		m.panic("ConvertArrayFunc")
	}
	return m.ConvertArrayFunc()
}

func (m *HaRulesMock) ConvertMap() map[HaRuleID]HaRule {
	if m.ConvertMapFunc == nil {
		m.panic("ConvertMapFunc")
	}
	return m.ConvertMapFunc()
}

type HaRuleMock struct {
	GetCommentFunc          func() string
	GetDigestFunc           func() [sha1.Size]byte
	GetEnabledFunc          func() bool
	GetIDFunc               func() HaRuleID
	GetNodeAffinityFunc     func() (RawHaNodeAffinityRule, bool)
	GetResourceAffinityFunc func() (RawHaResourceAffinityRule, bool)
	KindFunc                func() HaRuleKind
}

func (m *HaRuleMock) panic(field string) { panic(field + " not set in HaRuleMock") }

func (m *HaRuleMock) GetComment() string {
	if m.GetCommentFunc == nil {
		m.panic("GetCommentFunc")
	}
	return m.GetCommentFunc()
}

func (m *HaRuleMock) GetDigest() [sha1.Size]byte {
	if m.GetDigestFunc == nil {
		m.panic("GetDigestFunc")
	}
	return m.GetDigestFunc()
}

func (m *HaRuleMock) GetEnabled() bool {
	if m.GetEnabledFunc == nil {
		m.panic("GetEnabledFunc")
	}
	return m.GetEnabledFunc()
}

func (m *HaRuleMock) GetID() HaRuleID {
	if m.GetIDFunc == nil {
		m.panic("GetIDFunc")
	}
	return m.GetIDFunc()
}

func (m *HaRuleMock) GetNodeAffinity() (RawHaNodeAffinityRule, bool) {
	if m.GetNodeAffinityFunc == nil {
		m.panic("GetNodeAffinityFunc")
	}
	return m.GetNodeAffinityFunc()
}

func (m *HaRuleMock) GetResourceAffinity() (RawHaResourceAffinityRule, bool) {
	if m.GetResourceAffinityFunc == nil {
		m.panic("GetResourceAffinityFunc")
	}
	return m.GetResourceAffinityFunc()
}

func (m *HaRuleMock) Kind() HaRuleKind {
	if m.KindFunc == nil {
		m.panic("KindFunc")
	}
	return m.KindFunc()
}

type RawHaNodeAffinityRuleMock struct {
	GetFunc        func() HaNodeAffinityRule
	GetCommentFunc func() string
	GetDigestFunc  func() [sha1.Size]byte
	GetEnabledFunc func() bool
	GetGuestsFunc  func() []VmRef
	GetIDFunc      func() HaRuleID
	GetNodesFunc   func() []HaNode
	GetStrictFunc  func() bool
}

func (m *RawHaNodeAffinityRuleMock) panic(field string) {
	panic(field + " not set in RawHaNodeAffinityRuleMock")
}

func (m *RawHaNodeAffinityRuleMock) Get() HaNodeAffinityRule {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetComment() string {
	if m.GetCommentFunc == nil {
		m.panic("GetCommentFunc")
	}
	return m.GetCommentFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetDigest() [sha1.Size]byte {
	if m.GetDigestFunc == nil {
		m.panic("GetDigestFunc")
	}
	return m.GetDigestFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetEnabled() bool {
	if m.GetEnabledFunc == nil {
		m.panic("GetEnabledFunc")
	}
	return m.GetEnabledFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetGuests() []VmRef {
	if m.GetGuestsFunc == nil {
		m.panic("GetGuestsFunc")
	}
	return m.GetGuestsFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetID() HaRuleID {
	if m.GetIDFunc == nil {
		m.panic("GetIDFunc")
	}
	return m.GetIDFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetNodes() []HaNode {
	if m.GetNodesFunc == nil {
		m.panic("GetNodesFunc")
	}
	return m.GetNodesFunc()
}

func (m *RawHaNodeAffinityRuleMock) GetStrict() bool {
	if m.GetStrictFunc == nil {
		m.panic("GetStrictFunc")
	}
	return m.GetStrictFunc()
}

func (m *RawHaNodeAffinityRuleMock) get() HaNodeAffinityRule {
	panic("get not implemented in RawHaNodeAffinityRuleMock")
}

func (m *RawHaNodeAffinityRuleMock) getDigest() digest {
	panic("getDigest not implemented in RawHaNodeAffinityRuleMock")
}

type RawHaResourceAffinityRuleMock struct {
	GetFunc         func() HaResourceAffinityRule
	GetAffinityFunc func() HaAffinity
	GetCommentFunc  func() string
	GetDigestFunc   func() [sha1.Size]byte
	GetEnabledFunc  func() bool
	GetGuestsFunc   func() []VmRef
	GetIDFunc       func() HaRuleID
}

func (m *RawHaResourceAffinityRuleMock) panic(field string) {
	panic(field + " not set in RawHaResourceAffinityRuleMock")
}

func (m *RawHaResourceAffinityRuleMock) Get() HaResourceAffinityRule {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetAffinity() HaAffinity {
	if m.GetAffinityFunc == nil {
		m.panic("GetAffinityFunc")
	}
	return m.GetAffinityFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetComment() string {
	if m.GetCommentFunc == nil {
		m.panic("GetCommentFunc")
	}
	return m.GetCommentFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetDigest() [sha1.Size]byte {
	if m.GetDigestFunc == nil {
		m.panic("GetDigestFunc")
	}
	return m.GetDigestFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetEnabled() bool {
	if m.GetEnabledFunc == nil {
		m.panic("GetEnabledFunc")
	}
	return m.GetEnabledFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetGuests() []VmRef {
	if m.GetGuestsFunc == nil {
		m.panic("GetGuestsFunc")
	}
	return m.GetGuestsFunc()
}

func (m *RawHaResourceAffinityRuleMock) GetID() HaRuleID {
	if m.GetIDFunc == nil {
		m.panic("GetIDFunc")
	}
	return m.GetIDFunc()
}

func (m *RawHaResourceAffinityRuleMock) get() HaResourceAffinityRule {
	panic("get not implemented in RawHaResourceAffinityRuleMock")
}

func (m *RawHaResourceAffinityRuleMock) getDigest() digest {
	panic("getDigest not implemented in RawHaResourceAffinityRuleMock")
}
