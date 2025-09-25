package parse

import (
	"errors"
	"strconv"
)

const (
	InvalidType = "invalid type"
	Negative    = "negative value"
	Empty       = "empty value"
)

func Bool(input any) (bool, error) {
	switch input := input.(type) {
	case float64:
		return input == 1, nil
	case string:
		if input == "" {
			return false, errors.New(Empty)
		}
		return input[0] == '1', nil
	}
	return false, errors.New(InvalidType)
}

func Int(input any) (int, error) {
	switch input := input.(type) {
	case float64:
		return int(input), nil
	case string:
		mem, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return 0, err
		}
		return int(mem), nil
	}
	return 0, errors.New(InvalidType)
}

func Uint(input any) (uint, error) {
	tmpInt, err := Int(input)
	if err != nil {
		return 0, err
	}
	if tmpInt < 0 {
		return 0, errors.New(Negative)
	}
	return uint(tmpInt), nil
}
