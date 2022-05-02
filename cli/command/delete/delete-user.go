package delete

import (
	"github.com/spf13/cobra"
)

var delete_userCmd = &cobra.Command{
	Use:   "user USERID",
	Short: "Deletes the speciefied User",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		err = DeleteID(args, "User")
		return
	},
}

func init() {
	deleteCmd.AddCommand(delete_userCmd)
}
