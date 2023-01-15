package create

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var create_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount ACMEACCOUNTID",
	Short: "Creates a new AcmeAccount",
	Long: `Creates a new AcmeAccount.
The config can be set with the --file flag or piped from stdin.
For config examples see "example acmeaccount"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "AcmeAccountID")
		config, err := proxmox.NewConfigAcmeAccountFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		c := cli.NewClient()
		err = config.CreateAcmeAccount(id, c)
		if err != nil {
			return
		}
		cli.PrintItemCreated(CreateCmd.OutOrStdout(), id, "AcmeAccount")
		return
	},
}

func init() {
	CreateCmd.AddCommand(create_acmeaccountCmd)
}
