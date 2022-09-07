package get

import (
	"github.com/spf13/cobra"
)

var get_metricserverCmd = &cobra.Command{
	Use:   "metricserver METRICSID",
	Short: "Gets the configuration of the specified MetricServer",
	RunE: func(cmd *cobra.Command, args []string) error {
		return GetConfig(args, "MetricServer")
	},
}

func init() {
	GetCmd.AddCommand(get_metricserverCmd)
}
