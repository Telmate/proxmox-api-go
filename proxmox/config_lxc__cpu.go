package proxmox

import (
	"errors"
	"strconv"
)

type LxcCPU struct {
	Cores *LxcCpuCores `json:"cores,omitempty"`
	Limit *LxcCpuLimit `json:"limit,omitempty"`
	Units *LxcCpuUnits `json:"units,omitempty"`
}

func (config LxcCPU) mapToApiCreate(params map[string]any) {
	if config.Cores != nil && *config.Cores != 0 {
		params[lxcApiKeyCores] = int(*config.Cores)
	}
	if config.Limit != nil && *config.Limit != 0 {
		params[lxcApiKeyCpuLimit] = int(*config.Limit)
	}
	if config.Units != nil && *config.Units != 0 {
		params[lxcApiKeyCpuUnits] = int(*config.Units)
	}

}

func (config LxcCPU) mapToApiUpdate(current LxcCPU, params map[string]any) (delete string) {
	if config.Cores != nil {
		if current.Cores != nil {
			if *config.Cores == 0 {
				delete += "," + lxcApiKeyCores
			} else if *config.Cores != *current.Cores {
				params[lxcApiKeyCores] = int(*config.Cores)
			}
		} else {
			if *config.Cores != 0 {
				params[lxcApiKeyCores] = int(*config.Cores)
			}
		}
	}
	if config.Limit != nil {
		if current.Limit != nil {
			if *config.Limit == 0 {
				delete += "," + lxcApiKeyCpuLimit
			} else if *config.Limit != *current.Limit {
				params[lxcApiKeyCpuLimit] = int(*config.Limit)
			}
		} else {
			if *config.Limit != 0 {
				params[lxcApiKeyCpuLimit] = int(*config.Limit)
			}
		}
	}
	if config.Units != nil {
		if current.Units != nil {
			if *config.Units == 0 {
				delete += "," + lxcApiKeyCpuUnits
			} else if *config.Units != *current.Units {
				params[lxcApiKeyCpuUnits] = int(*config.Units)
			}
		} else {
			if *config.Units != 0 {
				params[lxcApiKeyCpuUnits] = int(*config.Units)
			}
		}
	}
	return
}

func (raw RawConfigLXC) CPU() *LxcCPU {
	cpu := LxcCPU{}
	var parameterSet bool
	if v, isSet := raw[lxcApiKeyCores]; isSet {
		tmp := LxcCpuCores(v.(float64))
		cpu.Cores = &tmp
		parameterSet = true
	}
	if v, isSet := raw[lxcApiKeyCpuLimit]; isSet {
		tmp := LxcCpuLimit(v.(float64))
		cpu.Limit = &tmp
		parameterSet = true
	}
	if v, isSet := raw[lxcApiKeyCpuUnits]; isSet {
		tmp := LxcCpuUnits(v.(float64))
		cpu.Units = &tmp
		parameterSet = true
	}
	if parameterSet {
		return &cpu
	}
	return nil
}

func (cpu LxcCPU) Validate() error {
	if cpu.Cores != nil {
		if err := cpu.Cores.Validate(); err != nil {
			return err
		}
	}
	if cpu.Limit != nil {
		if err := cpu.Limit.Validate(); err != nil {
			return err
		}
	}
	if cpu.Units != nil {
		if err := cpu.Units.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// 0-8192, 0 means no limit
type LxcCpuCores uint16

const (
	LxcCpuCoresMaximum        = 8192
	LxcCpuCores_Error_Invalid = "cpu cores should be in the range 0-8192"
)

func (cores LxcCpuCores) String() string {
	return strconv.FormatUint(uint64(cores), 10)
}

func (cores LxcCpuCores) Validate() error {
	if cores > LxcCpuCoresMaximum {
		return errors.New(LxcCpuCores_Error_Invalid)
	}
	return nil
}

// 0-8192, 0 means no limit
type LxcCpuLimit float32

const (
	LxcCpuLimitMaximum        = 8192
	LxcCpuLimit_Error_Invalid = "cpu limit should be in the range 0-8192"
)

func (limit LxcCpuLimit) String() string {
	return strconv.FormatFloat(float64(limit), 'f', -1, 32)
}

func (limit LxcCpuLimit) Validate() error {
	if limit > LxcCpuLimitMaximum {
		return errors.New(LxcCpuLimit_Error_Invalid)
	}
	return nil
}

// 0-100000, 0 means no limit
type LxcCpuUnits uint32

const (
	LxcCpuUnitsMaximum        = 100000
	LxcCpuUnits_Error_Invalid = "cpu units has a maximum of 100000"
)

func (units LxcCpuUnits) String() string {
	return strconv.FormatUint(uint64(units), 10)
}

func (units LxcCpuUnits) Validate() error {
	if units > LxcCpuUnitsMaximum {
		return errors.New(LxcCpuUnits_Error_Invalid)
	}
	return nil
}
