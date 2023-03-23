package group

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var group_setCmd = &cobra.Command{
	Use:   "set GROUP [ USERIDS ]",
	Short: "Set the members of the specified group",
	Long: `Adds and removes users, so the only the specified users will be members of the group.
	When no users are provided all users will be removed from the group.`,
	Example: "clear myGroup admin@pve,root@pam",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var users *[]proxmox.UserID
		if len(args) == 2 {
			users, err = proxmox.NewUserIDs(args[1])
			if err != nil {
				return
			}
		}
		c := cli.NewClient()
		err = proxmox.GroupName(args[0]).SetMembers(users, c)
		if err != nil {
			return
		}
		cli.PrintItemUpdated(member_GroupCmd.OutOrStdout(), args[0], "Group membership of")
		return
	},
}

func init() {
	member_GroupCmd.AddCommand(group_setCmd)
}
