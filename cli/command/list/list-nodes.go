package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Prints a list of Nodes in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := cli.NewClient().New().Node.List(cli.Context())
		if err != nil {
			return err
		}
		raws := raw.AsArray()
		nodes := make([]proxmox.NodeInfo, len(raws))
		for i := range raws {
			nodes[i] = raws[i].Get()
		}
		cli.PrintRawJson(listCmd.OutOrStdout(), nodes)
		return nil
	},
}

func init() {
	listCmd.AddCommand(list_nodesCmd)
}
