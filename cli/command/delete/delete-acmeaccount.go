package delete

import (
	"github.com/spf13/cobra"
)

var delete_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount ACMEACCOUNTID",
	Short: "Deletes the Specified AcmeAccount",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(args, "AcmeAccount")
	},
}

func init() {
	deleteCmd.AddCommand(delete_acmeaccountCmd)
}
