package id

import (
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var id_nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Returns the lowest available ID",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		id, err := c.GetNextID(0)
		if err != nil {
			return
		}
		fmt.Fprintf(idCmd.OutOrStdout(), "Getting Next Free ID: %d\n", id)
		return
	},
}

func init() {
	idCmd.AddCommand(id_nextCmd)
}
