package list

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all items of the same kind from proxmox",
}

func init() {
	cli.RootCmd.AddCommand(listCmd)
}

func listRaw(IDtype string) {
	c := cli.NewClient()
	var list interface{}
	var err error
	switch IDtype {
	case "AcmeAccounts":
		list, err = c.GetAcmeAccountList()
	case "AcmePlugins":
		list, err = c.GetAcmePluginList()
	case "MetricServers":
		list, err = c.GetMetricsServerList()
	case "Nodes":
		list, err = c.GetNodeList()
	case "Storages":
		list, err = c.GetStorageList()
	}
	cli.LogFatalListing(IDtype, err)
	cli.PrintRawJson(listCmd.OutOrStdout(), list)
}
