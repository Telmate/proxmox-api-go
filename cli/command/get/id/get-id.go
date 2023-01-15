package id

import (
	"github.com/perimeter-81/proxmox-api-go/cli/command/get"
	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Commands to get information about guestIDs",
}

func init() {
	get.GetCmd.AddCommand(idCmd)
}
