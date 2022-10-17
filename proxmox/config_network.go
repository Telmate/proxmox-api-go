package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
)

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

func NewConfigNetworkFromJson(input []byte) (config *ConfigNetwork, err error) {
	config = &ConfigNetwork{}
	err = json.Unmarshal([]byte(input), config)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (config ConfigNetwork) MapToAPIParams() map[string]interface{} {
	params, _ := json.Marshal(&config)
	var paramMap map[string]interface{}
	json.Unmarshal(params, &paramMap)
	return paramMap
}

func (config ConfigNetwork) CreateNetwork(client *Client) (err error) {
	paramMap := config.MapToAPIParams()

	exitStatus, err := client.CreateNetwork(config.Node, paramMap)
	if err != nil {
		params, _ := json.Marshal(&paramMap)
		return fmt.Errorf("error creating network: %v\n\t\t API Error: %s\n\t\t Params: %v", err, exitStatus, string(params))
	}
	return
}
