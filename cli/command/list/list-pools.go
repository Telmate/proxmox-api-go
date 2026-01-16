package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_poolsCmd = &cobra.Command{
	Use:   "pools",
	Short: "Prints a list of Pools in raw json format",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := cli.NewClient().New().Pool.List(cli.Context())
		if err != nil {
			return err
		}
		rawPools := raw.AsArray()

		pools := make([]proxmox.PoolName, len(rawPools))
		for i := range rawPools {
			pools[i] = rawPools[i].GetName()
		}
		cli.PrintFormattedJson(listCmd.OutOrStdout(), pools)
		return nil
	}}

func init() { listCmd.AddCommand(list_poolsCmd) }
