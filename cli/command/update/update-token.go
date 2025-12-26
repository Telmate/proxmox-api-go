package update

import (
	"encoding/json"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var update_tokenCmd = &cobra.Command{
	Use:   "token USERID",
	Short: "Updates the configuration of the specified API Token.",
	Long: `Updates the configuration of the specified API Token.
The config can be set with the --file flag or piped from stdin.
For config examples see "example token"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cli.RequiredIDset(args, 0, "UserID")
		var user proxmox.UserID
		if err := user.Parse(id); err != nil {
			return err
		}
		var config proxmox.ApiTokenConfig
		json.Unmarshal(cli.NewConfig(), &config)
		if err := cli.NewClient().New().ApiToken.Update(cli.Context(), user, config); err != nil {
			return err
		}
		cli.PrintItemUpdated(updateCmd.OutOrStdout(), proxmox.ApiTokenID{
			TokenName: config.Name,
			User:      user}.String(), "Token")
		return nil
	}}

func init() { updateCmd.AddCommand(update_tokenCmd) }
