package proxmox

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"regexp"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

// All code LXC and Qemu have in common should be placed here.

type GuestDNS struct {
	NameServers  *[]netip.Addr `json:"nameservers,omitempty"`
	SearchDomain *string       `json:"searchdomain,omitempty"` // we are not validating this field, as validating domain names is a complex topic.
}

func (config GuestDNS) mapToApiCreate(params map[string]any) {
	if config.NameServers != nil && len(*config.NameServers) > 0 {
		var nameservers string
		for _, ns := range *config.NameServers {
			nameservers += " " + ns.String()
		}
		params[guestApiKeyNameServer] = nameservers[1:]
	}
	if config.SearchDomain != nil && *config.SearchDomain != "" {
		params[guestApiKeySearchDomain] = *config.SearchDomain
	}
}

func (config GuestDNS) mapToApiUpdate(current GuestDNS, params map[string]any) (delete string) {
	if config.SearchDomain != nil {
		if *config.SearchDomain != "" {
			if current.SearchDomain == nil || *config.SearchDomain != *current.SearchDomain {
				params[guestApiKeySearchDomain] = *config.SearchDomain
			}
		} else if current.SearchDomain != nil {
			delete += "," + guestApiKeySearchDomain
		}
	}
	if config.NameServers != nil {
		if len(*config.NameServers) > 0 {
			var nameServers string
			for i := range *config.NameServers {
				nameServers += " " + (*config.NameServers)[i].String()
			}
			if current.NameServers != nil && len(*current.NameServers) > 0 {
				var currentNameServers string
				for i := range *current.NameServers {
					currentNameServers += " " + (*current.NameServers)[i].String()
				}
				if nameServers == currentNameServers {
					return
				}
			}
			params[guestApiKeyNameServer] = nameServers[1:]
		} else if current.NameServers != nil {
			delete += "," + guestApiKeyNameServer
		}
	}
	return
}

func (GuestDNS) mapToSDK(params map[string]any) *GuestDNS {
	var dnsSet bool
	var nameservers []netip.Addr
	if v, isSet := params[guestApiKeyNameServer]; isSet {
		tmp := strings.Split(v.(string), " ")
		nameservers = make([]netip.Addr, len(tmp))
		for i, e := range tmp {
			nameservers[i], _ = netip.ParseAddr(e)
		}
		dnsSet = true
	}
	var domain string
	if v, isSet := params[guestApiKeySearchDomain]; isSet {
		if len(v.(string)) > 1 {
			domain = v.(string)
			dnsSet = true
		}
	}
	if !dnsSet {
		return nil
	}
	return &GuestDNS{
		SearchDomain: &domain,
		NameServers:  &nameservers}
}

const (
	guestApiKeyNameServer   string = "nameserver"
	guestApiKeySearchDomain string = "searchdomain"
)

// GuestName has a maximum length of 128 characters.
// First character must be a letter or number, the rest can be letters, numbers or hyphens.
// Regex: ^([a-z]|[A-Z]|[0-9])([a-z]|[A-Z]|[0-9]|-){127,}$
type GuestName string

const (
	guestNameMaxLength    = 128
	GuestNameErrorEmpty   = `name cannot be empty`
	GuestNameErrorInvalid = `name did not match the following regex '^(?=.{1,127}$)(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)\.)*(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)$'`
	GuestNameErrorLength  = `name has a maximum length of 128`
	GuestNameErrorStart   = `name cannot start with a hyphen (-) or dot (.)`
	GuestNameErrorEnd     = `name cannot end with a hyphen (-) or dot (.)`
)

var guestNameRegex = regexp.MustCompile(`^(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)\.)*(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)$`)

func (name GuestName) String() string { return string(name) } // String is for fmt.Stringer.

func (name GuestName) Validate() error {
	if len(name) == 0 {
		return errors.New(GuestNameErrorEmpty)
	}
	if len(name) > guestNameMaxLength {
		return errors.New(GuestNameErrorLength)
	}
	switch name[0] {
	case '-', '.':
		return errors.New(GuestNameErrorStart)
	}
	switch name[len(name)-1] {
	case '-', '.':
		return errors.New(GuestNameErrorEnd)
	}
	if !guestNameRegex.MatchString(string(name)) {
		return errors.New(GuestNameErrorInvalid)
	}
	return nil
}

// 0 to 10240000, where 0 means no limit
type GuestNetworkRate uint32

const (
	GuestNetworkRate_Error_Invalid = "network rate must be in the range 0 to 10240000"
	GuestNetworkRateMaximum        = 10240000
	GuestNetworkRateUnlimited      = GuestNetworkRate(0)
)

// unsafe requires caller to check for nil
func (rate GuestNetworkRate) mapToAPI() string {
	if rate == GuestNetworkRateUnlimited {
		return ""
	}
	rawRate := strconv.Itoa(int(rate))
	length := len(rawRate)
	if length > 3 {
		// Insert a decimal point three places from the end
		if rate%1000 == 0 {
			return ",rate=" + rawRate[:length-3]
		} else {
			return strings.TrimRight(",rate="+rawRate[:length-3]+"."+rawRate[length-3:], "0")
		}
	}
	// Prepend zeros to ensure decimal places
	prefixRate := "000" + rawRate
	return strings.TrimRight(",rate=0."+prefixRate[length:], "0")
}

func (GuestNetworkRate) mapToSDK(rawRate string) *GuestNetworkRate {
	splitRate := strings.Split(rawRate, ".")
	var rate int
	switch len(splitRate) {
	case 1:
		if splitRate[0] != "0" {
			rate, _ = strconv.Atoi(splitRate[0] + "000")
		}
	case 2:
		// Pad the fractional part to ensure it has at least 3 digits
		fractional := splitRate[1] + "000"
		rate, _ = strconv.Atoi(splitRate[0] + fractional[:3])
	}
	return util.Pointer(GuestNetworkRate(rate))
}

func (rate GuestNetworkRate) Validate() error {
	if rate > GuestNetworkRateMaximum {
		return errors.New(GuestNetworkRate_Error_Invalid)
	}
	return nil
}

type GuestResource struct {
	CpuCores           uint       `json:"cpu_cores"`
	CpuUsage           float64    `json:"cpu_usage"`
	DiskReadTotal      uint       `json:"disk_read"`
	DiskSizeInBytes    uint       `json:"disk_size"`
	DiskUsedInBytes    uint       `json:"disk_used"`
	DiskWriteTotal     uint       `json:"disk_write"`
	HaState            string     `json:"hastate"` // TODO custom type?
	Id                 uint       `json:"id"`
	MemoryTotalInBytes uint       `json:"memory_total"`
	MemoryUsedInBytes  uint       `json:"memory_used"`
	Name               GuestName  `json:"name"`
	NetworkIn          uint       `json:"network_in"`
	NetworkOut         uint       `json:"network_out"`
	Node               string     `json:"node"` // TODO custom type
	Pool               PoolName   `json:"pool"`
	Status             PowerState `json:"status"`
	Tags               Tags       `json:"tags"`
	Template           bool       `json:"template"`
	Type               GuestType  `json:"type"`
	UptimeInSeconds    uint       `json:"uptime"`
}

const (
	guestApiKeyName string = "name"
)

// https://pve.proxmox.com/pve-docs/api-viewer/#/cluster/resources
func (GuestResource) mapToStruct(params []interface{}) []GuestResource {
	if len(params) == 0 {
		return nil
	}
	resources := make([]GuestResource, len(params))
	for i := range params {
		tmpParams := params[i].(map[string]interface{})
		if _, isSet := tmpParams["maxcpu"]; isSet {
			resources[i].CpuCores = uint(tmpParams["maxcpu"].(float64))
		}
		if _, isSet := tmpParams["cpu"]; isSet {
			resources[i].CpuUsage = tmpParams["cpu"].(float64)
		}
		if _, isSet := tmpParams["diskread"]; isSet {
			resources[i].DiskReadTotal = uint(tmpParams["diskread"].(float64))
		}
		if _, isSet := tmpParams["maxdisk"]; isSet {
			resources[i].DiskSizeInBytes = uint(tmpParams["maxdisk"].(float64))
		}
		if _, isSet := tmpParams["disk"]; isSet {
			resources[i].DiskUsedInBytes = uint(tmpParams["disk"].(float64))
		}
		if _, isSet := tmpParams["diskwrite"]; isSet {
			resources[i].DiskWriteTotal = uint(tmpParams["diskwrite"].(float64))
		}
		if _, isSet := tmpParams["hastate"]; isSet {
			resources[i].HaState = tmpParams["hastate"].(string)
		}
		if _, isSet := tmpParams["vmid"]; isSet {
			resources[i].Id = uint(tmpParams["vmid"].(float64))
		}
		if _, isSet := tmpParams["maxmem"]; isSet {
			resources[i].MemoryTotalInBytes = uint(tmpParams["maxmem"].(float64))
		}
		if _, isSet := tmpParams["mem"]; isSet {
			resources[i].MemoryUsedInBytes = uint(tmpParams["mem"].(float64))
		}
		if v, isSet := tmpParams[guestApiKeyName]; isSet {
			resources[i].Name = GuestName(v.(string))
		}
		if _, isSet := tmpParams["netin"]; isSet {
			resources[i].NetworkIn = uint(tmpParams["netin"].(float64))
		}
		if _, isSet := tmpParams["netout"]; isSet {
			resources[i].NetworkOut = uint(tmpParams["netout"].(float64))
		}
		if _, isSet := tmpParams["node"]; isSet {
			resources[i].Node = tmpParams["node"].(string)
		}
		if _, isSet := tmpParams["status"]; isSet {
			resources[i].Status = PowerState(0).parse(tmpParams["status"].(string))
		}
		if _, isSet := tmpParams["tags"]; isSet {
			resources[i].Tags = Tags{}.mapToSDK(tmpParams["tags"].(string))
		}
		if _, isSet := tmpParams["template"]; isSet {
			resources[i].Template = Itob(int(tmpParams["template"].(float64)))
		}
		if _, isSet := tmpParams["type"]; isSet {
			resources[i].Type = GuestType(tmpParams["type"].(string))
		}
		if _, isSet := tmpParams["uptime"]; isSet {
			resources[i].UptimeInSeconds = uint(tmpParams["uptime"].(float64))
		}
	}
	return resources
}

// Enum
type GuestFeature string

const (
	GuestFeature_Clone    GuestFeature = "clone"
	GuestFeature_Copy     GuestFeature = "copy"
	GuestFeature_Snapshot GuestFeature = "snapshot"
)

func (GuestFeature) Error() error {
	return errors.New("value should be one of (" + string(GuestFeature_Clone) + " ," + string(GuestFeature_Copy) + " ," + string(GuestFeature_Snapshot) + ")")
}

func (GuestFeature) mapToStruct(params map[string]interface{}) bool {
	if value, isSet := params["hasFeature"]; isSet {
		return Itob(int(value.(float64)))
	}
	return false
}

func (feature GuestFeature) Validate() error {
	switch feature {
	case GuestFeature_Copy, GuestFeature_Clone, GuestFeature_Snapshot:
		return nil
	}
	return GuestFeature("").Error()
}

type GuestFeatures struct {
	Clone    bool `json:"clone"`
	Copy     bool `json:"copy"`
	Snapshot bool `json:"snapshot"`
}

// Positive number between 100 and 1000000000
type GuestID uint32

const (
	GuestID_Error_Maximum string  = "guestID should be less than 1000000000"
	GuestID_Error_Minimum string  = "guestID should be greater than 99"
	GuestIdMaximum        GuestID = 999999999
	GuestIdMinimum        GuestID = 100
)

func (id GuestID) String() string {
	return strconv.Itoa(int(id))
}

func (id GuestID) Validate() error {
	if id < GuestIdMinimum {
		return errors.New(GuestID_Error_Minimum)
	}
	if id > GuestIdMaximum {
		return errors.New(GuestID_Error_Maximum)
	}
	return nil
}

type GuestType string

const (
	GuestLXC  GuestType = "lxc"
	GuestQemu GuestType = "qemu"
)

// check if the guest has the specified feature.
func GuestHasFeature(ctx context.Context, vmr *VmRef, client *Client, feature GuestFeature) (bool, error) {
	err := feature.Validate()
	if err != nil {
		return false, err
	}
	err = client.CheckVmRef(ctx, vmr)
	if err != nil {
		return false, err
	}
	return guestHasFeature(ctx, vmr, client, feature)
}

func guestHasFeature(ctx context.Context, vmr *VmRef, client *Client, feature GuestFeature) (bool, error) {
	var params map[string]interface{}
	params, err := client.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/feature?feature=snapshot", "guest", "FEATURES")
	if err != nil {
		return false, err
	}
	return GuestFeature("").mapToStruct(params), nil
}

// Check if there are any pending changes that require a reboot to be applied.
func GuestHasPendingChanges(ctx context.Context, vmr *VmRef, client *Client) (bool, error) {
	params, err := pendingGuestConfigFromApi(ctx, vmr, client)
	if err != nil {
		return false, err
	}
	return keyExists(params, "pending") || keyExists(params, "delete"), nil
}

// Reboot the specified guest
func GuestReboot(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	_, err = client.RebootVm(ctx, vmr)
	return
}

func guestSetPoolNoCheck(ctx context.Context, c *Client, guestID uint, newPool PoolName, currentPool *PoolName, version Version) (err error) {
	if newPool == "" {
		if currentPool != nil && *currentPool != "" { // leave pool
			if err = (*currentPool).removeGuestsNoCheck(ctx, c, []uint{guestID}, version); err != nil {
				return
			}
		}
	} else {
		if currentPool == nil || *currentPool == "" { // join pool
			if version.Encode() < version_8_0_0 {
				if err = newPool.addGuestsNoCheckV7(ctx, c, []uint{guestID}); err != nil {
					return
				}
			} else {
				newPool.addGuestsNoCheckV8(ctx, c, []uint{guestID})
			}
		} else if newPool != *currentPool { // change pool
			if version.Encode() < version_8_0_0 {
				if err = (*currentPool).removeGuestsNoCheck(ctx, c, []uint{guestID}, version); err != nil {
					return
				}
				if err = newPool.addGuestsNoCheckV7(ctx, c, []uint{guestID}); err != nil {
					return
				}
			} else {
				if err = newPool.addGuestsNoCheckV8(ctx, c, []uint{guestID}); err != nil {
					return
				}
			}
		}
	}
	return
}

func GuestShutdown(ctx context.Context, vmr *VmRef, client *Client, force bool) (err error) {
	if err = client.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	var params map[string]interface{}
	if force {
		params = map[string]interface{}{"forceStop": force}
	}
	_, err = client.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/status/shutdown")
	return
}

func GuestStart(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	_, err = client.StartVm(ctx, vmr)
	return
}

// List all features the guest has.
func ListGuestFeatures(ctx context.Context, vmr *VmRef, client *Client) (features GuestFeatures, err error) {
	err = client.CheckVmRef(ctx, vmr)
	if err != nil {
		return
	}
	features.Clone, err = guestHasFeature(ctx, vmr, client, GuestFeature_Clone)
	if err != nil {
		return
	}
	features.Copy, err = guestHasFeature(ctx, vmr, client, GuestFeature_Copy)
	if err != nil {
		return
	}
	features.Snapshot, err = guestHasFeature(ctx, vmr, client, GuestFeature_Snapshot)
	return
}

// List all guest the user has viewing rights for in the cluster
func ListGuests(ctx context.Context, client *Client) ([]GuestResource, error) {
	list, err := client.GetResourceList(ctx, "vm")
	if err != nil {
		return nil, err
	}
	return GuestResource{}.mapToStruct(list), nil
}

func pendingGuestConfigFromApi(ctx context.Context, vmr *VmRef, client *Client) ([]interface{}, error) {
	if err := client.CheckVmRef(ctx, vmr); err != nil {
		return nil, err
	}
	return client.GetItemConfigInterfaceArray(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/pending", "Guest", "PENDING CONFIG")
}

const guest_ApiError_AlreadyExists string = "config file already exists"

// Keep trying to create/clone a VM until we get a unique ID
func guestCreateLoop(ctx context.Context, idKey, url string, params map[string]interface{}, c *Client) (GuestID, error) {
	c.guestCreationMutex.Lock()
	defer c.guestCreationMutex.Unlock()
	for {
		guestID, err := c.GetNextIdNoCheck(ctx, nil)
		if err != nil {
			return 0, err
		}
		params[idKey] = int(guestID)
		var exitStatus string
		exitStatus, err = c.PostWithTask(ctx, params, url)
		if err != nil {
			if !strings.Contains(err.Error(), guest_ApiError_AlreadyExists) {
				return 0, fmt.Errorf("error creating Guest: %v, error status: %s (params: %v)", err, exitStatus, params)
			}
		} else {
			return guestID, nil
		}
	}
}
