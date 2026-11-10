package proxmox

import "errors"

func deepUnwrap(err error) error {
	var newErr error
	for {
		newErr = errors.Unwrap(err)
		if newErr == nil {
			return err
		}
		err = newErr
	}
}
