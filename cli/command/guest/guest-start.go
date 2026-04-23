package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guest_statusCmd = &cobra.Command{
	Use:   "start GUESTID",
	Short: "Starts the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (error) {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		if err := cli.NewClient().New().Guest.Start(cli.Context(), *vmr);err == nil {
			cli.PrintGuestStatus(GuestCmd.OutOrStdout(), vmr.VmId(), "started")
		}
		return nil
	},
}

func init() {
	GuestCmd.AddCommand(guest_statusCmd)
}
