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
	var task proxmox.Task
	id := cli.RequiredIDset(args, 0, IDtype+"ID")
	c := cli.NewClient()
	switch IDtype {
	case "AcmeAccount":
		task, err = c.DeleteAcmeAccount(ctx, id)
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
		return fmt.Errorf("error deleting %s (%s): %v", IDtype, id, err)
	}
	task.WaitForCompletion(ctx, c)

	cli.PrintItemDeleted(deleteCmd.OutOrStdout(), id, IDtype)
	return
}
