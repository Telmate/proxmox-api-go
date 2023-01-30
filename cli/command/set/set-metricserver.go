package set

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var set_metricserverCmd = &cobra.Command{
	Use:   "metricserver METRICSID",
	Short: "Sets the current state of a MetricServer",
	Long: `Sets the current state of a MetricServer.
Depending on the current state of the MetricServer, the MetricServer will be created or updated.
The config can be set with the --file flag or piped from stdin.
For config examples see "example metricserver"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		id := cli.RequiredIDset(args, 0, "MetricServerID")
		config, err := proxmox.NewConfigMetricsFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		c := cli.NewClient()
		err = config.SetMetrics(id, c)
		if err != nil {
			return
		}
		cli.PrintItemSet(setCmd.OutOrStdout(), id, "MericServer")
		return
	},
}

func init() {
	setCmd.AddCommand(set_metricserverCmd)
}
