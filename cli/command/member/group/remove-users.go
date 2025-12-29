package group

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/cli/helpers"
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var group_removeCmd = &cobra.Command{
	Use:     "remove GROUP USERIDS",
	Short:   "Remove members from the specified group",
	Example: "remove myGroup admin@pve,root@pam",
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		users, err := helpers.ParseUserIDs(args[1])
		if err != nil {
			return err
		}
		if err = cli.NewClient().New().Group.RemoveMembers(cli.Context(), []pveSDK.GroupName{pveSDK.GroupName(args[0])}, *users); err != nil {
			return err
		}
		cli.PrintItemUpdated(member_GroupCmd.OutOrStdout(), args[0], "Group membership of")
		return nil
	}}

func init() {
	member_GroupCmd.AddCommand(group_removeCmd)
}
