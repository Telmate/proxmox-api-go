package update

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var update_poolCmd = &cobra.Command{
	Use:   "poolcomment POOLID [COMMENT]",
	Short: "Updates the comment on the specified pool",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "PoolID")
		var comment *string
		if len(args) > 1 {
			comment = &args[1]
		}
		c := cli.NewClient()
		err = proxmox.ConfigPool{
			Name:    proxmox.PoolName(id),
			Comment: comment,
		}.Update(cli.Context(), c)
		if err != nil {
			return
		}
		cli.PrintItemUpdated(updateCmd.OutOrStdout(), id, "PoolComment")
		return
	},
}

func init() {
	updateCmd.AddCommand(update_poolCmd)
}
