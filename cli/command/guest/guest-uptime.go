package guest

import (
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guest_uptimeCmd = &cobra.Command{
	Use:   "uptime GUESTID",
	Short: "Gets the uptime of the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		c := cli.NewClient()
		vmState, err := c.GetVmState(cli.Context(), vmr)
		if err == nil {
			fmt.Fprintf(GuestCmd.OutOrStdout(), "Uptime of guest with id (%d) is %d\n", vmr.VmId(), int(vmState["uptime"].(float64)))
		}
		return
	},
}

func init() {
	GuestCmd.AddCommand(guest_uptimeCmd)
}
