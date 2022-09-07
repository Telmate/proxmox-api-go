package id

import (
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var id_checkCmd = &cobra.Command{
	Use:   "check ID",
	Short: "Checks if a ID is availible",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateIntIDset(args, "ID")
		c := cli.NewClient()
		exixst, err := c.VMIdExists(id)
		if err != nil {
			return
		}
		if exixst {
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
