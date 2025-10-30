package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_acmepluginCmd = &cobra.Command{
	Use:   "acmeplugin PLUGINID",
	Short: "Deletes the Specified AcmePlugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "AcmePlugin")
	},
}

func init() {
	deleteCmd.AddCommand(delete_acmepluginCmd)
}
