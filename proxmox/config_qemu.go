package proxmox

import (
	"encoding/json"
	"io"
	"log"
	"strconv"
)

type ConfigQemu struct {
	Name         string  `json:"name"`
	Description  string  `json:"desc"`
	Memory       int     `json:"memory"`
	DiskSize     float64 `json:"diskGB"`
	Storage      string  `json:"storage"`
	QemuOs       string  `json:"os"`
	QemuCores    int     `json:"cores"`
	QemuSockets  int     `json:"sockets"`
	QemuIso      string  `json:"iso"`
	QemuNicModel string  `json:"nic"`
	QemuBrige    string  `json:"bridge"`
	QemuVlanTag  int     `json:"vlan"`
}

func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	network := config.QemuNicModel + ",bridge=" + config.QemuBrige
	if config.QemuVlanTag > 0 {
		network = network + ",tag=" + string(config.QemuVlanTag)
	}
	params := map[string]string{
		"vmid":        strconv.Itoa(vmr.vmId),
		"name":        config.Name,
		"ide2":        config.QemuIso + ",media=cdrom",
		"ostype":      config.QemuOs,
		"virtio0":     config.Storage + ":" + strconv.FormatFloat(config.DiskSize, 'f', -1, 64),
		"sockets":     strconv.Itoa(config.QemuSockets),
		"cores":       strconv.Itoa(config.QemuCores),
		"cpu":         "host",
		"memory":      strconv.Itoa(config.Memory),
		"net0":        network,
		"description": config.Description}

	_, err = client.CreateQemuVm(vmr.node, params)
	return
}

func NewConfigQemuFromJson(io io.Reader) (config *ConfigQemu, err error) {
	config = &ConfigQemu{QemuVlanTag: -1}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Println(config)
	return
}
