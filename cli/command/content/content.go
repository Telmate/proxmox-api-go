package content

import (
	"github.com/Bluearchive/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var ContentCmd = &cobra.Command{
	Use:   "content",
	Short: "With this command you can manage storage content",
}

func init() {
	cli.RootCmd.AddCommand(ContentCmd)
}
