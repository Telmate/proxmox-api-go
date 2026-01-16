package update

import (
	"encoding/json"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var update_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Updates the comment on the specified pool",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cli.RequiredIDset(args, 0, "PoolID")
		var config proxmox.ConfigPool
		if err := json.Unmarshal(cli.NewConfig(), &config); err != nil {
			return err
		}
		config.Name = proxmox.PoolName(id)
		if err := cli.NewClient().New().Pool.Update(cli.Context(), config); err != nil {
			return err
		}
		cli.PrintItemUpdated(updateCmd.OutOrStdout(), id, "Pool")
		return nil
	}}

func init() { updateCmd.AddCommand(update_poolCmd) }
