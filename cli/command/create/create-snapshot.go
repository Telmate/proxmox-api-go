package create

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	// flag needs to be reset, as this value will persist during tests
	memory             bool
	create_snapshotCmd = &cobra.Command{
		Use:              "snapshot GUESTID SNAPSHOTNAME [DESCRIPTION]",
		Short:            "Creates a new snapshot of the specified guest",
		TraverseChildren: true,
		Args:             cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id := cli.ValidateIntIDset(args, "GuestID")
			snapName := cli.RequiredIDset(args, 1, "SnapshotName")
			config := proxmox.ConfigSnapshot{
				Name:        snapName,
				Description: cli.OptionalIDset(args, 2),
				VmState:     memory,
			}
			memory = false
			client := cli.NewClient()
			vmr := proxmox.NewVmRef(id)
			_, err = client.GetVmInfo(vmr)
			if err != nil {
				return
			}
			err = config.CreateSnapshot(client, vmr)
			if err != nil {
				return
			}
			cli.PrintItemCreated(CreateCmd.OutOrStdout(), snapName, "Snapshot")
			return
		},
	}
)

func init() {
	CreateCmd.AddCommand(create_snapshotCmd)
	create_snapshotCmd.Flags().BoolVar(&memory, "memory", false, "Snapshot memory")
}
