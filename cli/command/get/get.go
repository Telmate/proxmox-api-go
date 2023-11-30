package get

import (
	"github.com/Bluearchive/proxmox-api-go/cli"
	"github.com/Bluearchive/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "get shows the current configuration an item in proxmox",
}

func init() {
	cli.RootCmd.AddCommand(GetCmd)
}

func getConfig(args []string, IDtype string) (err error) {
	id := cli.RequiredIDset(args, 0, IDtype+"ID")
	c := cli.NewClient()
	var config interface{}
	switch IDtype {
	case "AcmeAccount":
		config, err = proxmox.NewConfigAcmeAccountFromApi(id, c)
	case "Group":
		config, err = proxmox.NewConfigGroupFromApi(proxmox.GroupName(id), c)
	case "MetricServer":
		config, err = proxmox.NewConfigMetricsFromApi(id, c)
	case "Pool":
		config, err = c.GetPoolInfo(id)
	case "Storage":
		config, err = proxmox.NewConfigStorageFromApi(id, c)
	case "User":
		var userId proxmox.UserID
		userId, err = proxmox.NewUserID(id)
		if err != nil {
			return
		}
		config, err = proxmox.NewConfigUserFromApi(userId, c)
	}
	if err != nil {
		return
	}
	cli.PrintFormattedJson(GetCmd.OutOrStdout(), config)
	return
}
