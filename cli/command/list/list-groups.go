package list

import (
	"encoding/json"
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Prints a list of groups in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		groups, err := proxmox.ListGroups(c)
		if err != nil {
			return
		}
		output, err := json.Marshal(groups)
		if err != nil {
			return
		}
		fmt.Fprintln(listCmd.OutOrStdout(), string(output))
		return
	},
}

func init() {
	listCmd.AddCommand(list_groupsCmd)
}
