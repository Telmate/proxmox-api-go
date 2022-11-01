package get

import (
	"github.com/spf13/cobra"
)

var get_storageCmd = &cobra.Command{
	Use:   "storage STORAGEID",
	Short: "Gets the configuration of the specified Storage backend",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getConfig(args, "Storage")
	},
}

func init() {
	GetCmd.AddCommand(get_storageCmd)
}
