package proxmox

import (
	"encoding/json"
	"errors"
)

type responseData[T any] struct {
	Data *T `json:"data"` // Throws error when missing
}

func (item *responseData[T]) UnmarshalJSON(data []byte) error {
	type tmp[T any] struct {
		Data *T `json:"data,omitempty"`
	}
	var required tmp[T]
	if err := json.Unmarshal(data, &required); err != nil {
		return err
	}
	if required.Data == nil {
		return errors.New("json: field data missing")
	}
	item.Data = required.Data
	return nil
}
