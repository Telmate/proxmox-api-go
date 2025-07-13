package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
)

type ConfigSDNVNet struct {
	VNet      string `json:"vnet"`
	Zone      string `json:"zone"`
	Alias     string `json:"alias,omitempty"`
	Delete    string `json:"delete,omitempty"`
	Tag       int    `json:"tag,omitempty"`
	VLANAware bool   `json:"vlanaware,omitempty"`
	// Digest allows for a form of optimistic locking
	Digest string `json:"digest,omitempty"`
}

func NewConfigSDNVNetFromJson(input []byte) (config *ConfigSDNVNet, err error) {
	config = &ConfigSDNVNet{}
	err = json.Unmarshal([]byte(input), config)
	return
}

func (config *ConfigSDNVNet) CreateWithValidate(ctx context.Context, id string, client *Client) (err error) {
	err = config.Validate(ctx, id, true, client)
	if err != nil {
		return
	}
	return config.Create(ctx, id, client)
}

func (config *ConfigSDNVNet) Create(ctx context.Context, id string, client *Client) (err error) {
	config.VNet = id
	params := config.mapToApiValues()
	return client.CreateSDNVNet(ctx, params)
}

func (config *ConfigSDNVNet) UpdateWithValidate(ctx context.Context, id string, client *Client) (err error) {
	err = config.Validate(ctx, id, false, client)
	if err != nil {
		return
	}
	return config.Update(ctx, id, client)
}

func (config *ConfigSDNVNet) Update(ctx context.Context, id string, client *Client) (err error) {
	config.VNet = id
	params := config.mapToApiValues()
	err = client.UpdateSDNVNet(ctx, id, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating SDN VNet: %v, (params: %v)", err, string(params))
	}
	return
}

func (c *ConfigSDNVNet) Validate(ctx context.Context, id string, create bool, client *Client) (err error) {
	exists, err := client.CheckSDNVNetExistance(ctx, id)
	if err != nil {
		return
	}
	if exists && create {
		return ErrorItemExists(id, "vnet")
	}
	if !exists && !create {
		return ErrorItemNotExists(id, "vnet")
	}
	zoneExists, err := client.CheckSDNZoneExistance(ctx, c.Zone)
	if err != nil {
		return
	}
	if !zoneExists {
		return fmt.Errorf("vnet must be associated to an existing zone. zone %s could not be found", c.Zone)
	}
	if c.Alias != "" {
		regex, _ := regexp.Compile(`^(?i:[\(\)-_.\w\d\s]{0,256})$`)
		if !regex.Match([]byte(c.Alias)) {
			return fmt.Errorf(`alias must match the validation regular expression: ^(?i:[\(\)-_.\w\d\s]{0,256})$`)
		}
	}
	err = ValidateIntGreater(0, c.Tag, "tag")
	if err != nil {
		return
	}

	return
}

func (config *ConfigSDNVNet) mapToApiValues() (params map[string]interface{}) {
	d, _ := json.Marshal(config)
	json.Unmarshal(d, &params)

	if v, has := params["vlanaware"]; has {
		params["vlanaware"] = Btoi(v.(bool))
	}

	return
}
