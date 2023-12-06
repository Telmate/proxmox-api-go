package proxmox

import (
	"encoding/json"
	"fmt"
)

// ConfigSDNDNS describes the SDN DNS configurable element
type ConfigSDNDNS struct {
	DNS  string `json:"dns"`
	Key  string `json:"key"`
	Type string `json:"type"`
	URL  string `json:"url"`
	TTL  int    `json:"ttl,omitempty"`
	// The SDN Plugin schema contains ReverseV6Mask attribute while the
	// PowerDNS plugin schema contains the ReverseMaskV6 attribute
	// This is probably a bug that crept into the Proxmox implementation.a
	// Checked in libpve-network-perl=0.7.3
	ReverseMaskV6 int `json:"reversemaskv6,omitempty"`
	ReverseV6Mask int `json:"reversev6mask,omitempty"`
	// Digest allows for a form of optimistic locking
	Digest string `json:"digest,omitempty"`
}

func NewConfigSDNDNSFromJson(input []byte) (config *ConfigSDNDNS, err error) {
	config = &ConfigSDNDNS{}
	err = json.Unmarshal([]byte(input), config)
	return
}

func (config *ConfigSDNDNS) CreateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, true, client)
	if err != nil {
		return
	}
	return config.Create(id, client)
}

func (config *ConfigSDNDNS) Create(id string, client *Client) (err error) {
	config.DNS = id
	params := config.mapToApiValues()
	return client.CreateSDNDNS(params)
}

func (config *ConfigSDNDNS) UpdateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, false, client)
	if err != nil {
		return
	}
	return config.Update(id, client)
}

func (config *ConfigSDNDNS) Update(id string, client *Client) (err error) {
	config.DNS = id
	params := config.mapToApiValues()
	err = client.UpdateSDNDNS(id, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating SDN DNS: %v, (params: %v)", err, string(params))
	}
	return
}

func (c *ConfigSDNDNS) Validate(id string, create bool, client *Client) (err error) {
	exists, err := client.CheckSDNDNSExistance(id)
	if err != nil {
		return
	}
	if exists && create {
		return ErrorItemExists(id, "dns")
	}
	if !exists && !create {
		return ErrorItemNotExists(id, "dns")
	}

	err = ValidateStringInArray([]string{"powerdns"}, c.Type, "type")
	if err != nil {
		return
	}
	err = ValidateIntGreater(0, c.TTL, "ttl")
	if err != nil {
		return
	}
	return
}

func (config *ConfigSDNDNS) mapToApiValues() (params map[string]interface{}) {
	d, _ := json.Marshal(config)
	json.Unmarshal(d, &params)
	return
}
