package list

import (
	"github.com/spf13/cobra"
)

var list_acmepluginsCmd = &cobra.Command{
	Use:   "acmeplugins",
	Short: "Prints a list of AcmePlugins in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		ListRaw("AcmePlugins")
	},
}


func init() {
	listCmd.AddCommand(list_acmepluginsCmd)
}
