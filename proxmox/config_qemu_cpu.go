package proxmox

import (
	"errors"
)

type QemuCPU struct {
	Cores *QemuCpuCores `json:"cores,omitempty"` // Required during creation
}

func (cpu QemuCPU) mapToApi(params map[string]interface{}) {
	if cpu.Cores != nil {
		params["cores"] = int(*cpu.Cores)
	}
}

func (QemuCPU) mapToSDK(params map[string]interface{}) *QemuCPU {
	var cpu QemuCPU
	if v, isSet := params["cores"]; isSet {
		tmp := QemuCpuCores(v.(float64))
		cpu.Cores = &tmp
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
