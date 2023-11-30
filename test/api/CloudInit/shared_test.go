package api_test

import (
	pxapi "github.com/Bluearchive/proxmox-api-go/proxmox"
)

func _create_vmref() (ref *pxapi.VmRef) {
	ref = pxapi.NewVmRef(101)
	ref.SetNode("pve")
	ref.SetVmType("qemu")
	return ref
}

func _create_vm_spec(network bool) pxapi.ConfigQemu {

	disks := make(pxapi.QemuDevices)

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
		Name:         "test-qemu01",
		Bios:         "seabios",
		Tablet:       pxapi.PointerBool(true),
		Memory:       2048,
		QemuOs:       "l26",
		QemuCores:    1,
		QemuSockets:  1,
		QemuCpu:      "kvm64",
		QemuNuma:     pxapi.PointerBool(false),
		QemuKVM:      pxapi.PointerBool(true),
		Hotplug:      "network,disk,usb",
		QemuNetworks: networks,
		QemuIso:      "none",
		Boot:         "order=ide2;net0",
		Scsihw:       "virtio-scsi-pci",
		QemuDisks:    disks,
	}

	return config
}

func _create_network_spec() pxapi.ConfigNetwork {
	config := pxapi.ConfigNetwork{
		Type:      "bridge",
		Iface:     "vmbr0",
		Node:      "pve",
		Autostart: true,
	}

	return config
}
