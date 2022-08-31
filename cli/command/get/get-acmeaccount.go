package get

import (
	"github.com/spf13/cobra"
)

var get_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount ACMEACCOUNTID",
	Short: "Gets the configuration of the specified AcmeAccount",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return GetConfig(args, "AcmeAccount")
	},
}

func init() {
	getCmd.AddCommand(get_acmeaccountCmd)
}
