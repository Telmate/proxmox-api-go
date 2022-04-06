package proxmox

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Currently ZFS local, LVM, Ceph RBD, CephFS, Directory and virtio-scsi-pci are considered.
// Other formats are not verified, but could be added if they're needed.
const rxStorageTypes = `(zfspool|lvm|rbd|cephfs|dir|virtio-scsi-pci)`
const machineModels = `(pc|q35|pc-i440fx)`

type (
	QemuDevices     map[int]map[string]interface{}
	QemuDevice      map[string]interface{}
	QemuDeviceParam []string
)

// ConfigQemu - Proxmox API QEMU options
type ConfigQemu struct {
	VmID            int         `json:"vmid"`
	Name            string      `json:"name"`
	Description     string      `json:"desc"`
	Pool            string      `json:"pool,omitempty"`
	Bios            string      `json:"bios"`
	EFIDisk         string      `json:"efidisk,omitempty"`
	Machine         string      `json:"machine,omitempty"`
	Onboot          bool        `json:"onboot"`
	Tablet          bool        `json:"tablet"`
	Agent           int         `json:"agent"`
	Memory          int         `json:"memory"`
	Balloon         int         `json:"balloon"`
	QemuOs          string      `json:"os"`
	QemuCores       int         `json:"cores"`
	QemuSockets     int         `json:"sockets"`
	QemuVcpus       int         `json:"vcpus"`
	QemuCpu         string      `json:"cpu"`
	QemuNuma        bool        `json:"numa"`
	QemuKVM         bool        `json:"kvm"`
	Hotplug         string      `json:"hotplug"`
	QemuIso         string      `json:"iso"`
	QemuPxe         bool        `json:"pxe"`
	FullClone       *int        `json:"fullclone"`
	Boot            string      `json:"boot"`
	BootDisk        string      `json:"bootdisk,omitempty"`
	Scsihw          string      `json:"scsihw,omitempty"`
	QemuDisks       QemuDevices `json:"disk"`
	QemuUnusedDisks QemuDevices `json:"unused_disk"`
	QemuVga         QemuDevice  `json:"vga,omitempty"`
	QemuNetworks    QemuDevices `json:"network"`
	QemuSerials     QemuDevices `json:"serial,omitempty"`
	QemuUsbs        QemuDevices `json:"usb,omitempty"`
	HaState         string      `json:"hastate,omitempty"`
	HaGroup         string      `json:"hagroup,omitempty"`
	Tags            string      `json:"tags"`
	Args            string      `json:"args"`

	// Deprecated single disk.
	DiskSize    float64 `json:"diskGB"`
	Storage     string  `json:"storage"`
	StorageType string  `json:"storageType"` // virtio|scsi (cloud-init defaults to scsi)

	// Deprecated single nic.
	QemuNicModel string `json:"nic"`
	QemuBrige    string `json:"bridge"`
	QemuVlanTag  int    `json:"vlan"`
	QemuMacAddr  string `json:"mac"`

	// cloud-init options
	CIuser     string `json:"ciuser"`
	CIpassword string `json:"cipassword"`
	CIcustom   string `json:"cicustom"`

	Searchdomain string `json:"searchdomain"`
	Nameserver   string `json:"nameserver"`
	Sshkeys      string `json:"sshkeys"`

	// arrays are hard, support 16 interfaces for now
	Ipconfig0  string `json:"ipconfig0"`
	Ipconfig1  string `json:"ipconfig1"`
	Ipconfig2  string `json:"ipconfig2"`
	Ipconfig3  string `json:"ipconfig3"`
	Ipconfig4  string `json:"ipconfig4"`
	Ipconfig5  string `json:"ipconfig5"`
	Ipconfig6  string `json:"ipconfig6"`
	Ipconfig7  string `json:"ipconfig7"`
	Ipconfig8  string `json:"ipconfig8"`
	Ipconfig9  string `json:"ipconfig9"`
	Ipconfig10 string `json:"ipconfig10"`
	Ipconfig11 string `json:"ipconfig11"`
	Ipconfig12 string `json:"ipconfig12"`
	Ipconfig13 string `json:"ipconfig13"`
	Ipconfig14 string `json:"ipconfig14"`
	Ipconfig15 string `json:"ipconfig15"`
}

// CreateVm - Tell Proxmox API to make the VM
func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	if config.HasCloudInit() {
		return fmt.Errorf("cloud-init parameters only supported on clones or updates")
	}
	vmr.SetVmType("qemu")

	params := map[string]interface{}{
		"vmid":        vmr.vmId,
		"name":        config.Name,
		"onboot":      config.Onboot,
		"tablet":      config.Tablet,
		"agent":       config.Agent,
		"ostype":      config.QemuOs,
		"sockets":     config.QemuSockets,
		"cores":       config.QemuCores,
		"cpu":         config.QemuCpu,
		"numa":        config.QemuNuma,
		"kvm":         config.QemuKVM,
		"hotplug":     config.Hotplug,
		"memory":      config.Memory,
		"boot":        config.Boot,
		"description": config.Description,
		"tags":        config.Tags,
		"machine":     config.Machine,
		"args":        config.Args,
	}

	if config.QemuIso != "" {
		params["ide2"] = config.QemuIso + ",media=cdrom"
	}

	if config.Bios != "" {
		params["bios"] = config.Bios
	}

	if config.Balloon >= 1 {
		params["balloon"] = config.Balloon
	}

	if config.QemuVcpus >= 1 {
		params["vcpus"] = config.QemuVcpus
	}

	if vmr.pool != "" {
		params["pool"] = vmr.pool
	}

	if config.BootDisk != "" {
		params["bootdisk"] = config.BootDisk
	}

	if config.Scsihw != "" {
		params["scsihw"] = config.Scsihw
	}

	err = config.CreateQemuMachineParam(params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create disks config.
	err = config.CreateQemuDisksParams(vmr.vmId, params, false)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	//Creat additional efi disk
	if config.EFIDisk != "" {
		params["efidisk0"] = fmt.Sprintf("%s:1", config.EFIDisk)
	}

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		params["vga"] = strings.Join(vgaParam, ",")
	}

	// Create networks config.
	err = config.CreateQemuNetworksParams(vmr.vmId, params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create serial interfaces
	err = config.CreateQemuSerialsParams(vmr.vmId, params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create usb interfaces
	err = config.CreateQemuUsbsParams(vmr.vmId, params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	exitStatus, err := client.CreateQemuVm(vmr.node, params)
	if err != nil {
		return fmt.Errorf("error creating VM: %v, error status: %s (params: %v)", err, exitStatus, params)
	}

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	return
}

// HasCloudInit - are there cloud-init options?
func (config ConfigQemu) HasCloudInit() bool {
	return config.CIuser != "" ||
		config.CIpassword != "" ||
		config.Searchdomain != "" ||
		config.Nameserver != "" ||
		config.Sshkeys != "" ||
		config.Ipconfig0 != "" ||
		config.Ipconfig1 != "" ||
		config.Ipconfig2 != "" ||
		config.Ipconfig3 != "" ||
		config.Ipconfig4 != "" ||
		config.Ipconfig5 != "" ||
		config.Ipconfig6 != "" ||
		config.Ipconfig7 != "" ||
		config.Ipconfig8 != "" ||
		config.Ipconfig9 != "" ||
		config.Ipconfig10 != "" ||
		config.Ipconfig11 != "" ||
		config.Ipconfig12 != "" ||
		config.Ipconfig13 != "" ||
		config.Ipconfig14 != "" ||
		config.Ipconfig15 != "" ||
		config.CIcustom != ""
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
	storage := config.Storage
	if disk0Storage, ok := config.QemuDisks[0]["storage"].(string); ok && len(disk0Storage) > 0 {
		storage = disk0Storage
	}
	params := map[string]interface{}{
		"newid":  vmr.vmId,
		"target": vmr.node,
		"name":   config.Name,
		"full":   fullclone,
	}
	if vmr.pool != "" {
		params["pool"] = vmr.pool
	}

	if fullclone == "1" {
		params["storage"] = storage
	}

	_, err = client.CloneQemuVm(sourceVmr, params)
	return err
}

func (config ConfigQemu) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	configParams := map[string]interface{}{
		"name":        config.Name,
		"description": config.Description,
		"tags":        config.Tags,
		"args":        config.Args,
		"onboot":      config.Onboot,
		"tablet":      config.Tablet,
		"agent":       config.Agent,
		"sockets":     config.QemuSockets,
		"cores":       config.QemuCores,
		"cpu":         config.QemuCpu,
		"numa":        config.QemuNuma,
		"kvm":         config.QemuKVM,
		"hotplug":     config.Hotplug,
		"memory":      config.Memory,
		"boot":        config.Boot,
	}

	//Array to list deleted parameters
	deleteParams := []string{}

	if config.Bios != "" {
		configParams["bios"] = config.Bios
	}

	if config.Balloon >= 1 {
		configParams["balloon"] = config.Balloon
	} else {
		deleteParams = append(deleteParams, "balloon")
	}

	if config.QemuVcpus >= 1 {
		configParams["vcpus"] = config.QemuVcpus
	} else {
		deleteParams = append(deleteParams, "vcpus")
	}

	if config.BootDisk != "" {
		configParams["bootdisk"] = config.BootDisk
	}

	if config.Scsihw != "" {
		configParams["scsihw"] = config.Scsihw
	}

	// Create disks config.
	configParamsDisk := map[string]interface{}{
		"vmid": vmr.vmId,
	}
	// TODO keep going if error=
	err = config.CreateQemuDisksParams(vmr.vmId, configParamsDisk, false)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}
	// TODO keep going if error=
	_, err = client.createVMDisks(vmr.node, configParamsDisk)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}
	//Copy the disks to the global configParams
	for key, value := range configParamsDisk {
		//vmid is only required in createVMDisks
		if key != "vmid" {
			configParams[key] = value
		}
	}

	// Create networks config.
	err = config.CreateQemuNetworksParams(vmr.vmId, configParams)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		configParams["vga"] = strings.Join(vgaParam, ",")
	} else {
		deleteParams = append(deleteParams, "vga")
	}

	// Create serial interfaces
	err = config.CreateQemuSerialsParams(vmr.vmId, configParams)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create usb interfaces
	err = config.CreateQemuUsbsParams(vmr.vmId, configParams)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// cloud-init options
	if config.CIuser != "" {
		configParams["ciuser"] = config.CIuser
	}
	if config.CIpassword != "" {
		configParams["cipassword"] = config.CIpassword
	}
	if config.CIcustom != "" {
		configParams["cicustom"] = config.CIcustom
	}
	if config.Searchdomain != "" {
		configParams["searchdomain"] = config.Searchdomain
	}
	if config.Nameserver != "" {
		configParams["nameserver"] = config.Nameserver
	}
	if config.Sshkeys != "" {
		sshkeyEnc := url.PathEscape(config.Sshkeys + "\n")
		sshkeyEnc = strings.Replace(sshkeyEnc, "+", "%2B", -1)
		sshkeyEnc = strings.Replace(sshkeyEnc, "@", "%40", -1)
		sshkeyEnc = strings.Replace(sshkeyEnc, "=", "%3D", -1)
		configParams["sshkeys"] = sshkeyEnc
	}
	if config.Ipconfig0 != "" {
		configParams["ipconfig0"] = config.Ipconfig0
	}
	if config.Ipconfig1 != "" {
		configParams["ipconfig1"] = config.Ipconfig1
	}
	if config.Ipconfig2 != "" {
		configParams["ipconfig2"] = config.Ipconfig2
	}
	if config.Ipconfig3 != "" {
		configParams["ipconfig3"] = config.Ipconfig3
	}
	if config.Ipconfig4 != "" {
		configParams["ipconfig4"] = config.Ipconfig4
	}
	if config.Ipconfig5 != "" {
		configParams["ipconfig5"] = config.Ipconfig5
	}
	if config.Ipconfig6 != "" {
		configParams["ipconfig6"] = config.Ipconfig6
	}
	if config.Ipconfig7 != "" {
		configParams["ipconfig7"] = config.Ipconfig7
	}
	if config.Ipconfig8 != "" {
		configParams["ipconfig8"] = config.Ipconfig8
	}
	if config.Ipconfig9 != "" {
		configParams["ipconfig9"] = config.Ipconfig9
	}
	if config.Ipconfig10 != "" {
		configParams["ipconfig10"] = config.Ipconfig10
	}
	if config.Ipconfig11 != "" {
		configParams["ipconfig11"] = config.Ipconfig11
	}
	if config.Ipconfig12 != "" {
		configParams["ipconfig12"] = config.Ipconfig12
	}
	if config.Ipconfig13 != "" {
		configParams["ipconfig13"] = config.Ipconfig13
	}
	if config.Ipconfig14 != "" {
		configParams["ipconfig14"] = config.Ipconfig14
	}
	if config.Ipconfig15 != "" {
		configParams["ipconfig15"] = config.Ipconfig15
	}

	if len(deleteParams) > 0 {
		configParams["delete"] = strings.Join(deleteParams, ", ")
	}

	_, err = client.SetVmConfig(vmr, configParams)
	if err != nil {
		log.Print(err)
		return err
	}

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	_, err = client.UpdateVMPool(vmr, config.Pool)

	return err
}

func NewConfigQemuFromJson(io io.Reader) (config *ConfigQemu, err error) {
	config = &ConfigQemu{QemuVlanTag: -1, QemuKVM: true}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return
}

var (
	rxIso            = regexp.MustCompile(`(.*?),media`)
	rxDeviceID       = regexp.MustCompile(`\d+`)
	rxDiskName       = regexp.MustCompile(`(virtio|scsi)\d+`)
	rxDiskType       = regexp.MustCompile(`\D+`)
	rxUnusedDiskName = regexp.MustCompile(`^(unused)\d+`)
	rxNicName        = regexp.MustCompile(`net\d+`)
	rxMpName         = regexp.MustCompile(`mp\d+`)
	rxSerialName     = regexp.MustCompile(`serial\d+`)
	rxUsbName        = regexp.MustCompile(`usb\d+`)
)

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
		return nil, fmt.Errorf("vm locked, could not obtain config")
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
	if _, isSet := vmConfig["name"]; isSet {
		name = vmConfig["name"].(string)
	}
	description := ""
	if _, isSet := vmConfig["description"]; isSet {
		description = vmConfig["description"].(string)
	}
	tags := ""
	if _, isSet := vmConfig["tags"]; isSet {
		tags = vmConfig["tags"].(string)
	}
	args := ""
	if _, isSet := vmConfig["args"]; isSet {
		args = vmConfig["args"].(string)
	}

	bios := "seabios"
	if _, isSet := vmConfig["bios"]; isSet {
		bios = vmConfig["bios"].(string)
	}
	efidisk := ""
	if _, isSet := vmConfig["efidisk0"]; isSet {
		efidisk = vmConfig["efidisk0"].(string)
	}
	onboot := true
	if _, isSet := vmConfig["onboot"]; isSet {
		onboot = Itob(int(vmConfig["onboot"].(float64)))
	}
	tablet := true
	if _, isSet := vmConfig["tablet"]; isSet {
		tablet = Itob(int(vmConfig["tablet"].(float64)))
	}

	agent := 0
	if _, isSet := vmConfig["agent"]; isSet {
		switch vmConfig["agent"].(type) {
		case float64:
			agent = int(vmConfig["agent"].(float64))
		case string:
			agent, _ = strconv.Atoi(vmConfig["agent"].(string))
		}

	}
	ostype := "other"
	if _, isSet := vmConfig["ostype"]; isSet {
		ostype = vmConfig["ostype"].(string)
	}
	memory := 0.0
	if _, isSet := vmConfig["memory"]; isSet {
		memory = vmConfig["memory"].(float64)
	}
	balloon := 0.0
	if _, isSet := vmConfig["balloon"]; isSet {
		balloon = vmConfig["balloon"].(float64)
	}
	cores := 1.0
	if _, isSet := vmConfig["cores"]; isSet {
		cores = vmConfig["cores"].(float64)
	}
	vcpus := 0.0
	if _, isSet := vmConfig["vcpus"]; isSet {
		vcpus = vmConfig["vcpus"].(float64)
	}
	sockets := 1.0
	if _, isSet := vmConfig["sockets"]; isSet {
		sockets = vmConfig["sockets"].(float64)
	}
	cpu := "host"
	if _, isSet := vmConfig["cpu"]; isSet {
		cpu = vmConfig["cpu"].(string)
	}
	numa := false
	if _, isSet := vmConfig["numa"]; isSet {
		numa = Itob(int(vmConfig["numa"].(float64)))
	}
	//Can be network,disk,cpu,memory,usb
	hotplug := "network,disk,usb"
	if _, isSet := vmConfig["hotplug"]; isSet {
		hotplug = vmConfig["hotplug"].(string)
	}
	//boot by default from hard disk (c), CD-ROM (d), network (n).
	boot := "cdn"
	if _, isSet := vmConfig["boot"]; isSet {
		boot = vmConfig["boot"].(string)
	}
	bootdisk := ""
	if _, isSet := vmConfig["bootdisk"]; isSet {
		bootdisk = vmConfig["bootdisk"].(string)
	}
	kvm := true
	if _, isSet := vmConfig["kvm"]; isSet {
		kvm = Itob(int(vmConfig["kvm"].(float64)))
	}
	scsihw := "lsi"
	if _, isSet := vmConfig["scsihw"]; isSet {
		scsihw = vmConfig["scsihw"].(string)
	}

	config = &ConfigQemu{
		Name:            name,
		Description:     strings.TrimSpace(description),
		Tags:            strings.TrimSpace(tags),
		Args:            strings.TrimSpace(args),
		Bios:            bios,
		EFIDisk:         efidisk,
		Onboot:          onboot,
		Tablet:          tablet,
		Agent:           agent,
		QemuOs:          ostype,
		Memory:          int(memory),
		QemuCores:       int(cores),
		QemuSockets:     int(sockets),
		QemuCpu:         cpu,
		QemuNuma:        numa,
		QemuKVM:         kvm,
		Hotplug:         hotplug,
		QemuVlanTag:     -1,
		Boot:            boot,
		BootDisk:        bootdisk,
		Scsihw:          scsihw,
		QemuDisks:       QemuDevices{},
		QemuUnusedDisks: QemuDevices{},
		QemuVga:         QemuDevice{},
		QemuNetworks:    QemuDevices{},
		QemuSerials:     QemuDevices{},
		QemuUsbs:        QemuDevices{},
	}

	if balloon >= 1 {
		config.Balloon = int(balloon)
	}
	if vcpus >= 1 {
		config.QemuVcpus = int(vcpus)
	}

	if vmConfig["ide2"] != nil {
		isoMatch := rxIso.FindStringSubmatch(vmConfig["ide2"].(string))
		config.QemuIso = isoMatch[1]
	}

	if _, isSet := vmConfig["ciuser"]; isSet {
		config.CIuser = vmConfig["ciuser"].(string)
	}
	if _, isSet := vmConfig["cipassword"]; isSet {
		config.CIpassword = vmConfig["cipassword"].(string)
	}
	if _, isSet := vmConfig["cicustom"]; isSet {
		config.CIcustom = vmConfig["cicustom"].(string)
	}
	if _, isSet := vmConfig["searchdomain"]; isSet {
		config.Searchdomain = vmConfig["searchdomain"].(string)
	}
	if _, isSet := vmConfig["nameserver"]; isSet {
		config.Nameserver = vmConfig["nameserver"].(string)
	}
	if _, isSet := vmConfig["sshkeys"]; isSet {
		config.Sshkeys, _ = url.PathUnescape(vmConfig["sshkeys"].(string))
	}
	if _, isSet := vmConfig["ipconfig0"]; isSet {
		config.Ipconfig0 = vmConfig["ipconfig0"].(string)
	}
	if _, isSet := vmConfig["ipconfig1"]; isSet {
		config.Ipconfig1 = vmConfig["ipconfig1"].(string)
	}
	if _, isSet := vmConfig["ipconfig2"]; isSet {
		config.Ipconfig2 = vmConfig["ipconfig2"].(string)
	}
	if _, isSet := vmConfig["ipconfig3"]; isSet {
		config.Ipconfig3 = vmConfig["ipconfig3"].(string)
	}
	if _, isSet := vmConfig["ipconfig4"]; isSet {
		config.Ipconfig4 = vmConfig["ipconfig4"].(string)
	}
	if _, isSet := vmConfig["ipconfig5"]; isSet {
		config.Ipconfig5 = vmConfig["ipconfig5"].(string)
	}
	if _, isSet := vmConfig["ipconfig6"]; isSet {
		config.Ipconfig6 = vmConfig["ipconfig6"].(string)
	}
	if _, isSet := vmConfig["ipconfig7"]; isSet {
		config.Ipconfig7 = vmConfig["ipconfig7"].(string)
	}
	if _, isSet := vmConfig["ipconfig8"]; isSet {
		config.Ipconfig8 = vmConfig["ipconfig8"].(string)
	}
	if _, isSet := vmConfig["ipconfig9"]; isSet {
		config.Ipconfig9 = vmConfig["ipconfig9"].(string)
	}
	if _, isSet := vmConfig["ipconfig10"]; isSet {
		config.Ipconfig10 = vmConfig["ipconfig10"].(string)
	}
	if _, isSet := vmConfig["ipconfig11"]; isSet {
		config.Ipconfig11 = vmConfig["ipconfig11"].(string)
	}
	if _, isSet := vmConfig["ipconfig12"]; isSet {
		config.Ipconfig12 = vmConfig["ipconfig12"].(string)
	}
	if _, isSet := vmConfig["ipconfig13"]; isSet {
		config.Ipconfig13 = vmConfig["ipconfig13"].(string)
	}
	if _, isSet := vmConfig["ipconfig14"]; isSet {
		config.Ipconfig14 = vmConfig["ipconfig14"].(string)
	}
	if _, isSet := vmConfig["ipconfig15"]; isSet {
		config.Ipconfig15 = vmConfig["ipconfig15"].(string)
	}

	// Add disks.
	diskNames := []string{}

	for k := range vmConfig {
		if diskName := rxDiskName.FindStringSubmatch(k); len(diskName) > 0 {
			diskNames = append(diskNames, diskName[0])
		}
	}

	for _, diskName := range diskNames {
		diskConfStr := vmConfig[diskName].(string)

		id := rxDeviceID.FindStringSubmatch(diskName)
		diskID, _ := strconv.Atoi(id[0])
		diskType := rxDiskType.FindStringSubmatch(diskName)[0]

		diskConfMap := ParsePMConf(diskConfStr, "volume")
		diskConfMap["slot"] = diskID
		diskConfMap["type"] = diskType

		storageName, fileName := ParseSubConf(diskConfMap["volume"].(string), ":")
		diskConfMap["storage"] = storageName
		diskConfMap["file"] = fileName

		filePath := diskConfMap["volume"]

		// Get disk format
		storageContent, err := client.GetStorageContent(vmr, storageName)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		var storageFormat string
		contents := storageContent["data"].([]interface{})
		for content := range contents {
			storageContentMap := contents[content].(map[string]interface{})
			if storageContentMap["volid"] == filePath {
				storageFormat = storageContentMap["format"].(string)
				break
			}
		}
		diskConfMap["format"] = storageFormat

		// Get storage type for disk
		var storageStatus map[string]interface{}
		storageStatus, err = client.GetStorageStatus(vmr, storageName)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		storageType := storageStatus["type"]

		diskConfMap["storage_type"] = storageType

		// cloud-init disks not always have the size sent by the API, which results in a crash
		if diskConfMap["size"] == nil && strings.Contains(fileName.(string), "cloudinit") {
			diskConfMap["size"] = "4M" // default cloud-init disk size
		}

		var sizeInTerabytes = regexp.MustCompile(`[0-9]+T`)
		// Convert to gigabytes if disk size was received in terabytes
		matched := sizeInTerabytes.MatchString(diskConfMap["size"].(string))
		if matched {
			diskConfMap["size"] = fmt.Sprintf("%.0fG", DiskSizeGB(diskConfMap["size"]))
		}

		// And device config to disks map.
		if len(diskConfMap) > 0 {
			config.QemuDisks[diskID] = diskConfMap
		}
	}

	// Add unused disks
	// unused0:local:100/vm-100-disk-1.qcow2
	unusedDiskNames := []string{}
	for k := range vmConfig {
		// look for entries from the config in the format "unusedX:<storagepath>" where X is an integer
		if unusedDiskName := rxUnusedDiskName.FindStringSubmatch(k); len(unusedDiskName) > 0 {
			unusedDiskNames = append(unusedDiskNames, unusedDiskName[0])
		}
	}
	if len(unusedDiskNames) > 0 {
		log.Printf("[DEBUG] unusedDiskNames: %v", unusedDiskNames)
	}

	for _, unusedDiskName := range unusedDiskNames {
		unusedDiskConfStr := vmConfig[unusedDiskName].(string)
		finalDiskConfMap := QemuDevice{}

		// parse "unused0" to get the id '0' as an int
		id := rxDeviceID.FindStringSubmatch(unusedDiskName)
		diskID, err := strconv.Atoi(id[0])
		if err != nil {
			return nil, fmt.Errorf(fmt.Sprintf("Unable to parse unused disk id from input string '%v' tried to convert '%v' to integer.", unusedDiskName, diskID))
		}
		finalDiskConfMap["slot"] = diskID

		// parse the attributes from the unused disk
		// extract the storage and file path from the unused disk entry
		parsedUnusedDiskMap := ParsePMConf(unusedDiskConfStr, "storage+file")
		storageName, fileName := ParseSubConf(parsedUnusedDiskMap["storage+file"].(string), ":")
		finalDiskConfMap["storage"] = storageName
		finalDiskConfMap["file"] = fileName

		config.QemuUnusedDisks[diskID] = finalDiskConfMap
	}

	//Display
	if vga, isSet := vmConfig["vga"]; isSet {
		vgaList := strings.Split(vga.(string), ",")
		vgaMap := QemuDevice{}

		// TODO: keep going if error?
		err = vgaMap.readDeviceConfig(vgaList)
		if err != nil {
			log.Printf("[ERROR] %q", err)
		}
		if len(vgaMap) > 0 {
			config.QemuVga = vgaMap
		}
	}

	// Add networks.
	nicNames := []string{}

	for k := range vmConfig {
		if nicName := rxNicName.FindStringSubmatch(k); len(nicName) > 0 {
			nicNames = append(nicNames, nicName[0])
		}
	}

	for _, nicName := range nicNames {
		nicConfStr := vmConfig[nicName]
		nicConfList := strings.Split(nicConfStr.(string), ",")

		id := rxDeviceID.FindStringSubmatch(nicName)
		nicID, _ := strconv.Atoi(id[0])
		model, macaddr := ParseSubConf(nicConfList[0], "=")

		// Add model and MAC address.
		nicConfMap := QemuDevice{
			"id":      nicID,
			"model":   model,
			"macaddr": macaddr,
		}

		// Add rest of device config.
		err = nicConfMap.readDeviceConfig(nicConfList[1:])
		if err != nil {
			log.Printf("[ERROR] %q", err)
		}
		if nicConfMap["firewall"] == 1 {
			nicConfMap["firewall"] = true
		} else if nicConfMap["firewall"] == 0 {
			nicConfMap["firewall"] = false
		}
		if nicConfMap["link_down"] == 1 {
			nicConfMap["link_down"] = true
		} else if nicConfMap["link_down"] == 0 {
			nicConfMap["link_down"] = false
		}

		// And device config to networks.
		if len(nicConfMap) > 0 {
			config.QemuNetworks[nicID] = nicConfMap
		}
	}

	// Add serials
	serialNames := []string{}

	for k := range vmConfig {
		if serialName := rxSerialName.FindStringSubmatch(k); len(serialName) > 0 {
			serialNames = append(serialNames, serialName[0])
		}
	}

	for _, serialName := range serialNames {
		id := rxDeviceID.FindStringSubmatch(serialName)
		serialID, _ := strconv.Atoi(id[0])

		serialConfMap := QemuDevice{
			"id":   serialID,
			"type": vmConfig[serialName],
		}

		// And device config to serials map.
		if len(serialConfMap) > 0 {
			config.QemuSerials[serialID] = serialConfMap
		}
	}

	// Add usbs
	usbNames := []string{}

	for k := range vmConfig {
		if usbName := rxUsbName.FindStringSubmatch(k); len(usbName) > 0 {
			usbNames = append(usbNames, usbName[0])
		}
	}

	for _, usbName := range usbNames {
		usbConfStr := vmConfig[usbName]
		usbConfList := strings.Split(usbConfStr.(string), ",")
		id := rxDeviceID.FindStringSubmatch(usbName)
		usbID, _ := strconv.Atoi(id[0])
		_, host := ParseSubConf(usbConfList[0], "=")

		usbConfMap := QemuDevice{
			"id":   usbID,
			"host": host,
		}

		err = usbConfMap.readDeviceConfig(usbConfList[1:])
		if err != nil {
			log.Printf("[ERROR] %q", err)
		}
		if usbConfMap["usb3"] == 1 {
			usbConfMap["usb3"] = true
		}

		// And device config to usbs map.
		if len(usbConfMap) > 0 {
			config.QemuUsbs[usbID] = usbConfMap
		}
	}

	// hastate is return by the api for a vm resource type but not the hagroup
	err = client.ReadVMHA(vmr)
	if err == nil {
		config.HaState = vmr.HaState()
		config.HaGroup = vmr.HaGroup()
	} else {
		log.Printf("[DEBUG] VM %d(%s) has no HA config", vmr.vmId, vmConfig["hostname"])
		return config, nil
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
	return fmt.Errorf("not shutdown within wait time")
}

// This is because proxmox create/config API won't let us make usernet devices
func SshForwardUsernet(vmr *VmRef, client *Client) (sshPort string, err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return "", err
	}
	if vmState["status"] == "stopped" {
		return "", fmt.Errorf("VM must be running first")
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
		return fmt.Errorf("VM must be running first")
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
	max = 100
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
		return fmt.Errorf("VM must be running first")
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
		time.Sleep(1 * time.Millisecond)
	}
	return nil
}

// Given a QemuDevice, return a param string to give to ProxMox
func formatDeviceParam(device QemuDevice) string {
	deviceConfParams := QemuDeviceParam{}
	deviceConfParams = deviceConfParams.createDeviceParam(device, nil)
	return strings.Join(deviceConfParams, ",")
}

// Given a QemuDevice (represesting a disk), return a param string to give to ProxMox
func FormatDiskParam(disk QemuDevice) string {
	diskConfParam := QemuDeviceParam{}

	if volume, ok := disk["volume"]; ok && volume != "" {
		diskConfParam = append(diskConfParam, volume.(string))
		diskConfParam = append(diskConfParam, fmt.Sprintf("size=%v", disk["size"]))
	} else {
		volumeInit := fmt.Sprintf("%v:%v", disk["storage"], DiskSizeGB(disk["size"]))
		diskConfParam = append(diskConfParam, volumeInit)
	}

	// Set cache if not none (default).
	if cache, ok := disk["cache"]; ok && cache != "none" {
		diskCache := fmt.Sprintf("cache=%v", disk["cache"])
		diskConfParam = append(diskConfParam, diskCache)
	}

	// Mountoptions
	if mountoptions, ok := disk["mountoptions"]; ok {
		options := []string{}
		for opt, enabled := range mountoptions.(map[string]interface{}) {
			if enabled.(bool) {
				options = append(options, opt)
			}
		}
		diskMountOpts := fmt.Sprintf("mountoptions=%v", strings.Join(options, ";"))
		diskConfParam = append(diskConfParam, diskMountOpts)
	}

	// Keys that are not used as real/direct conf.
	ignoredKeys := []string{"key", "slot", "type", "storage", "file", "size", "cache", "volume", "container", "vm", "mountoptions", "storage_type"}

	// Rest of config.
	diskConfParam = diskConfParam.createDeviceParam(disk, ignoredKeys)

	return strings.Join(diskConfParam, ",")
}

// Given a QemuDevice (represesting a usb), return a param string to give to ProxMox
func FormatUsbParam(usb QemuDevice) string {
	usbConfParam := QemuDeviceParam{}

	usbConfParam = usbConfParam.createDeviceParam(usb, []string{})

	return strings.Join(usbConfParam, ",")
}

// Create parameters for each Nic device.
func (c ConfigQemu) CreateQemuNetworksParams(vmID int, params map[string]interface{}) error {

	// For backward compatibility.
	if len(c.QemuNetworks) == 0 && len(c.QemuNicModel) > 0 {
		deprecatedStyleMap := QemuDevice{
			"id":      0,
			"model":   c.QemuNicModel,
			"bridge":  c.QemuBrige,
			"macaddr": c.QemuMacAddr,
		}

		if c.QemuVlanTag > 0 {
			deprecatedStyleMap["tag"] = strconv.Itoa(c.QemuVlanTag)
		}

		c.QemuNetworks[0] = deprecatedStyleMap
	}

	// For new style with multi net device.
	for nicID, nicConfMap := range c.QemuNetworks {

		nicConfParam := QemuDeviceParam{}

		// Set Nic name.
		qemuNicName := "net" + strconv.Itoa(nicID)

		// Set Mac address.
		var macAddr string
		switch nicConfMap["macaddr"] {
		case nil, "":
			// Generate random Mac based on time
			macaddr := make(net.HardwareAddr, 6)
			rand.Seed(time.Now().UnixNano())
			rand.Read(macaddr)
			macaddr[0] = (macaddr[0] | 2) & 0xfe // fix from github issue #18
			macAddr = strings.ToUpper(fmt.Sprintf("%v", macaddr))

			// Add Mac to source map so it will be returned. (useful for some use case like Terraform)
			nicConfMap["macaddr"] = macAddr
		case "repeatable":
			// Generate deterministic Mac based on VmID and NicID
			// Assume that rare VM has more than 32 nics
			macaddr := make(net.HardwareAddr, 6)
			pairing := vmID<<5 | nicID
			// Linux MAC vendor - 00:18:59
			macaddr[0] = 0x00
			macaddr[1] = 0x18
			macaddr[2] = 0x59
			macaddr[3] = byte((pairing >> 16) & 0xff)
			macaddr[4] = byte((pairing >> 8) & 0xff)
			macaddr[5] = byte(pairing & 0xff)
			// Convert to string
			macAddr = strings.ToUpper(fmt.Sprintf("%v", macaddr))

			// Add Mac to source map so it will be returned. (useful for some use case like Terraform)
			nicConfMap["macaddr"] = macAddr
		default:
			macAddr = nicConfMap["macaddr"].(string)
		}

		// use model=mac format for older proxmox compatability as the parameters which will be sent to Proxmox API.
		nicConfParam = append(nicConfParam, fmt.Sprintf("%v=%v", nicConfMap["model"], macAddr))

		// Set bridge if not nat.
		if nicConfMap["bridge"].(string) != "nat" {
			bridge := fmt.Sprintf("bridge=%v", nicConfMap["bridge"])
			nicConfParam = append(nicConfParam, bridge)
		}

		// Keys that are not used as real/direct conf.
		ignoredKeys := []string{"id", "bridge", "macaddr", "model"}

		// Rest of config.
		nicConfParam = nicConfParam.createDeviceParam(nicConfMap, ignoredKeys)

		// Add nic to Qemu prams.
		params[qemuNicName] = strings.Join(nicConfParam, ",")
	}

	return nil
}

// Create parameters for each disk.
func (c ConfigQemu) CreateQemuDisksParams(
	vmID int,
	params map[string]interface{},
	cloned bool,
) error {

	// For backward compatibility.
	if len(c.QemuDisks) == 0 && len(c.Storage) > 0 {

		dType := c.StorageType
		if dType == "" {
			if c.HasCloudInit() {
				dType = "scsi"
			} else {
				dType = "virtio"
			}
		}
		deprecatedStyleMap := QemuDevice{
			"id":           0,
			"type":         dType,
			"storage":      c.Storage,
			"size":         c.DiskSize,
			"storage_type": "lvm",  // default old style
			"cache":        "none", // default old value
		}
		if c.QemuDisks == nil {
			c.QemuDisks = make(QemuDevices)
		}
		c.QemuDisks[0] = deprecatedStyleMap
	}

	// For new style with multi disk device.
	for diskID, diskConfMap := range c.QemuDisks {
		// skip the first disk for clones (may not always be right, but a template probably has at least 1 disk)
		if diskID == 0 && cloned {
			continue
		}

		// Device name.
		deviceType := diskConfMap["type"].(string)
		qemuDiskName := deviceType + strconv.Itoa(diskID)

		// Add back to Qemu prams.
		params[qemuDiskName] = FormatDiskParam(diskConfMap)
	}

	return nil
}

// Create parameters for serial interface
func (c ConfigQemu) CreateQemuSerialsParams(
	vmID int,
	params map[string]interface{},
) error {

	// For new style with multi disk device.
	for serialID, serialConfMap := range c.QemuSerials {
		// Device name.
		deviceType := serialConfMap["type"].(string)
		qemuSerialName := "serial" + strconv.Itoa(serialID)

		// Add back to Qemu prams.
		params[qemuSerialName] = deviceType
	}

	return nil
}

// Create parameters for usb interface
func (c ConfigQemu) CreateQemuUsbsParams(
	vmID int,
	params map[string]interface{},
) error {
	for usbID, usbConfMap := range c.QemuUsbs {
		qemuUsbName := "usb" + strconv.Itoa(usbID)

		// Add back to Qemu prams.
		params[qemuUsbName] = FormatUsbParam(usbConfMap)
	}

	return nil
}

// Create parameters for serial interface
func (c ConfigQemu) CreateQemuMachineParam(
	params map[string]interface{},
) error {
	if c.Machine == "" {
		return nil
	}
	if matched, _ := regexp.MatchString(machineModels, c.Machine); matched {
		params["machine"] = c.Machine
		return nil
	}
	return fmt.Errorf("unsupported machine type, fall back to default")
}

func (p QemuDeviceParam) createDeviceParam(
	deviceConfMap QemuDevice,
	ignoredKeys []string,
) QemuDeviceParam {

	for key, value := range deviceConfMap {
		if ignored := inArray(ignoredKeys, key); !ignored {
			var confValue interface{}
			if bValue, ok := value.(bool); ok && bValue {
				confValue = "1"
			} else if sValue, ok := value.(string); ok && len(sValue) > 0 {
				confValue = sValue
			} else if iValue, ok := value.(int); ok && iValue > 0 {
				confValue = iValue
			}
			if confValue != nil {
				deviceConf := fmt.Sprintf("%v=%v", key, confValue)
				p = append(p, deviceConf)
			}
		}
	}

	return p
}

// readDeviceConfig - get standard sub-conf strings where `key=value` and update conf map.
func (confMap QemuDevice) readDeviceConfig(confList []string) error {
	// Add device config.
	for _, conf := range confList {
		key, value := ParseSubConf(conf, "=")
		confMap[key] = value
	}
	return nil
}

func (c ConfigQemu) String() string {
	jsConf, _ := json.Marshal(c)
	return string(jsConf)
}

// VMIdExists - If you pass an VMID that exists it will raise an error otherwise it will return the vmID
func (c *Client) VMIdExists(vmID int) (id int, err error) {
	_, err = c.session.Get(fmt.Sprintf("/cluster/nextid?vmid=%d", vmID), nil, nil)
	if err != nil {
		return -1, err
	}
	return vmID, nil
}
