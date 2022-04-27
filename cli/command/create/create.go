package create

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "With this command you can create new items in proxmox",
}

func init() {
	cli.RootCmd.AddCommand(createCmd)
}
