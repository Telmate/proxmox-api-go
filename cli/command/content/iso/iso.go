package iso

import (
	"github.com/perimeter-81/proxmox-api-go/cli/command/content"
	"github.com/spf13/cobra"
)

var isoCmd = &cobra.Command{
	Use:   "iso",
	Short: "With this command you can download iso files to a Proxmox storage.",
}

func init() {
	content.ContentCmd.AddCommand(isoCmd)
}
