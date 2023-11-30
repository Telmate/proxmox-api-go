package list

import (
	"github.com/Bluearchive/proxmox-api-go/cli"
	"github.com/Bluearchive/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_qemuguestsCmd = &cobra.Command{
	Use:   "guests",
	Short: "Prints a list of Qemu/Lxc Guests in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c := cli.NewClient()
		guests, err := proxmox.ListGuests(c)
		cli.LogFatalListing("Guests", err)
		cli.PrintRawJson(listCmd.OutOrStdout(), guests)
	},
}

func init() {
	listCmd.AddCommand(list_qemuguestsCmd)
}
