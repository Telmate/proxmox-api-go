package delete

import (
	"github.com/spf13/cobra"
)

var delete_storageCmd = &cobra.Command{
	Use:   "storage STORAGEID",
	Short: "Deletes the specified Storage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(args, "Storage")
	},
}

func init() {
	deleteCmd.AddCommand(delete_storageCmd)
}
