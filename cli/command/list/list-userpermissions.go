package list

import (
	"sort"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_userPermissionsCmd = &cobra.Command{
	Use:   "userpermissions USER PATH",
	Short: "Prints the list of permissions for the specified USER and PATH",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		path := args[1]
		var userID proxmox.UserID
		if err = userID.Parse(args[0]); err != nil {
			return
		}

		client := cli.NewClient()
		permissions, err := client.GetUserPermissions(cli.Context(), userID, path)
		sort.Strings(permissions)
		cli.PrintRawJson(listCmd.OutOrStdout(), permissions)
		return
	}}

func init() {
	listCmd.AddCommand(list_userPermissionsCmd)
}
