package template

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var template_listCmd = &cobra.Command{
	Use:   "list NODE",
	Short: "Prints a list of all LXC templates available for download in raw json format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		templates, err := proxmox.ListTemplates(cli.NewClient(), args[0])
		if err != nil {
			return
		}
		cli.PrintRawJson(templateCmd.OutOrStdout(), format(templates))
		return
	},
}

func init() {
	templateCmd.AddCommand(template_listCmd)
}

func format(templates *[]proxmox.TemplateItem) *[]string {
	list := make([]string, len(*templates))
	for i, e := range *templates {
		list[i] = e.Template
	}
	return &list
}
