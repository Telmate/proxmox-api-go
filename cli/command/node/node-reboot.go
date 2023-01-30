package node

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var reboot_nodeCmd = &cobra.Command{
	Use:   "reboot NODE",
	Short: "Reboots the specified node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		node := cli.RequiredIDset(args, 0, "node")
		c := cli.NewClient()
		_, err = c.RebootNode(node)
		if err != nil {
			return
		}
		cli.RootCmd.Printf("Node %s is rebooting", node)
		return
	},
}

func init() {
	nodeCmd.AddCommand(reboot_nodeCmd)
}
