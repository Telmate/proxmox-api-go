package proxmox

import "time"

// RawGuestResourceMock is a mock implementation of the RawGuestResource interface
type RawGuestResourceMock struct {
	GetFunc                   func() GuestResource
	GetCPUcoresFunc           func() uint
	GetCPUusageFunc           func() float64
	GetDiskReadTotalFunc      func() uint
	GetDiskSizeInBytesFunc    func() uint
	GetDiskUsedInBytesFunc    func() uint
	GetDiskWriteTotalFunc     func() uint
	GetHaStateFunc            func() string
	GetIDFunc                 func() GuestID
	GetMemoryTotalInBytesFunc func() uint
	GetMemoryUsedInBytesFunc  func() uint
	GetNameFunc               func() GuestName
	GetNetworkInFunc          func() uint
	GetNetworkOutFunc         func() uint
	GetNodeFunc               func() NodeName
	GetPoolFunc               func() PoolName
	GetStatusFunc             func() PowerState
	GetTagsFunc               func() Tags
	GetTemplateFunc           func() bool
	GetTypeFunc               func() GuestType
	GetUptimeFunc             func() time.Duration
}

func (m *RawGuestResourceMock) panic(field string) { panic(field + " not set in RawGuestResourceMock") }

// Interface methods

func (m *RawGuestResourceMock) Get() GuestResource {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc()
}

func (m *RawGuestResourceMock) GetCPUcores() uint {
	if m.GetCPUcoresFunc == nil {
		m.panic("GetCPUcoresFunc")
	}
	return m.GetCPUcoresFunc()
}

func (m *RawGuestResourceMock) GetCPUusage() float64 {
	if m.GetCPUusageFunc == nil {
		m.panic("GetCPUusageFunc")
	}
	return m.GetCPUusageFunc()
}

func (m *RawGuestResourceMock) GetDiskReadTotal() uint {
	if m.GetDiskReadTotalFunc == nil {
		m.panic("GetDiskReadTotalFunc")
	}
	return m.GetDiskReadTotalFunc()
}

func (m *RawGuestResourceMock) GetDiskSizeInBytes() uint {
	if m.GetDiskSizeInBytesFunc == nil {
		m.panic("GetDiskSizeInBytesFunc")
	}
	return m.GetDiskSizeInBytesFunc()
}

func (m *RawGuestResourceMock) GetDiskUsedInBytes() uint {
	if m.GetDiskUsedInBytesFunc == nil {
		m.panic("GetDiskUsedInBytesFunc")
	}
	return m.GetDiskUsedInBytesFunc()
}

func (m *RawGuestResourceMock) GetDiskWriteTotal() uint {
	if m.GetDiskWriteTotalFunc == nil {
		m.panic("GetDiskWriteTotalFunc")
	}
	return m.GetDiskWriteTotalFunc()
}

func (m *RawGuestResourceMock) GetHaState() string {
	if m.GetHaStateFunc == nil {
		m.panic("GetHaStateFunc")
	}
	return m.GetHaStateFunc()
}

func (m *RawGuestResourceMock) GetID() GuestID {
	if m.GetIDFunc == nil {
		m.panic("GetIDFunc")
	}
	return m.GetIDFunc()
}

func (m *RawGuestResourceMock) GetMemoryTotalInBytes() uint {
	if m.GetMemoryTotalInBytesFunc == nil {
		m.panic("GetMemoryTotalInBytesFunc")
	}
	return m.GetMemoryTotalInBytesFunc()
}

func (m *RawGuestResourceMock) GetMemoryUsedInBytes() uint {
	if m.GetMemoryUsedInBytesFunc == nil {
		m.panic("GetMemoryUsedInBytesFunc")
	}
	return m.GetMemoryUsedInBytesFunc()
}

func (m *RawGuestResourceMock) GetName() GuestName {
	if m.GetNameFunc == nil {
		m.panic("GetNameFunc")
	}
	return m.GetNameFunc()
}

func (m *RawGuestResourceMock) GetNetworkIn() uint {
	if m.GetNetworkInFunc == nil {
		m.panic("GetNetworkInFunc")
	}
	return m.GetNetworkInFunc()
}

func (m *RawGuestResourceMock) GetNetworkOut() uint {
	if m.GetNetworkOutFunc == nil {
		m.panic("GetNetworkOutFunc")
	}
	return m.GetNetworkOutFunc()
}

func (m *RawGuestResourceMock) GetNode() NodeName {
	if m.GetNodeFunc == nil {
		m.panic("GetNodeFunc")
	}
	return m.GetNodeFunc()
}

func (m *RawGuestResourceMock) GetPool() PoolName {
	if m.GetPoolFunc == nil {
		m.panic("GetPoolFunc")
	}
	return m.GetPoolFunc()
}

func (m *RawGuestResourceMock) GetStatus() PowerState {
	if m.GetStatusFunc == nil {
		m.panic("GetStatusFunc")
	}
	return m.GetStatusFunc()
}

func (m *RawGuestResourceMock) GetTags() Tags {
	if m.GetTagsFunc == nil {
		m.panic("GetTagsFunc")
	}
	return m.GetTagsFunc()
}

func (m *RawGuestResourceMock) GetTemplate() bool {
	if m.GetTemplateFunc == nil {
		m.panic("GetTemplateFunc")
	}
	return m.GetTemplateFunc()
}

func (m *RawGuestResourceMock) GetType() GuestType {
	if m.GetTypeFunc == nil {
		m.panic("GetTypeFunc")
	}
	return m.GetTypeFunc()
}

func (m *RawGuestResourceMock) GetUptime() time.Duration {
	if m.GetUptimeFunc == nil {
		m.panic("GetUptimeFunc")
	}
	return m.GetUptimeFunc()
}
