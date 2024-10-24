package proxmox

import "errors"

type MTU uint16 // minimum value 576 - 65520

const MTU_Error_Invalid string = "mtu must be in the range 576-65520"

func (mtu MTU) Validate() error {
	if mtu == 0 || (mtu > 575 && mtu < 65521) {
		return nil
	}
	return errors.New(MTU_Error_Invalid)
}
