package get

import (
	"github.com/spf13/cobra"
)

var get_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount ACMEACCOUNTID",
	Short: "Gets the configuration of the specified AcmeAccount",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return getConfig(args, "AcmeAccount")
	},
}

func init() {
	GetCmd.AddCommand(get_acmeaccountCmd)
}
