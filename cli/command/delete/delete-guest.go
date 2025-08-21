package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var delete_guestCmd = &cobra.Command{
	Use:   "guest GUESTID",
	Short: "Deletes the Specified Guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateGuestIDset(args, "GuestID")
		vmr := proxmox.NewVmRef(id)
		c := cli.NewClient()
		if err = vmr.Delete(cli.Context(), c); err != nil {
			return
		}
		cli.PrintItemDeleted(deleteCmd.OutOrStdout(), id.String(), "GuestID")
		return
	},
}

func init() {
	deleteCmd.AddCommand(delete_guestCmd)
}
