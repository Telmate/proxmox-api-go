package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"regexp"
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

var rxStorage = regexp.MustCompile("(.*?):.*?,size=(\\d+)G")
var rxIso = regexp.MustCompile("(.*?),media")
var rxNetwork = regexp.MustCompile("(.*?)=.*?,bridge=([^,]+)(?:,tag=)?(.*)")

func NewConfigQemuFromApi(vmr *VmRef, client *Client) (config *ConfigQemu, err error) {
	vmConfig, err := client.GetVmConfig(vmr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// vmConfig Sample: map[ cpu:host
	// net0:virtio=62:DF:XX:XX:XX:XX,bridge=vmbr0
	// ide2:local:iso/xxx-xx.iso,media=cdrom memory:2048
	// smbios1:uuid=8b3bf833-aad8-4545-xxx-xxxxxxx digest:aa6ce5xxxxx1b9ce33e4aaeff564d4 sockets:1
	// name:terraform-ubuntu1404-template bootdisk:virtio0
	// virtio0:ProxmoxxxxISCSI:vm-1014-disk-2,size=4G
	// description:Base image
	// cores:2 ostype:l26 ]
	config = &ConfigQemu{
		Name:        vmConfig["name"].(string),
		Description: vmConfig["description"].(string),
		QemuOs:      vmConfig["ostype"].(string),
		Memory:      int(vmConfig["memory"].(float64)),
		QemuCores:   int(vmConfig["cores"].(float64)),
		QemuSockets: int(vmConfig["sockets"].(float64)),
		QemuVlanTag: -1,
	}

	storageMatch := rxStorage.FindStringSubmatch(vmConfig["virtio0"].(string))
	config.Storage = storageMatch[1]
	config.DiskSize, _ = strconv.ParseFloat(storageMatch[2], 64)

	isoMatch := rxIso.FindStringSubmatch(vmConfig["ide2"].(string))
	config.QemuIso = isoMatch[1]

	netMatch := rxNetwork.FindStringSubmatch(vmConfig["net0"].(string))
	config.QemuNicModel = netMatch[1]
	config.QemuBrige = netMatch[2]
	if netMatch[3] != "" {
		config.QemuVlanTag, _ = strconv.Atoi(netMatch[3])
	}

	return
}

// Useful waiting for ISO install to complete
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

// This is because proxmox create/config API won't let us make usernet devices
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
