package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_metricserverCmd = &cobra.Command{
	Use:   "metricserver METRICSID",
	Short: "Deletes the specified MetricServer",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "MetricServer")
	},
}

func init() {
	deleteCmd.AddCommand(delete_metricserverCmd)
}
