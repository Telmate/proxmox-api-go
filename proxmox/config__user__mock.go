package proxmox

// RawConfigUserMock is a mock implementation of the RawConfigUser interface
type RawConfigUserMock struct {
	GetFunc          func() *ConfigUser
	GetCommentFunc   func() string
	GetEmailFunc     func() string
	GetEnableFunc    func() bool
	GetExpireFunc    func() uint
	GetFirstNameFunc func() string
	GetGroupsFunc    func() []GroupName
	GetKeysFunc      func() string
	GetLastNameFunc  func() string
	GetUserFunc      func() UserID
}

func (m *RawConfigUserMock) panic(field string) { panic(field + " not set in RawConfigUserMock") }

// Interface methods

func (m *RawConfigUserMock) Get() *ConfigUser {
	if m.GetFunc == nil {
		m.panic("GetFunc")
	}
	return m.GetFunc()
}

func (m *RawConfigUserMock) GetComment() string {
	if m.GetCommentFunc == nil {
		m.panic("GetCommentFunc")
	}
	return m.GetCommentFunc()
}

func (m *RawConfigUserMock) GetEmail() string {
	if m.GetEmailFunc == nil {
		m.panic("GetEmailFunc")
	}
	return m.GetEmailFunc()
}

func (m *RawConfigUserMock) GetEnable() bool {
	if m.GetEnableFunc == nil {
		m.panic("GetEnableFunc")
	}
	return m.GetEnableFunc()
}

func (m *RawConfigUserMock) GetExpire() uint {
	if m.GetExpireFunc == nil {
		m.panic("GetExpireFunc")
	}
	return m.GetExpireFunc()
}

func (m *RawConfigUserMock) GetFirstName() string {
	if m.GetFirstNameFunc == nil {
		m.panic("GetFirstNameFunc")
	}
	return m.GetFirstNameFunc()
}

func (m *RawConfigUserMock) GetGroups() []GroupName {
	if m.GetGroupsFunc == nil {
		m.panic("GetGroupsFunc")
	}
	return m.GetGroupsFunc()
}

func (m *RawConfigUserMock) GetKeys() string {
	if m.GetKeysFunc == nil {
		m.panic("GetKeysFunc")
	}
	return m.GetKeysFunc()
}

func (m *RawConfigUserMock) GetLastName() string {
	if m.GetLastNameFunc == nil {
		m.panic("GetLastNameFunc")
	}
	return m.GetLastNameFunc()
}

func (m *RawConfigUserMock) GetUser() UserID {
	if m.GetUserFunc == nil {
		m.panic("GetUserFunc")
	}
	return m.GetUserFunc()
}
