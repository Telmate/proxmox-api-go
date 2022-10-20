package list

import (
	"github.com/spf13/cobra"
)

var list_usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Prints a list of Users in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		listRaw("Users")
	},
}

func init() {
	listCmd.AddCommand(list_usersCmd)
}
