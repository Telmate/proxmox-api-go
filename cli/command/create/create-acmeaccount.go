package create

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var create_acmeaccountCmd = &cobra.Command{
	Use:   "acmeaccount",
	Short: "Placeholder",
	Long:  `A longer Placeholder`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateIDset(args, 0, "AcmeAccountID")
		config, err := proxmox.NewConfigAcmeAccountFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		c := cli.NewClient()
		err = config.CreateAcmeAccount(id, c)
		if err != nil {
			return
		}
		cli.PrintItemCreated(createCmd.OutOrStdout(), id, "AcmeAccount")
		return
	},
}

func init() {
	createCmd.AddCommand(create_acmeaccountCmd)
}
