package list

import (
	"encoding/json"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_groupsCmd = &cobra.Command{
	Use:   "groups",
	Short: "Prints a list of groups in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := cli.NewClient().New().Group.List(cli.Context())
		if err != nil {
			return err
		}
		rawGroups := raw.FormatArray()
		groups := make([]proxmox.ConfigGroup, len(rawGroups))
		for i := range rawGroups {
			groups[i] = rawGroups[i].Get()
		}
		var output []byte
		output, err = json.Marshal(groups)
		if err != nil {
			return err
		}
		fmt.Fprintln(listCmd.OutOrStdout(), string(output))
		return nil
	}}

func init() { listCmd.AddCommand(list_groupsCmd) }
