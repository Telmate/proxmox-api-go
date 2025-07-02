package guest

import (
	"fmt"
	"strconv"

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
		raw, err := vmr.GetRawGuestStatus(cli.Context(), c)
		if err == nil {
			fmt.Fprintln(GuestCmd.OutOrStdout(), "Uptime of guest with id "+vmr.VmId().String()+" is "+strconv.Itoa(int(raw.Uptime().Seconds())))
		}
		return
	},
}

func init() {
	GuestCmd.AddCommand(guest_uptimeCmd)
}
