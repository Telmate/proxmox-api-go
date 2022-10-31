package delete

import (
	"github.com/spf13/cobra"
)

var delete_metricserverCmd = &cobra.Command{
	Use:   "metricserver METRICSID",
	Short: "Deletes the specified MetricServer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(args, "MetricServer")
	},
}

func init() {
	deleteCmd.AddCommand(delete_metricserverCmd)
}
