package group

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/cli/helpers"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var group_setCmd = &cobra.Command{
	Use:   "set GROUP [ USERIDS ]",
	Short: "Set the members of the specified group",
	Long: `Adds and removes users, so the only the specified users will be members of the group.
	When no users are provided all users will be removed from the group.`,
	Example: "clear myGroup admin@pve,root@pam",
	Args:    cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var users *[]pveSDK.UserID
		var err error
		if len(args) == 2 {
			users, err = helpers.ParseUserIDs(args[1])
			if err != nil {
				return err
			}
		}
		if err = cli.NewClient().New().Group.Set(cli.Context(), pveSDK.ConfigGroup{Name: pveSDK.GroupName(args[0]), Members: users}); err != nil {
			return err
		}
		cli.PrintItemUpdated(member_GroupCmd.OutOrStdout(), args[0], "Group membership of")
		return nil
	}}

func init() {
	member_GroupCmd.AddCommand(group_setCmd)
}
