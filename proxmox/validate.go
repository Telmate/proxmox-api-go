package proxmox

import (
	"fmt"
	"path/filepath"
)

func ValidateIntInRange(min, max, value int, text string) error {
	if value >= min && value <= max {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be between %d and %d", text, min, max)
}

func ValidateIntGreaterOrEquals(min, value int, text string) error {
	if value >= min {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be greater or equal to %d", text, min)
}

func ValidateIntGreater(min, value int, text string) error {
	if value > min {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be greater than %d", text, min)
}

func ValidateStringInArray(array []string, value, text string) error {
	err := ValidateStringNotEmpty(value, text)
	if err != nil {
		return err
	}
	if inArray(array, value) {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be one of %s", text, ArrayToCSV(array))
}

func ValidateStringNotEmpty(value, text string) error {
	if value != "" {
		return nil
	}
	return ErrorKeyEmpty(text)
}

// check if a key is allowed to be changed after creation.
func ValidateStringsEqual(value1, value2, text string) error {
	if value1 == value2 {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) may not be changed during update", text)
}

func ValidateFilePath(path, text string) error {
	err := ValidateStringNotEmpty(path, text)
	if err != nil {
		return err
	}
	if filepath.IsAbs(path) {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) is not a valid file absolute path", text)
}

func ValidateArrayNotEmpty(array interface{}, text string) error {
	if len(array.([]string)) > 0 {
		return nil
	}
	return ErrorKeyEmpty(text)
}

func ValidateArrayEven(array interface{}, text string) error {
	if len(array.([]string))%2 == 0 {
		return nil
	}
	return ErrorKeyEmpty(text)
}

func ErrorKeyEmpty(text string) error {
	return fmt.Errorf("error the value of key (%s) may not be empty", text)
}

func ErrorKeyNotSet(text string) error {
	return fmt.Errorf("error the key (%s) must be set", text)
}

func ErrorItemExists(item, text string) error {
	return fmt.Errorf("error %s with id ( %s ) already exists", text, item)
}

func ErrorItemNotExists(item, text string) error {
	return fmt.Errorf("error %s with id ( %s ) does not exist", text, item)
}
