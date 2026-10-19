package proxmox

import (
	"strconv"
)

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
		Order:           new(GuestStartupOrderAny),
		ShutdownTimeout: new(TimeDurationDefault),
		StartupDelay:    new(TimeDurationDefault)}
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

// Negative value means any order
type GuestStartupOrder int

const GuestStartupOrderAny GuestStartupOrder = -1

func (o GuestStartupOrder) String() string { return strconv.Itoa(int(o)) } // String is for fmt.Stringer.

// Negative value means default
type TimeDuration int

const TimeDurationDefault TimeDuration = -1

func (d TimeDuration) String() string { return strconv.Itoa(int(d)) } // String is for fmt.Stringer.
