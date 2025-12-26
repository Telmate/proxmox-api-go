package delete

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var delete_tokenCmd = &cobra.Command{
	Use:   "user TOKENID",
	Short: "Deletes the specified API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteID(cli.Context(), args, "TOKEN")
	}}

func init() { deleteCmd.AddCommand(delete_tokenCmd) }
