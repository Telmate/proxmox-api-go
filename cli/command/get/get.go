package get

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
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
	var config any
	switch IDtype {
	case "AcmeAccount":
		config, err = proxmox.NewConfigAcmeAccountFromApi(cli.Context(), id, c)
		if err != nil {
			return
		}
	case "AcmePlugin":
		config, err = proxmox.NewConfigAcmePluginFromApi(cli.Context(), id, c)
		if err != nil {
			return
		}
	case "Group":
		var rawConfig proxmox.RawGroupConfig
		rawConfig, err = c.New().Group.Read(cli.Context(), proxmox.GroupName(id))
		if err != nil {
			return
		}
		config = rawConfig.Get()
	case "MetricServer":
		config, err = proxmox.NewConfigMetricsFromApi(cli.Context(), id, c)
		if err != nil {
			return
		}
	case "Pool":
		var rawConfig proxmox.RawConfigPool
		rawConfig, err = proxmox.PoolName(id).Get(cli.Context(), c)
		if err != nil {
			return
		}
		config = rawConfig.Get()
	case "Storage":
		config, err = proxmox.NewConfigStorageFromApi(cli.Context(), id, c)
		if err != nil {
			return
		}
	case "Token":
		var token proxmox.ApiTokenID
		if err = token.Parse(id); err != nil {
			return
		}
		var rawConfig proxmox.RawApiTokenConfig
		rawConfig, err = c.New().ApiToken.Read(cli.Context(), token)
		if err != nil {
			return
		}
		config = rawConfig.Get()
	case "User":
		var userID proxmox.UserID
		if err = userID.Parse(id); err != nil {
			return
		}
		var rawConfig proxmox.RawConfigUser
		rawConfig, err = c.New().User.Read(cli.Context(), userID)
		if err != nil {
			return
		}
		config = rawConfig.Get()
	}
	cli.PrintFormattedJson(GetCmd.OutOrStdout(), config)
	return
}
