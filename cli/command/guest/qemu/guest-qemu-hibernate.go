package qemu

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var qemu_hibernateCmd = &cobra.Command{
	Use:   "hibernate GUESTID",
	Short: "Hibernates the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		vmr := proxmox.NewVmRef(cli.ValidateIntIDset(args, "GuestID"))
		c := cli.NewClient()
		_, err = c.HibernateVm(vmr)
		if err == nil {
			cli.PrintGuestStatus(qemuCmd.OutOrStdout(), vmr.VmId(), "suspended to disk")
		}
		return
	},
}

func init() {
	qemuCmd.AddCommand(qemu_hibernateCmd)
}
