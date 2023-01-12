package proxmox

import (
	"encoding/json"
	"fmt"
)

// ConfigNetwork maps go variables to API parameters.
type ConfigNetwork struct {
	Iface              string `json:"iface,omitempty"`
	Node               string `json:"node,omitempty"`
	Type               string `json:"type,omitempty"`
	Address            string `json:"address,omitempty"`
	Address6           string `json:"address6,omitempty"`
	Autostart          bool   `json:"autostart,omitempty"`
	BondPrimary        string `json:"bond-primary,omitempty"`
	BondMode           string `json:"bond_mode,omitempty"`
	BondXmitHashPolicy string `json:"bond_xmit_hash_policy,omitempty"`
	BridgePorts        string `json:"bridge_ports,omitempty"`
	BridgeVlanAware    bool   `json:"bridge_vlan_aware,omitempty"`
	CIDR               string `json:"cidr,omitempty"`
	CIDR6              string `json:"cidr6,omitempty"`
	Comments           string `json:"comments,omitempty"`
	Comments6          string `json:"comments6,omitempty"`
	Gateway            string `json:"gateway,omitempty"`
	Gateway6           string `json:"gateway6,omitempty"`
	MTU                int    `json:"mtu,omitempty"`
	Netmask            string `json:"netmask,omitempty"`
	Netmask6           int    `json:"netmask6,omitempty"`
	OVSBonds           string `json:"ovs_bonds,omitempty"`
	OVSBridge          string `json:"ovs_bridge,omitempty"`
	OVSOptions         string `json:"ovs_options,omitempty"`
	OVSPorts           string `json:"ovs_ports,omitempty"`
	OVSTag             int    `json:"ovs_tag,omitempty"`
	Slaves             string `json:"slaves,omitempty"`
	VlanID             int    `json:"vlan-id,omitempty"`
	VlanRawDevice      string `json:"vlan-raw-device,omitempty"`
}

// NewConfigNetworkFromJSON takes in a byte array from a json encoded network
// configuration and stores it in config.
// It returns the newly created config with the passed in configuration stored
// and an error if one occurs unmarshalling the input data.
func NewConfigNetworkFromJSON(input []byte) (config *ConfigNetwork, err error) {
	config = &ConfigNetwork{}
	err = json.Unmarshal([]byte(input), config)
	return
}

// mapToApiValues converts the stored config into a parameter map to be
// sent to the API.
func (config ConfigNetwork) mapToApiValues() map[string]interface{} {
	params, _ := json.Marshal(&config)
	var paramMap map[string]interface{}
	json.Unmarshal(params, &paramMap)
	return paramMap
}

// CreateNetwork creates a network on the Proxmox host with the stored
// config.
// It returns an error if the creation of the network fails.
func (config ConfigNetwork) CreateNetwork(client *Client) (err error) {
	paramMap := config.mapToApiValues()

	exitStatus, err := client.CreateNetwork(config.Node, paramMap)
	if err != nil {
		params, _ := json.Marshal(&paramMap)
		return fmt.Errorf("error creating network: %v\n\t\t api response: %s\n\t\t params: %v", err, exitStatus, string(params))
	}
	return
}

// UpdateNetwork updates a network on the Proxmox host with the stored
// config.
// It returns an error if updating the network fails.
func (config ConfigNetwork) UpdateNetwork(client *Client) (err error) {
	paramMap := config.mapToApiValues()

	exitStatus, err := client.UpdateNetwork(config.Node, config.Iface, paramMap)
	if err != nil {
		params, _ := json.Marshal(paramMap)
		return fmt.Errorf("error creating network: %v\n\t\t api response: %s\n\t\t params: %v", err, exitStatus, string(params))
	}
	return
}
