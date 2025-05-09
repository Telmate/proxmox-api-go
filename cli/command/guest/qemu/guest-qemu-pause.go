package qemu

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var qemu_pauseCmd = &cobra.Command{
	Use:   "pause GUESTID",
	Short: "Pauses the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		c := cli.NewClient()
		_, err = c.PauseVm(cli.Context(), vmr)
		if err == nil {
			cli.PrintGuestStatus(qemuCmd.OutOrStdout(), vmr.VmId(), "paused")
		}
		return
	},
}

func init() {
	qemuCmd.AddCommand(qemu_pauseCmd)
}
