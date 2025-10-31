package template

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var template_deleteCmd = &cobra.Command{
	Use:   "delete NODE STORAGE TEMPLATE",
	Short: "delete the specified LXC template",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		config := proxmox.ConfigContent_Template{
			Node:     args[0],
			Storage:  args[1],
			Template: args[2],
		}
		if err = config.Validate(); err != nil {
			return
		}
		if err = config.Delete(cli.Context(), c); err != nil {
			return
		}
		cli.PrintItemDeleted(templateCmd.OutOrStdout(), config.Template, "LXC Template")
		return
	},
}

func init() {
	templateCmd.AddCommand(template_deleteCmd)
}
