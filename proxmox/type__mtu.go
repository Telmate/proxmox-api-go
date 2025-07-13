package proxmox

import (
	"errors"
	"strconv"
)

type MTU uint16 // minimum value 576 - 65520

const (
	MTU_Error_Invalid = "mtu must be in the range 576-65520"
	mtu_Maximum       = 65520
	mtu_Minimum       = 576
)

func (mtu MTU) String() string { // String is for fmt.Stringer.
	if mtu < mtu_Minimum || mtu > mtu_Maximum {
		return ""
	}
	return strconv.Itoa(int(mtu))
}

func (mtu MTU) string() string {
	return strconv.Itoa(int(mtu))
}

func (mtu MTU) Validate() error {
	if mtu != 0 && (mtu < mtu_Minimum || mtu > mtu_Maximum) {
		return errors.New(MTU_Error_Invalid)
	}
	return nil
}
