package proxmox

// minimum length: 2
// ,maximum length: 128
// ,regex: ^\w(\w|\d|_|-){1,127}$
type ResourceMappingUsbID string

func (id ResourceMappingUsbID) String() string {
	return string(id)
}
