package list

import (
	"context"

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

func listRaw(ctx context.Context, IDtype string) {
	c := cli.NewClient()
	var list interface{}
	var err error
	switch IDtype {
	case "AcmeAccounts":
		list, err = c.GetAcmeAccountList(ctx)
	case "AcmePlugins":
		list, err = c.GetAcmePluginList(ctx)
	case "MetricServers":
		list, err = c.GetMetricsServerList(ctx)
	case "Nodes":
		list, err = c.GetNodeList(ctx)
	case "Resources":
		list, err = c.GetResourceList(ctx, "")
	case "Storages":
		list, err = c.GetStorageList(ctx)
	}
	cli.LogFatalListing(IDtype, err)
	cli.PrintRawJson(listCmd.OutOrStdout(), list)
}
