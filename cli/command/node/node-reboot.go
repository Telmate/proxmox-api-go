package node

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var reboot_nodeCmd = &cobra.Command{
	Use:   "reboot NODE",
	Short: "Reboots the specified node",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		node := cli.RequiredIDset(args, 0, "node")
		c := cli.NewClient()
		_, err = c.RebootNode(node)
		if err != nil {
			return
		}
		cli.RootCmd.Printf("Node %s is shutting down", node)
		return
	},
}

func init() {
	nodeCmd.AddCommand(reboot_nodeCmd)
}
