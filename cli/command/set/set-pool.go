package set

import (
	"encoding/json"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var set_poolCmd = &cobra.Command{
	Use:   "pool POOLID",
	Short: "Sets the current state of the specified pool",
	Long: `Sets the current state of the specified pool.
Depending on the current state of the pool, the pool will be created or updated.
The config can be set with the --file flag or piped from stdin.
For config examples see "example pool"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := cli.RequiredIDset(args, 0, "PoolID")
		var config proxmox.ConfigPool
		if err := json.Unmarshal(cli.NewConfig(), &config); err != nil {
			return err
		}
		config.Name = proxmox.PoolName(id)
		if err := cli.NewClient().New().Pool.Set(cli.Context(), config); err != nil {
			return err
		}
		cli.PrintItemSet(setCmd.OutOrStdout(), id, "Pool")
		return nil
	}}

func init() { setCmd.AddCommand(set_poolCmd) }
