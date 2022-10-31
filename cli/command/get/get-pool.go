package get

import (
	"github.com/spf13/cobra"
)

var get_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Gets the configuration of the specified Pool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getConfig(args, "Pool")
	},
}

func init() {
	GetCmd.AddCommand(get_poolCmd)
}
