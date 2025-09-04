package proxmox

import "crypto/sha1"

// RawConfigLXCMock is a mock implementation of the RawConfigLXC interface
type RawConfigLXCMock struct {
	GetFunc                func(vmr VmRef, state PowerState) *ConfigLXC
	GetArchitectureFunc    func() CpuArchitecture
	GetBootMountFunc       func() *LxcBootMount
	GetDNSFunc             func() *GuestDNS
	GetDescriptionFunc     func() *string
	GetDigestFunc          func() [sha1.Size]byte
	GetMemoryFunc          func() LxcMemory
	GetMountsFunc          func() LxcMounts
	GetNameFunc            func() GuestName
	GetOperatingSystemFunc func() OperatingSystem
	GetPrivilegedFunc      func() bool
	GetProtectionFunc      func() bool
	GetSwapFunc            func() LxcSwap
	GetTagsFunc            func() *Tags
}

func (m *RawConfigLXCMock) panic(field string) { panic(field + " not set in RawConfigLXCMock") }

// Interface methods

func (m *RawConfigLXCMock) Get(vmr VmRef, state PowerState) *ConfigLXC {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc(vmr, state)
}

func (m *RawConfigLXCMock) GetArchitecture() CpuArchitecture {
	if m.GetArchitectureFunc == nil {
		m.panic("GetArchitectureFunc")
	}
	return m.GetArchitectureFunc()
}

func (m *RawConfigLXCMock) GetBootMount() *LxcBootMount {
	if m.GetBootMountFunc == nil {
		m.panic("GetBootMountFunc")
	}
	return m.GetBootMountFunc()
}

func (m *RawConfigLXCMock) GetDNS() *GuestDNS {
	if m.GetDNSFunc == nil {
		m.panic("GetDNSFunc")
	}
	return m.GetDNSFunc()
}

func (m *RawConfigLXCMock) GetDescription() *string {
	if m.GetDescriptionFunc == nil {
		m.panic("GetDescriptionFunc")
	}
	return m.GetDescriptionFunc()
}

func (m *RawConfigLXCMock) GetDigest() [sha1.Size]byte {
	if m.GetDigestFunc == nil {
		m.panic("GetDigestFunc")
	}
	return m.GetDigestFunc()
}

func (m *RawConfigLXCMock) GetMemory() LxcMemory {
	if m.GetMemoryFunc == nil {
		m.panic("GetMemoryFunc")
	}
	return m.GetMemoryFunc()
}

func (m *RawConfigLXCMock) GetMounts() LxcMounts {
	if m.GetMountsFunc == nil {
		m.panic("GetMountsFunc")
	}
	return m.GetMountsFunc()
}

func (m *RawConfigLXCMock) GetName() GuestName {
	if m.GetNameFunc == nil {
		m.panic("GetNameFunc")
	}
	return m.GetNameFunc()
}

func (m *RawConfigLXCMock) GetOperatingSystem() OperatingSystem {
	if m.GetOperatingSystemFunc == nil {
		m.panic("GetOperatingSystemFunc")
	}
	return m.GetOperatingSystemFunc()
}

func (m *RawConfigLXCMock) GetPrivileged() bool {
	if m.GetPrivilegedFunc == nil {
		m.panic("GetPrivilegedFunc")
	}
	return m.GetPrivilegedFunc()
}

func (m *RawConfigLXCMock) GetProtection() bool {
	if m.GetProtectionFunc == nil {
		m.panic("GetProtectionFunc")
	}
	return m.GetProtectionFunc()
}

func (m *RawConfigLXCMock) GetSwap() LxcSwap {
	if m.GetSwapFunc == nil {
		m.panic("GetSwapFunc")
	}
	return m.GetSwapFunc()
}

func (m *RawConfigLXCMock) GetTags() *Tags {
	if m.GetTagsFunc == nil {
		m.panic("GetTagsFunc")
	}
	return m.GetTagsFunc()
}

func (m *RawConfigLXCMock) get(vmr VmRef) *ConfigLXC {
	panic("get not implemented in RawConfigLXCMock")
}

func (m *RawConfigLXCMock) getBootMount(privileged bool) *LxcBootMount {
	panic("getBootMount not implemented in RawConfigLXCMock")
}

func (m *RawConfigLXCMock) getDigest() [sha1.Size]byte {
	panic("getDigest not implemented in RawConfigLXCMock")
}

func (m *RawConfigLXCMock) getMounts(privileged bool) LxcMounts {
	panic("getMounts not implemented in RawConfigLXCMock")
}
