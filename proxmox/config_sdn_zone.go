package proxmox

import (
	"encoding/json"
	"fmt"
)

// ConfigSDNZone describes the Zone configurable element
type ConfigSDNZone struct {
	Type                     string `json:"type"`
	Zone                     string `json:"zone"`
	AdvertiseSubnets         bool   `json:"advertise-subnets,omitempty"`
	Bridge                   string `json:"bridge,omitempty"`
	BridgeDisableMacLearning bool   `json:"bridge-disable-mac-learning,omitempty"`
	Controller               string `json:"controller,omitempty"`
	DisableARPNDSuppression  bool   `json:"disable-arp-nd-suppression,omitempty"`
	DNS                      string `json:"dns,omitempty"`
	DNSZone                  string `json:"dnszone,omitempty"`
	DPID                     int    `json:"dp-id,omitempty"`
	ExitNodes                string `json:"exitnodes,omitempty"`
	ExitNodesLocalRouting    bool   `json:"exitnodes-local-routing,omitempty"`
	ExitNodesPrimary         string `json:"exitnodes-primary,omitempty"`
	IPAM                     string `json:"ipam,omitempty"`
	MAC                      string `json:"mac,omitempty"`
	MTU                      int    `json:"mtu,omitempty"`
	Nodes                    string `json:"nodes,omitempty"`
	Peers                    string `json:"peers,omitempty"`
	ReverseDNS               string `json:"reversedns,omitempty"`
	RTImport                 string `json:"rt-import,omitempty"`
	Tag                      int    `json:"tag,omitempty"`
	VlanProtocol             string `json:"vlan-protocol,omitempty"`
	VrfVxlan                 int    `json:"vrf-vxlan,omitempty"`
	// Pass a string of attributes to be deleted from the remote object
	Delete string `json:"delete,omitempty"`
	// Digest allows for a form of optimistic locking
	Digest string `json:"digest,omitempty"`
}

// NewConfigNetworkFromJSON takes in a byte array from a json encoded SDN Zone
// configuration and stores it in config.
// It returns the newly created config with the passed in configuration stored
// and an error if one occurs unmarshalling the input data.
func NewConfigSDNZoneFromJson(input []byte) (config *ConfigSDNZone, err error) {
	config = &ConfigSDNZone{}
	err = json.Unmarshal([]byte(input), config)
	return
}

func (config *ConfigSDNZone) CreateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, true, client)
	if err != nil {
		return
	}
	return config.Create(id, client)
}

func (config *ConfigSDNZone) Create(id string, client *Client) (err error) {
	config.Zone = id
	params := config.mapToApiValues()
	return client.CreateSDNZone(params)
}

func (config *ConfigSDNZone) UpdateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, false, client)
	if err != nil {
		return
	}
	return config.Update(id, client)
}

func (config *ConfigSDNZone) Update(id string, client *Client) (err error) {
	config.Zone = id
	params := config.mapToApiValues()
	err = client.UpdateSDNZone(id, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating SDN Zone: %v, (params: %v)", err, string(params))
	}
	return
}

func (c *ConfigSDNZone) Validate(id string, create bool, client *Client) (err error) {
	exists, err := client.CheckSDNZoneExistance(id)
	if err != nil {
		return
	}
	if exists && create {
		return ErrorItemExists(id, "zone")
	}
	if !exists && !create {
		return ErrorItemNotExists(id, "zone")
	}

	err = ValidateStringInArray([]string{"evpn", "qinq", "simple", "vlan", "vxlan"}, c.Type, "type")
	if err != nil {
		return
	}
	switch c.Type {
	case "simple":
	case "vlan":
		if create {
			if c.Bridge == "" {
				return ErrorKeyEmpty("bridge")
			}
		}
	case "qinq":
		if create {
			if c.Bridge == "" {
				return ErrorKeyEmpty("bridge")
			}
			if c.Tag <= 0 {
				return ErrorKeyEmpty("tag")
			}
			if c.VlanProtocol == "" {
				return ErrorKeyEmpty("vlan-protocol")
			}
		}
	case "vxlan":
		if create {
			if c.Peers == "" {
				return ErrorKeyEmpty("peers")
			}
		}
	case "evpn":
		if create {
			if c.VrfVxlan < 0 {
				return ErrorKeyEmpty("vrf-vxlan")
			}
			if c.Controller == "" {
				return ErrorKeyEmpty("controller")
			}
		}
	}
	if c.VlanProtocol != "" {
		err = ValidateStringInArray([]string{"802.1q", "802.1ad"}, c.VlanProtocol, "vlan-protocol")
		if err != nil {
			return
		}
	}
	return
}

func (config *ConfigSDNZone) mapToApiValues() (params map[string]interface{}) {

	d, _ := json.Marshal(config)
	json.Unmarshal(d, &params)

	boolsToFix := []string{
		"advertise-subnets",
		"bridge-disable-mac-learning",
		"disable-arp-nd-suppression",
		"exitnodes-local-routing",
	}
	for _, key := range boolsToFix {
		if v, has := params[key]; has {
			params[key] = Btoi(v.(bool))
		}
	}
	// Remove the zone and type (path parameters) from the map
	delete(params, "zone")
	delete(params, "type")
	return
}
