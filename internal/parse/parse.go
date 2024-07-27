package parse

import (
	"errors"
	"strconv"
)

func Int(input interface{}) (int, error) {
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
	return 0, nil
}

func Uint(input interface{}) (uint, error) {
	tmpInt, err := Int(input)
	if err != nil {
		return 0, err
	}
	if tmpInt < 0 {
		return 0, errors.New("negative value")
	}
	return uint(tmpInt), nil
}
