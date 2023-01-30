package list

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_FileCmd = &cobra.Command{
	Use:   "files NODE STORAGE TYPE",
	Short: "Prints a list of Files of the specified type in raw json format",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		templates, err := proxmox.ListFiles(c, args[0], args[1], proxmox.ContentType(args[2]))
		if err != nil {
			return
		}
		cli.PrintRawJson(listCmd.OutOrStdout(), templates)
		return
	},
}

func init() {
	listCmd.AddCommand(list_FileCmd)
}
