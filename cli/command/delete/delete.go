package delete

import (
	"fmt"

	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "With this command you can delete existing items from proxmox",
}

func init() {
	cli.RootCmd.AddCommand(deleteCmd)
}

func deleteID(args []string, IDtype string) (err error) {
	var exitStatus string
	id := cli.RequiredIDset(args, 0, IDtype+"ID")
	c := cli.NewClient()
	switch IDtype {
	case "AcmeAccount":
		exitStatus, err = c.DeleteAcmeAccount(id)
	case "Group":
		err = proxmox.GroupName(id).Delete(c)
	case "MetricServer":
		err = c.DeleteMetricServer(id)
	case "Pool":
		err = c.DeletePool(id)
	case "Storage":
		err = c.DeleteStorage(id)
	case "User":
		var userId proxmox.UserID
		userId, err = proxmox.NewUserID(id)
		if err != nil {
			return
		}
		err = proxmox.ConfigUser{User: userId}.DeleteUser(c)
	}
	if err != nil {
		if exitStatus != "" {
			err = fmt.Errorf("error deleting %s (%s): %v, error status: %s ", IDtype, id, err, exitStatus)
		}
		return
	}
	cli.PrintItemDeleted(deleteCmd.OutOrStdout(), id, IDtype)
	return
}
