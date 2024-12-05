package proxmox

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
)

type ConfigSDNSubnet struct {
	// For creation purposes - Subnet is a CIDR
	// Once a subnet has been created, the Subnet is an identifier with the format
	// "<zone>-<ip>-<mask>"
	Subnet string `json:"subnet"`

	DNSZonePrefix string `json:"dnszoneprefix,omitempty"`
	Gateway       string `json:"gateway,omitempty"`
	SNAT          bool   `json:"snat,omitempty"`

	// Delete is a string of attributes to be deleted from the object
	Delete string `json:"delete,omitempty"`
	// Type must always hold the string "subnet"
	Type string `json:"type"`
	// Digest allows for a form of optimistic locking
	Digest string `json:"digest,omitempty"`
}

// NewConfigSDNSubnetFromJSON takes in a byte array from a json encoded SDN Subnet
// configuration and stores it in config.
// It returns the newly created config with the passed in configuration stored
// and an error if one occurs unmarshalling the input data.
func NewConfigSDNSubnetFromJson(input []byte) (config *ConfigSDNSubnet, err error) {
	config = &ConfigSDNSubnet{}
	err = json.Unmarshal([]byte(input), config)
	return
}

func (config *ConfigSDNSubnet) CreateWithValidate(ctx context.Context, vnet, id string, client *Client) (err error) {
	err = config.Validate(ctx, vnet, id, true, client)
	if err != nil {
		return
	}
	return config.Create(ctx, vnet, id, client)
}

func (config *ConfigSDNSubnet) Create(ctx context.Context, vnet, id string, client *Client) (err error) {
	config.Subnet = id
	config.Type = "subnet"
	params := config.mapToApiValues()
	return client.CreateSDNSubnet(ctx, vnet, params)
}

func (config *ConfigSDNSubnet) UpdateWithValidate(ctx context.Context, vnet, id string, client *Client) (err error) {
	err = config.Validate(ctx, vnet, id, false, client)
	if err != nil {
		return
	}
	return config.Update(ctx, vnet, id, client)
}

func (config *ConfigSDNSubnet) Update(ctx context.Context, vnet, id string, client *Client) (err error) {
	config.Subnet = id
	config.Type = "" // For some reason, this shouldn't be sent on update. Only on create.
	params := config.mapToApiValues()
	err = client.UpdateSDNSubnet(ctx, vnet, id, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating SDN Subnet: %v, (params: %v)", err, string(params))
	}
	return
}

func (c *ConfigSDNSubnet) Validate(ctx context.Context, vnet, id string, create bool, client *Client) (err error) {
	vnetExists, err := client.CheckSDNVNetExistance(ctx, vnet)
	if err != nil {
		return
	}
	if !vnetExists {
		return fmt.Errorf("subnet must be created in an existing vnet. vnet (%s) wasn't found", vnet)
	}
	exists, err := client.CheckSDNSubnetExistance(ctx, vnet, id)
	if err != nil {
		return
	}
	if exists && create {
		return ErrorItemExists(id, "subnet")
	}
	if !exists && !create {
		return ErrorItemNotExists(id, "subnet")
	}

	// if this is an update, the Subnet is an identifier of the form <zone>-<ip>-<mask>
	// and therefore shouldn't be validated or changed
	if create {
		// Make sure that the CIDR is actually a valid CIDR
		_, _, err = net.ParseCIDR(c.Subnet)
		if err != nil {
			return
		}
	}

	if c.Gateway != "" {
		ip := net.ParseIP(c.Gateway)
		if ip == nil {
			return fmt.Errorf("error gateway (%s) is not a valid IP", c.Gateway)
		}
	}

	return
}

func (config *ConfigSDNSubnet) mapToApiValues() (params map[string]interface{}) {

	d, _ := json.Marshal(config)
	json.Unmarshal(d, &params)

	if v, has := params["snat"]; has {
		params["snat"] = Btoi(v.(bool))
	}
	// Remove the subnet and vnet (path parameters) from the map
	delete(params, "subnet")
	delete(params, "vnet")
	return
}
