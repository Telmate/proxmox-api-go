package list

import (
	"github.com/spf13/cobra"
)

var list_nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Prints a list of Nodes in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw("Nodes")
	},
}

func init() {
	listCmd.AddCommand(list_nodesCmd)
}
