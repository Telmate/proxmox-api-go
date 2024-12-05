package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_poolsCmd = &cobra.Command{
	Use:   "pools",
	Short: "Prints a list of Pools in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		pools, err := proxmox.ListPools(cli.Context(), c)
		if err != nil {
			return
		}
		cli.PrintFormattedJson(listCmd.OutOrStdout(), pools)
		return
	},
}

func init() {
	listCmd.AddCommand(list_poolsCmd)
}
