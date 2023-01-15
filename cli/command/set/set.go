package set

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "This command sets the current configuration of an item",
	Long: `This command sets the current configuration of an item.
Depending on if the item already exists the item will be created or updated.`,
}

func init() {
	cli.RootCmd.AddCommand(setCmd)
}
