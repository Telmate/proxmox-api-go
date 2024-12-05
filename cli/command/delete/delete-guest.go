package delete

import (
	"strconv"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var delete_guestCmd = &cobra.Command{
	Use:   "guest GUESTID",
	Short: "Deletes the Specified Guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateIntIDset(args, "GuestID")
		vmr := proxmox.NewVmRef(id)
		c := cli.NewClient()
		_, err = c.StopVm(cli.Context(), vmr)
		if err != nil {
			return
		}
		_, err = c.DeleteVm(cli.Context(), vmr)
		if err != nil {
			return
		}
		cli.PrintItemDeleted(deleteCmd.OutOrStdout(), strconv.Itoa(id), "GuestID")
		return
	},
}

func init() {
	deleteCmd.AddCommand(delete_guestCmd)
}
