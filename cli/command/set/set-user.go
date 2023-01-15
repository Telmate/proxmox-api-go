package set

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
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
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "UserID")
		userId, err := proxmox.NewUserID(id)
		if err != nil {
			return
		}
		config, err := proxmox.NewConfigUserFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		var password proxmox.UserPassword
		if len(args) > 1 {
			password = proxmox.UserPassword(args[1])
		}
		c := cli.NewClient()
		err = config.SetUser(userId, password, c)
		if err != nil {
			return
		}
		cli.PrintItemSet(setCmd.OutOrStdout(), id, "User")
		return
	},
}

func init() {
	setCmd.AddCommand(set_userCmd)
}
