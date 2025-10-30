package set

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var set_acmepluginCmd = &cobra.Command{
	Use:   "acmeplugin PLUGINID",
	Short: "Sets the configuration of the specified AcmePlugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		config, err := proxmox.NewConfigAcmePluginFromJson(cli.NewConfig())
		if err != nil {
			return
		}

		id := cli.RequiredIDset(args, 0, "PLUGINID")
		c := cli.NewClient()
		if err = config.SetAcmePlugin(cli.Context(), id, c); err != nil {
			return
		}

		cli.PrintItemSet(setCmd.OutOrStdout(), id, "AcmePlugin")
		return
	},
}

func init() {
	setCmd.AddCommand(set_acmepluginCmd)
}
