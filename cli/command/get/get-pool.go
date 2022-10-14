package get

import (
	"github.com/spf13/cobra"
)

var get_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Gets the configuration of the specified Pool",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return GetConfig(args, "Pool")
	},
}

func init() {
	GetCmd.AddCommand(get_poolCmd)
}
