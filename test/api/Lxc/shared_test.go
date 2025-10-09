package api_test

import (
	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

const lxcOsTemplate = test.CtStorage + ":vztmpl/" + test.DownloadedLXCTemplate;

func _create_vmref() (ref *pxapi.VmRef) {
	ref = pxapi.NewVmRef(200)
	ref.SetNode("pve")
	ref.SetVmType(pxapi.GuestLxc)
	return ref
}

func _create_lxc_spec(network bool, osTemplate string) pxapi.ConfigLxc {

	disks := make(pxapi.QemuDevices)
	disks[0] = make(map[string]interface{})
	disks[0]["type"] = "virtio"
	disks[0]["storage"] = "local-lvm"
	disks[0]["size"] = "1G"

	networks := make(pxapi.QemuDevices)

	config := pxapi.ConfigLxc{
		Hostname:     "test-lxc01",
		Cores:        1,
		Memory:       128,
		Password:     "SuperSecretPassword",
		Ostemplate:   osTemplate,
		Storage:      "local",
		RootFs:       disks[0],
		Networks:     networks,
		Arch:         "amd64",
		CMode:        "tty",
		Console:      true,
		CPULimit:     0,
		CPUUnits:     1024,
		OnBoot:       false,
		Protection:   false,
		Start:        false,
		Swap:         512,
		Template:     false,
		Tty:          2,
		Unprivileged: false,
	}

	return config
}
