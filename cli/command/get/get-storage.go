package get

import (
	"github.com/spf13/cobra"
)

var get_storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Gets the configuration of the specified Storage backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		return GetConfig(args, "Storage")
	},
}

func init() {
	GetCmd.AddCommand(get_storageCmd)
}
