package list

import (
	"github.com/spf13/cobra"
)

var list_poolsCmd = &cobra.Command{
	Use:   "pools",
	Short: "Prints a list of Pools in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		ListRaw("Pools")
	},
}


func init() {
	listCmd.AddCommand(list_poolsCmd)
}
