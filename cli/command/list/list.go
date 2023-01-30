package list

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
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
	case "Guests":
		list, err = c.GetVmList()
	case "MetricServers":
		list, err = c.GetMetricsServerList()
	case "Nodes":
		list, err = c.GetNodeList()
	case "Pools":
		list, err = c.GetPoolList()
	case "Storages":
		list, err = c.GetStorageList()
	case "Users":
		list, err = proxmox.ListUsers(c)
	}
	cli.LogFatalListing(IDtype, err)
	cli.PrintRawJson(listCmd.OutOrStdout(), list)
}
