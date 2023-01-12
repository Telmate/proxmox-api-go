package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// LXC options for the Proxmox API
type ConfigLxc struct {
	Ostemplate         string      `json:"ostemplate"`
	Arch               string      `json:"arch"`
	BWLimit            int         `json:"bwlimit,omitempty"`
	Clone              string      `json:"clone,omitempty"`
	CloneStorage       string      `json:"clone-storage,omitempty"`
	CMode              string      `json:"cmode"`
	Console            bool        `json:"console"`
	Cores              int         `json:"cores,omitempty"`
	CPULimit           int         `json:"cpulimit"`
	CPUUnits           int         `json:"cpuunits"`
	Description        string      `json:"description,omitempty"`
	Features           QemuDevice  `json:"features,omitempty"`
	Force              bool        `json:"force,omitempty"`
	Full               bool        `json:"full,omitempty"`
	HaState            string      `json:"hastate,omitempty"`
	HaGroup            string      `json:"hagroup,omitempty"`
	Hookscript         string      `json:"hookscript,omitempty"`
	Hostname           string      `json:"hostname,omitempty"`
	IgnoreUnpackErrors bool        `json:"ignore-unpack-errors,omitempty"`
	Lock               string      `json:"lock,omitempty"`
	Memory             int         `json:"memory"`
	Mountpoints        QemuDevices `json:"mountpoints,omitempty"`
	Nameserver         string      `json:"nameserver,omitempty"`
	Networks           QemuDevices `json:"networks,omitempty"`
	OnBoot             bool        `json:"onboot"`
	OsType             string      `json:"ostype,omitempty"`
	Password           string      `json:"password,omitempty"`
	Pool               string      `json:"pool,omitempty"`
	Protection         bool        `json:"protection"`
	Restore            bool        `json:"restore,omitempty"`
	RootFs             QemuDevice  `json:"rootfs,omitempty"`
	SearchDomain       string      `json:"searchdomain,omitempty"`
	Snapname           string      `json:"snapname,omitempty"`
	SSHPublicKeys      string      `json:"ssh-public-keys,omitempty"`
	Start              bool        `json:"start"`
	Startup            string      `json:"startup,omitempty"`
	Storage            string      `json:"storage"`
	Swap               int         `json:"swap"`
	Template           bool        `json:"template,omitempty"`
	Tty                int         `json:"tty"`
	Unique             bool        `json:"unique,omitempty"`
	Unprivileged       bool        `json:"unprivileged"`
	Tags               string      `json:"tags"`
	Unused             []string    `json:"unused,omitempty"`
}

func NewConfigLxc() ConfigLxc {
	return ConfigLxc{
		Arch:         "amd64",
		CMode:        "tty",
		Console:      true,
		CPULimit:     0,
		CPUUnits:     1024,
		Memory:       512,
		OnBoot:       false,
		Protection:   false,
		Start:        false,
		Storage:      "local",
		Swap:         512,
		Template:     false,
		Tty:          2,
		Unprivileged: false,
	}
}

func NewConfigLxcFromJson(input []byte) (config ConfigLxc, err error) {
	config = NewConfigLxc()
	err = json.Unmarshal([]byte(input), &config)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func NewConfigLxcFromApi(vmr *VmRef, client *Client) (config *ConfigLxc, err error) {
	// prepare json map to receive the information from the api
	var lxcConfig map[string]interface{}
	lxcConfig, err = client.GetVmConfig(vmr)
	if err != nil {
		return nil, err
	}

	// prepare a new lxc config to store and return\
	// the information from api
	newConfig := NewConfigLxc()
	config = &newConfig

	arch := ""
	if _, isSet := lxcConfig["arch"]; isSet {
		arch = lxcConfig["arch"].(string)
	}
	cmode := ""
	if _, isSet := lxcConfig["cmode"]; isSet {
		cmode = lxcConfig["cmode"].(string)
	}
	console := true
	if _, isSet := lxcConfig["console"]; isSet {
		console = Itob(int(lxcConfig["console"].(float64)))
	}
	cores := 0
	if _, isSet := lxcConfig["cores"]; isSet {
		cores = int(lxcConfig["cores"].(float64))
	}
	cpulimit := 0
	if _, isSet := lxcConfig["cpulimit"]; isSet {
		cpulimit, _ = strconv.Atoi(lxcConfig["cpulimit"].(string))
	}
	cpuunits := 1024
	if _, isSet := lxcConfig["cpuunits"]; isSet {
		cpuunits = int(lxcConfig["cpuunits"].(float64))
	}
	description := ""
	if _, isSet := lxcConfig["description"]; isSet {
		description = lxcConfig["description"].(string)
	}

	// add features, if any
	if features, isSet := lxcConfig["features"]; isSet {
		featureList := strings.Split(features.(string), ",")

		// create new device map to store features
		featureMap := QemuDevice{}
		// add all features to device map
		featureMap.readDeviceConfig(featureList)
		// prepare empty feature map
		if config.Features == nil {
			config.Features = QemuDevice{}
		}
		// and device config to networks
		if len(featureMap) > 0 {
			config.Features = featureMap
		}
	}
	hookscript := ""
	if _, isSet := lxcConfig["hookscript"]; isSet {
		hookscript = lxcConfig["hookscript"].(string)
	}
	hostname := ""
	if _, isSet := lxcConfig["hostname"]; isSet {
		hostname = lxcConfig["hostname"].(string)
	}
	lock := ""
	if _, isSet := lxcConfig["lock"]; isSet {
		lock = lxcConfig["lock"].(string)
	}
	memory := 512
	if _, isSet := lxcConfig["memory"]; isSet {
		memory = int(lxcConfig["memory"].(float64))
	}

	// add rootfs
	rootfs := QemuDevice{}
	if rootfsStr, isSet := lxcConfig["rootfs"]; isSet {
		rootfs = ParsePMConf(rootfsStr.(string), "volume")
	}

	// add mountpoints
	mpNames := []string{}

	for k := range lxcConfig {
		if mpName := rxMpName.FindStringSubmatch(k); len(mpName) > 0 {
			mpNames = append(mpNames, mpName[0])
		}
	}

	for _, mpName := range mpNames {
		mpConfStr := lxcConfig[mpName].(string)
		mpConfMap := ParseLxcDisk(mpConfStr)

		// add mp id
		id := rxDeviceID.FindStringSubmatch(mpName)
		mpID, _ := strconv.Atoi(id[0])
		mpConfMap["slot"] = mpID

		// 5 potential boolean flags need to be converted
		for _, key := range []string{"acl", "backup", "quota", "replicate", "shared"} {
			// if flag is set, need to convert int to bool
			if _, isSet := mpConfMap[key]; isSet {
				mpConfMap[key] = Itob(mpConfMap[key].(int))
			}
		}

		// prepare empty mountpoint map
		if config.Mountpoints == nil {
			config.Mountpoints = QemuDevices{}
		}
		// and device config to mountpoints
		if len(mpConfMap) > 0 {
			config.Mountpoints[mpID] = mpConfMap
		}
	}

	nameserver := ""
	if _, isSet := lxcConfig["nameserver"]; isSet {
		nameserver = lxcConfig["nameserver"].(string)
	}

	// add networks
	nicNames := []string{}

	for k := range lxcConfig {
		if nicName := rxNicName.FindStringSubmatch(k); len(nicName) > 0 {
			nicNames = append(nicNames, nicName[0])
		}
	}

	for _, nicName := range nicNames {
		nicConfStr := lxcConfig[nicName]
		nicConfList := strings.Split(nicConfStr.(string), ",")

		id := rxDeviceID.FindStringSubmatch(nicName)
		nicID, _ := strconv.Atoi(id[0])
		// add nic id
		nicConfMap := QemuDevice{
			"id": nicID,
		}
		// add rest of device config
		nicConfMap.readDeviceConfig(nicConfList)

		// if firewall flag is set, need to convert int to bool
		if _, isSet := nicConfMap["firewall"]; isSet {
			nicConfMap["firewall"] = Itob(nicConfMap["firewall"].(int))
		}

		// prepare empty network map
		if config.Networks == nil {
			config.Networks = QemuDevices{}
		}
		// and device config to networks
		if len(nicConfMap) > 0 {
			config.Networks[nicID] = nicConfMap
		}
	}

	onboot := false
	if _, isSet := lxcConfig["onboot"]; isSet {
		onboot = Itob(int(lxcConfig["onboot"].(float64)))
	}
	ostype := ""
	if _, isSet := lxcConfig["ostype"]; isSet {
		ostype = lxcConfig["ostype"].(string)
	}
	protection := false
	if _, isSet := lxcConfig["protection"]; isSet {
		protection = Itob(int(lxcConfig["protection"].(float64)))
	}
	searchdomain := ""
	if _, isSet := lxcConfig["searchdomain"]; isSet {
		searchdomain = lxcConfig["searchdomain"].(string)
	}
	startup := ""
	if _, isSet := lxcConfig["startup"]; isSet {
		startup = lxcConfig["startup"].(string)
	}
	swap := 512
	if _, isSet := lxcConfig["swap"]; isSet {
		swap = int(lxcConfig["swap"].(float64))
	}
	template := false
	if _, isSet := lxcConfig["template"]; isSet {
		template = Itob(int(lxcConfig["template"].(float64)))
	}
	tty := 2
	if _, isSet := lxcConfig["tty"]; isSet {
		tty = int(lxcConfig["tty"].(float64))
	}
	unprivileged := false
	if _, isset := lxcConfig["unprivileged"]; isset {
		unprivileged = Itob(int(lxcConfig["unprivileged"].(float64)))
	}
	tags := ""
	if _, isSet := lxcConfig["tags"]; isSet {
		tags = lxcConfig["tags"].(string)
	}
	unused := []string{}
	for k := range lxcConfig {
		// look for entries from the config in the format "unusedX:<storagepath>" where X is an integer
		if unusedDiskName := rxUnusedDiskName.FindStringSubmatch(k); len(unusedDiskName) > 0 {
			unused = append(unused, unusedDiskName[0])
		}
	}

	config.Arch = arch
	config.CMode = cmode
	config.Console = console
	config.Cores = cores
	config.CPULimit = cpulimit
	config.CPUUnits = cpuunits
	config.Description = description
	config.OnBoot = onboot
	config.Hookscript = hookscript
	config.Hostname = hostname
	config.Lock = lock
	config.Memory = memory
	config.Nameserver = nameserver
	config.OnBoot = onboot
	config.OsType = ostype
	config.Protection = protection
	config.RootFs = rootfs
	config.SearchDomain = searchdomain
	config.Startup = startup
	config.Swap = swap
	config.Template = template
	config.Tty = tty
	config.Unprivileged = unprivileged
	config.Unused = unused
	config.Tags = tags

	err = client.ReadVMHA(vmr)
	if err == nil {
		config.HaState = vmr.HaState()
		config.HaGroup = vmr.HaGroup()
	} else {
		//log.Printf("[DEBUG] Container %d(%s) has no HA config", vmr.vmId, lxcConfig["hostname"])
		return config, nil
	}

	return
}

// create LXC container using the Proxmox API
func (config ConfigLxc) CreateLxc(vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("lxc")
	paramMap := config.mapToApiValues()

	// amend vmid
	paramMap["vmid"] = vmr.vmId

	exitStatus, err := client.CreateLxcContainer(vmr.node, paramMap)
	if err != nil {
		params, _ := json.Marshal(&paramMap)
		return fmt.Errorf("error creating LXC container: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		return fmt.Errorf("[ERROR] %q", err)
	}

	return
}

func (config ConfigLxc) CloneLxc(vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("lxc")

	//map the clone specific parameters
	paramMap := map[string]interface{}{
		"newid":  vmr.vmId,
		"vmid":   config.Clone,
		"node":   vmr.node,
		"target": vmr.node,
		"full":   config.Full,
	}

	if config.BWLimit != 0 {
		paramMap["bwlimit"] = config.Hostname
	}

	if config.CloneStorage != "" {
		paramMap["storage"] = config.CloneStorage
	}

	if config.Description != "" {
		paramMap["description"] = config.Description
	}

	if config.Hostname != "" {
		paramMap["hostname"] = config.Hostname
	}

	if config.Pool != "" {
		paramMap["pool"] = config.Pool
	}

	if config.Snapname != "" {
		paramMap["snapname"] = config.Snapname
	}

	exitStatus, err := client.CloneLxcContainer(vmr, paramMap)
	if err != nil {
		params, _ := json.Marshal(&paramMap)
		return fmt.Errorf("error cloning LXC container: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		return fmt.Errorf("[ERROR] %q", err)
	}

	return
}

func (config ConfigLxc) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	paramMap := config.mapToApiValues()

	// delete parameters which are not supported in updated operations
	delete(paramMap, "pool")
	delete(paramMap, "storage")
	delete(paramMap, "password")
	delete(paramMap, "ostemplate")
	delete(paramMap, "start")
	delete(paramMap, "clone")
	delete(paramMap, "full")

	// even though it is listed as a PUT option in the API documentation
	// we remove it here because "it should not be modified manually";
	// also, error "500 unable to modify read-only option: 'unprivileged'"
	delete(paramMap, "unprivileged")

	_, err = client.UpdateVMHA(vmr, config.HaState, config.HaGroup)
	if err != nil {
		return err
	}

	_, err = client.SetLxcConfig(vmr, paramMap)
	return err
}

func ParseLxcDisk(diskStr string) QemuDevice {
	disk := ParsePMConf(diskStr, "volume")

	// add features, if any
	if mountoptions, isSet := disk["mountoptions"]; isSet {
		moList := strings.Split(mountoptions.(string), ";")
		moMap := map[string]bool{}
		for _, mo := range moList {
			moMap[mo] = true
		}
		disk["mountoptions"] = moMap
	}

	storageName, fileName := ParseSubConf(disk["volume"].(string), ":")
	disk["storage"] = storageName
	disk["file"] = fileName

	return disk
}

func (config ConfigLxc) mapToApiValues() map[string]interface{} {
	// convert config to map
	params, _ := json.Marshal(&config)
	var paramMap map[string]interface{}
	json.Unmarshal(params, &paramMap)

	// build list of features
	// add features as parameter list to lxc parameters
	// this overwrites the original formatting with a
	// comma separated list of "key=value" pairs
	paramMap["features"] = formatDeviceParam(config.Features)

	// format rootfs params as expected
	if rootfs := config.RootFs; rootfs != nil {
		paramMap["rootfs"] = FormatDiskParam(rootfs)
	}

	// build list of mountpoints
	// this does the same as for the feature list
	// except that there can be multiple of these mountpoint sets
	// and each mountpoint set comes with a new id
	for _, mpConfMap := range config.Mountpoints {
		// add mp to lxc parameters
		mpID := mpConfMap["slot"]
		mpName := fmt.Sprintf("mp%v", mpID)
		paramMap[mpName] = FormatDiskParam(mpConfMap)
	}

	// build list of network parameters
	for nicID, nicConfMap := range config.Networks {
		// add nic to lxc parameters
		nicName := fmt.Sprintf("net%v", nicID)
		paramMap[nicName] = formatDeviceParam(nicConfMap)
	}

	// build list of unused volumes for sake of completeness,
	// even if it is not recommended to change these volumes manually
	for volID, vol := range config.Unused {
		// add volume to lxc parameters
		volName := fmt.Sprintf("unused%v", volID)
		paramMap[volName] = vol
	}

	// now that we concatenated the key value parameter
	// list for the networks, mountpoints and unused volumes,
	// remove the original keys, since the Proxmox API does
	// not know how to handle this key
	delete(paramMap, "networks")
	delete(paramMap, "mountpoints")
	delete(paramMap, "unused")

	// also delete the hastate & hagroup key which is used elsewhere
	delete(paramMap, "hastate")
	delete(paramMap, "hagroup")

	return paramMap
}
