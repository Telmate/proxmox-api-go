package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var featureCmd = &cobra.Command{
	Use:   "feature GUESTID",
	Short: "Gets the available features of the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateGuestIDset(args, "GuestID")
		vmr := proxmox.NewVmRef(id)
		c := cli.NewClient()
		err = c.CheckVmRef(cli.Context(), vmr)
		if err != nil {
			return
		}
		var features proxmox.GuestFeatures
		if features, err = c.New().Guest.ListFeatures(cli.Context(), *vmr); err != nil {
			return
		}
		cli.PrintFormattedJson(guestCmd.OutOrStdout(), features)
		return
	}}

func init() {
	guestCmd.AddCommand(featureCmd)
}
