package proxmox

// RawConfigPoolMock is a mock implementation of the RawConfigPool interface
type RawConfigPoolMock struct {
	GetFunc        func() ConfigPool
	GetNameFunc    func() PoolName
	GetCommentFunc func() string
	GetGuestsFunc  func() *[]GuestID
}

func (m *RawConfigPoolMock) panic(field string) { panic(field + " not set in RawConfigPoolMock") }

// Interface methods

func (m *RawConfigPoolMock) Get() ConfigPool {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc()
}

func (m *RawConfigPoolMock) GetName() PoolName {
	if m.GetNameFunc == nil {
		m.panic("GetNameFunc")
	}
	return m.GetNameFunc()
}

func (m *RawConfigPoolMock) GetComment() string {
	if m.GetCommentFunc == nil {
		m.panic("GetCommentFunc")
	}
	return m.GetCommentFunc()
}

func (m *RawConfigPoolMock) GetGuests() *[]GuestID {
	if m.GetGuestsFunc == nil {
		m.panic("GetGuestsFunc")
	}
	return m.GetGuestsFunc()
}
