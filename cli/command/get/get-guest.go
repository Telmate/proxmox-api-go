package get

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var get_guestCmd = &cobra.Command{
	Use:   "guest GUESTID",
	Short: "Gets the configuration of the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.ValidateIntIDset(args, "GuestID")
		vmr := proxmox.NewVmRef(id)
		c := cli.NewClient()
		err = c.CheckVmRef(vmr)
		if err != nil {
			return
		}
		vmType := vmr.GetVmType()
		var config interface{}
		switch vmType {
		case "qemu":
			config, err = proxmox.NewConfigQemuFromApi(vmr, c)
		case "lxc":
			config, err = proxmox.NewConfigLxcFromApi(vmr, c)
		}
		if err != nil {
			return
		}
		cli.PrintFormattedJson(GetCmd.OutOrStdout(), config)
		return
	},
}

func init() {
	GetCmd.AddCommand(get_guestCmd)
}
