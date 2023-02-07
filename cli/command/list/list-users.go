package list

import (
	"encoding/json"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Prints a list of Users in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		groups, _ := cmd.Flags().GetBool("groups")
		users, err := proxmox.ListUsers(c, groups)
		if err != nil {
			return
		}
		output, err := json.Marshal(users)
		if err != nil {
			return
		}
		fmt.Fprintln(listCmd.OutOrStdout(), string(output))
		return
	},
}

func init() {
	listCmd.AddCommand(list_usersCmd)
	list_usersCmd.PersistentFlags().Bool("groups", false, "Result will include group membership")
}
