package set

import (
	"encoding/json"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var set_userCmd = &cobra.Command{
	Use:   "user USERID PASSWORD",
	Short: "Sets the current state of a user",
	Long: `Sets the current state of a user.
Depending on the current state of the user, the user will be created or updated.
The config can be set with the --file flag or piped from stdin.
For config examples see "example user"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cli.RequiredIDset(args, 0, "UserID")
		var userID proxmox.UserID
		if err := userID.Parse(id); err != nil {
			return err
		}
		var config proxmox.ConfigUser
		if err := json.Unmarshal(cli.NewConfig(), &config); err != nil {
			return err
		}
		if len(args) > 1 {
			v := proxmox.UserPassword(args[1])
			config.Password = &v
		}
		config.User = userID
		if err := cli.NewClient().New().User.Set(cli.Context(), config); err != nil {
			return err
		}
		cli.PrintItemSet(setCmd.OutOrStdout(), id, "User")
		return nil
	}}

func init() {
	setCmd.AddCommand(set_userCmd)
}
