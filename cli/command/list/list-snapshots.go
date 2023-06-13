package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	// flag needs to be reset, as this value will persist during tests
	noTree            bool
	list_snapshotsCmd = &cobra.Command{
		Use:              "snapshots GuestID",
		Short:            "Prints a list of QemuSnapshots in json format",
		TraverseChildren: true,
		Args:             cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id := cli.ValidateExistingGuestID(args, 0)
			rawSnapshots, err := proxmox.ListSnapshots(cli.NewClient(), proxmox.NewVmRef(id))
			if err != nil {
				noTree = false
				return
			}
			var list []*proxmox.Snapshot
			if noTree {
				noTree = false
				list = rawSnapshots.FormatSnapshotsList()
			} else {
				list = rawSnapshots.FormatSnapshotsTree()
			}
			if len(list) == 0 {
				listCmd.Printf("Guest with ID (%d) has no snapshots", id)
			} else {
				cli.PrintFormattedJson(listCmd.OutOrStdout(), list)
			}
			return
		},
	}
)

func init() {
	listCmd.AddCommand(list_snapshotsCmd)
	list_snapshotsCmd.Flags().BoolVar(&noTree, "no-tree", false, "Format output as list instead of a tree.")
}
