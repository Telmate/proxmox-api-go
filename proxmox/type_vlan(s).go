package proxmox

import (
	"errors"
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

func (config *Vlans) mapToApiUnsafe() string {
	// Use a map to track seen elements and remove duplicates.
	seen := make(map[Vlan]bool)
	result := make([]int, 0, len(*config))
	// Iterate over the input slice and add unique elements to the result slice.
	for _, value := range *config {
		if _, ok := seen[value]; !ok {
			seen[value] = true
			result = append(result, int(value))
		}
	}
	builder := strings.Builder{}
	for _, vlan := range result {
		builder.WriteString(";" + strconv.Itoa(vlan))
	}
	return builder.String()
}

func (config Vlans) Validate() error {
	for _, vlan := range config {
		if err := vlan.Validate(); err != nil {
			return err
		}
	}
	return nil
}
