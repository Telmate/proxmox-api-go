package api_test

import (
	"github.com/Telmate/proxmox-api-go/internal/util"
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
)

func _create_vmref() (ref *pxapi.VmRef) {
	ref = pxapi.NewVmRef(100)
	ref.SetNode("pve")
	ref.SetVmType("qemu")
	return ref
}

func _create_vm_spec(network bool) pxapi.ConfigQemu {

	disks := make(pxapi.QemuDevices)
	disks[0] = make(map[string]interface{})
	disks[0]["type"] = "virtio"
	disks[0]["storage"] = "local"
	disks[0]["size"] = "1G"

	networks := make(pxapi.QemuDevices)
	if network {
		networks[0] = make(map[string]interface{})
		networks[0]["bridge"] = "vmbr0"
		networks[0]["firewall"] = "true"
		networks[0]["id"] = "0"
		networks[0]["macaddr"] = "B6:8F:9D:7C:8F:BC"
		networks[0]["model"] = "virtio"
	}

	config := pxapi.ConfigQemu{
		Name:   "test-qemu01",
		Bios:   "seabios",
		Tablet: util.Pointer(true),
		Memory: &pxapi.QemuMemory{CapacityMiB: util.Pointer(pxapi.QemuMemoryCapacity(128))},
		QemuOs: "l26",
		CPU: &pxapi.QemuCPU{
			Cores:   util.Pointer(pxapi.QemuCpuCores(1)),
			Numa:    util.Pointer(false),
			Sockets: util.Pointer(pxapi.QemuCpuSockets(1)),
		},
		QemuCpu:      "kvm64",
		QemuKVM:      util.Pointer(true),
		Hotplug:      "network,disk,usb",
		QemuNetworks: networks,
		QemuIso:      "none",
		Boot:         "order=ide2;net0",
		Scsihw:       "virtio-scsi-pci",
		QemuDisks:    disks,
	}

	return config
}
