package delete

import (
	"github.com/spf13/cobra"
)

var delete_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount ACMEACCOUNTID",
	Short: "Deletes the Speciefied acmeaccount",
	RunE: func(cmd *cobra.Command, args []string) error {
		return DeleteID(args, "AcmeAccount")
	},
}

func init() {
	deleteCmd.AddCommand(delete_acmeaccountCmd)
}
