package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Deletes the Specified pool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "Pool")
	},
}

func init() {
	deleteCmd.AddCommand(delete_poolCmd)
}
