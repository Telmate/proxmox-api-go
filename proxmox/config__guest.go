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
// Has the same syntax as a DNS name.
// Domain sections may not start or end with a hyphen (-) or dot (.).
// Valid characters are letters, numbers, hyphens (-) and dots (.).
// Regex: ^(?=.{1,127}$)(?:(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)\.)*(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]*[a-zA-Z0-9])?)$
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
	GuestID_Error_Maximum = "guestID should be less than 1000000000"
	GuestID_Error_Minimum = "guestID should be greater than 99"
	GuestIdMaximum        = 999999999
	GuestIdMinimum        = 100
)

// DeleteHaResource deletes an HA resource. Returns false if the resource does not exist and was not deleted, true if successfully deleted.
func (id GuestID) DeleteHaResource(ctx context.Context, c *Client) (bool, error) {
	err := id.Validate()
	if err != nil {
		return false, err
	}
	return c.new().haDeleteResource(ctx, id)
}

func (c *clientNew) haDeleteResource(ctx context.Context, id GuestID) (bool, error) {
	return id.deleteHaResource(ctx, c.apiGet())
}

func (id GuestID) deleteHaResource(ctx context.Context, c clientApiInterface) (bool, error) {
	err := c.deleteHaResource(ctx, id)
	if err != nil {
		if strings.HasPrefix(err.Error(), "500 cannot delete service") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (id GuestID) errorContext() string {
	return "ID " + id.String()
}

func (id GuestID) Exists(ctx context.Context, c *Client) (bool, error) {
	if err := c.checkInitialized(); err != nil {
		return false, err
	}
	if err := id.Validate(); err != nil {
		return false, err
	}
	return id.exists_Unsafe(ctx, c)
}

func (id GuestID) ExistsNoCheck(ctx context.Context, c *Client) (bool, error) {
	if err := c.checkInitialized(); err != nil {
		return false, err
	}
	return id.exists_Unsafe(ctx, c)
}

func (id GuestID) exists_Unsafe(ctx context.Context, c *Client) (bool, error) {
	_, err := c.GetItemConfigString(ctx, url_NextID+"?vmid="+id.String(), "API", "cluster/nextid",
		func(err error) bool {
			return err.Error() == guestIDexistsError
		})
	if err != nil {
		if err.Error() == guestIDexistsError {
			return true, nil
		}
	}
	return false, err
}

func (id GuestID) String() string { return strconv.Itoa(int(id)) } // String is for fmt.Stringer.

func (id GuestID) Validate() error {
	if id < GuestIdMinimum {
		return errors.New(GuestID_Error_Minimum)
	}
	if id > GuestIdMaximum {
		return errors.New(GuestID_Error_Maximum)
	}
	return nil
}

type StartupAndShutdown struct {
	Order           *GuestStartupOrder `json:"order,omitempty"`            // Never nil when returned.
	ShutdownTimeout *TimeDuration      `json:"shutdown_timeout,omitempty"` // Number of seconds of the shutdown timeout. Never nil when returned.
	StartupDelay    *TimeDuration      `json:"startup_delay,omitempty"`    // Number of seconds of the startup delay. Never nil when returned.
}

func startAtNodeBootMapToSDK(params map[string]any) bool {
	if v, isSet := params[guestApiKeyOnBoot]; isSet {
		return int(v.(float64)) == 1
	}
	return false
}

func startAtNodeBootMapToApiCreate(params map[string]any, startup bool) {
	if startup {
		params[guestApiKeyOnBoot] = 1
	}
}

func startAtNodeBootMapToApiUpdate(params map[string]any, startup bool, current bool) (delete string) {
	if startup == current {
		return
	}
	if startup {
		params[guestApiKeyOnBoot] = 1
		return
	}
	return "," + guestApiKeyOnBoot
}

const (
	guestApiKeyOnBoot  string = "onboot"
	guestApiKeyStartup string = "startup"
)

func (config StartupAndShutdown) combine(current *StartupAndShutdown) StartupAndShutdown {
	newConfig := StartupAndShutdown{
		Order:           current.Order,
		StartupDelay:    current.StartupDelay,
		ShutdownTimeout: current.ShutdownTimeout}
	if config.Order != nil {
		newConfig.Order = config.Order
	}
	if config.StartupDelay != nil {
		newConfig.StartupDelay = config.StartupDelay
	}
	if config.ShutdownTimeout != nil {
		newConfig.ShutdownTimeout = config.ShutdownTimeout
	}
	return newConfig
}

func (config StartupAndShutdown) mapToApiCreate(params map[string]any) {
	if v := config.mapToApiStartupCreate(); len(v) > 0 {
		params[guestApiKeyStartup] = v[1:] // remove leading comma
	}
}

func (config StartupAndShutdown) mapToApiStartupCreate() (parts string) {
	if config.Order != nil && *config.Order > GuestStartupOrderAny {
		parts += ",order=" + config.Order.String()
	}
	if config.StartupDelay != nil && *config.StartupDelay > TimeDurationDefault {
		parts += ",up=" + config.StartupDelay.String()
	}
	if config.ShutdownTimeout != nil && *config.ShutdownTimeout > TimeDurationDefault {
		parts += ",down=" + config.ShutdownTimeout.String()
	}
	return
}

func (config StartupAndShutdown) mapToApiUpdate(current *StartupAndShutdown, params map[string]any) (delete string) {
	usedConfig := config.combine(current)
	rawStartup := usedConfig.mapToApiStartupCreate()
	if rawStartup != current.mapToApiStartupCreate() {
		if rawStartup != "" {
			params[guestApiKeyStartup] = rawStartup[1:] // remove leading comma
		} else {
			delete += "," + guestApiKeyStartup
		}
	}
	return
}

func (StartupAndShutdown) mapToSDK(params map[string]any) *StartupAndShutdown {
	config := StartupAndShutdown{
		Order:           util.Pointer(GuestStartupOrderAny),
		ShutdownTimeout: util.Pointer(TimeDurationDefault),
		StartupDelay:    util.Pointer(TimeDurationDefault)}
	if v, ok := params[guestApiKeyStartup]; ok {
		settings := splitStringOfSettings(v.(string))
		if vv, ok := settings["order"]; ok {
			order, _ := strconv.ParseUint(vv, 10, 64)
			*config.Order = GuestStartupOrder(order)
		}
		if vv, ok := settings["up"]; ok {
			up, _ := strconv.ParseUint(vv, 10, 64)
			*config.StartupDelay = TimeDuration(up)
		}
		if vv, ok := settings["down"]; ok {
			down, _ := strconv.ParseUint(vv, 10, 64)
			*config.ShutdownTimeout = TimeDuration(down)
		}
		return &config
	}
	return nil
}

// Negative value means default
type TimeDuration int

const TimeDurationDefault TimeDuration = -1

func (d TimeDuration) String() string { return strconv.Itoa(int(d)) } // String is for fmt.Stringer.

// Negative value means any order
type GuestStartupOrder int

const GuestStartupOrderAny GuestStartupOrder = -1

func (o GuestStartupOrder) String() string { return strconv.Itoa(int(o)) } // String is for fmt.Stringer.

// GuestType is an enum for the type of guest (lxc or qemu)
type GuestType uint8

const (
	guestUnknown GuestType = 0
	GuestLxc     GuestType = 1
	GuestQemu    GuestType = 2
)

const GuestType_Error_Invalid = "guest type should be one of (lxc, qemu)"

func (t *GuestType) Parse(guestType string) error {
	switch guestType {
	case "lxc":
		*t = GuestLxc
		return nil
	case "qemu":
		*t = GuestQemu
		return nil
	}
	return errors.New(GuestType_Error_Invalid)
}

func (t *GuestType) parse(guestType string) {
	switch guestType {
	case "lxc":
		*t = GuestLxc
	case "qemu":
		*t = GuestQemu
	}
}

func (t GuestType) String() string {
	switch t {
	case GuestLxc:
		return "lxc"
	case GuestQemu:
		return "qemu"
	}
	return "unknown"
}

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
	params, err := client.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/feature?feature=snapshot", "guest", "FEATURES")
	if err != nil {
		return false, err
	}
	return GuestFeature("").mapToStruct(params), nil
}

// Check if there are any pending changes that require a reboot to be applied.
// Same as VmRef.PendingChanges.
func GuestHasPendingChanges(ctx context.Context, vmr *VmRef, client *Client) (bool, error) {
	return vmr.PendingChanges(ctx, client)
}

// Reboot the specified guest
func GuestReboot(ctx context.Context, vmr *VmRef, client *Client) (err error) {
	_, err = client.RebootVm(ctx, vmr)
	return
}

func guestSetPoolNoCheck(ctx context.Context, c *Client, guestID GuestID, newPool PoolName, currentPool *PoolName, version EncodedVersion) (err error) {
	if newPool == "" {
		if currentPool != nil && *currentPool != "" { // leave pool
			if err = (*currentPool).removeGuestsNoCheck(ctx, c, []GuestID{guestID}, version); err != nil {
				return
			}
		}
	} else {
		if currentPool == nil || *currentPool == "" { // join pool
			if version < version_8_0_0 {
				if err = newPool.addGuestsNoCheckV7(ctx, c, []GuestID{guestID}); err != nil {
					return
				}
			} else {
				newPool.addGuestsNoCheckV8(ctx, c, []GuestID{guestID})
			}
		} else if newPool != *currentPool { // change pool
			if version < version_8_0_0 {
				if err = (*currentPool).removeGuestsNoCheck(ctx, c, []GuestID{guestID}, version); err != nil {
					return
				}
				if err = newPool.addGuestsNoCheckV7(ctx, c, []GuestID{guestID}); err != nil {
					return
				}
			} else {
				if err = newPool.addGuestsNoCheckV8(ctx, c, []GuestID{guestID}); err != nil {
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
	_, err = client.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/status/shutdown")
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

const guest_ApiError_AlreadyExists string = "config file already exists"

// Keep trying to create/clone a VM until we get a unique ID
func guestCreateLoop_Unsafe(ctx context.Context, idKey, url string, params map[string]any, c *Client) (GuestID, error) {
	c.guestCreationMutex.Lock()
	defer c.guestCreationMutex.Unlock()
	for {
		guestID, err := c.getNextID_Unsafe(ctx)
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
