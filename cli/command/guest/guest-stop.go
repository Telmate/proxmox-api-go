package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guest_stopCmd = &cobra.Command{
	Use:   "stop GUESTID",
	Short: "Stops the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		if err := cli.NewClient().New().Guest.Stop(cli.Context(), *vmr, true); err == nil {
			cli.PrintGuestStatus(GuestCmd.OutOrStdout(), vmr.VmId(), "stopped")
		}
		return nil
	},
}

func init() {
	GuestCmd.AddCommand(guest_stopCmd)
}
