package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/internal/util"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
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
			id := cli.ValidateGuestIDset(args, "GuestID")
			rawSnapshots, err := cli.NewClient().New().Snapshot.List(cli.Context(), *pveSDK.NewVmRef(id))
			if err != nil {
				noTree = false
				return
			}
			type snapshotArray struct {
				Name        pveSDK.SnapshotName  `json:"name,omitempty"`
				Time        *int64               `json:"time,omitempty"`
				Description string               `json:"description,omitempty"`
				VmState     *bool                `json:"state,omitempty"`
				Parent      *pveSDK.SnapshotName `json:"parent,omitempty"`
			}
			type snapshotTree struct {
				Name        pveSDK.SnapshotName `json:"name,omitempty"`
				Time        *int64              `json:"time,omitempty"`
				Description string              `json:"description,omitempty"`
				VmState     *bool               `json:"state,omitempty"`
				Children    []*snapshotTree     `json:"children,omitempty"`
			}

			// Helper function to convert Snapshot to snapshotTree recursively
			var convertToTree func(s *pveSDK.Snapshot) *snapshotTree
			convertToTree = func(s *pveSDK.Snapshot) *snapshotTree {
				if s == nil {
					return nil
				}
				node := &snapshotTree{
					Name:        s.Name,
					Description: s.Description,
					VmState:     s.VmState,
				}
				if s.Time != nil {
					node.Time = util.Pointer(s.Time.Unix())
				}
				// Convert children recursively
				if len(s.Children) > 0 {
					node.Children = make([]*snapshotTree, len(s.Children))
					for i, child := range s.Children {
						node.Children[i] = convertToTree(child)
					}
				}
				return node
			}

			if noTree {
				rawArray := rawSnapshots.AsArray()
				if len(rawArray) > 0 {
					array := make([]snapshotArray, len(rawArray))
					for i := range rawArray {
						info := rawArray[i].Get()
						array[i] = snapshotArray{
							Name:        info.Name,
							Description: info.Description,
							VmState:     info.VmState,
							Parent:      info.Parent,
						}
						if info.Time != nil {
							array[i].Time = util.Pointer(info.Time.Unix())
						}
					}
					cli.PrintFormattedJson(listCmd.OutOrStdout(), array)
					return
				}
			} else {
				tree := rawSnapshots.Tree()
				roots := tree.Root()
				if len(roots) > 1 || roots[0] != tree.Current() {
					treeList := make([]*snapshotTree, len(roots))
					for i, root := range roots {
						treeList[i] = convertToTree(root)
					}
					cli.PrintFormattedJson(listCmd.OutOrStdout(), treeList)
					return
				}
			}
			listCmd.Printf("Guest with ID (%d) has no snapshots", id)
			return
		},
	}
)

func init() {
	listCmd.AddCommand(list_snapshotsCmd)
	list_snapshotsCmd.Flags().BoolVar(&noTree, "no-tree", false, "Format output as list instead of a tree.")
}
