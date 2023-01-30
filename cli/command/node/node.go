package node

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "With this command you can manage existing nodes within proxmox",
}

func init() {
	cli.RootCmd.AddCommand(nodeCmd)
}
