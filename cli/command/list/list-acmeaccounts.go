package list

import (
	"github.com/spf13/cobra"
)

var list_acmeaccountsCmd = &cobra.Command{
	Use:   "acmeaccounts",
	Short: "Prints a list of AcmeAccounts in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		ListRaw("AcmeAccounts")
	},
}


func init() {
	listCmd.AddCommand(list_acmeaccountsCmd)
}
