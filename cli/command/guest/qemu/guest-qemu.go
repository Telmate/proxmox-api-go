package qemu

import (
	"github.com/perimeter-81/proxmox-api-go/cli/command/guest"
	"github.com/spf13/cobra"
)

var qemuCmd = &cobra.Command{
	Use:   "qemu",
	Short: "Commands to interact with the qemu guests on Proxmox",
}

func init() {
	guest.GuestCmd.AddCommand(qemuCmd)
}
