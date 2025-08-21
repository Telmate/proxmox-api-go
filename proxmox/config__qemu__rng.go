package proxmox

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type VirtIoRNG struct {
	Limit  *uint          `json:"limit,omitempty"`
	Period *time.Duration `json:"period,omitempty"`
	Source *EntropySource `json:"source,omitempty"` // Never nil when returned
	Delete bool           `json:"delete,omitempty"`
}

const (
	VirtIoRNGErrorSourceNotSet = "source must be set on creation"
)

const (
	qemuAPISettingRngLimit  = "max_bytes"
	qemuAPISettingRngPeriod = "period"
	qemuAPISettingRngSource = "source"
)

func (config VirtIoRNG) combine(current VirtIoRNG) VirtIoRNG {
	var newConfig VirtIoRNG
	if config.Source != nil {
		newConfig.Source = config.Source
	} else {
		newConfig.Source = current.Source
	}
	if config.Limit != nil {
		newConfig.Limit = config.Limit
	} else {
		newConfig.Limit = current.Limit
	}
	if config.Period != nil {
		newConfig.Period = config.Period
	} else {
		newConfig.Period = current.Period
	}
	return newConfig
}

func (config VirtIoRNG) mapToAPICreate(params map[string]any) {
	if !config.Delete {
		params[qemuApiKeyRandomnessDevice] = config.string()
	}
}

func (config VirtIoRNG) mapToAPIUpdateUnsafe(current *VirtIoRNG, params map[string]any) (delete string) {
	if config.Delete {
		return "," + qemuApiKeyRandomnessDevice
	}
	new := config.combine(*current).string()
	if new != current.string() {
		params[qemuApiKeyRandomnessDevice] = new
	}
	return ""
}

func (config VirtIoRNG) Validate(current *VirtIoRNG) error {
	if current != nil {
		return config.validateUpdate()
	}
	return config.validateCreate()
}

func (config VirtIoRNG) validateCreate() error {
	if config.Delete {
		return nil // Deletion is always valid, as it doesn't do anything during creation.
	}
	if config.Source == nil {
		return errors.New(VirtIoRNGErrorSourceNotSet)
	}
	return config.validateUpdate()
}

func (config VirtIoRNG) validateUpdate() error {
	if config.Source != nil {
		return config.Source.Validate()
	}
	return nil
}

func (config VirtIoRNG) string() string {
	var settings string
	if config.Source != nil {
		settings = "source=" + config.Source.String()
	}
	if config.Limit != nil && *config.Limit > 0 {
		settings += ",max_bytes=" + strconv.FormatUint(uint64(*config.Limit), 10)
	}
	if config.Period != nil {
		period := config.Period.Milliseconds()
		if period > 0 {
			settings += ",period=" + strconv.FormatInt(period, 10)
		}
	}
	return settings
}

// EntropySource is an Enum.
type EntropySource int

const (
	EntropySourceErrorInvalid = "invalid value for EntropySource"
)

const (
	entropySourceInvalid EntropySource = iota
	EntropySourceRandom
	EntropySourceURandom
	EntropySourceHwRNG
)

const (
	EntropySourceRawRandom  = "/dev/random"
	EntropySourceRawURandom = "/dev/urandom"
	EntropySourceRawHwRNG   = "/dev/hwrng"
)

func (source EntropySource) MarshalJSON() ([]byte, error) {
	str := source.String()
	if str == "" {
		return nil, errors.New(EntropySourceErrorInvalid)
	}
	return json.Marshal(str)
}

func (source *EntropySource) UnmarshalJSON(data []byte) error {
	// Trim the quotes from the JSON string value
	return source.Parse(strings.Trim(string(data), "\""))
}

func (source *EntropySource) Parse(raw string) error {
	switch raw {
	case EntropySourceRawRandom:
		*source = EntropySourceRandom
	case EntropySourceRawURandom:
		*source = EntropySourceURandom
	case EntropySourceRawHwRNG:
		*source = EntropySourceHwRNG
	default:
		return errors.New(EntropySourceErrorInvalid)
	}
	return nil
}

func (source EntropySource) String() string { // for fmt.Stringer interface.
	switch source {
	case EntropySourceRandom:
		return EntropySourceRawRandom
	case EntropySourceURandom:
		return EntropySourceRawURandom
	case EntropySourceHwRNG:
		return EntropySourceRawHwRNG
	}
	return ""
}

func (source EntropySource) Validate() error {
	switch source {
	case EntropySourceRandom, EntropySourceURandom, EntropySourceHwRNG:
		return nil
	}
	return errors.New(EntropySourceErrorInvalid)
}

func (raw RawConfigQemu) GetRandomnessDevice() *VirtIoRNG {
	if v, isSet := raw[qemuApiKeyRandomnessDevice]; isSet {
		var config VirtIoRNG
		settings := splitStringOfSettings(v.(string))
		var source EntropySource
		config.Source = &source
		if v, isSet := settings[qemuAPISettingRngSource]; isSet {
			source.Parse(v) // Can never fail because we get it from the API
		}
		if v, isSet := settings[qemuAPISettingRngLimit]; isSet {
			limit, _ := strconv.ParseUint(v, 10, 64)
			config.Limit = util.Pointer(uint(limit))
		}
		if v, isSet := settings[qemuAPISettingRngPeriod]; isSet {
			period, _ := strconv.ParseInt(v, 10, 64)
			config.Period = util.Pointer(time.Duration(period) * time.Millisecond)
		}
		return &config
	}
	return nil
}
