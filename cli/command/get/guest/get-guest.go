package guest

import (
	"github.com/Telmate/proxmox-api-go/cli/command/get"
	"github.com/spf13/cobra"
)

var guestCmd = &cobra.Command{
	Use:   "guest",
	Short: "Commands to get information of guests on Proxmox",
}

func init() {
	get.GetCmd.AddCommand(guestCmd)
}
