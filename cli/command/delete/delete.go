package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "With this command you can delete existing items from proxmox",
}

func init() {
	cli.RootCmd.AddCommand(deleteCmd)
}

func DeleteID(args []string, IDtype string) (err error){
	id := cli.ValidateIDset(args, 0, IDtype+"ID")
	c := cli.NewClient()
	switch IDtype {
	case "MetricServer" :
		err = c.DeleteMetricServer(id)
	case "Pool" :
		err = c.DeletePool(id)
	case "Storage" :
		err = c.DeleteStorage(id)
	case "User" :
		err = c.DeleteUser(id)
	}
	if err != nil {
		return
	}
	cli.PrintItemDeleted(deleteCmd.OutOrStdout(), id, IDtype)
	return
}