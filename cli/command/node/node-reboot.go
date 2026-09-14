package node

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var reboot_nodeCmd = &cobra.Command{
	Use:   "reboot NODE",
	Short: "Reboots the specified node",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		node := cli.RequiredIDset(args, 0, "node")
		if err = cli.NewClient().New().Node.Reboot(cli.Context(), proxmox.NodeName(node)); err != nil {
			return
		}
		cli.RootCmd.Printf("Node %s is rebooting", node)
		return
	},
}

func init() {
	nodeCmd.AddCommand(reboot_nodeCmd)
}
