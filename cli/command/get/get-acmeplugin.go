package get

import (
	"github.com/spf13/cobra"
)

var get_acmepluginCmd = &cobra.Command{
	Use:   "acmeplugin PLUGINID",
	Short: "Gets the configuration of the specified AcmePlugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return getConfig(args, "AcmePlugin")
	},
}

func init() {
	GetCmd.AddCommand(get_acmepluginCmd)
}
