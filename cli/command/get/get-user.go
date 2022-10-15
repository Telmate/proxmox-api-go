package get

import (
	"github.com/spf13/cobra"
)

var get_userCmd = &cobra.Command{
	Use:   "user USERID",
	Short: "Gets the configuration of the specified User",
	RunE: func(cmd *cobra.Command, args []string) error {
		return getConfig(args, "User")
	},
}

func init() {
	GetCmd.AddCommand(get_userCmd)
}
