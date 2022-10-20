package list

import (
	"github.com/spf13/cobra"
)

var list_qemuguestsCmd = &cobra.Command{
	Use:   "guests",
	Short: "Prints a list of Qemu/Lxc Guests in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		listRaw("Guests")
	},
}

func init() {
	listCmd.AddCommand(list_qemuguestsCmd)
}
