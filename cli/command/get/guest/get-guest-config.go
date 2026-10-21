package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config GUESTID",
	Short: "Gets the configuration of the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateGuestIDset(args, "GuestID")
		vmr := proxmox.NewVmRef(id)
		c := cli.NewClient()
		err = c.CheckVmRef(cli.Context(), vmr)
		if err != nil {
			return
		}
		vmType := vmr.GetVmType()
		var config any
		var raw proxmox.RawConfigQemu
		switch vmType {
		case proxmox.GuestQemu:
			raw, err = c.New().QemuGuest.Read(cli.Context(), vmr)
			if err != nil {
				return err
			}
			config, err = raw.Get(*vmr)
		case proxmox.GuestLxc:
			config, err = proxmox.NewConfigLxcFromApi(cli.Context(), vmr, c)
		}
		if err != nil {
			return
		}
		cli.PrintFormattedJson(guestCmd.OutOrStdout(), config)
		return
	}}

func init() {
	guestCmd.AddCommand(configCmd)
}
