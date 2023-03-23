package member

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

// TODO add feature to change pool membership
var MemberCmd = &cobra.Command{
	Use:   "member",
	Short: "Change Group and Pool membership",
}

func init() {
	cli.RootCmd.AddCommand(MemberCmd)
}
