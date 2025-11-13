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
		config, err = proxmox.NewConfigGroupFromApi(cli.Context(), proxmox.GroupName(id), c)
		if err != nil {
			return
		}
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
	case "User":
		var userId proxmox.UserID
		userId, err = proxmox.NewUserID(id)
		if err != nil {
			return
		}
		var rawConfig proxmox.RawConfigUser
		rawConfig, err = proxmox.NewRawConfigUserFromApi(cli.Context(), userId, c)
		if err != nil {
			return
		}
		config = rawConfig.Get()
	case "Vminfo":
		var vmr *proxmox.VmRef
		vmr = proxmox.NewVmRef(cli.ValidateGuestIDset(args, "GuestID"))

		c := cli.NewClient()
		c.CheckVmRef(cli.Context(), vmr)
		config, err = proxmox.NewConfigQemuFromApi(cli.Context(), vmr, c)
	}
	cli.PrintFormattedJson(GetCmd.OutOrStdout(), config)
	return
}
