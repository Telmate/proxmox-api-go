package set

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var (
	set_groupCmd = &cobra.Command{
		Use:   "group GROUPNAME [COMMENT]",
		Short: "Sets the current state of a group",
		Long: `Sets the current state of a group.
	Depending on the current state of the group, the user will be created or updated.
	When the "members" flag in not specified the group memberships will not be updated.
	Specifying the "members" flag and not populating it will remove all members from the group.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tmpGroupName := cli.RequiredIDset(args, 0, "Groupname")
			groupname := proxmox.GroupName(tmpGroupName)
			var comment string
			if len(args) > 1 {
				comment = args[1]
			}
			var formattedMembers *[]proxmox.UserID
			if cmd.Flags().Changed("members") {
				members, _ := cmd.Flags().GetString("members")
				formattedMembers, err = proxmox.NewUserIDs(members)
				if err != nil {
					return
				}
			}
			config := proxmox.ConfigGroup{
				Name:    groupname,
				Comment: comment,
				Members: formattedMembers,
			}
			err = config.Set(cli.NewClient())
			if err != nil {
				return
			}
			cli.PrintItemSet(setCmd.OutOrStdout(), tmpGroupName, "Group")
			return
		},
	}
)

func init() {
	setCmd.AddCommand(set_groupCmd)
	set_groupCmd.PersistentFlags().String("members", "", "Group members")
}
