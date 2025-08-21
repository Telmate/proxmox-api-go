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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		c := cli.NewClient()
		if err = vmr.Stop(cli.Context(), c); err == nil {
			cli.PrintGuestStatus(GuestCmd.OutOrStdout(), vmr.VmId(), "stopped")
		}
		return
	},
}

func init() {
	GuestCmd.AddCommand(guest_stopCmd)
}
