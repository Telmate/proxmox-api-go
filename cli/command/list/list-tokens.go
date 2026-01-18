package list

import (
	"encoding/json"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_tokensCmd = &cobra.Command{
	Use:   "tokens USERID",
	Short: "Prints a list of API tokens in raw json format",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cli.RequiredIDset(args, 0, "USERID")
		var user proxmox.UserID
		if err := user.Parse(id); err != nil {
			return err
		}
		data, err := cli.NewClient().New().ApiToken.List(cli.Context(), user)
		if err != nil {
			return err
		}
		rawArray := data.AsArray()
		tokens := make([]proxmox.ApiTokenConfig, len(rawArray))
		for i := range rawArray {
			tokens[i] = rawArray[i].Get()
		}
		var output []byte
		output, err = json.Marshal(tokens)
		if err != nil {
			return err
		}
		fmt.Fprintln(listCmd.OutOrStdout(), string(output))
		return nil
	}}

func init() { listCmd.AddCommand(list_tokensCmd) }
