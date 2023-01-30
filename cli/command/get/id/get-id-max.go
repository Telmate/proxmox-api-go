package id

import (
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var id_maxCmd = &cobra.Command{
	Use:   "max",
	Short: "Returns the maximum in use ID number",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		id, err := proxmox.MaxVmId(c)
		if err != nil {
			return
		}
		fmt.Fprintf(idCmd.OutOrStdout(), "Max in use ID: %d\n", id)
		return
	},
}

func init() {
	idCmd.AddCommand(id_maxCmd)
}
