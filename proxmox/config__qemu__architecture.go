package proxmox

import "errors"

// Enum
//
//	const (
//		QemuCpuArchitectureAmd64
//		QemuCpuArchitectureArm64
//	)
type QemuCpuArchitecture string

const QemuCpuArchitecture_Error = "invalid cpu architecture"

const (
	QemuCpuArchitectureAmd64 QemuCpuArchitecture = "x86_64"
	QemuCpuArchitectureArm64 QemuCpuArchitecture = "aarch64"
)

func (arch QemuCpuArchitecture) String() string { return string(arch) } // String is for fmt.Stringer.

func (arch QemuCpuArchitecture) Validate() error {
	switch arch {
	case QemuCpuArchitectureAmd64, QemuCpuArchitectureArm64:
		return nil
	}
	return errors.New(QemuCpuArchitecture_Error)
}

func (raw *rawConfigQemu) GetArchitecture() *QemuCpuArchitecture {
	if v, isSet := raw.a[qemuApiKeyArchitecture]; isSet {
		return new(QemuCpuArchitecture(v.(string)))
	}
	return nil

}
