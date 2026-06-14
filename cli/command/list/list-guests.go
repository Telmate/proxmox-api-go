package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_qemuguestsCmd = &cobra.Command{
	Use:   "guests",
	Short: "Prints a list of Qemu/Lxc Guests in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		c := cli.NewClient()
		rawGuests, err := c.New().Guest.List(cli.Context())
		cli.LogFatalListing("Guests", err)
		guests := make([]pveSDK.GuestResource, rawGuests.Len())
		var index int
		for e := range rawGuests.Iter() {
			guests[index] = e.Get()
			index++
		}
		cli.PrintRawJson(listCmd.OutOrStdout(), guests)
	},
}

func init() {
	listCmd.AddCommand(list_qemuguestsCmd)
}
