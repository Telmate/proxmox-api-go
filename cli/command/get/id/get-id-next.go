package id

import (
	"fmt"
	"strconv"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var id_nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Returns the lowest available ID",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		var currentID *proxmox.GuestID
		if len(args) > 0 {
			currentIdInt, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			id := proxmox.GuestID(currentIdInt)
			currentID = &id
		}
		id, err := c.GetNextID(cli.Context(), currentID)
		if err != nil {
			return
		}
		fmt.Fprintf(idCmd.OutOrStdout(), "Getting Next Free ID: %d\n", id)
		return
	},
}

func init() {
	idCmd.AddCommand(id_nextCmd)
}
