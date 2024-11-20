package proxmox

import (
	"errors"
	"regexp"
	"unicode"
)

var resourceMappingID = regexp.MustCompile(`^(\w|\d|_|-)+$`)

// minimum length: 2
// ,maximum length: 128
// ,regex: ^\w(\w|\d|_|-){1,127}$
type ResourceMappingUsbID string

const (
	resourceMappingUsbKey                string = "usb"
	ResourceMappingUsbID_Error_MaxLength string = resourceMappingUsbKey + mappingID_Error_MaxLength
	ResourceMappingUsbID_Error_MinLength string = resourceMappingUsbKey + mappingID_Error_MinLength
	ResourceMappingUsbID_Error_Start     string = resourceMappingUsbKey + mappingID_Error_Start
	ResourceMappingUsbID_Error_Invalid   string = resourceMappingUsbKey + mappingID_Error_Invalid
)

func (id ResourceMappingUsbID) String() string {
	return string(id)
}

func (id ResourceMappingUsbID) Validate() error {
	return mappingID(id).Validate(resourceMappingUsbKey)
}

// minimum length: 2
// ,maximum length: 128
// ,regex: ^\w(\w|\d|_|-){1,127}$
type ResourceMappingPciID string

const (
	resourceMappingPciKey                string = "pcie"
	ResourceMappingPciID_Error_MaxLength string = resourceMappingPciKey + mappingID_Error_MaxLength
	ResourceMappingPciID_Error_MinLength string = resourceMappingPciKey + mappingID_Error_MinLength
	ResourceMappingPciID_Error_Start     string = resourceMappingPciKey + mappingID_Error_Start
	ResourceMappingPciID_Error_Invalid   string = resourceMappingPciKey + mappingID_Error_Invalid
)

func (id ResourceMappingPciID) String() string {
	return string(id)
}

func (id ResourceMappingPciID) Validate() error {
	return mappingID(id).Validate(resourceMappingPciKey)
}

type mappingID string

const (
	mappingID_Error_MaxLength string = " mapping id is too long"
	mappingID_Error_MinLength string = " mapping id is too short"
	mappingID_Error_Start     string = " mapping id must start with a letter"
	mappingID_Error_Invalid   string = ` mapping id should match the following regex: '^\w(\w|\d|_|-){1,127}$'`
)

func (id mappingID) Validate(kind string) error {
	if len(id) < 2 {
		return errors.New(kind + mappingID_Error_MinLength)
	}
	if len(id) > 128 {
		return errors.New(kind + mappingID_Error_MaxLength)
	}
	if !unicode.IsLetter(rune(id[0])) {
		return errors.New(kind + mappingID_Error_Start)
	}
	if !resourceMappingID.MatchString(string(id)) {
		return errors.New(kind + mappingID_Error_Invalid)
	}
	return nil
}
