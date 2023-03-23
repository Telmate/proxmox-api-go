package get

import (
	"github.com/spf13/cobra"
)

var get_groupCmd = &cobra.Command{
	Use:   "group GROUP",
	Short: "Gets the configuration of the specified Group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getConfig(args, "Group")
	},
}

func init() {
	GetCmd.AddCommand(get_groupCmd)
}
