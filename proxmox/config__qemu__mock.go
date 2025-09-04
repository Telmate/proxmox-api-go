package proxmox

// RawConfigQemuMock is a mock implementation of the RawConfigQemu interface
type RawConfigQemuMock struct {
	GetAgentFunc            func() *QemuGuestAgent
	GetCPUFunc              func() *QemuCPU
	GetCloudInitFunc        func() *CloudInit
	GetDescriptionFunc      func() string
	GetFunc                 func(vmr *VmRef) (*ConfigQemu, error)
	GetMemoryFunc           func() *QemuMemory
	GetNameFunc             func() GuestName
	GetNetworksFunc         func() QemuNetworkInterfaces
	GetPciDevicesFunc       func() QemuPciDevices
	GetProtectionFunc       func() bool
	GetRandomnessDeviceFunc func() *VirtIoRNG
	GetSerialsFunc          func() SerialInterfaces
	GetTabletFunc           func() bool
	GetTagsFunc             func() *Tags
	GetUSBsFunc             func() QemuUSBs
}

func (m *RawConfigQemuMock) panic(field string) { panic(field + " not set in RawConfigQemuMock") }

// Interface methods

func (m *RawConfigQemuMock) Get(vmr *VmRef) (*ConfigQemu, error) {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc(vmr)
}

func (m *RawConfigQemuMock) GetAgent() *QemuGuestAgent {
	if m.GetAgentFunc == nil {
		m.panic("GetAgentFunc")
	}
	return m.GetAgentFunc()
}

func (m *RawConfigQemuMock) GetCPU() *QemuCPU {
	if m.GetCPUFunc == nil {
		m.panic("GetCPUFunc")
	}
	return m.GetCPUFunc()
}

func (m *RawConfigQemuMock) GetCloudInit() *CloudInit {
	if m.GetCloudInitFunc == nil {
		m.panic("GetCloudInitFunc")
	}
	return m.GetCloudInitFunc()
}

func (m *RawConfigQemuMock) GetDescription() string {
	if m.GetDescriptionFunc == nil {
		m.panic("GetDescriptionFunc")
	}
	return m.GetDescriptionFunc()
}

func (m *RawConfigQemuMock) GetMemory() *QemuMemory {
	if m.GetMemoryFunc == nil {
		m.panic("GetMemoryFunc")
	}
	return m.GetMemoryFunc()
}

func (m *RawConfigQemuMock) GetName() GuestName {
	if m.GetNameFunc == nil {
		m.panic("GetNameFunc")
	}
	return m.GetNameFunc()
}

func (m *RawConfigQemuMock) GetNetworks() QemuNetworkInterfaces {
	if m.GetNetworksFunc == nil {
		m.panic("GetNetworksFunc")
	}
	return m.GetNetworksFunc()
}

func (m *RawConfigQemuMock) GetPciDevices() QemuPciDevices {
	if m.GetPciDevicesFunc == nil {
		m.panic("GetPciDevicesFunc")
	}
	return m.GetPciDevicesFunc()
}

func (m *RawConfigQemuMock) GetProtection() bool {
	if m.GetProtectionFunc == nil {
		m.panic("GetProtectionFunc")
	}
	return m.GetProtectionFunc()
}

func (m *RawConfigQemuMock) GetRandomnessDevice() *VirtIoRNG {
	if m.GetRandomnessDeviceFunc == nil {
		m.panic("GetRandomnessDeviceFunc")
	}
	return m.GetRandomnessDeviceFunc()
}

func (m *RawConfigQemuMock) GetSerials() SerialInterfaces {
	if m.GetSerialsFunc == nil {
		m.panic("GetSerialsFunc")
	}
	return m.GetSerialsFunc()
}

func (m *RawConfigQemuMock) GetTablet() bool {
	if m.GetTabletFunc == nil {
		m.panic("GetTabletFunc")
	}
	return m.GetTabletFunc()
}

func (m *RawConfigQemuMock) GetTags() *Tags {
	if m.GetTagsFunc == nil {
		m.panic("GetTagsFunc")
	}
	return m.GetTagsFunc()
}

func (m *RawConfigQemuMock) GetUSBs() QemuUSBs {
	if m.GetUSBsFunc == nil {
		m.panic("GetUSBsFunc")
	}
	return m.GetUSBsFunc()
}

func (m *RawConfigQemuMock) get(vmr *VmRef) (*ConfigQemu, error) {
	panic("get not implemented in RawConfigQemuMock")
}
