package proxmox

import (
	"fmt"
)

func ValidateIntInRange(min, max, value int, text string) error{
	if value >= min && value <= max{
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be between %d and %d", text, min, max)
}

func ValidateIntGreater(min, value int, text string) error{
	if value >= min {
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be greater than %d", text, min)
}

func ValidateStringInArray(array []string, value, text string) error{
	if inArray(array, value){
		return nil
	}
	return fmt.Errorf("error the value of key (%s) must be one of %s", text, ArrayToCSV(array))
}

func ValidateStringNotEmpty(value, text string) error{
	if value != ""{
		return nil
	}
	return fmt.Errorf("error the value of key (%s) may not be empty", text)
}