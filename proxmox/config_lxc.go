package proxmox

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// ConfigLxc - Proxmox API LXC options
type ConfigLxc struct {
	Ostemplate      string       `json:"ostemplate"`
	Storage         string       `json:"storage"`
	Pool            string       `json:"pool"`
	Password        string       `json:"password"`
	Hostname        string       `json:"hostname"`
	Networks        QemuDevices  `json:"networks"`
}

func NewConfigLxcFromJson(io io.Reader) (config *ConfigLxc, err error) {
	config = &ConfigLxc{}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Println(config)
	return
}

// create LXC container using the Proxmox API
func (config ConfigLxc) CreateLxc(vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("lxc")

        // convert config to map
        params, _ := json.Marshal(&config)
        var paramMap map[string]interface{}
        json.Unmarshal(params, &paramMap)

        // build list of network parameters
	for nicID, nicConfMap := range config.Networks {
		nicConfParam := QemuDeviceParam{}
		nicConfParam = nicConfParam.createDeviceParam(nicConfMap, nil)

		// add nic to lxc parameters
		nicName := fmt.Sprintf("net%v", nicID)
		paramMap[nicName] = strings.Join(nicConfParam, ",")
        }

        // now that we concatenated the key value parameter
        // list for the networks, remove the original network key
        // since the Proxmox API does not know how to handle this key
        delete(paramMap, "networks")

        // amend vmid
        paramMap["vmid"] = vmr.vmId

	exitStatus, err := client.CreateLxcContainer(vmr.node, paramMap)
	if err != nil {
		return fmt.Errorf("Error creating LXC container: %v, error status: %s (params: %v)", err, exitStatus, params)
	}
	return
}
