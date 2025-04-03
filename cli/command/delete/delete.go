package delete

import (
	"context"
	"fmt"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "With this command you can delete existing items from proxmox",
}

func init() {
	cli.RootCmd.AddCommand(deleteCmd)
}

func deleteID(ctx context.Context, args []string, IDtype string) (err error) {
	var exitStatus string
	id := cli.RequiredIDset(args, 0, IDtype+"ID")
	c := cli.NewClient()
	var task proxmox.Task
	switch IDtype {
	case "AcmeAccount":
		task, err = c.DeleteAcmeAccount(ctx, id)
		if err != nil {
			return
		}
		err = task.WaitForCompletion()
	case "Group":
		err = proxmox.GroupName(id).Delete(ctx, c)
	case "MetricServer":
		err = c.DeleteMetricServer(ctx, id)
	case "Pool":
		err = proxmox.PoolName(id).Delete(ctx, c)
	case "Storage":
		err = c.DeleteStorage(ctx, id)
	case "User":
		var userId proxmox.UserID
		userId, err = proxmox.NewUserID(id)
		if err != nil {
			return
		}
		err = proxmox.ConfigUser{User: userId}.DeleteUser(ctx, c)
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
