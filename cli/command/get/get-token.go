package get

import (
	"github.com/spf13/cobra"
)

var get_tokenCmd = &cobra.Command{
	Use:   "token TOKENID",
	Short: "Gets the configuration of the specified API Token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getConfig(args, "Token")
	},
}

func init() {
	GetCmd.AddCommand(get_tokenCmd)
}
