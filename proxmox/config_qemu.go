package proxmox

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
// const rxStorageTypes = `(zfspool|lvm|rbd|cephfs|dir|virtio-scsi-pci)`
const machineModels = `(pc|q35|pc-i440fx)`

type (
	QemuDevices     map[int]map[string]interface{}
	QemuDevice      map[string]interface{}
	QemuDeviceParam []string
	IpconfigMap     map[int]interface{}
)

type AgentNetworkInterface struct {
	MACAddress  string
	IPAddresses []net.IP
	Name        string
	Statistics  map[string]int64
}

// ConfigQemu - Proxmox API QEMU options
type ConfigQemu struct {
	Agent           int           `json:"agent,omitempty"` // TODO should probably be a bool
	Args            string        `json:"args,omitempty"`
	Balloon         int           `json:"balloon,omitempty"` // TODO should probably be a bool
	Bios            string        `json:"bios,omitempty"`
	Boot            string        `json:"boot,omitempty"`       // TODO should be an array of custom enums
	BootDisk        string        `json:"bootdisk,omitempty"`   // TODO discuss deprecation? Only returned as it's deprecated in the proxmox api
	CIcustom        string        `json:"cicustom,omitempty"`   // TODO should be part of a cloud-init struct (cloud-init option)
	CIpassword      string        `json:"cipassword,omitempty"` // TODO should be part of a cloud-init struct (cloud-init option)
	CIuser          string        `json:"ciuser,omitempty"`     // TODO should be part of a cloud-init struct (cloud-init option)
	Description     string        `json:"description,omitempty"`
	Disks           *QemuStorages `json:"disks,omitempty"`
	EFIDisk         QemuDevice    `json:"efidisk,omitempty"`   // TODO should be a struct
	RNGDrive        QemuDevice    `json:"rng0,omitempty"`      // TODO should be a struct
	FullClone       *int          `json:"fullclone,omitempty"` // TODO should probably be a bool
	HaGroup         string        `json:"hagroup,omitempty"`
	HaState         string        `json:"hastate,omitempty"` // TODO should be custom type with enum
	Hookscript      string        `json:"hookscript,omitempty"`
	Hotplug         string        `json:"hotplug,omitempty"`    // TODO should be a struct
	Ipconfig        IpconfigMap   `json:"ipconfig,omitempty"`   // TODO should be part of a cloud-init struct (cloud-init option)
	Iso             *IsoFile      `json:"iso,omitempty"`        // Same as Disks.Ide.Disk_2.CdRom.Iso
	LinkedVmId      uint          `json:"linked_id,omitempty"`  // Only returned setting it has no effect
	Machine         string        `json:"machine,omitempty"`    // TODO should be custom type with enum
	Memory          int           `json:"memory,omitempty"`     // TODO should be uint
	Name            string        `json:"name,omitempty"`       // TODO should be custom type as there are character and length limitations
	Nameserver      string        `json:"nameserver,omitempty"` // TODO should be part of a cloud-init struct (cloud-init option)
	Node            string        `json:"node,omitempty"`
	Onboot          *bool         `json:"onboot,omitempty"`
	Pool            string        `json:"pool,omitempty"`    // TODO should be custom type as there are character and length limitations
	QemuCores       int           `json:"cores,omitempty"`   // TODO should be uint
	QemuCpu         string        `json:"cpu,omitempty"`     // TODO should be custom type with enum
	QemuDisks       QemuDevices   `json:"disk,omitempty"`    // DEPRECATED use Disks *QemuStorages instead
	QemuIso         string        `json:"qemuiso,omitempty"` // DEPRECATED use Iso *IsoFile instead
	QemuKVM         *bool         `json:"kvm,omitempty"`
	QemuNetworks    QemuDevices   `json:"network,omitempty"` // TODO should be a struct
	QemuNuma        *bool         `json:"numa,omitempty"`
	QemuOs          string        `json:"ostype,omitempty"`
	QemuPCIDevices  QemuDevices   `json:"hostpci,omitempty"` // TODO should be a struct
	QemuPxe         bool          `json:"pxe,omitempty"`
	QemuSerials     QemuDevices   `json:"serial,omitempty"`       // TODO should be a struct
	QemuSockets     int           `json:"sockets,omitempty"`      // TODO should be uint
	QemuUnusedDisks QemuDevices   `json:"unused,omitempty"`       // TODO should be a struct
	QemuUsbs        QemuDevices   `json:"usb,omitempty"`          // TODO should be a struct
	QemuVcpus       int           `json:"vcpus,omitempty"`        // TODO should be uint
	QemuVga         QemuDevice    `json:"vga,omitempty"`          // TODO should be a struct
	Scsihw          string        `json:"scsihw,omitempty"`       // TODO should be custom type with enum
	Searchdomain    string        `json:"searchdomain,omitempty"` // TODO should be part of a cloud-init struct (cloud-init option)
	Smbios1         string        `json:"smbios1,omitempty"`      // TODO should be custom type with enum?
	Sshkeys         string        `json:"sshkeys,omitempty"`      // TODO should be an array of strings
	Startup         string        `json:"startup,omitempty"`      // TODO should be a struct?
	Tablet          *bool         `json:"tablet,omitempty"`
	Tags            string        `json:"tags,omitempty"` // TODO should be an array of a custom type as there are character and length limitations
	VmID            int           `json:"vmid,omitempty"` // TODO should be a custom type as there are limitations
}

// Create - Tell Proxmox API to make the VM
func (config ConfigQemu) Create(vmr *VmRef, client *Client) (err error) {
	_, err = config.setAdvanced(nil, false, vmr, client)
	return
}

// DEPRECATED use ConfigQemu{}.Create Instead.
//
// CreateVm - Tell Proxmox API to make the VM
func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	if config.HasCloudInit() {
		return fmt.Errorf("cloud-init parameters only supported on clones or updates")
	}
	vmr.SetVmType("qemu")

	params := map[string]interface{}{
		"vmid":        vmr.vmId,
		"name":        config.Name,
		"startup":     config.Startup,
		"agent":       config.Agent,
		"ostype":      config.QemuOs,
		"sockets":     config.QemuSockets,
		"cores":       config.QemuCores,
		"cpu":         config.QemuCpu,
		"hotplug":     config.Hotplug,
		"memory":      config.Memory,
		"boot":        config.Boot,
		"description": config.Description,
		"tags":        config.Tags,
		"machine":     config.Machine,
		"args":        config.Args,
	}

	if config.QemuNuma != nil {
		params["numa"] = *config.QemuNuma
	}

	if config.QemuKVM != nil {
		params["kvm"] = *config.QemuKVM
	}

	if config.Tablet != nil {
		params["tablet"] = *config.Tablet
	}

	if config.Onboot != nil {
		params["onboot"] = *config.Onboot
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
	if config.Boot != "" {
		params["boot"] = config.Boot
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
	config.CreateQemuDisksParams(params, false)

	// Create EFI disk
	config.CreateQemuEfiParams(params)

	// Create VirtIO RNG
	config.CreateQemuRngParams(params)

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		params["vga"] = strings.Join(vgaParam, ",")
	}

	// Create networks config.
	config.CreateQemuNetworksParams(params)

	// Create ipconfig.
	err = config.CreateIpconfigParams(params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create serial interfaces
	config.CreateQemuSerialsParams(params)

	config.CreateQemuPCIsParams(params)

	// Create usb interfaces
	config.CreateQemuUsbsParams(params)

	exitStatus, err := client.CreateQemuVm(vmr.node, params)
	if err != nil {
		return fmt.Errorf("error creating VM: %v, error status: %s (params: %v)", err, exitStatus, params)
	}

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		return fmt.Errorf("[ERROR] %q", err)
	}

	return
}

func (config *ConfigQemu) defaults() {
	if config == nil {
		return
	}
	if config.Boot == "" {
		config.Boot = "cdn"
	}
	if config.Bios == "" {
		config.Bios = "seabios"
	}
	if config.RNGDrive == nil {
		config.RNGDrive = QemuDevice{}
	}
	if config.EFIDisk == nil {
		config.EFIDisk = QemuDevice{}
	}
	if config.Onboot == nil {
		config.Onboot = PointerBool(true)
	}
	if config.Hotplug == "" {
		config.Hotplug = "network,disk,usb"
	}
	if config.Ipconfig == nil {
		config.Ipconfig = IpconfigMap{}
	}
	if config.QemuCores == 0 {
		config.QemuCores = 1
	}
	if config.QemuCpu == "" {
		config.QemuCpu = "host"
	}
	if config.QemuDisks == nil {
		config.QemuDisks = QemuDevices{}
	}
	if config.QemuKVM == nil {
		config.QemuKVM = PointerBool(true)
	}
	if config.QemuNetworks == nil {
		config.QemuNetworks = QemuDevices{}
	}
	if config.QemuOs == "" {
		config.QemuOs = "other"
	}
	if config.QemuPCIDevices == nil {
		config.QemuPCIDevices = QemuDevices{}
	}
	if config.QemuSerials == nil {
		config.QemuSerials = QemuDevices{}
	}
	if config.QemuSockets == 0 {
		config.QemuSockets = 1
	}
	if config.QemuUnusedDisks == nil {
		config.QemuUnusedDisks = QemuDevices{}
	}
	if config.QemuUsbs == nil {
		config.QemuUsbs = QemuDevices{}
	}
	if config.QemuVga == nil {
		config.QemuVga = QemuDevice{}
	}
	if config.Scsihw == "" {
		config.Scsihw = "lsi"
	}
	if config.Tablet == nil {
		config.Tablet = PointerBool(true)
	}

}

func (config ConfigQemu) mapToApiValues(currentConfig ConfigQemu) (rebootRequired bool, params map[string]interface{}, err error) {
	// TODO check if cloudInit settings changed, they require a reboot to take effect.
	var itemsToDelete string

	params = map[string]interface{}{}

	if config.VmID != 0 {
		params["vmid"] = config.VmID
	}
	if config.Args != "" {
		params["args"] = config.Args
	}
	if config.Agent != 0 {
		params["agent"] = config.Agent
	}
	if config.Balloon >= 1 {
		params["balloon"] = config.Balloon
	}
	if config.Bios != "" {
		params["bios"] = config.Bios
	}
	if config.Boot != "" {
		params["boot"] = config.Boot
	}
	if config.CIcustom != "" {
		params["cicustom"] = config.CIcustom
	}
	if config.CIpassword != "" {
		params["cipassword"] = config.CIpassword
	}
	if config.CIuser != "" {
		params["ciuser"] = config.CIuser
	}
	if config.QemuCores != 0 {
		params["cores"] = config.QemuCores
	}
	if config.QemuCpu != "" {
		params["cpu"] = config.QemuCpu
	}
	if config.Description != "" {
		params["description"] = config.Description
	}
	if config.Hookscript != "" {
		params["hookscript"] = config.Hookscript
	}
	if config.Hotplug != "" {
		params["hotplug"] = config.Hotplug
	}
	if config.QemuKVM != nil {
		params["kvm"] = *config.QemuKVM
	}
	if config.Machine != "" {
		params["machine"] = config.Machine
	}
	if config.Memory != 0 {
		params["memory"] = config.Memory
	}
	if config.Name != "" {
		params["name"] = config.Name
	}
	if config.Nameserver != "" {
		params["nameserver"] = config.Nameserver
	}
	if config.QemuNuma != nil {
		params["numa"] = *config.QemuNuma
	}
	if config.Onboot != nil {
		params["onboot"] = *config.Onboot
	}
	if config.QemuOs != "" {
		params["ostype"] = config.QemuOs
	}
	if config.Pool != "" {
		params["pool"] = config.Pool
	}
	if config.Scsihw != "" {
		params["scsihw"] = config.Scsihw
	}
	if config.Searchdomain != "" {
		params["searchdomain"] = config.Searchdomain
	}
	if config.QemuSockets != 0 {
		params["sockets"] = config.QemuSockets
	}
	if config.Sshkeys != "" {
		params["sshkeys"] = sshKeyUrlEncode(config.Sshkeys)
	}
	if config.Startup != "" {
		params["startup"] = config.Startup
	}
	if config.Tablet != nil {
		params["tablet"] = *config.Tablet
	}
	if config.Tags != "" {
		params["tags"] = config.Tags
	}
	if config.QemuVcpus >= 1 {
		params["vcpus"] = config.QemuVcpus
	}
	if config.Smbios1 != "" {
		params["smbios1"] = config.Smbios1
	}

	if config.Iso != nil {
		if config.Disks == nil {
			config.Disks = &QemuStorages{}
		}
		if config.Disks.Ide == nil {
			config.Disks.Ide = &QemuIdeDisks{}
		}
		if config.Disks.Ide.Disk_2 == nil {
			config.Disks.Ide.Disk_2 = &QemuIdeStorage{}
		}
		if config.Disks.Ide.Disk_2.CdRom == nil {
			config.Disks.Ide.Disk_2.CdRom = &QemuCdRom{Iso: config.Iso}
		}
	}
	// Disks
	if currentConfig.Disks != nil {
		if config.Disks != nil {
			// Create,Update,Delete
			delete := config.Disks.mapToApiValues(*currentConfig.Disks, uint(config.VmID), currentConfig.LinkedVmId, params)
			if delete != "" {
				itemsToDelete = AddToList(itemsToDelete, delete)
			}
		}
	} else {
		if config.Disks != nil {
			// Create
			config.Disks.mapToApiValues(QemuStorages{}, uint(config.VmID), 0, params)
		}
	}

	// Create EFI disk
	config.CreateQemuEfiParams(params)

	// Create VirtIO RNG
	config.CreateQemuRngParams(params)

	// Create networks config.
	config.CreateQemuNetworksParams(params)

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		params["vga"] = strings.Join(vgaParam, ",")
	}
	// Create serial interfaces
	config.CreateQemuSerialsParams(params)

	// Create usb interfaces
	config.CreateQemuUsbsParams(params)

	config.CreateQemuPCIsParams(params)

	err = config.CreateIpconfigParams(params)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	if itemsToDelete != "" {
		params["delete"] = itemsToDelete
	}
	return
}

func (ConfigQemu) mapToStruct(params map[string]interface{}) (*ConfigQemu, error) {
	// vmConfig Sample: map[ cpu:host
	// net0:virtio=62:DF:XX:XX:XX:XX,bridge=vmbr0
	// ide2:local:iso/xxx-xx.iso,media=cdrom memory:2048
	// smbios1:uuid=8b3bf833-aad8-4545-xxx-xxxxxxx digest:aa6ce5xxxxx1b9ce33e4aaeff564d4 sockets:1
	// name:terraform-ubuntu1404-template bootdisk:virtio0
	// virtio0:ProxmoxxxxISCSI:vm-1014-disk-2,size=4G
	// description:Base image
	// cores:2 ostype:l26

	config := ConfigQemu{}

	if _, isSet := params["agent"]; isSet {
		switch params["agent"].(type) {
		case float64:
			config.Agent = int(params["agent"].(float64))
		case string:
			AgentConfList := strings.Split(params["agent"].(string), ",")
			config.Agent, _ = strconv.Atoi(AgentConfList[0])
		}
	}
	if _, isSet := params["args"]; isSet {
		config.Args = strings.TrimSpace(params["args"].(string))
	}
	if _, isSet := params["balloon"]; isSet {
		balloon := int(params["balloon"].(float64))
		if balloon > 0 {
			config.Balloon = balloon
		}
	}
	//boot by default from hard disk (c), CD-ROM (d), network (n).
	if _, isSet := params["boot"]; isSet {
		config.Boot = params["boot"].(string)
	}
	if _, isSet := params["bootdisk"]; isSet {
		config.BootDisk = params["bootdisk"].(string)
	}
	if _, isSet := params["bios"]; isSet {
		config.Bios = params["bios"].(string)
	}
	if _, isSet := params["cicustom"]; isSet {
		config.CIcustom = params["cicustom"].(string)
	}
	if _, isSet := params["cipassword"]; isSet {
		config.CIpassword = params["cipassword"].(string)
	}
	if _, isSet := params["ciuser"]; isSet {
		config.CIuser = params["ciuser"].(string)
	}
	if _, isSet := params["description"]; isSet {
		config.Description = strings.TrimSpace(params["description"].(string))
	}
	//Can be network,disk,cpu,memory,usb
	if _, isSet := params["hotplug"]; isSet {
		config.Hotplug = params["hotplug"].(string)
	}
	if _, isSet := params["hookscript"]; isSet {
		config.Hookscript = params["hookscript"].(string)
	}
	if _, isSet := params["memory"]; isSet {
		config.Memory = int(params["memory"].(float64))
	}
	if _, isSet := params["name"]; isSet {
		config.Name = params["name"].(string)
	}
	if _, isSet := params["nameserver"]; isSet {
		config.Nameserver = params["nameserver"].(string)
	}
	if _, isSet := params["onboot"]; isSet {
		config.Onboot = PointerBool(Itob(int(params["onboot"].(float64))))
	}
	if _, isSet := params["cores"]; isSet {
		config.QemuCores = int(params["cores"].(float64))
	}
	if _, isSet := params["cpu"]; isSet {
		config.QemuCpu = params["cpu"].(string)
	}
	if _, isSet := params["kvm"]; isSet {
		config.QemuKVM = PointerBool(Itob(int(params["kvm"].(float64))))
	}
	if _, isSet := params["numa"]; isSet {
		config.QemuNuma = PointerBool(Itob(int(params["numa"].(float64))))
	}
	if _, isSet := params["ostype"]; isSet {
		config.QemuOs = params["ostype"].(string)
	}
	if _, isSet := params["sockets"]; isSet {
		config.QemuSockets = int(params["sockets"].(float64))
	}
	if _, isSet := params["vcpus"]; isSet {
		vCpu := int(params["vcpus"].(float64))
		if vCpu > 0 {
			config.QemuVcpus = vCpu
		}
	}
	if _, isSet := params["scsihw"]; isSet {
		config.Scsihw = params["scsihw"].(string)
	}
	if _, isSet := params["searchdomain"]; isSet {
		config.Searchdomain = params["searchdomain"].(string)
	}
	if _, isSet := params["sshkeys"]; isSet {
		config.Sshkeys, _ = url.PathUnescape(params["sshkeys"].(string))
	}
	if _, isSet := params["startup"]; isSet {
		config.Startup = params["startup"].(string)
	}
	if _, isSet := params["tablet"]; isSet {
		config.Tablet = PointerBool(Itob(int(params["tablet"].(float64))))
	}
	if _, isSet := params["tags"]; isSet {
		config.Tags = strings.TrimSpace(params["tags"].(string))
	}
	if _, isSet := params["smbios1"]; isSet {
		config.Smbios1 = params["smbios1"].(string)
	}

	ipconfigNames := []string{}

	for k := range params {
		if ipconfigName := rxIpconfigName.FindStringSubmatch(k); len(ipconfigName) > 0 {
			ipconfigNames = append(ipconfigNames, ipconfigName[0])
		}
	}

	if len(ipconfigNames) > 0 {
		config.Ipconfig = IpconfigMap{}
		for _, ipconfigName := range ipconfigNames {
			ipConfStr := params[ipconfigName]
			id := rxDeviceID.FindStringSubmatch(ipconfigName)
			ipconfigID, _ := strconv.Atoi(id[0])
			config.Ipconfig[ipconfigID] = ipConfStr
		}
	}

	linkedVmId := uint(0)
	config.Disks = QemuStorages{}.mapToStruct(params, &linkedVmId)
	if linkedVmId != 0 {
		config.LinkedVmId = linkedVmId
	}

	if config.Disks != nil && config.Disks.Ide != nil && config.Disks.Ide.Disk_2 != nil && config.Disks.Ide.Disk_2.CdRom != nil {
		config.Iso = config.Disks.Ide.Disk_2.CdRom.Iso
	}

	// Add unused disks
	// unused0:local:100/vm-100-disk-1.qcow2
	unusedDiskNames := []string{}
	for k := range params {
		// look for entries from the config in the format "unusedX:<storagepath>" where X is an integer
		if unusedDiskName := rxUnusedDiskName.FindStringSubmatch(k); len(unusedDiskName) > 0 {
			unusedDiskNames = append(unusedDiskNames, unusedDiskName[0])
		}
	}
	// if len(unusedDiskNames) > 0 {
	// 	log.Printf("[DEBUG] unusedDiskNames: %v", unusedDiskNames)
	// }

	if len(unusedDiskNames) > 0 {
		config.QemuUnusedDisks = QemuDevices{}
		for _, unusedDiskName := range unusedDiskNames {
			unusedDiskConfStr := params[unusedDiskName].(string)
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
			config.QemuUnusedDisks[diskID] = finalDiskConfMap
			config.QemuUnusedDisks[diskID] = finalDiskConfMap
		}
	}
	//Display

	if vga, isSet := params["vga"]; isSet {
		vgaList := strings.Split(vga.(string), ",")
		vgaMap := QemuDevice{}

		vgaMap.readDeviceConfig(vgaList)
		if len(vgaMap) > 0 {
			config.QemuVga = vgaMap
		}
	}

	// Add networks.
	nicNames := []string{}

	for k := range params {
		if nicName := rxNicName.FindStringSubmatch(k); len(nicName) > 0 {
			nicNames = append(nicNames, nicName[0])
		}
	}

	if len(nicNames) > 0 {
		config.QemuNetworks = QemuDevices{}
		for _, nicName := range nicNames {
			nicConfStr := params[nicName]
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
			nicConfMap.readDeviceConfig(nicConfList[1:])
			switch nicConfMap["firewall"] {
			case 1:
				nicConfMap["firewall"] = true
			case 0:
				nicConfMap["firewall"] = false
			}
			switch nicConfMap["link_down"] {
			case 1:
				nicConfMap["link_down"] = true
			case 0:
				nicConfMap["link_down"] = false
			}

			// And device config to networks.
			if len(nicConfMap) > 0 {
				config.QemuNetworks[nicID] = nicConfMap
			}
		}
	}

	// Add serials
	serialNames := []string{}

	for k := range params {
		if serialName := rxSerialName.FindStringSubmatch(k); len(serialName) > 0 {
			serialNames = append(serialNames, serialName[0])
		}
	}

	if len(serialNames) > 0 {
		config.QemuSerials = QemuDevices{}
		for _, serialName := range serialNames {
			id := rxDeviceID.FindStringSubmatch(serialName)
			serialID, _ := strconv.Atoi(id[0])

			serialConfMap := QemuDevice{
				"id":   serialID,
				"type": params[serialName],
			}

			// And device config to serials map.
			if len(serialConfMap) > 0 {
				config.QemuSerials[serialID] = serialConfMap
			}
		}
	}

	// Add usbs
	usbNames := []string{}

	for k := range params {
		if usbName := rxUsbName.FindStringSubmatch(k); len(usbName) > 0 {
			usbNames = append(usbNames, usbName[0])
		}
	}

	if len(usbNames) > 0 {
		config.QemuUsbs = QemuDevices{}
		for _, usbName := range usbNames {
			usbConfStr := params[usbName]
			usbConfList := strings.Split(usbConfStr.(string), ",")
			id := rxDeviceID.FindStringSubmatch(usbName)
			usbID, _ := strconv.Atoi(id[0])
			_, host := ParseSubConf(usbConfList[0], "=")

			usbConfMap := QemuDevice{
				"id":   usbID,
				"host": host,
			}

			usbConfMap.readDeviceConfig(usbConfList[1:])
			if usbConfMap["usb3"] == 1 {
				usbConfMap["usb3"] = true
			}

			// And device config to usbs map.
			if len(usbConfMap) > 0 {
				config.QemuUsbs[usbID] = usbConfMap
			}
		}
	}

	// hostpci
	hostPCInames := []string{}

	for k := range params {
		if hostPCIname := rxPCIName.FindStringSubmatch(k); len(hostPCIname) > 0 {
			hostPCInames = append(hostPCInames, hostPCIname[0])
		}
	}

	if len(hostPCInames) > 0 {
		config.QemuPCIDevices = QemuDevices{}
		for _, hostPCIname := range hostPCInames {
			hostPCIConfStr := params[hostPCIname]
			hostPCIConfList := strings.Split(hostPCIConfStr.(string), ",")
			id := rxPCIName.FindStringSubmatch(hostPCIname)
			hostPCIID, _ := strconv.Atoi(id[0])
			hostPCIConfMap := QemuDevice{
				"id": hostPCIID,
			}
			hostPCIConfMap.readDeviceConfig(hostPCIConfList)
			// And device config to usbs map.
			if len(hostPCIConfMap) > 0 {
				config.QemuPCIDevices[hostPCIID] = hostPCIConfMap
			}
		}
	}

	return &config, nil
}

func (newConfig ConfigQemu) Update(rebootIfNeeded bool, vmr *VmRef, client *Client) (rebootRequired bool, err error) {
	currentConfig, err := NewConfigQemuFromApi(vmr, client)
	if err != nil {
		return
	}
	return newConfig.setAdvanced(currentConfig, rebootIfNeeded, vmr, client)
}

func (config *ConfigQemu) setVmr(vmr *VmRef) (err error) {
	if config == nil {
		return errors.New("config may not be nil")
	}
	if err = vmr.nilCheck(); err != nil {
		return
	}
	vmr.SetVmType("qemu")
	config.VmID = vmr.vmId
	return
}

func (newConfig ConfigQemu) setAdvanced(currentConfig *ConfigQemu, rebootIfNeeded bool, vmr *VmRef, client *Client) (rebootRequired bool, err error) {
	err = newConfig.setVmr(vmr)
	if err != nil {
		return
	}
	err = newConfig.Validate()
	if err != nil {
		return
	}

	var params map[string]interface{}
	var exitStatus string

	if currentConfig != nil {
		// Update
		if newConfig.Disks != nil && currentConfig.Disks != nil {
			markedDisks := newConfig.Disks.markDiskChanges(*currentConfig.Disks)
			// move disk to different storage or change disk format
			for _, e := range markedDisks.Move {
				_, err = e.move(true, vmr, client)
				if err != nil {
					return
				}
			}
			// increase Disks in size
			for _, e := range markedDisks.Resize {
				_, err = e.resize(vmr, client)
				if err != nil {
					return
				}
			}
			// Moving disks changes the disk id. we need to get the config again if any disk was moved
			if len(markedDisks.Move) != 0 {
				currentConfig, err = NewConfigQemuFromApi(vmr, client)
				if err != nil {
					return
				}
			}
		}

		// Migrate VM
		if newConfig.Node != currentConfig.Node {
			vmr.SetNode(currentConfig.Node)
			_, err = client.MigrateNode(vmr, newConfig.Node, true)
			if err != nil {
				return
			}
			// Set node to the node the VM was migrated to
			vmr.SetNode(newConfig.Node)
		}

		rebootRequired, params, err = newConfig.mapToApiValues(*currentConfig)
		if err != nil {
			return
		}
		exitStatus, err = client.PutWithTask(params, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/config")
		if err != nil {
			return false, fmt.Errorf("error updating VM: %v, error status: %s (params: %v)", err, exitStatus, params)
		}

		if !rebootRequired {
			rebootRequired, err = GuestHasPendingChanges(vmr, client)
			if err != nil {
				return
			}
		}

		if rebootRequired && rebootIfNeeded {
			if err = GuestReboot(vmr, client); err != nil {
				return
			}
			rebootRequired = false
		}
	} else {
		// Create

		_, params, err = newConfig.mapToApiValues(ConfigQemu{})
		if err != nil {
			return
		}
		exitStatus, err = client.CreateQemuVm(vmr.node, params)
		if err != nil {
			return false, fmt.Errorf("error creating VM: %v, error status: %s (params: %v)", err, exitStatus, params)
		}
	}

	_, err = client.UpdateVMHA(vmr, newConfig.HaState, newConfig.HaGroup)
	if err != nil {
		return
	}

	_, err = client.UpdateVMPool(vmr, newConfig.Pool)
	return
}

func (config ConfigQemu) Validate() (err error) {
	// TODO test all other use cases
	// TODO has no context about changes caused by updating the vm
	if config.Disks != nil {
		err = config.Disks.Validate()
		if err != nil {
			return
		}
	}

	return
}

// HasCloudInit - are there cloud-init options?
func (config ConfigQemu) HasCloudInit() bool {
	for _, config := range config.Ipconfig {
		if config != nil && config != "" {
			return true
		}
	}
	return config.CIuser != "" ||
		config.CIpassword != "" ||
		config.Searchdomain != "" ||
		config.Nameserver != "" ||
		config.Sshkeys != "" ||
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
	var storage string
	fullClone := "1"
	if config.FullClone != nil {
		fullClone = strconv.Itoa(*config.FullClone)
	}
	if disk0Storage, ok := config.QemuDisks[0]["storage"].(string); ok && len(disk0Storage) > 0 {
		storage = disk0Storage
	}
	params := map[string]interface{}{
		"newid":  vmr.vmId,
		"target": vmr.node,
		"name":   config.Name,
		"full":   fullClone,
	}
	if vmr.pool != "" {
		params["pool"] = vmr.pool
	}

	if fullClone == "1" && storage != "" {
		params["storage"] = storage
	}

	_, err = client.CloneQemuVm(sourceVmr, params)
	return err
}

// DEPRECATED use ConfigQemu.Update instead
func (config ConfigQemu) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	configParams := map[string]interface{}{}

	//Array to list deleted parameters
	//deleteParams := []string{}

	if config.Agent != 0 {
		configParams["agent"] = config.Agent
	}
	if config.QemuOs != "" {
		configParams["ostype"] = config.QemuOs
	}
	if config.QemuCores != 0 {
		configParams["cores"] = config.QemuCores
	}
	if config.Memory != 0 {
		configParams["memory"] = config.Memory
	}

	if config.QemuSockets != 0 {
		configParams["sockets"] = config.QemuSockets
	}

	if config.QemuKVM != nil {
		configParams["kvm"] = *config.QemuKVM
	}

	if config.QemuNuma != nil {
		configParams["numa"] = *config.QemuNuma
	}

	if config.Onboot != nil {
		configParams["onboot"] = *config.Onboot
	}

	if config.Tablet != nil {
		configParams["tablet"] = *config.Tablet
	}

	if config.Args != "" {
		configParams["args"] = config.Args
	}

	if config.Tags != "" {
		configParams["tags"] = config.Tags
	}

	if config.Startup != "" {
		configParams["startup"] = config.Startup
	}

	if config.QemuIso != "" {
		configParams["ide2"] = config.QemuIso + ",media=cdrom"
	}

	if config.Bios != "" {
		configParams["bios"] = config.Bios
	}

	if config.Hotplug != "" {
		configParams["hotplug"] = config.Hotplug
	}

	if config.Name != "" {
		configParams["name"] = config.Name
	}

	if config.Description != "" {
		configParams["description"] = config.Description
	}

	if config.Balloon >= 1 {
		configParams["balloon"] = config.Balloon
	}

	if config.QemuVcpus >= 1 {
		configParams["vcpus"] = config.QemuVcpus
	}

	if config.Boot != "" {
		configParams["boot"] = config.Boot
	}

	if config.BootDisk != "" {
		configParams["bootdisk"] = config.BootDisk
	}

	if config.Hookscript != "" {
		configParams["hookscript"] = config.Hookscript
	}

	if config.QemuCpu != "" {
		configParams["cpu"] = config.QemuCpu
	}

	if config.Scsihw != "" {
		configParams["scsihw"] = config.Scsihw
	}

	err = config.CreateQemuMachineParam(configParams)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// Create disks config.
	configParamsDisk := map[string]interface{}{
		"vmid": vmr.vmId,
	}
	config.CreateQemuDisksParams(configParamsDisk, false)
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
	config.VmID = vmr.vmId
	config.CreateQemuNetworksParams(configParams)

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		configParams["vga"] = strings.Join(vgaParam, ",")
	}
	// Create serial interfaces
	config.CreateQemuSerialsParams(configParams)

	// Create usb interfaces
	config.CreateQemuUsbsParams(configParams)

	config.CreateQemuPCIsParams(configParams)

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
	if config.Smbios1 != "" {
		configParams["smbios1"] = config.Smbios1
	}
	if config.Sshkeys != "" {
		configParams["sshkeys"] = sshKeyUrlEncode(config.Sshkeys)
	}
	err = config.CreateIpconfigParams(configParams)
	if err != nil {
		log.Printf("[ERROR] %q", err)
	}

	// if len(deleteParams) > 0 {
	// 	configParams["delete"] = strings.Join(deleteParams, ", ")
	// }

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

func NewConfigQemuFromJson(input []byte) (config *ConfigQemu, err error) {
	config = &ConfigQemu{}
	err = json.Unmarshal([]byte(input), config)
	if err != nil {
		log.Fatal(err)
	}
	return
}

var (
	rxDeviceID       = regexp.MustCompile(`\d+`)
	rxUnusedDiskName = regexp.MustCompile(`^(unused)\d+`)
	rxNicName        = regexp.MustCompile(`net\d+`)
	rxMpName         = regexp.MustCompile(`mp\d+`)
	rxSerialName     = regexp.MustCompile(`serial\d+`)
	rxUsbName        = regexp.MustCompile(`usb\d+`)
	rxPCIName        = regexp.MustCompile(`hostpci\d+`)
	rxIpconfigName   = regexp.MustCompile(`ipconfig\d+`)
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
	config, err = ConfigQemu{}.mapToStruct(vmConfig)
	if err != nil {
		return
	}
	config.defaults()
	// HAstate is return by the api for a vm resource type but not the HAgroup
	err = client.ReadVMHA(vmr)
	if err == nil {
		config.HaState = vmr.HaState()
		config.HaGroup = vmr.HaGroup()
	} else {
		//log.Printf("[DEBUG] VM %d(%s) has no HA config", vmr.vmId, vmConfig["hostname"])
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

// URL encodes the ssh keys
func sshKeyUrlEncode(keys string) (encodedKeys string) {
	encodedKeys = url.PathEscape(keys + "\n")
	encodedKeys = strings.Replace(encodedKeys, "+", "%2B", -1)
	encodedKeys = strings.Replace(encodedKeys, "@", "%40", -1)
	encodedKeys = strings.Replace(encodedKeys, "=", "%3D", -1)
	encodedKeys = strings.Replace(encodedKeys, ":", "%3A", -1)
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
	vms, err := client.GetResourceList(resourceListGuest)
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

// Given a QemuDevice (representing a disk), return a param string to give to ProxMox
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

	// Backup
	if backup, ok := disk["backup"].(bool); ok {
		// Backups are enabled by default (backup=1)
		// Only set the parameter if backups are explicitly disabled
		if !backup {
			diskConfParam = append(diskConfParam, "backup=0")
		}
	}

	// Keys that are not used as real/direct conf.
	ignoredKeys := []string{"backup", "key", "slot", "type", "storage", "file", "size", "cache", "volume", "container", "vm", "mountoptions", "storage_type"}

	// Rest of config.
	diskConfParam = diskConfParam.createDeviceParam(disk, ignoredKeys)

	return strings.Join(diskConfParam, ",")
}

// Given a QemuDevice (representing a usb), return a param string to give to ProxMox
func FormatUsbParam(usb QemuDevice) string {
	usbConfParam := QemuDeviceParam{}

	usbConfParam = usbConfParam.createDeviceParam(usb, []string{})

	return strings.Join(usbConfParam, ",")
}

// Create parameters for each Nic device.
func (c ConfigQemu) CreateQemuNetworksParams(params map[string]interface{}) {
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
			pairing := c.VmID<<5 | nicID
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

		// use model=mac format for older proxmox compatibility as the parameters which will be sent to Proxmox API.
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
}

// Create parameters for each Cloud-Init ipconfig entry.
func (c ConfigQemu) CreateIpconfigParams(params map[string]interface{}) error {

	for ID, config := range c.Ipconfig {
		if ID > 15 {
			return fmt.Errorf("only up to 16 Cloud-Init network configurations supported (ipconfig[0-15]), skipping ipconfig%d", ID)
		}

		if config != "" {
			ipconfigName := "ipconfig" + strconv.Itoa(ID)
			params[ipconfigName] = config
		}
	}

	return nil
}

// Create RNG parameter.
func (c ConfigQemu) CreateQemuRngParams(params map[string]interface{}) {
	rngParam := QemuDeviceParam{}
	rngParam = rngParam.createDeviceParam(c.RNGDrive, nil)

	if len(rngParam) > 0 {
		rng_info := []string{}
		rng := ""
		for _, param := range rngParam {
			key := strings.Split(param, "=")
			rng_info = append(rng_info, fmt.Sprintf("%s=%s", key[0], key[1]))
		}
		if len(rng_info) > 0 {
			rng = strings.Join(rng_info, ",")
			params["rng0"] = rng
		}
	}
}

// Create efi parameter.
func (c ConfigQemu) CreateQemuEfiParams(params map[string]interface{}) {
	efiParam := QemuDeviceParam{}
	efiParam = efiParam.createDeviceParam(c.EFIDisk, nil)

	if len(efiParam) > 0 {
		storage_info := []string{}
		storage := ""
		for _, param := range efiParam {
			key := strings.Split(param, "=")
			if key[0] == "storage" {
				// Proxmox format for disk creation
				storage = fmt.Sprintf("%s:1", key[1])
			} else {
				storage_info = append(storage_info, param)
			}
		}
		if len(storage_info) > 0 {
			storage = fmt.Sprintf("%s,%s", storage, strings.Join(storage_info, ","))
		}
		params["efidisk0"] = storage
	}
}

// Create parameters for each disk.
func (c ConfigQemu) CreateQemuDisksParams(params map[string]interface{}, cloned bool) {
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
}

// Create parameters for each PCI Device
func (c ConfigQemu) CreateQemuPCIsParams(params map[string]interface{}) {
	// For new style with multi pci device.
	for pciConfID, pciConfMap := range c.QemuPCIDevices {
		qemuPCIName := "hostpci" + strconv.Itoa(pciConfID)
		var pcistring bytes.Buffer
		for elem := range pciConfMap {
			pcistring.WriteString(elem)
			pcistring.WriteString("=")
			pcistring.WriteString(fmt.Sprintf("%v", pciConfMap[elem]))
			pcistring.WriteString(",")
		}

		// Add back to Qemu prams.
		params[qemuPCIName] = strings.TrimSuffix(pcistring.String(), ",")
	}
}

// Create parameters for serial interface
func (c ConfigQemu) CreateQemuSerialsParams(params map[string]interface{}) {
	// For new style with multi disk device.
	for serialID, serialConfMap := range c.QemuSerials {
		// Device name.
		deviceType := serialConfMap["type"].(string)
		qemuSerialName := "serial" + strconv.Itoa(serialID)

		// Add back to Qemu prams.
		params[qemuSerialName] = deviceType
	}
}

// Create parameters for usb interface
func (c ConfigQemu) CreateQemuUsbsParams(params map[string]interface{}) {
	for usbID, usbConfMap := range c.QemuUsbs {
		qemuUsbName := "usb" + strconv.Itoa(usbID)

		// Add back to Qemu prams.
		params[qemuUsbName] = FormatUsbParam(usbConfMap)
	}
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
			} else if iValue, ok := value.(float64); ok && iValue > 0 {
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
func (confMap QemuDevice) readDeviceConfig(confList []string) {
	// Add device config.
	for _, conf := range confList {
		key, value := ParseSubConf(conf, "=")
		confMap[key] = value
	}
}

func (c ConfigQemu) String() string {
	jsConf, _ := json.Marshal(c)
	return string(jsConf)
}
