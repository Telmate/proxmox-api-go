package proxmox

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
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
	QemuMacAddr  string  `json:"mac"`
	FullClone    *int    `json:"fullclone"`
}

func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("qemu")
	network := config.QemuNicModel + ",bridge=" + config.QemuBrige
	if config.QemuMacAddr != "" {
		network = network + ",macaddr=" + config.QemuMacAddr
	}
	if config.QemuVlanTag > 0 {
		network = network + ",tag=" + strconv.Itoa(config.QemuVlanTag)
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
	fullclone := "1"
	if config.FullClone != nil {
		fullclone = strconv.Itoa(*config.FullClone)
	}
	params := map[string]string{
		"newid":   strconv.Itoa(vmr.vmId),
		"target":  vmr.node,
		"name":    config.Name,
		"storage": config.Storage,
		"full":    fullclone,
	}
	_, err = client.CloneQemuVm(sourceVmr, params)
	if err != nil {
		return
	}
	return config.UpdateConfig(vmr, client)
}

func (config ConfigQemu) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	network := config.QemuNicModel + ",bridge=" + config.QemuBrige
	if config.QemuMacAddr != "" {
		network = network + ",macaddr=" + config.QemuMacAddr
	}
	if config.QemuVlanTag > 0 {
		network = network + ",tag=" + strconv.Itoa(config.QemuVlanTag)
	}
	configParams := map[string]string{
		"sockets":     strconv.Itoa(config.QemuSockets),
		"cores":       strconv.Itoa(config.QemuCores),
		"memory":      strconv.Itoa(config.Memory),
		"net0":        network,
		"description": config.Description,
	}
	_, err = client.SetVmConfig(vmr, configParams)
	return err
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
var rxNetwork = regexp.MustCompile("(.*?)=(.*?),bridge=([^,]+)(?:,tag=)?(.*)")

func NewConfigQemuFromApi(vmr *VmRef, client *Client) (config *ConfigQemu, err error) {
	var vmConfig map[string]interface{}
	for ii := 0; ii < 3; ii++ {
		vmConfig, err = client.GetVmConfig(vmr)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		// this can happen:
		// {"data":{"lock":"clone","digest":"eb54fb9d9f120ba0c3bdf694f73b10002c375c38","description":" qmclone temporary file\n"}})
		if vmConfig["lock"] == nil {
			break
		} else {
			time.Sleep(8 * time.Second)
		}
	}

	if vmConfig["lock"] != nil {
		return nil, errors.New("vm locked, could not obtain config")
	}

	// vmConfig Sample: map[ cpu:host
	// net0:virtio=62:DF:XX:XX:XX:XX,bridge=vmbr0
	// ide2:local:iso/xxx-xx.iso,media=cdrom memory:2048
	// smbios1:uuid=8b3bf833-aad8-4545-xxx-xxxxxxx digest:aa6ce5xxxxx1b9ce33e4aaeff564d4 sockets:1
	// name:terraform-ubuntu1404-template bootdisk:virtio0
	// virtio0:ProxmoxxxxISCSI:vm-1014-disk-2,size=4G
	// description:Base image
	// cores:2 ostype:l26

	name := ""
	if _, is_set := vmConfig["name"]; is_set {
		name = vmConfig["name"].(string)
	}
	description := ""
	if _, is_set := vmConfig["description"]; is_set {
		description = vmConfig["description"].(string)
	}
	ostype := ""
	if _, is_set := vmConfig["ostype"]; is_set {
		ostype = vmConfig["ostype"].(string)
	}
	memory := 0.0
	if _, is_set := vmConfig["memory"]; is_set {
		memory = vmConfig["memory"].(float64)
	}
	cores := 0.0
	if _, is_set := vmConfig["cores"]; is_set {
		cores = vmConfig["cores"].(float64)
	}
	sockets := 0.0
	if _, is_set := vmConfig["sockets"]; is_set {
		sockets = vmConfig["sockets"].(float64)
	}
	config = &ConfigQemu{
		Name:        name,
		Description: strings.TrimSpace(description),
		QemuOs:      ostype,
		Memory:      int(memory),
		QemuCores:   int(cores),
		QemuSockets: int(sockets),
		QemuVlanTag: -1,
	}

	if vmConfig["virtio0"] == nil {
		return nil, errors.New("virtio0 (required) not found in current config")
	}

	storageMatch := rxStorage.FindStringSubmatch(vmConfig["virtio0"].(string))
	config.Storage = storageMatch[1]
	config.DiskSize, _ = strconv.ParseFloat(storageMatch[2], 64)

	if vmConfig["ide2"] != nil {
		isoMatch := rxIso.FindStringSubmatch(vmConfig["ide2"].(string))
		config.QemuIso = isoMatch[1]
	}

	if vmConfig["net0"] == nil {
		return nil, errors.New("net0 (required) not found in current config")
	}

	netMatch := rxNetwork.FindStringSubmatch(vmConfig["net0"].(string))
	config.QemuNicModel = netMatch[1]
	config.QemuMacAddr = netMatch[2]
	config.QemuBrige = netMatch[3]
	if netMatch[4] != "" {
		config.QemuVlanTag, _ = strconv.Atoi(netMatch[4])
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

// device_del net1
// netdev_del net1
func RemoveSshForwardUsernet(vmr *VmRef, client *Client) (err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return err
	}
	if vmState["status"] == "stopped" {
		return errors.New("VM must be running first")
	}
	_, err = client.MonitorCmd(vmr, "device_del net1")
	if err != nil {
		return err
	}
	_, err = client.MonitorCmd(vmr, "netdev_del net1")
	if err != nil {
		return err
	}
	return nil
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

func SendKeysString(vmr *VmRef, client *Client, keys string) (err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return err
	}
	if vmState["status"] == "stopped" {
		return errors.New("VM must be running first")
	}
	for _, r := range keys {
		c := string(r)
		lower := strings.ToLower(c)
		if c != lower {
			c = "shift-" + lower
		} else {
			switch c {
			case "!":
				c = "shift-1"
			case "@":
				c = "shift-2"
			case "#":
				c = "shift-3"
			case "$":
				c = "shift-4"
			case "%%":
				c = "shift-5"
			case "^":
				c = "shift-6"
			case "&":
				c = "shift-7"
			case "*":
				c = "shift-8"
			case "(":
				c = "shift-9"
			case ")":
				c = "shift-0"
			case "_":
				c = "shift-minus"
			case "+":
				c = "shift-equal"
			case " ":
				c = "spc"
			case "/":
				c = "slash"
			case "\\":
				c = "backslash"
			case ",":
				c = "comma"
			case "-":
				c = "minus"
			case "=":
				c = "equal"
			case ".":
				c = "dot"
			case "?":
				c = "shift-slash"
			}
		}
		_, err = client.MonitorCmd(vmr, "sendkey "+c)
		if err != nil {
			return err
		}
		time.Sleep(100)
	}
	return nil
}
