package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_storageCmd = &cobra.Command{
	Use:   "storage STORAGEID",
	Short: "Deletes the specified Storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "Storage")
	},
}

func init() {
	deleteCmd.AddCommand(delete_storageCmd)
}
