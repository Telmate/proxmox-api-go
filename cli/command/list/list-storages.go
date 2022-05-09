package list

import (
	"github.com/spf13/cobra"
)

var list_storagesCmd = &cobra.Command{
	Use:   "storages",
	Short: "Prints a list of Storages in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		ListRaw("Storages")
	},
}


func init() {
	listCmd.AddCommand(list_storagesCmd)
}
