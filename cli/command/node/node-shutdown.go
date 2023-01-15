package node

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var shutdown_nodeCmd = &cobra.Command{
	Use:   "shutdown NODE",
	Short: "Shuts the specified node down",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		node := cli.RequiredIDset(args, 0, "node")
		c := cli.NewClient()
		_, err = c.ShutdownNode(node)
		if err != nil {
			return
		}
		cli.RootCmd.Printf("Node %s is shutting down", node)
		return
	},
}

func init() {
	nodeCmd.AddCommand(shutdown_nodeCmd)
}
