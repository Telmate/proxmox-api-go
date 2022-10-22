package proxmox

import (
	"encoding/json"
	"fmt"
)

type ConfigSnapshot struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	VmState     bool   `json:"ram,omitempty"`
}

func (config *ConfigSnapshot) mapValues() map[string]interface{} {
	return map[string]interface{}{
		"snapname":    config.Name,
		"description": config.Description,
		"vmstate":     config.VmState,
	}
}

func (config *ConfigSnapshot) CreateSnapshot(guestId uint, c *Client) (err error) {
	params := config.mapValues()
	_, err = c.CreateSnapshot(NewVmRef(int(guestId)), params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Snapshot: %v, (params: %v)", err, string(params))
	}
	return
}
