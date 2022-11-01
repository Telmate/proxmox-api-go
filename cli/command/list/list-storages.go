package list

import (
	"github.com/spf13/cobra"
)

var list_storagesCmd = &cobra.Command{
	Use:   "storages",
	Short: "Prints a list of Storages in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw("Storages")
	},
}

func init() {
	listCmd.AddCommand(list_storagesCmd)
}
