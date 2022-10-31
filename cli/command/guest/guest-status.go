package guest

import (
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guest_startCmd = &cobra.Command{
	Use:   "status GUESTID",
	Short: "Gets the status of the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
		c := cli.NewClient()
		vmState, err := c.GetVmState(vmr)
		if err == nil {
			fmt.Fprintf(GuestCmd.OutOrStdout(), "Status of guest with id (%d) is %s\n", vmr.VmId(), vmState["status"].(string))
		}
		return
	},
}

func init() {
	GuestCmd.AddCommand(guest_startCmd)
}
