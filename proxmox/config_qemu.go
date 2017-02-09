package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"
	"time"
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
	vmr.SetVmType("qemu")
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
		"description": config.Description,
	}

	_, err = client.CreateQemuVm(vmr.node, params)
	return
}

/*

CloneVm
Example: Request

nodes/proxmox1-xx/qemu/1012/clone

newid:145
name:tf-clone1
target:proxmox1-xx
full:1
storage:xxx

*/
func (config ConfigQemu) CloneVm(sourceVmr *VmRef, vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("qemu")
	params := map[string]string{
		"newid":   strconv.Itoa(vmr.vmId),
		"target":  vmr.node,
		"name":    config.Name,
		"storage": config.Storage,
		"full":    "1",
	}
	_, err = client.CloneQemuVm(sourceVmr, params)
	if err != nil {
		return
	}
	configParams := map[string]string{
		"sockets":     strconv.Itoa(config.QemuSockets),
		"cores":       strconv.Itoa(config.QemuCores),
		"memory":      strconv.Itoa(config.Memory),
		"description": config.Description,
	}
	_, err = client.SetVmConfig(vmr, configParams)
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

func WaitForShutdown(vmr *VmRef, client *Client) (err error) {
	for ii := 0; ii < 100; ii++ {
		vmState, err := client.GetVmState(vmr)
		if err != nil {
			log.Print("Wait error:")
			log.Println(err)
		} else if vmState["status"] == "stopped" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("Not shutdown within wait time")
}

func SshForwardUsernet(vmr *VmRef, client *Client) (sshPort string, err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return "", err
	}
	if vmState["status"] == "stopped" {
		return "", errors.New("VM must be running first")
	}
	sshPort = strconv.Itoa(vmr.VmId() + 22000)
	_, err = client.MonitorCmd(vmr, "netdev_add user,id=net1,hostfwd=tcp::"+sshPort+"-:22")
	if err != nil {
		return "", err
	}
	_, err = client.MonitorCmd(vmr, "device_add virtio-net-pci,id=net1,netdev=net1,addr=0x13")
	if err != nil {
		return "", err
	}
	return
}

func MaxVmId(client *Client) (max int, err error) {
	resp, err := client.GetVmList()
	vms := resp["data"].([]interface{})
	max = 0
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		vmid := int(vm["vmid"].(float64))
		if vmid > max {
			max = vmid
		}
	}
	return
}
