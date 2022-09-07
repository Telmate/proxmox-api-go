package get

import (
	"encoding/json"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var get_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Gets the configuration of the specified Pool",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateIDset(args, 0, "PoolID")
		c := cli.NewClient()
		poolinfo, err := c.GetPoolInfo(id)
		if err != nil {
			return
		}
		poolList, err := json.Marshal(poolinfo)
		if err != nil {
			return
		}
		fmt.Fprintln(GetCmd.OutOrStdout(), string(poolList))
		return
	},
}

func init() {
	GetCmd.AddCommand(get_poolCmd)
}
