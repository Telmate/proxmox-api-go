package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_userCmd = &cobra.Command{
	Use:   "user USERID",
	Short: "Deletes the specified User",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "User")
	},
}

func init() {
	deleteCmd.AddCommand(delete_userCmd)
}
