package proxmox

import (
	"errors"
	"fmt"
)

func ValidateIntInRange(min, max, value int, text string) error{
	if value >= min && value <= max{return nil}
	return errors.New(fmt.Sprintf("error the value of key %s must be between %d and %d", text, min, max))
}