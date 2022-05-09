package delete

import (
	"github.com/spf13/cobra"
)

var delete_metricserverCmd = &cobra.Command{
	Use:   "metricserver METRICSID",
	Short: "Deletes the speciefied MetricServer",
	RunE: func(cmd *cobra.Command, args []string) error {
		return DeleteID(args, "MetricServer")
	},
}

func init() {
	deleteCmd.AddCommand(delete_metricserverCmd)
}
