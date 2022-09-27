package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guest_statusCmd = &cobra.Command{
	Use:   "start GUESTID",
	Short: "Starts the speciefid guest",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
		c := cli.NewClient()
		_, err = c.StartVm(vmr)
		if err == nil {
			cli.PrintGuestStatus(GuestCmd.OutOrStdout(), vmr.VmId(), "started")
		}
		return
	},
}

func init() {
	GuestCmd.AddCommand(guest_statusCmd)
}