package version

import (
	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the version of proxmox-ve",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		versionInfo, err := c.GetVersion(cli.Context())
		if err != nil {
			return
		}

		cli.PrintRawJson(cli.RootCmd.OutOrStdout(), versionInfo)
		return
	},
}

func init() {
	cli.RootCmd.AddCommand(VersionCmd)
}
