package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var list_acmeaccountsCmd = &cobra.Command{
	Use:   "acmeaccounts",
	Short: "Prints a list of AcmeAccounts in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw(cli.Context(), "AcmeAccounts")
	},
}

func init() {
	listCmd.AddCommand(list_acmeaccountsCmd)
}
