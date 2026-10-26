package qemu

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var qemu_resumeCmd = &cobra.Command{
	Use:   "resume GUESTID",
	Short: "Resumes the specified guest",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmr := proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))
		c := cli.NewClient().New()
		if err := c.QemuGuest.Resume(cli.Context(), *vmr); err == nil {
			cli.PrintGuestStatus(qemuCmd.OutOrStdout(), vmr.VmId(), "resumed")
		}
		return nil
	}}

func init() {
	qemuCmd.AddCommand(qemu_resumeCmd)
}
