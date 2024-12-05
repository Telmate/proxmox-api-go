package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_groupCmd = &cobra.Command{
	Use:   "group GROUP",
	Short: "Deletes the Specified group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "Group")
	},
}

func init() {
	deleteCmd.AddCommand(delete_groupCmd)
}
