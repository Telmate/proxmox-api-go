package api_test

import (
	"net"

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

	mac, _ := net.ParseMAC("B6:8F:9D:7C:8F:BC")

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
			Type:    util.Pointer(pxapi.CpuType_QemuKvm64),
		},
		QemuKVM: util.Pointer(true),
		Hotplug: "network,disk,usb",
		Networks: pxapi.QemuNetworkInterfaces{
			pxapi.QemuNetworkInterfaceID0: pxapi.QemuNetworkInterface{
				Bridge:   util.Pointer("vmbr0"),
				Firewall: util.Pointer(true),
				Model:    util.Pointer(pxapi.QemuNetworkModelVirtIO),
				MAC:      &mac}},
		QemuIso: "none",
		Boot:    "order=ide2;net0",
		Scsihw:  "virtio-scsi-pci",
	}

	return config
}
