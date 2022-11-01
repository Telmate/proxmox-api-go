package list

import (
	"github.com/spf13/cobra"
)

var list_metricserversCmd = &cobra.Command{
	Use:   "metricservers",
	Short: "Prints a list of MetricServers in raw json format",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		listRaw("MetricServers")
	},
}

func init() {
	listCmd.AddCommand(list_metricserversCmd)
}
