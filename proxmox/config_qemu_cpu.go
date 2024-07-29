package proxmox

import (
	"errors"
)

type QemuCPU struct {
	Cores   *QemuCpuCores   `json:"cores,omitempty"` // Required during creation
	Numa    *bool           `json:"numa,omitempty"`
	Sockets *QemuCpuSockets `json:"sockets,omitempty"`
}

func (cpu QemuCPU) mapToApi(params map[string]interface{}) {
	if cpu.Cores != nil {
		params["cores"] = int(*cpu.Cores)
	}
	if cpu.Numa != nil {
		params["numa"] = Btoi(*cpu.Numa)
	}
	if cpu.Sockets != nil {
		params["sockets"] = int(*cpu.Sockets)
	}
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
	return &cpu
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
