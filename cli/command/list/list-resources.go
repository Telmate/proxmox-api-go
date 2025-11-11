package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var list_resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Prints a list of Resources in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw(cli.Context(), "Resources")
	},
}

func init() {
	listCmd.AddCommand(list_resourcesCmd)
}
