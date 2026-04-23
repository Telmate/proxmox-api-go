package qemu

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var qemu_resetCmd = &cobra.Command{
	Use:   "reset GUESTID",
	Short: "Resets the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		if err := cli.NewClient().New().Guest.Start(cli.Context(), *vmr); err != nil {
			return err
		}
		cli.PrintGuestStatus(qemuCmd.OutOrStdout(), vmr.VmId(), "reset")
		return nil
	},
}

func init() {
	qemuCmd.AddCommand(qemu_resetCmd)
}
