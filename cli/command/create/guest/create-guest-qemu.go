package guest

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var guest_qemuCmd = &cobra.Command{
	Use:   "qemu GUESTID NODEID",
	Short: "Creates a new Guest System of the type Qemu on the specified Node",
	Long: `Creates a new Guest System of the type Qemu on the specified Node.
	The config can be set with the --file flag or piped from stdin.
	For config examples see "example guest qemu"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return createGuest(cli.Context(), args, "QemuGuest")
	},
}

func init() {
	guestCmd.AddCommand(guest_qemuCmd)
}
