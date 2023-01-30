package cli_guestqemu_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
)

func Test_GuestQemu_100_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:      true,
		ErrContains: "100",
		Args:        []string{"-i", "delete", "guest", "100"},
	}
	Test.StandardTest(t)
}

// TODO add more parameters to test
func Test_GuestQemu_100_Create(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"name": "test-qemu01",
	"bios": "seabios",
	"tablet": true,
	"memory": 128,
	"os": "l26",
	"cores": 1,
	"sockets": 1,
	"cpu": "host",
	"numa": false,
	"kvm": true,
	"hotplug": "network,disk,usb",
	"iso": "none",
	"boot": "order=ide2;net0",
	"scsihw": "virtio-scsi-pci",
	"network": {
		"0": {
			"bridge": "vmbr0",
			"firewall": true,
			"id": 0,
			"macaddr": "B6:8F:9D:7C:8F:BC",
			"model": "virtio"
		}
	}
}`,
		Expected: "(100)",
		Contains: true,
		Args:     []string{"-i", "create", "guest", "qemu", "100", "pve"},
	}
	Test.StandardTest(t)
}

func Test_GuestQemu_100_Get(t *testing.T) {
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-qemu01",
	"bios": "seabios",
	"onboot": true,
	"tablet": true,
	"memory": 128,
	"os": "l26",
	"cores": 1,
	"sockets": 1,
	"cpu": "host",
	"numa": false,
	"kvm": true,
	"hotplug": "network,disk,usb",
	"iso": "none",
	"boot": "order=ide2;net0",
	"scsihw": "virtio-scsi-pci",
	"network": {
		"0": {
			"bridge": "vmbr0",
			"firewall": true,
			"id": 0,
			"macaddr": "B6:8F:9D:7C:8F:BC",
			"model": "virtio"
		}
	}
}`,
		Args: []string{"-i", "get", "guest", "100"},
	}
	Test.StandardTest(t)
}

func Test_GuestQemu_100_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "guest", "100"},
	}
	Test.StandardTest(t)
}
