package proxmox

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
//github.com/fatih/structs
)

type (
	LxcDevices     []map[string]interface{}
	LxcDevice      map[string]interface{}
	LxcDeviceParam []string
)

// ConfigLxc - Proxmox API LXC options
type ConfigLxc struct {
	Ostemplate      string      `json:"ostemplate"`
	Storage         string      `json:"storage"`
	Pool            string      `json:"pool"`
	Password        string      `json:"password"`
	Hostname        string      `json:"hostname"`
	Networks        LxcDevices  `json:"networks"`
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

// CreateLxc - Tell Proxmox API to make the LXC container
func (config ConfigLxc) CreateLxc(vmr *VmRef, client *Client) (err error) {
	//vmr.SetVmType("lxc")

        // convert config to map
        params, _ := json.Marshal(&config)
        var paramMap map[string]interface{}
        json.Unmarshal(params, &paramMap)

        // build list network name list
        delete(paramMap, "networks")
	for networkId, network := range config.Networks {
                var networkList []string
	        for key, value := range network {
	                param := fmt.Sprintf("%v=%v", key, value)
                        networkList = append(networkList, param)
	        }
                networkName := fmt.Sprintf("net%v", networkId)
                paramMap[networkName] = strings.Join(networkList, ",")
        }

        // amend vmid
        paramMap["vmid"] = vmr.vmId

	exitStatus, err := client.CreateLxcContainer(vmr.node, paramMap)
	if err != nil {
		return fmt.Errorf("Error creating LXC container: %v, error status: %s (params: %v)", err, exitStatus, params)
	}
	return
}
