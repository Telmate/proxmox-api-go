package create

import (
	"encoding/json"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var create_tokenCmd = &cobra.Command{
	Use:   "token USERID",
	Short: "Creates a new API Token",
	Long: `Creates a new API Token.
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
		secret, err := cli.NewClient().New().ApiToken.Create(cli.Context(), user, config)
		if err != nil {
			return err
		}
		fmt.Fprintf(CreateCmd.OutOrStdout(), "%s\n", secret)
		return nil
	}}

func init() { CreateCmd.AddCommand(create_tokenCmd) }
