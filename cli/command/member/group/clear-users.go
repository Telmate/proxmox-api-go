package group

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var group_clearCmd = &cobra.Command{
	Use:   "clear GROUP",
	Short: "Remove all user from the specified group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		err = proxmox.GroupName(args[0]).RemoveAllUsersFromGroup(c)
		if err != nil {
			return
		}
		cli.PrintItemUpdated(member_GroupCmd.OutOrStdout(), args[0], "Group membership of")
		return
	},
}

func init() {
	member_GroupCmd.AddCommand(group_clearCmd)
}
