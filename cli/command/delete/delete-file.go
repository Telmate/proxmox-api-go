package delete

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var delete_fileCmd = &cobra.Command{
	Use:   "file NODE STORAGE TYPE FILE",
	Short: "Deletes the specified File",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		Type := proxmox.ContentType(args[2])
		if Type.Validate() != nil {
			return
		}
		err = proxmox.DeleteFile(c, args[0], proxmox.Content_File{
			Storage:     args[1],
			ContentType: Type,
			FilePath:    args[3],
		})
		if err != nil {
			return
		}
		cli.PrintItemDeleted(deleteCmd.OutOrStdout(), args[3], "File")
		return
	},
}

func init() {
	deleteCmd.AddCommand(delete_fileCmd)
}
