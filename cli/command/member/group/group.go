package group

import (
	"github.com/perimeter-81/proxmox-api-go/cli/command/member"

	"github.com/spf13/cobra"
)

var member_GroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Change Group membership",
}

func init() {
	member.MemberCmd.AddCommand(member_GroupCmd)
}
