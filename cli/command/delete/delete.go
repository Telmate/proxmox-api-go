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
	switch IDtype {
	case "AcmeAccount":
		exitStatus, err = c.DeleteAcmeAccount(ctx, id)
	case "AcmePlugin":
		err = c.DeleteAcmePlugin(ctx, id)
	case "Group":
		_, err = c.New().Group.Delete(ctx, proxmox.GroupName(id))
	case "MetricServer":
		err = c.DeleteMetricServer(ctx, id)
	case "Pool":
		err = proxmox.PoolName(id).Delete(ctx, c)
	case "Storage":
		err = c.DeleteStorage(ctx, id)
	case "Token":
		var token proxmox.ApiTokenID
		if err = token.Parse(id); err != nil {
			return
		}
		_, err = c.New().ApiToken.Delete(ctx, token)
	case "User":
		var user proxmox.UserID
		if err = user.Parse(id); err != nil {
			return
		}
		c.New().User.Delete(ctx, user)
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
