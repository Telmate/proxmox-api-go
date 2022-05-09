package get

import (
	"github.com/spf13/cobra"
)

var get_userCmd = &cobra.Command{
	Use:   "user USERID",
	Short: "Gets the configuration of the specified User",
	RunE: func(cmd *cobra.Command, args []string) error {
		return GetConfig(args, "User")
	},
}

func init() {
	getCmd.AddCommand(get_userCmd)
}
