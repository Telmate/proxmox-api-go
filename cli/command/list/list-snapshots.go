package list

import (
	"fmt"
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var list_snapshotsCmd = &cobra.Command{
	Use:   "snapshots GuestID",
	Short: "Prints a list of QemuSnapshots in raw json format",
	Run: func(cmd *cobra.Command, args []string) {
		id := cli.ValidateExistinGuestID(args, 0)
		c := cli.NewClient()
		vmr := proxmox.NewVmRef(id)
		_, err := c.GetVmInfo(vmr)
		cli.LogFatalError(err)
		jbody, _, err := c.ListQemuSnapshot(vmr)
		cli.LogFatalError(err)
		temp := jbody["data"].([]interface{})
		if len(temp) == 1 {
			fmt.Printf("Guest with ID (%d) has no snapshots",id)
		} else {
			for _, e := range temp {
				snapshotName := e.(map[string]interface{})["name"].(string)
				if snapshotName != "current" {
					fmt.Println(snapshotName)
				}
			}
		}
	},
}

func init() {
	listCmd.AddCommand(list_snapshotsCmd)
}
