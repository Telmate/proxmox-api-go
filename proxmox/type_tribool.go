package proxmox

import (
	"encoding/json"
	"errors"
	"strings"
)

type TriBool int8

const (
	TriBoolFalse          TriBool = -1
	TriBoolNone           TriBool = 0
	TriBoolTrue           TriBool = 1
	TriBool_Error_Invalid string  = "invalid value for TriBool"
)

func (b TriBool) MarshalJSON() ([]byte, error) {
	var str string
	switch b {
	case TriBoolTrue:
		str = "true"
	case TriBoolFalse:
		str = "false"
	case TriBoolNone:
		str = "none"
	default:
		return nil, errors.New(TriBool_Error_Invalid)
	}
	return json.Marshal(str)
}

func (b *TriBool) UnmarshalJSON(data []byte) error {
	// Trim the quotes from the JSON string value
	str := strings.Trim(string(data), "\"")
	for _, v := range []string{"true", "yes", "on"} {
		if strings.EqualFold(str, v) {
			*b = TriBoolTrue
			return nil
		}
	}
	for _, v := range []string{"false", "no", "off"} {
		if strings.EqualFold(str, v) {
			*b = TriBoolFalse
			return nil
		}
	}
	for _, v := range []string{"none", ""} {
		if strings.EqualFold(str, v) {
			*b = TriBoolNone
			return nil
		}
	}
	return errors.New(TriBool_Error_Invalid)
}

func (b TriBool) Validate() error {
	switch b {
	case TriBoolTrue, TriBoolFalse, TriBoolNone:
		return nil
	default:
		return errors.New(TriBool_Error_Invalid)
	}
}
