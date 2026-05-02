package proxmox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	QemuGuestInterface interface {
		Create(context.Context, ConfigQemu) (*VmRef, error)
		CreateNoCheck(context.Context, ConfigQemu) (*VmRef, error)

		// When allowRestart is false an error is return if the update rewuires a reboot or shutdown.
		Update(ctx context.Context, vmr VmRef, allowRestart bool, allowForceStop bool, config ConfigQemu) error
		UpdateNoCheck(ctx context.Context, vmr VmRef, allowRestart bool, allowForceStop bool, config ConfigQemu) error
	}

	qemuGuestClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ QemuGuestInterface = (*qemuGuestClient)(nil)

func (c *qemuGuestClient) Create(ctx context.Context, config ConfigQemu) (*VmRef, error) {
	client := c.oldClient
	version, err := client.Version(ctx)
	if err != nil {
		return nil, err
	}
	if err = config.Validate(nil, version); err != nil {
		return nil, err
	}

	var vmr *VmRef
	vmr, err = config.create(ctx, client, c.api, version)
	if err != nil {
		return nil, err
	}

	if err = client.insertCachedPermission(ctx, permissionPath(permissionCategory_GuestPath)+"/"+permissionPath(vmr.vmId.String())); err != nil {
		return nil, err
	}
	return vmr, nil
}

func (c *qemuGuestClient) CreateNoCheck(ctx context.Context, config ConfigQemu) (*VmRef, error) {
	client := c.oldClient
	version, err := client.Version(ctx)
	if err != nil {
		return nil, err
	}

	return config.create(ctx, client, c.api, version)

}

func (c *qemuGuestClient) Update(
	ctx context.Context, vmr VmRef,
	allowRestart bool, allowForceStop bool,
	config ConfigQemu,
) (err error) {
	if err = config.setVmr(&vmr); err != nil {
		return
	}

	client := c.oldClient

	var rawConfig *rawConfigQemu
	if rawConfig, err = guestGetQemuConfig(ctx, &vmr, client); err != nil {
		return
	}

	currentLegacy, err := rawConfig.Get(vmr) // LEGACY we shouldn't need full config as we can grab this from the raw config as needed.
	if err != nil {
		return
	}
	var version Version
	if version, err = client.Version(ctx); err != nil {
		return
	}
	if err = config.Validate(currentLegacy, version); err != nil {
		return
	}
	_, err = config.updateNoCheck(ctx, c.api, version, client, allowRestart, &vmr, currentLegacy, configQemuUpdate{raw: rawConfig})
	return
}

func (c *qemuGuestClient) UpdateNoCheck(
	ctx context.Context, vmr VmRef,
	allowRestart bool, allowForceStop bool,
	config ConfigQemu,
) (err error) {
	if err = config.setVmr(&vmr); err != nil {
		return
	}

	client := c.oldClient

	var rawConfig *rawConfigQemu
	if rawConfig, err = guestGetQemuConfig(ctx, &vmr, client); err != nil {
		return
	}

	currentLegacy, err := rawConfig.Get(vmr) // LEGACY we shouldn't need full config as we can grab this from the raw config as needed.
	if err != nil {
		return
	}

	var version Version
	if version, err = client.Version(ctx); err != nil {
		return
	}

	config.updateNoCheck(ctx, c.api, version, client, allowRestart, &vmr, currentLegacy, configQemuUpdate{})

	return
}

// Currently ZFS local, LVM, Ceph RBD, CephFS, Directory and virtio-scsi-pci are considered.
// Other formats are not verified, but could be added if they're needed.
// const rxStorageTypes = `(zfspool|lvm|rbd|cephfs|dir|virtio-scsi-pci)`

type (
	// TODO phase this out
	QemuDevices map[int]map[string]interface{}
	// TODO phase this out
	QemuDevice map[string]interface{}
	// TODO phase this out
	QemuDeviceParam []string
)

// ConfigQemu - Proxmox API QEMU options
type ConfigQemu struct {
	ID               *GuestID              `json:"id,omitempty"`   // Required for creation, cannot be changed
	Node             *NodeName             `json:"node,omitempty"` // Required for creation
	Agent            *QemuGuestAgent       `json:"agent,omitempty"`
	Args             string                `json:"args,omitempty"`
	Bios             string                `json:"bios,omitempty"`
	Boot             string                `json:"boot,omitempty"`     // TODO should be an array of custom enums
	BootDisk         string                `json:"bootdisk,omitempty"` // TODO discuss deprecation? Only returned as it's deprecated in the proxmox api
	CPU              *QemuCPU              `json:"cpu,omitempty"`      // never nil when returned
	CloudInit        *CloudInit            `json:"cloudinit,omitempty"`
	Description      *string               `json:"description,omitempty"` // never nil when returned
	Disks            *QemuStorages         `json:"disks,omitempty"`
	EfiDisk          *EfiDisk              `json:"efidisk,omitempty"`
	FullClone        *int                  `json:"fullclone,omitempty"` // Deprecated
	HaGroup          string                `json:"hagroup,omitempty"`
	HaState          string                `json:"hastate,omitempty"` // TODO should be custom type with enum
	Hookscript       string                `json:"hookscript,omitempty"`
	Hotplug          string                `json:"hotplug,omitempty"`   // TODO should be a struct
	Iso              *IsoFile              `json:"iso,omitempty"`       // Same as Disks.Ide.Disk_2.CdRom.Iso
	LinkedID         *GuestID              `json:"linked_id,omitempty"` // Only returned setting it has no effect
	Machine          string                `json:"machine,omitempty"`   // TODO should be custom type with enum
	Memory           *QemuMemory           `json:"memory,omitempty"`
	Name             *GuestName            `json:"name,omitempty"` // never nil when returned
	Networks         QemuNetworkInterfaces `json:"networks,omitempty"`
	Pool             *PoolName             `json:"pool,omitempty"`
	Protection       *bool                 `json:"protection,omitempty"` // never nil when returned
	QemuDisks        QemuDevices           `json:"disk,omitempty"`       // Deprecated use Disks *QemuStorages instead
	QemuIso          string                `json:"qemuiso,omitempty"`    // Deprecated use Iso *IsoFile instead
	QemuKVM          *bool                 `json:"kvm,omitempty"`
	QemuOs           string                `json:"ostype,omitempty"`
	PciDevices       QemuPciDevices        `json:"pci_devices,omitempty"`
	QemuPxe          bool                  `json:"pxe,omitempty"`
	QemuUnusedDisks  QemuDevices           `json:"unused,omitempty"` // TODO should be a struct
	USBs             QemuUSBs              `json:"usbs,omitempty"`
	QemuVga          QemuDevice            `json:"vga,omitempty"`    // TODO should be a struct
	Scsihw           string                `json:"scsihw,omitempty"` // TODO should be custom type with enum
	Serials          SerialInterfaces      `json:"serials,omitempty"`
	Smbios1          string                `json:"smbios1,omitempty"`            // TODO should be custom type with enum?
	StartAtNodeBoot  *bool                 `json:"start_at_node_boot,omitempty"` // Never nil when returned
	StartupShutdown  *StartupAndShutdown   `json:"startup_shutdown,omitempty"`
	State            *PowerState           `json:"state,omitempty"`   // Never returned
	Storage          string                `json:"storage,omitempty"` // this value is only used when doing a full clone and is never returned
	TPM              *TpmState             `json:"tpm,omitempty"`
	Tablet           *bool                 `json:"tablet,omitempty"` // never nil when returned
	Tags             *Tags                 `json:"tags,omitempty"`   // Never nil when returned
	RandomnessDevice *VirtIoRNG            `json:"randomness_device,omitempty"`
}

const (
	ConfigQemu_Error_UnableToUpdateWithoutReboot string = "unable to update vm without rebooting"
	ConfigQemu_Error_CpuRequired                 string = "cpu is required during creation"
	ConfigQemu_Error_MemoryRequired              string = "memory is required during creation"
	ConfigQemu_Error_NodeRequired                string = "node is required during creation"
)

// Deprecated: use QemuGuestInterface.Create() instead.
// Create - Tell Proxmox API to make the VM
func (config ConfigQemu) Create(ctx context.Context, client *Client) (*VmRef, error) {
	return client.New().QemuGuest.Create(ctx, config)
}

func (config ConfigQemu) create(ctx context.Context, client *Client, ca *clientAPI, version Version) (*VmRef, error) {
	params, body := config.mapToApiCreate(version)
	// pool field unsupported by /nodes/%s/vms/%d/config used by update (currentConfig != nil).
	// To be able to create directly in a configured pool, add pool to mapped params from ConfigQemu, before creating VM
	var pool PoolName
	if config.Pool != nil && *config.Pool != "" {
		params["pool"] = *config.Pool
	}
	var id GuestID
	var node NodeName
	if config.Node != nil {
		node = *config.Node
	}
	url := "/nodes/" + node.String() + "/qemu"
	var err error
	if config.ID == nil {
		id, err = guestCreateLoop_Unsafe(ctx, "vmid", url, params, body, client, ca)
		if err != nil {
			return nil, err
		}
	} else {
		id = *config.ID
		params["vmid"] = int(id)
		if err = ca.postRawTask(ctx, url, combineParamsAndBody(params, body)); err != nil {
			return nil, err
		}
	}

	vmr := &VmRef{
		node:   node,
		vmId:   id,
		pool:   pool,
		vmType: GuestQemu,
	}
	if err = resizeNewDisks(ctx, vmr, client, config.Disks, nil); err != nil {
		return nil, err
	}
	if config.State != nil && *config.State == PowerStateRunning {
		if err = vmr.start_Unsafe(ctx, ca); err != nil {
			return nil, err
		}
	}
	_, err = client.UpdateVMHA(ctx, vmr, config.HaState, config.HaGroup)
	return vmr, err
}

// TODO this should not be done here, but should be done in the unmarshaling of each respective field
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
	if config.Hotplug == "" {
		config.Hotplug = "network,disk,usb"
	}
	if config.QemuDisks == nil {
		config.QemuDisks = QemuDevices{}
	}
	if config.QemuKVM == nil {
		config.QemuKVM = util.Pointer(true)
	}
	if config.QemuOs == "" {
		config.QemuOs = "other"
	}
	if config.QemuUnusedDisks == nil {
		config.QemuUnusedDisks = QemuDevices{}
	}
	if config.QemuVga == nil {
		config.QemuVga = QemuDevice{}
	}
	if config.Scsihw == "" {
		config.Scsihw = "lsi"
	}
}

func (config ConfigQemu) mapToAPI(currentConfig ConfigQemu, version Version) (params map[string]interface{}) {
	// TODO check if cloudInit settings changed, they require a reboot to take effect.
	var itemsToDelete string

	params = map[string]any{}

	var guestID GuestID
	if config.ID != nil {
		guestID = *config.ID
	}
	if config.Args != "" {
		params["args"] = config.Args
	}
	if config.Agent != nil {
		params[qemuApiKeyGuestAgent] = config.Agent.mapToAPI(currentConfig.Agent)
	}
	if config.Bios != "" {
		params["bios"] = config.Bios
	}
	if config.Boot != "" {
		params["boot"] = config.Boot
	}
	if config.Description != nil && (*config.Description != "" || currentConfig.Description != nil) {
		params[qemuApiKeyDescription] = *config.Description
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
	if config.Name != nil {
		if currentConfig.Name == nil || *config.Name != *currentConfig.Name {
			params[qemuApiKeyName] = config.Name.String()
		}
	}
	if config.Protection != nil {
		params["protection"] = *config.Protection
	}
	if config.QemuOs != "" {
		params["ostype"] = config.QemuOs
	}
	if config.Scsihw != "" {
		params["scsihw"] = config.Scsihw
	}
	if config.StartAtNodeBoot != nil {
		if currentConfig.StartAtNodeBoot != nil {
			itemsToDelete += startAtNodeBootMapToApiUpdate(params, *config.StartAtNodeBoot, *currentConfig.StartAtNodeBoot)
		} else {
			startAtNodeBootMapToApiCreate(params, *config.StartAtNodeBoot)
		}
	}
	if config.StartupShutdown != nil {
		if currentConfig.StartupShutdown != nil {
			itemsToDelete += config.StartupShutdown.mapToApiUpdate(currentConfig.StartupShutdown, params)
		} else {
			config.StartupShutdown.mapToApiCreate(params)
		}
	}
	if config.Tablet != nil {
		params[qemuApiKeyTablet] = *config.Tablet
	}
	if config.Smbios1 != "" {
		params["smbios1"] = config.Smbios1
	}
	if config.TPM != nil {
		if delete := config.TPM.mapToApi(params, currentConfig.TPM); delete != "" {
			itemsToDelete = AddToList(itemsToDelete, delete)
		}
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
			var linkedID GuestID
			if currentConfig.LinkedID != nil {
				linkedID = *currentConfig.LinkedID
			}
			delete := config.Disks.mapToApiValues(*currentConfig.Disks, guestID, linkedID, params)
			if delete != "" {
				itemsToDelete = AddToList(itemsToDelete, delete)
			}
		}
	} else {
		if config.Disks != nil {
			// Create
			config.Disks.mapToApiValues(QemuStorages{}, guestID, 0, params)
		}
	}

	if config.CPU != nil {
		itemsToDelete += config.CPU.mapToApi(currentConfig.CPU, params, version)
	}
	if config.CloudInit != nil {
		itemsToDelete += config.CloudInit.mapToAPI(currentConfig.CloudInit, params, version)
	}
	if config.Memory != nil {
		itemsToDelete += config.Memory.mapToAPI(currentConfig.Memory, params)
	}
	if config.Serials != nil {
		itemsToDelete += config.Serials.mapToAPI(currentConfig.Serials, params)
	}

	if config.RandomnessDevice != nil {
		if currentConfig.RandomnessDevice != nil {
			itemsToDelete += config.RandomnessDevice.mapToAPIUpdateUnsafe(currentConfig.RandomnessDevice, params)
		} else {
			config.RandomnessDevice.mapToAPICreate(params)
		}
	}

	// Create networks config.
	itemsToDelete += config.Networks.mapToAPI(currentConfig.Networks, params)

	// Create vga config.
	vgaParam := QemuDeviceParam{}
	vgaParam = vgaParam.createDeviceParam(config.QemuVga, nil)
	if len(vgaParam) > 0 {
		params["vga"] = strings.Join(vgaParam, ",")
	}

	if config.USBs != nil {
		itemsToDelete += config.USBs.mapToAPI(currentConfig.USBs, params)
	}

	if config.PciDevices != nil {
		itemsToDelete += config.PciDevices.mapToAPI(currentConfig.PciDevices, params)
	}

	if itemsToDelete != "" {
		params["delete"] = strings.TrimPrefix(itemsToDelete, ",")
	}
	return
}

func (config ConfigQemu) mapToApiCreate(version Version) (params map[string]any, body *[]byte) {
	params = config.mapToAPI(ConfigQemu{}, version)
	if len(params) == 0 {
		params = nil
	}
	builder := strings.Builder{}
	if config.EfiDisk != nil {
		config.EfiDisk.mapToApiCreate(&builder)
	}
	if config.State != nil && *config.State == PowerStateRunning {
		builder.WriteString("&start=1")
	}
	if config.Tags != nil {
		if v := config.Tags.mapToApiCreateLower(); v != "" {
			builder.WriteString("&" + qemuApiKeyTags + "=")
			builder.WriteString(v)
		}
	}
	if builder.Len() > 0 {
		body = util.Pointer(bytes.NewBufferString(builder.String()[1:]).Bytes())
	}
	return
}

func (config ConfigQemu) mapToApiUpdate(currentLegacy *ConfigQemu, current configQemuUpdate, version Version) (params map[string]any, body *[]byte) {
	params = config.mapToAPI(*currentLegacy, version)
	if len(params) == 0 {
		params = nil
	}
	builder := strings.Builder{}
	delete := strings.Builder{}
	if config.EfiDisk != nil {
		if current.efiDisk != nil {
			config.EfiDisk.mapToApiUpdate(current.efiDisk, &builder, &delete)
		} else {
			config.EfiDisk.mapToApiCreate(&builder)
		}
	}
	if config.Tags != nil {
		if cur := current.raw.GetTags(); len(cur) != 0 {
			if v, ok := config.Tags.mapToApiUpdate(cur); ok {
				builder.WriteString("&" + qemuApiKeyTags + "=")
				builder.WriteString(v)
			}
		} else {
			if v := config.Tags.mapToApiCreate(); v != "" {
				builder.WriteString("&" + qemuApiKeyTags + "=")
				builder.WriteString(v)
			}
		}
	}

	if delete.Len() > 0 {
		if v, ok := params["delete"]; ok {
			params["delete"] = v.(string) + delete.String()
		} else {
			builder.WriteString("&delete=")
			builder.WriteString(delete.String()[1:]) // remove leading comma
		}
	}
	if builder.Len() > 0 {
		body = util.Pointer(bytes.NewBufferString(builder.String()[1:]).Bytes())
	}
	return
}

func (config *ConfigQemu) mapToStruct(vmr *VmRef, params map[string]interface{}) error {
	// vmConfig Sample: map[ cpu:host
	// net0:virtio=62:DF:XX:XX:XX:XX,bridge=vmbr0
	// ide2:local:iso/xxx-xx.iso,media=cdrom memory:2048
	// smbios1:uuid=8b3bf833-aad8-4545-xxx-xxxxxxx digest:aa6ce5xxxxx1b9ce33e4aaeff564d4 sockets:1
	// name:terraform-ubuntu1404-template bootdisk:virtio0
	// virtio0:ProxmoxxxxISCSI:vm-1014-disk-2,size=4G
	// description:Base image
	// cores:2 ostype:l26

	if vmr != nil {
		if vmr.node != "" {
			nodeCopy := vmr.node
			config.Node = &nodeCopy
		}
		if vmr.pool != "" {
			poolCopy := PoolName(vmr.pool)
			config.Pool = &poolCopy
		}
		if vmr.vmId != 0 {
			idCopy := vmr.vmId
			config.ID = &idCopy
		}
	}

	if _, isSet := params["args"]; isSet {
		config.Args = strings.TrimSpace(params["args"].(string))
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
	//Can be network,disk,cpu,memory,usb
	if _, isSet := params["hotplug"]; isSet {
		config.Hotplug = params["hotplug"].(string)
	}
	if _, isSet := params["hookscript"]; isSet {
		config.Hookscript = params["hookscript"].(string)
	}
	if _, isSet := params["machine"]; isSet {
		config.Machine = params["machine"].(string)
	}
	if itemValue, isSet := params["tpmstate0"]; isSet {
		config.TPM = TpmState{}.mapToSDK(itemValue.(string))
	}
	if _, isSet := params["kvm"]; isSet {
		config.QemuKVM = util.Pointer(Itob(int(params["kvm"].(float64))))
	}
	if _, isSet := params["ostype"]; isSet {
		config.QemuOs = params["ostype"].(string)
	}
	if _, isSet := params["scsihw"]; isSet {
		config.Scsihw = params["scsihw"].(string)
	}
	if _, isSet := params["smbios1"]; isSet {
		config.Smbios1 = params["smbios1"].(string)
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
				return fmt.Errorf("unable to parse unused disk id from input string '%v' tried to convert '%v' to integer", unusedDiskName, diskID)
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

	return nil
}

// Deprecated: use QemuGuestInterface.Update() instead.
func (config ConfigQemu) Update(ctx context.Context, rebootIfNeeded bool, vmr *VmRef, client *Client) (rebootRequired bool, err error) {
	if vmr == nil {
		vmr = &VmRef{}
	}
	err = client.New().QemuGuest.Update(ctx, *vmr, rebootIfNeeded, false, config)
	return
}

func (config ConfigQemu) updateNoCheck(
	ctx context.Context, c *clientAPI, version Version, client *Client,
	allowRestart bool, vmr *VmRef,
	currentLegacy *ConfigQemu, updateConfig configQemuUpdate,
) (rebootRequired bool, err error) {
	// TODO add digest during update to check if the config has changed

	ca := client.new().apiGet()

	urlPart := "/" + vmr.vmType.String() + "/" + vmr.vmId.String() + "/config"
	deleteBuilder := &strings.Builder{} // this is for items that should be removed before they can be created again e.g. cloud-init disks. (convert to array when needed)

	var currentState *PowerState
	var desiredState *PowerState
	if config.State != nil {
		desiredState = config.State
		var raw *rawGuestStatus
		raw, err = vmr.getRawGuestStatus_Unsafe(ctx, client)
		if err != nil {
			return false, err
		}
		tmpState := raw.GetState()
		if *desiredState == PowerStateStopped && *desiredState != tmpState {
			if err = vmr.shutdown_Unsafe(ctx, c); err != nil {
				return false, err
			}
		}
		currentState = &tmpState
	}

	var markedDisks qemuUpdateChanges
	if config.Disks != nil && currentLegacy.Disks != nil {
		markedDisks = *config.Disks.markDiskChanges(*currentLegacy.Disks)
		for _, e := range markedDisks.Move { // move disk to different storage or change disk format
			_, err = e.move(ctx, true, vmr, client)
			if err != nil {
				return
			}
		}
		if err = resizeDisks(ctx, vmr, client, markedDisks.Resize); err != nil { // increase Disks in size
			return false, err
		}
		config.Disks.cloudInitRemove(*currentLegacy.Disks, deleteBuilder)
	}

	if config.TPM != nil && currentLegacy.TPM != nil { // delete or move TPM
		if disk := config.TPM.markChanges(*currentLegacy.TPM, deleteBuilder); disk != nil { // move
			if _, err := disk.move(ctx, true, vmr, client); err != nil {
				return false, err
			}
		}
	}

	var pending bool

	if config.EfiDisk != nil {
		updateConfig.efiDisk = updateConfig.raw.GetEfiDisk()
		if updateConfig.efiDisk != nil {
			if disk := config.EfiDisk.markChangesUnsafe(updateConfig.efiDisk); disk != nil {
				if currentState == nil || *currentState != PowerStateStopped {
					desiredState, err = shutdownIfRunning(ctx, c, client, vmr, desiredState, allowRestart)
					if err != nil {
						return false, err
					}
					currentState = util.Pointer(PowerStateStopped)
				}
				if _, err := disk.move(ctx, true, vmr, client); err != nil {
					return false, err
				}
			}
		}
	}

	if deleteBuilder.Len() > 0 {
		pending = true
		itemsToDeleteBeforeUpdate := deleteBuilder.String()[len(comma):] // remove leading comma
		cl := client.new().apiRaw()
		if err := cl.putRawRetry(ctx, "/nodes/"+vmr.node.String()+urlPart, util.Pointer([]byte("delete="+itemsToDeleteBeforeUpdate)), 3); err != nil {
			return false, fmt.Errorf("error updating VM: %v", err)
		}
	}

	// Deleting items can create pending changes.
	// Moving disks changes the disk id. we need to get the config again if any disk was moved.
	if pending || len(markedDisks.Move) != 0 {
		var rawConfig map[string]any
		rawConfig, rebootRequired, err = vmr.pendingConfig(ctx, ca)
		if err != nil {
			return
		}
		currentLegacy, err = (&rawConfigQemu{a: rawConfig}).get(*vmr)
		if err != nil {
			return
		}
		if rebootRequired { // shutdown vm if reboot is required
			desiredState, err = shutdownIfRunning(ctx, c, client, vmr, desiredState, allowRestart)
			if err != nil {
				return false, err
			}
			rebootRequired = false
		}
	}

	if config.Node != nil && currentLegacy.Node != nil && *config.Node != *currentLegacy.Node { // Migrate VM
		if err = vmr.migrate_Unsafe(ctx, client, *config.Node, true); err != nil {
			return
		}

		// Set node to the node the VM was migrated to
		vmr.node = *config.Node

		// After migration, we must wait for the lock to be released.
		err = waitForMigrationLockRelease(ctx, client, vmr)
		if err != nil {
			return
		}
	}

	versionEncoded := version.Encode()
	params, body := config.mapToApiUpdate(currentLegacy, updateConfig, version)
	body = combineParamsAndBody(params, body)
	if body != nil {
		if err = c.putRawRetry(ctx, "/nodes/"+vmr.node.String()+urlPart, body, 3); err != nil {
			return false, fmt.Errorf("error updating VM: %v", err)
		}
		pending = true
	}

	if (currentState == nil || *currentState != PowerStateStopped) && pending { // Only check if we arent sure the guest is turned off, or it's not off.
		rebootRequired, err = vmr.pendingChanges(ctx, ca)
		if err != nil {
			return
		}
	}

	if err = resizeNewDisks(ctx, vmr, client, config.Disks, currentLegacy.Disks); err != nil {
		return
	}

	if rebootRequired {
		if allowRestart {
			return true, errors.New(ConfigQemu_Error_UnableToUpdateWithoutReboot)
		}
		if err = vmr.reboot_Unsafe(ctx, c); err != nil {
			return true, err
		}
	}

	if config.Pool != nil { // update pool membership
		if err = vmr.vmId.setPool(ctx, client.new().apiRaw(), *config.Pool, currentLegacy.Pool, versionEncoded); err != nil {
			return
		}
	}

	_, err = client.UpdateVMHA(ctx, vmr, config.HaState, config.HaGroup)

	if desiredState != nil && *desiredState == PowerStateRunning {
		if currentState == nil {
			var raw *rawGuestStatus
			raw, err = vmr.getRawGuestStatus_Unsafe(ctx, client)
			if err != nil {
				return false, nil
			}
			currentState = util.Pointer(raw.GetState())
		}
		if *desiredState != *currentState {
			if err = vmr.start_Unsafe(ctx, c); err != nil {
				return false, err
			}
		}
	}

	return
}

// returns the desired state as the current when no state was provided, so we can revert back to it later.
func shutdownIfRunning(ctx context.Context, c *clientAPI, client *Client, vmr *VmRef, desired *PowerState, allowRestart bool) (*PowerState, error) {
	raw, err := vmr.getRawGuestStatus_Unsafe(ctx, client)
	if err != nil {
		return nil, err
	}
	current := raw.GetState()
	if current != PowerStateStopped {
		if !allowRestart {
			return nil, errors.New(ConfigQemu_Error_UnableToUpdateWithoutReboot)
		}
		if desired == nil {
			desired = &current
		}
		if err = vmr.shutdown_Unsafe(ctx, c); err != nil {
			return nil, err
		}
	}
	return desired, nil
}

func (config *ConfigQemu) setVmr(vmr *VmRef) (err error) {
	if config == nil {
		return errors.New("config may not be nil")
	}
	if err = vmr.nilCheck(); err != nil {
		return
	}
	vmr.SetVmType(GuestQemu)
	idCopy := vmr.vmId
	config.ID = &idCopy
	return
}

func (config ConfigQemu) Validate(current *ConfigQemu, version Version) (err error) {
	// TODO test all other use cases
	// TODO has no context about changes caused by updating the vm
	if current == nil { // Create
		if config.ID != nil {
			if err = config.ID.Validate(); err != nil {
				return
			}
		}
		if config.Node == nil {
			return errors.New(ConfigQemu_Error_NodeRequired)
		}
		if err = config.Node.Validate(); err != nil {
			return
		}
		if config.CPU == nil {
			return errors.New(ConfigQemu_Error_CpuRequired)
		} else {
			if err = config.CPU.Validate(nil, version); err != nil {
				return
			}
		}
		if config.Memory == nil {
			return errors.New(ConfigQemu_Error_MemoryRequired)
		} else {
			if err = config.Memory.Validate(nil); err != nil {
				return
			}
		}
		if config.Networks != nil {
			if err = config.Networks.Validate(nil); err != nil {
				return
			}
		}
		if config.PciDevices != nil {
			if err = config.PciDevices.Validate(nil); err != nil {
				return
			}
		}
		if config.RandomnessDevice != nil {
			if err = config.RandomnessDevice.validateCreate(); err != nil {
				return
			}
		}
		if config.TPM != nil {
			if err = config.TPM.Validate(nil); err != nil {
				return
			}
		}
		if config.USBs != nil {
			if err = config.USBs.Validate(nil); err != nil {
				return
			}
		}
		if err = config.validateCreate(); err != nil {
			return
		}
	} else { // Update
		if config.Node != nil {
			if err = config.Node.Validate(); err != nil {
				return
			}
		}
		if config.CPU != nil {
			if err = config.CPU.Validate(current.CPU, version); err != nil {
				return
			}
		}
		if config.Memory != nil {
			if err = config.Memory.Validate(current.Memory); err != nil {
				return
			}
		}
		if config.Networks != nil {
			if err = config.Networks.Validate(current.Networks); err != nil {
				return
			}
		}
		if config.PciDevices != nil {
			if err = config.PciDevices.Validate(current.PciDevices); err != nil {
				return
			}
		}
		if config.RandomnessDevice != nil {
			if err = config.RandomnessDevice.Validate(current.RandomnessDevice); err != nil {
				return
			}
		}
		if config.TPM != nil {
			if err = config.TPM.Validate(current.TPM); err != nil {
				return
			}
		}
		if config.USBs != nil {
			if err = config.USBs.Validate(current.USBs); err != nil {
				return
			}
		}
		if err = config.validateUpdate(current); err != nil {
			return
		}
	}
	// Shared
	if config.Agent != nil {
		if err = config.Agent.Validate(); err != nil {
			return
		}
	}
	if config.CloudInit != nil {
		if err = config.CloudInit.Validate(version); err != nil {
			return
		}
	}
	if config.Disks != nil {
		err = config.Disks.Validate()
		if err != nil {
			return
		}
	}
	if config.Name != nil {
		if err = config.Name.Validate(); err != nil {
			return
		}
	}
	if config.Pool != nil && *config.Pool != "" {
		if err = config.Pool.Validate(); err != nil {
			return
		}
	}
	if len(config.Serials) > 0 {
		if err = config.Serials.Validate(); err != nil {
			return
		}
	}
	if config.Tags != nil {
		if err := (*config.Tags).Validate(); err != nil {
			return err
		}
	}
	return
}

func (config ConfigQemu) validateCreate() error {
	if config.EfiDisk != nil {
		if err := config.EfiDisk.validateCreate(); err != nil {
			return err
		}
	}
	return nil
}

func (config ConfigQemu) validateUpdate(current *ConfigQemu) error {
	if config.EfiDisk != nil {
		if current.EfiDisk != nil {
			if err := config.EfiDisk.validateUpdate(); err != nil {
				return err
			}
		} else {
			if err := config.EfiDisk.validateCreate(); err != nil {
				return err
			}
		}
	}
	return nil
}

/*
CloneVm
Example: request

nodes/proxmox1-xx/qemu/1012/clone

newid:145
name:tf-clone1
target:proxmox1-xx
full:1
storage:xxx
*/
// Deprecated: use VmRef.CloneQemu() instead
func (config ConfigQemu) CloneVm(ctx context.Context, sourceVmr *VmRef, vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType(GuestQemu)
	var storage string
	var format string
	fullClone := "1"
	if config.FullClone != nil {
		fullClone = strconv.Itoa(*config.FullClone)
	}
	if disk0Storage, ok := config.QemuDisks[0]["storage"].(string); ok && len(disk0Storage) > 0 {
		storage = disk0Storage
	}
	if disk0Format, ok := config.QemuDisks[0]["format"].(string); ok && len(disk0Format) > 0 {
		format = disk0Format
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

	if fullClone == "1" {
		if storage != "" {
			params["storage"] = storage
		}
		if format != "" {
			params["format"] = format
		}
	}

	_, err = client.CloneQemuVm(ctx, sourceVmr, params)
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
)

// Useful waiting for ISO install to complete
func WaitForShutdown(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	for ii := 0; ii < 100; ii++ {
		raw, err := vmr.GetRawGuestStatus(ctx, client)
		if err != nil {
			log.Print("Wait error:")
			log.Println(err)
		} else if raw.GetState() == PowerStateStopped {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("not shutdown within wait time")
}

// This is because proxmox create/config API won't let us make usernet devices
func SshForwardUsernet(ctx context.Context, vmr *VmRef, client *Client) (sshPort string, err error) {
	raw, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return "", err
	}
	if raw.GetState() == PowerStateStopped {
		return "", fmt.Errorf("VM must be running first")
	}
	sshPort = strconv.Itoa(int(vmr.VmId()) + 22000)
	_, err = client.MonitorCmd(ctx, vmr, "netdev_add user,id=net1,hostfwd=tcp::"+sshPort+"-:22")
	if err != nil {
		return "", err
	}
	_, err = client.MonitorCmd(ctx, vmr, "device_add virtio-net-pci,id=net1,netdev=net1,addr=0x13")
	if err != nil {
		return "", err
	}
	return
}

// device_del net1
// netdev_del net1
func RemoveSshForwardUsernet(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	raw, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return err
	}
	if raw.GetState() == PowerStateStopped {
		return fmt.Errorf("VM must be running first")
	}
	_, err = client.MonitorCmd(ctx, vmr, "device_del net1")
	if err != nil {
		return err
	}
	_, err = client.MonitorCmd(ctx, vmr, "netdev_del net1")
	if err != nil {
		return err
	}
	return nil
}

func MaxVmId(ctx context.Context, client *Client) (max int, err error) {
	vms, err := client.GetResourceList(ctx, resourceListGuest)
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

func SendKeysString(ctx context.Context, vmr *VmRef, client *Client, keys string) (err error) {
	raw, err := vmr.GetRawGuestStatus(ctx, client)
	if err != nil {
		return err
	}
	if raw.GetState() == PowerStateStopped {
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
		_, err = client.MonitorCmd(ctx, vmr, "sendkey "+c)
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

		if size, ok := disk["size"]; ok && size != "" {
			diskConfParam = append(diskConfParam, fmt.Sprintf("size=%v", disk["size"]))
		}
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

func (p QemuDeviceParam) createDeviceParam(
	deviceConfMap QemuDevice,
	ignoredKeys []string,
) QemuDeviceParam {

	for key, value := range deviceConfMap {
		if ignored := slices.Contains(ignoredKeys, key); !ignored {
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

func (c ConfigQemu) String() string { // String is for fmt.Stringer.
	jsConf, _ := json.Marshal(c)
	return string(jsConf)
}

// We have some special properties as they would have to be decoded more than once.
type configQemuUpdate struct {
	efiDisk *EfiDisk
	raw     *rawConfigQemu
}

type RawConfigQemu interface {
	Get(vmr VmRef) (*ConfigQemu, error)
	GetAgent() *QemuGuestAgent
	GetCPU() *QemuCPU
	GetCloudInit() *CloudInit
	GetDescription() string
	GetEfiDisk() *EfiDisk
	GetMemory() *QemuMemory
	GetName() GuestName
	GetNetworks() QemuNetworkInterfaces
	GetPciDevices() QemuPciDevices
	GetProtection() bool
	GetRandomnessDevice() *VirtIoRNG
	GetSerials() SerialInterfaces
	GetStartAtNodeBoot() bool
	GetStartupShutdown() *StartupAndShutdown
	GetTablet() bool
	GetTags() Tags
	GetUSBs() QemuUSBs
}

type rawConfigQemu struct{ a map[string]any }

func (raw *rawConfigQemu) Get(vmr VmRef) (*ConfigQemu, error) {
	config, err := raw.get(vmr)
	if err != nil {
		return nil, err
	}
	config.defaults()
	return config, nil
}

func (raw *rawConfigQemu) get(vmr VmRef) (*ConfigQemu, error) {
	config := ConfigQemu{
		Agent:            raw.GetAgent(),
		CPU:              raw.GetCPU(),
		CloudInit:        raw.GetCloudInit(),
		Description:      util.Pointer(raw.GetDescription()),
		EfiDisk:          raw.GetEfiDisk(),
		HaGroup:          vmr.HaGroup(),
		HaState:          vmr.HaState(),
		Memory:           raw.GetMemory(),
		Name:             util.Pointer(raw.GetName()),
		Networks:         raw.GetNetworks(),
		PciDevices:       raw.GetPciDevices(),
		Protection:       util.Pointer(raw.GetProtection()),
		RandomnessDevice: raw.GetRandomnessDevice(),
		Serials:          raw.GetSerials(),
		StartAtNodeBoot:  util.Pointer(raw.GetStartAtNodeBoot()),
		StartupShutdown:  raw.GetStartupShutdown(),
		Tablet:           util.Pointer(raw.GetTablet()),
		Tags:             new(raw.GetTags()),
		USBs:             raw.GetUSBs(),
	}
	config.Disks, config.LinkedID = raw.GetDisks()
	if err := config.mapToStruct(&vmr, raw.a); err != nil {
		return nil, err
	}
	return &config, nil
}

func (raw *rawConfigQemu) GetDescription() string {
	if v, isSet := raw.a[qemuApiKeyDescription]; isSet {
		return v.(string)
	}
	return ""
}

func (raw *rawConfigQemu) GetName() GuestName {
	if v, isSet := raw.a[qemuApiKeyName]; isSet {
		return GuestName(v.(string))
	}
	return ""
}

func (raw *rawConfigQemu) GetProtection() bool {
	if v, isSet := raw.a[qemuApiKeyProtection]; isSet {
		return int(v.(float64)) == 1
	}
	return false
}

func (raw *rawConfigQemu) GetStartAtNodeBoot() bool { return startAtNodeBootMapToSDK(raw.a) }

func (raw *rawConfigQemu) GetStartupShutdown() *StartupAndShutdown {
	return StartupAndShutdown{}.mapToSDK(raw.a)
}

func (raw *rawConfigQemu) GetTablet() bool {
	if v, isSet := raw.a[qemuApiKeyTablet]; isSet {
		return int(v.(float64)) == 1
	}
	return true
}

func (raw *rawConfigQemu) GetTags() Tags {
	var t Tags
	if v, isSet := raw.a[qemuApiKeyTags]; isSet {
		t.mapToSDK(v.(string))
	}
	return t
}

const (
	qemuApiKeyCloudInitCustom   = "cicustom"
	qemuApiKeyCloudInitPassword = "cipassword"
	qemuApiKeyCloudInitSshKeys  = "sshkeys"
	qemuApiKeyCloudInitUpgrade  = "ciupgrade"
	qemuApiKeyCloudInitUser     = "ciuser"
	qemuApiKeyCpuAffinity       = "affinity"
	qemuApiKeyCpuCores          = "cores"
	qemuApiKeyCpuLimit          = "cpulimit"
	qemuApiKeyCpuNuma           = "numa"
	qemuApiKeyCpuSockets        = "sockets"
	qemuApiKeyCpuType           = "cpu"
	qemuApiKeyCpuUnits          = "cpuunits"
	qemuApiKeyCpuVirtual        = "vcpus"
	qemuApiKeyDescription       = "description"
	qemuApiKeyEfiDisk           = "efidisk0"
	qemuApiKeyGuestAgent        = "agent"
	qemuApiKeyMemoryBallooning  = "balloon"
	qemuApiKeyMemoryCapacity    = "memory"
	qemuApiKeyMemoryShares      = "shares"
	qemuApiKeyName              = "name"
	qemuApiKeyProtection        = "protection"
	qemuApiKeyRandomnessDevice  = "rng0"
	qemuApiKeyTablet            = "tablet"
	qemuApiKeyTags              = "tags"
	qemuPrefixApiKeyDiskIde     = "ide"
	qemuPrefixApiKeyDiskSCSI    = "scsi"
	qemuPrefixApiKeyDiskSata    = "sata"
	qemuPrefixApiKeyDiskVirtIO  = "virtio"
	qemuPrefixApiKeyNetwork     = "net"
	qemuPrefixApiKeyPCI         = "hostpci"
	qemuPrefixApiKeySerial      = "serial"
	qemuPrefixApiKeyUSB         = "usb"
)

// NewRawConfigQemuFromApi returns the configuration of the Qemu guest.
// Including pending changes.
func NewRawConfigQemuFromApi(ctx context.Context, vmr *VmRef, client *Client) (RawConfigQemu, error) {
	if client == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return nil, err
	}
	return client.new().guestGetQemuRawConfig(ctx, vmr)
}

func guestGetRawQemuConfig_Unsafe(ctx context.Context, vmr *VmRef, c clientApiInterface) (*rawConfigQemu, error) {
	rawConfig, err := c.getGuestConfig(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return &rawConfigQemu{a: rawConfig}, nil
}

func (c *clientNewTest) guestGetQemuRawConfig(ctx context.Context, vmr *VmRef) (*rawConfigQemu, error) {
	return guestGetRawQemuConfig_Unsafe(ctx, vmr, c.api)
}

// NewActiveRawConfigQemuFromApi returns the active configuration of the Qemu guest.
// Without pending changes.
func NewActiveRawConfigQemuFromApi(ctx context.Context, vmr *VmRef, c *Client) (raw RawConfigQemu, pending bool, err error) {
	return c.new().guestGetQemuActiveRawConfig(ctx, vmr)
}

func guestGetActiveRawQemuConfig_Unsafe(ctx context.Context, vmr *VmRef, c clientApiInterface) (raw *rawConfigQemu, pending bool, err error) {
	var tmpConfig map[string]any
	tmpConfig, pending, err = vmr.pendingConfig(ctx, c)
	if err != nil {
		return nil, false, err
	}
	return &rawConfigQemu{a: tmpConfig}, pending, nil
}

func (c *clientNewTest) guestGetQemuActiveRawConfig(ctx context.Context, vmr *VmRef) (raw *rawConfigQemu, pending bool, err error) {
	return guestGetActiveRawQemuConfig_Unsafe(ctx, vmr, c.api)
}

func NewConfigQemuFromApi(ctx context.Context, vmr *VmRef, client *Client) (*ConfigQemu, error) {
	raw, err := guestGetQemuConfig(ctx, vmr, client)
	if err != nil {
		return nil, err
	}
	return raw.Get(*vmr)
}

func guestGetQemuConfig(ctx context.Context, vmr *VmRef, client *Client) (raw *rawConfigQemu, err error) {
	var vmInfo map[string]interface{}
	for ii := 0; ii < 3; ii++ {
		raw, err = client.new().guestGetQemuRawConfig(ctx, vmr)
		if err != nil {
			return nil, err
		}
		// TODO: this is a workaround for the issue that GetVmConfig will not always return the guest info
		vmInfo, err = client.GetVmInfo(ctx, vmr)
		if err != nil {
			return nil, err
		}
		// this can happen:
		// {"data":{"lock":"clone","digest":"eb54fb9d9f120ba0c3bdf694f73b10002c375c38","description":" qmclone temporary file\n"}})
		if vmInfo["lock"] == nil {
			break
		} else {
			time.Sleep(8 * time.Second)
		}
	}

	if vmInfo["lock"] != nil {
		return nil, fmt.Errorf("vm locked, could not obtain config")
	}
	if v, isSet := vmInfo["pool"]; isSet { // TODO: this is a workaround for the issue that GetVmConfig will not always return the guest info
		vmr.pool = PoolName(v.(string))
	}

	// HAstate is return by the api for a vm resource type but not the HAgroup
	err = client.ReadVMHA(ctx, vmr) // TODO: can be optimized, uses same API call as GetVmConfig and GetVmInfo
	if apiErr, ok := err.(*ApiError); ok {
		if strings.HasPrefix(apiErr.Message, "no such resource") {
			err = nil
		}
	}
	return
}

func waitForMigrationLockRelease(ctx context.Context, client *Client, vmr *VmRef) error {
	// Give a little time for the lock to potentially appear before we start polling.
	time.Sleep(5 * time.Second)

	timeout := time.After(10 * time.Minute) // 10 minutes, as large migrations can take a while.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	targetNode := vmr.Node()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timed out waiting for VM %d to unlock on node %s after migration", vmr.VmId(), targetNode)
		case <-ticker.C:
			// We need to create a new VmRef for GetVmInfo as it updates the node info in the ref.
			currentStatusVmr := NewVmRef(vmr.VmId())
			vmInfo, err := client.GetVmInfo(ctx, currentStatusVmr)
			if err != nil {
				log.Printf("[DEBUG] Waiting for VM %d to be available after migration, current error: %v", vmr.VmId(), err)
				continue
			}

			// Check if the VM is on the target node yet.
			if currentStatusVmr.Node() != targetNode {
				log.Printf("[DEBUG] Waiting for VM %d to appear on node %s, currently on %s", vmr.VmId(), targetNode, currentStatusVmr.Node())
				continue
			}

			// Once on the target node, check for a lock.
			if lock, ok := vmInfo["lock"]; !ok {
				log.Printf("[DEBUG] VM %d is on node %s and has no lock.", vmr.VmId(), targetNode)
				return nil // Success!
			} else {
				log.Printf("[DEBUG] Waiting for VM %d on node %s to unlock. Current lock: %v", vmr.VmId(), targetNode, lock)
			}
		}
	}
}
