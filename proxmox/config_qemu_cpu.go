package proxmox

import (
	"errors"
	"strconv"
)

// min value 0 is unset, max value 512. is QemuCpuCores * CpuSockets
type CpuVirtualCores uint16

func (cores CpuVirtualCores) Error() error {
	return errors.New("CpuVirtualCores may have a maximum of " + strconv.FormatInt(int64(cores), 10))
}

func (vCores CpuVirtualCores) Validate(cores *QemuCpuCores, sockets *QemuCpuSockets, current *QemuCPU) error {
	var usedCores, usedSockets CpuVirtualCores
	if cores != nil {
		usedCores = CpuVirtualCores(*cores)
	} else if current != nil && current.Cores != nil {
		usedCores = CpuVirtualCores(*current.Cores)
	}
	if sockets != nil {
		usedSockets = CpuVirtualCores(*sockets)
	} else if current != nil && current.Sockets != nil {
		usedSockets = CpuVirtualCores(*current.Sockets)
	}
	if vCores > usedCores*usedSockets {
		return (usedCores * usedSockets).Error()
	}
	return nil
}

type QemuCPU struct {
	Cores        *QemuCpuCores    `json:"cores,omitempty"` // Required during creation
	Numa         *bool            `json:"numa,omitempty"`
	Sockets      *QemuCpuSockets  `json:"sockets,omitempty"`
	VirtualCores *CpuVirtualCores `json:"vcores,omitempty"`
}

const (
	QemuCPU_Error_CoresRequired string = "cores is required"
)

func (cpu QemuCPU) mapToApi(current *QemuCPU, params map[string]interface{}) (delete string) {
	if cpu.Cores != nil {
		params["cores"] = int(*cpu.Cores)
	}
	if cpu.Numa != nil {
		params["numa"] = Btoi(*cpu.Numa)
	}
	if cpu.Sockets != nil {
		params["sockets"] = int(*cpu.Sockets)
	}
	if cpu.VirtualCores != nil {
		if *cpu.VirtualCores != 0 {
			params["vcpus"] = int(*cpu.VirtualCores)
		} else if current != nil && current.VirtualCores != nil {
			delete += ",vcpus"
		}
	}
	return
}

func (QemuCPU) mapToSDK(params map[string]interface{}) *QemuCPU {
	var cpu QemuCPU
	if v, isSet := params["cores"]; isSet {
		tmp := QemuCpuCores(v.(float64))
		cpu.Cores = &tmp
	}
	if v, isSet := params["numa"]; isSet {
		tmp := v.(float64) == 1
		cpu.Numa = &tmp
	}
	if v, isSet := params["sockets"]; isSet {
		tmp := QemuCpuSockets(v.(float64))
		cpu.Sockets = &tmp
	}
	if value, isSet := params["vcpus"]; isSet {
		tmp := CpuVirtualCores((value.(float64)))
		cpu.VirtualCores = &tmp
	}
	return &cpu
}

func (cpu QemuCPU) Validate(current *QemuCPU) (err error) {
	if cpu.Cores != nil {
		if err = cpu.Cores.Validate(); err != nil {
			return
		}
	} else if current == nil {
		return errors.New(QemuCPU_Error_CoresRequired)
	}
	if cpu.Sockets != nil {
		if err = cpu.Sockets.Validate(); err != nil {
			return
		}
	}
	if cpu.VirtualCores != nil {
		if err = cpu.VirtualCores.Validate(cpu.Cores, cpu.Sockets, current); err != nil {
			return
		}
	}
	return
}

// min value 1, max value of 128
type QemuCpuCores uint8

const (
	QemuCpuCores_Error_LowerBound string = "minimum value of QemuCpuCores is 1"
	QemuCpuCores_Error_UpperBound string = "maximum value of QemuCpuCores is 128"
)

func (cores QemuCpuCores) Validate() error {
	if cores < 1 {
		return errors.New(QemuCpuCores_Error_LowerBound)
	}
	if cores > 128 {
		return errors.New(QemuCpuCores_Error_UpperBound)
	}
	return nil
}

// min value 1, max value 4
type QemuCpuSockets uint8

const (
	QemuCpuSockets_Error_LowerBound string = "minimum value of QemuCpuSockets is 1"
	QemuCpuSockets_Error_UpperBound string = "maximum value of QemuCpuSockets is 4"
)

func (sockets QemuCpuSockets) Validate() error {
	if sockets < 1 {
		return errors.New(QemuCpuSockets_Error_LowerBound)
	}
	if sockets > 4 {
		return errors.New(QemuCpuSockets_Error_UpperBound)
	}
	return nil
}
