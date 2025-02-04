package id

import (
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var id_checkCmd = &cobra.Command{
	Use:   "check GuestID",
	Short: "Checks if a GuestID is available",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateGuestIDset(args, "GUESTID")
		c := cli.NewClient()
		exists, err := c.VMIdExists(cli.Context(), id)
		if err != nil {
			return
		}
		if exists {
			fmt.Fprintf(idCmd.OutOrStdout(), "Selected ID is in use: %d\n", id)
		} else {
			fmt.Fprintf(idCmd.OutOrStdout(), "Selected ID is free: %d\n", id)
		}
		return
	},
}

func init() {
	idCmd.AddCommand(id_checkCmd)
}
