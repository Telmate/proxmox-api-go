package group

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var group_addCmd = &cobra.Command{
	Use:     "add GROUP USERIDS",
	Short:   "Add members to the specified group",
	Example: "add myGroup admin@pve,root@pam",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		users, err := proxmox.NewUserIDs(args[1])
		if err != nil {
			return
		}
		c := cli.NewClient()
		err = proxmox.GroupName(args[0]).AddUsersToGroup(users, c)
		if err != nil {
			return
		}
		cli.PrintItemUpdated(member_GroupCmd.OutOrStdout(), args[0], "Group membership of")
		return
	},
}

func init() {
	member_GroupCmd.AddCommand(group_addCmd)
}
