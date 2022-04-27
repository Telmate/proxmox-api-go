package get

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "With shows the current configuration an item in proxmox",
}

func init() {
	cli.RootCmd.AddCommand(getCmd)
}
