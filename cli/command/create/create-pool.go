package create

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var create_poolCmd = &cobra.Command{
	Use:   "pool POOLID [COMMENT]",
	Short: "Creates a new pool",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "PoolID")
		var comment string
		if len(args) > 1 {
			comment = args[1]
		}
		c := cli.NewClient()
		err = c.CreatePool(id, comment)
		if err != nil {
			return
		}
		cli.PrintItemCreated(CreateCmd.OutOrStdout(), id, "Pool")
		return
	},
}

func init() {
	CreateCmd.AddCommand(create_poolCmd)
}
