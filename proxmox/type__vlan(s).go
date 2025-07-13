package proxmox

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

type Vlan uint16 // 0-4095, 0 means no vlan

const (
	VlanMaximum        Vlan   = 4095
	Vlan_Error_Invalid string = "vlan tag must be in the range 0-4095"
)

func (config Vlan) String() string {
	return strconv.FormatInt(int64(config), 10)
}

func (config Vlan) Validate() error {
	if config > VlanMaximum {
		return errors.New(Vlan_Error_Invalid)
	}
	return nil
}

type Vlans []Vlan

func (config Vlans) string() string {
	if len(config) == 0 {
		return ""
	}
	// Use a map to track seen elements and remove duplicates.
	uniqueMap := make(map[int]struct{})
	// Iterate over the input slice and add unique elements to the result slice.
	for i := range config {
		uniqueMap[int(config[i])] = struct{}{}
	}
	uniqueArr := make([]int, len(uniqueMap))
	var index int
	for key := range uniqueMap {
		uniqueArr[index] = key
		index++
	}
	slices.Sort(uniqueArr)
	builder := strings.Builder{}
	for i := range uniqueArr {
		builder.WriteString(";" + strconv.Itoa(uniqueArr[i]))
	}
	return builder.String()[1:] // Skip the leading semicolon
}

func (config Vlans) Validate() error {
	for _, vlan := range config {
		if err := vlan.Validate(); err != nil {
			return err
		}
	}
	return nil
}
