package get

import (
	"github.com/spf13/cobra"
)

var get_vminfoCmd = &cobra.Command{
	Use:   "vminfo GUESTID",
	Short: "Gets the infos of the specified VM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return getConfig(args, "Vminfo")
	},
}

func init() {
	GetCmd.AddCommand(get_vminfoCmd)
}
