package proxmox

import (
	"errors"
	"strings"
)

type Watchdog struct {
	Model  *WatchdogModel  `json:"model,omitempty"`  // Never nil when returned
	Action *WatchdogAction `json:"action,omitempty"` // Never nil when returned
	Delete bool
}

const (
	Watchdog_Error_ActionRequired = "watchdog action is required during create"
	Watchdog_Error_ModelRequired  = "watchdog model is required during create"
)

func (config Watchdog) mapToApi(b *strings.Builder) {
	if config.Model != nil {
		b.WriteString("model" + equal)
		b.WriteString(config.Model.String())
	}
	if config.Action != nil {
		b.WriteString(comma + "action" + equal)
		b.WriteString(config.Action.String())
	}
}

func (config Watchdog) mapToApiCreate(b *strings.Builder) {
	b.WriteString("&" + qemuApiKeyWatchdog + "=")
	config.mapToApi(b)
}

func (config Watchdog) mapToApiUpdate(current *Watchdog, builder, delete *strings.Builder) {
	if config.Delete {
		delete.WriteString("," + qemuApiKeyWatchdog)
		return
	}
	b := strings.Builder{}
	current.mapToApi(&b)
	currentStr := b.String()
	if config.Action != nil {
		current.Action = config.Action
	}
	if config.Model != nil {
		current.Model = config.Model
	}
	b = strings.Builder{}
	current.mapToApi(&b)
	configStr := b.String()
	if currentStr != configStr {
		builder.WriteString("&" + qemuApiKeyWatchdog + "=")
		builder.WriteString(configStr)
	}
}

func (config Watchdog) Validate(current *Watchdog) error {
	if current == nil { // create
		return config.validateCreate()
	}
	return config.validateUpdate()
}

func (config Watchdog) validateCreate() error {
	if config.Delete {
		return nil
	}
	if config.Model == nil {
		return errors.New(Watchdog_Error_ModelRequired)
	}
	if config.Action == nil {
		return errors.New(Watchdog_Error_ActionRequired)
	}
	if err := config.Model.Validate(); err != nil {
		return err
	}
	return config.Action.Validate()
}

func (config Watchdog) validateUpdate() error {
	if config.Delete {
		return nil
	}
	if config.Model != nil {
		if err := config.Model.Validate(); err != nil {
			return err
		}
	}
	if config.Action != nil {
		if err := config.Action.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (raw *rawConfigQemu) GetWatchdog() *Watchdog {
	if v, isSet := raw.a[qemuApiKeyWatchdog]; isSet {
		if v.(string) == " " {
			return nil
		}
		settings := splitStringOfSettings(v.(string))
		watchdog := Watchdog{
			Model:  new(WatchdogModel),
			Action: new(WatchdogAction),
		}
		if v, ok := settings["model"]; ok {
			*watchdog.Model = WatchdogModel(v)
		}
		if v, ok := settings["action"]; ok {
			*watchdog.Action = WatchdogAction(v)
		}
		return &watchdog
	}
	return nil

}

// Enum
//
//	const (
//		WatchdogModelI6300esb
//		WatchdogModelIb700
//	)
type WatchdogModel string

const (
	WatchdogModelI6300esb WatchdogModel = "i6300esb"
	WatchdogModelIb700    WatchdogModel = "ib700"
)

const WatchdogModel_Error = "invalid watchdog model"

func (model WatchdogModel) String() string { return string(model) } // String is for fmt.Stringer.

func (model WatchdogModel) Validate() error {
	switch model {
	case WatchdogModelI6300esb, WatchdogModelIb700:
		return nil
	}
	return errors.New(WatchdogModel_Error)
}

// Enum
//
//	const (
//		WatchdogActionDebug
//		WatchdogActionNone
//		WatchdogActionPause
//		WatchdogActionPoweroff
//		WatchdogActionReset
//		WatchdogActionShutdown
//	)
type WatchdogAction string

const (
	WatchdogActionDebug    WatchdogAction = "debug"
	WatchdogActionNone     WatchdogAction = "none"
	WatchdogActionPause    WatchdogAction = "pause"
	WatchdogActionPoweroff WatchdogAction = "poweroff"
	WatchdogActionReset    WatchdogAction = "reset"
	WatchdogActionShutdown WatchdogAction = "shutdown"
)

const WatchdogAction_Error = "invalid watchdog action"

func (action WatchdogAction) String() string { return string(action) } // String is for fmt.Stringer.

func (action WatchdogAction) Validate() error {
	switch action {
	case WatchdogActionDebug, WatchdogActionNone, WatchdogActionPause, WatchdogActionPoweroff, WatchdogActionReset, WatchdogActionShutdown:
		return nil
	}
	return errors.New(WatchdogAction_Error)
}
