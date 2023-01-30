package example

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "This function show examples of fully populated config files",
}

func init() {
	cli.RootCmd.AddCommand(exampleCmd)
}
