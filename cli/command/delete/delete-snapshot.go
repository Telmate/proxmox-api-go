package delete

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	delete_snapshotCmd = &cobra.Command{
		Use:   "snapshot GUESTID SNAPSHOTNAME",
		Short: "Deletes the Specified snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id := cli.ValidateIntIDset(args, "GuestID")
			snapName := cli.RequiredIDset(args, 1, "SnapshotName")
			_, err = proxmox.DeleteSnapshot(cli.NewClient(), proxmox.NewVmRef(id), snapName)
			if err != nil {
				return
			}
			cli.PrintItemDeleted(deleteCmd.OutOrStdout(), snapName, "Snapshot")
			return
		},
	}
)

func init() {
	deleteCmd.AddCommand(delete_snapshotCmd)
}
