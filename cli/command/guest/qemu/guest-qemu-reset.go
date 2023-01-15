package qemu

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var qemu_resetCmd = &cobra.Command{
	Use:   "reset GUESTID",
	Short: "Resets the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
		c := cli.NewClient()
		_, err = c.StartVm(vmr)
		if err == nil {
			cli.PrintGuestStatus(qemuCmd.OutOrStdout(), vmr.VmId(), "reset")
		}
		return
	},
}

func init() {
	qemuCmd.AddCommand(qemu_resetCmd)
}
