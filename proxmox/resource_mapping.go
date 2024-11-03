package proxmox

import (
	"errors"
	"regexp"
	"unicode"
)

// minimum length: 2
// ,maximum length: 128
// ,regex: ^\w(\w|\d|_|-){1,127}$
type ResourceMappingUsbID string

var resourceMappingUsbID = regexp.MustCompile(`^(\w|\d|_|-)+$`)

const (
	ResourceMappingUsbID_Error_MaxLength string = "usb id is too long"
	ResourceMappingUsbID_Error_MinLength string = "usb id is too short"
	ResourceMappingUsbID_Error_Start     string = "usb id must start with a letter"
	ResourceMappingUsbID_Error_Invalid   string = "usb id should match the following regex: '^\\w(\\w|\\d|_|-){1,127}$'"
)

func (id ResourceMappingUsbID) String() string {
	return string(id)
}

func (id ResourceMappingUsbID) Validate() error {
	if len(id) < 2 {
		return errors.New(ResourceMappingUsbID_Error_MinLength)
	}
	if len(id) > 128 {
		return errors.New(ResourceMappingUsbID_Error_MaxLength)
	}
	if !unicode.IsLetter(rune(id[0])) {
		return errors.New(ResourceMappingUsbID_Error_Start)
	}
	if !resourceMappingUsbID.MatchString(string(id)) {
		return errors.New(ResourceMappingUsbID_Error_Invalid)
	}
	return nil
}
