package template

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var template_downloadCmd = &cobra.Command{
	Use:   "download NODE STORAGE TEMPLATE",
	Short: "download te specified LXC template",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		config := proxmox.ConfigContent_Template{
			Node:     args[0],
			Storage:  args[1],
			Template: args[2],
		}
		err = config.Validate()
		if err != nil {
			return
		}
		err = proxmox.DownloadLxcTemplate(c, config)
		if err != nil {
			return
		}
		cli.PrintItemCreated(templateCmd.OutOrStdout(), config.Template, "LXC Template")
		return
	},
}

func init() {
	templateCmd.AddCommand(template_downloadCmd)
}
