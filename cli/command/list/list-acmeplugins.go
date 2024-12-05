package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var list_acmepluginsCmd = &cobra.Command{
	Use:   "acmeplugins",
	Short: "Prints a list of AcmePlugins in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw(cli.Context(), "AcmePlugins")
	},
}

func init() {
	listCmd.AddCommand(list_acmepluginsCmd)
}
