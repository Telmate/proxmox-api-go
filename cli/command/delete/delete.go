package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "With this command you can delete existing items from proxmox",
}

func init() {
	cli.RootCmd.AddCommand(deleteCmd)
}
