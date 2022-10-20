package delete

import (
	"github.com/spf13/cobra"
)

var delete_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Deletes the Speciefied pool",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(args, "Pool")
	},
}

func init() {
	deleteCmd.AddCommand(delete_poolCmd)
}
