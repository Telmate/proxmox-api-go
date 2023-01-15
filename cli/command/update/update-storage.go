package update

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var update_storageCmd = &cobra.Command{
	Use:   "storage STORAGEID",
	Short: "Updates the configuration of the specified Storage Backend.",
	Long: `Updates the configuration of the specified Storage Backend.
The config can be set with the --file flag or piped from stdin.
For config examples see "example storage"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "StorageID")
		config, err := proxmox.NewConfigStorageFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		c := cli.NewClient()
		err = config.UpdateWithValidate(id, c)
		if err != nil {
			return
		}
		cli.PrintItemUpdated(updateCmd.OutOrStdout(), id, "Storage")
		return
	},
}

func init() {
	updateCmd.AddCommand(update_storageCmd)
}
