package iso

import (
	"github.com/perimeter-81/proxmox-api-go/cli"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var iso_downloadCmd = &cobra.Command{
	Use:   "download NODE STORAGE URL FILENAME [CHECKSUMALGORITHM] [CHECKSUM]",
	Short: "download iso file from URL",
	Args:  cobra.RangeArgs(4, 7),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		c := cli.NewClient()
		config := proxmox.ConfigContent_Iso{
			Node:              cli.RequiredIDset(args, 0, "Node"),
			Storage:           cli.RequiredIDset(args, 1, "Storage"),
			DownloadUrl:       cli.RequiredIDset(args, 2, "URL"),
			Filename:          cli.RequiredIDset(args, 3, "Filename"),
			ChecksumAlgorithm: cli.OptionalIDset(args, 4),
			Checksum:          cli.OptionalIDset(args, 5),
		}
		err = config.Validate()
		if err != nil {
			return
		}
		err = proxmox.DownloadIsoFromUrl(c, config)
		if err != nil {
			return
		}
		cli.PrintItemCreated(isoCmd.OutOrStdout(), config.Filename, "ISO file")
		return
	},
}

func init() {
	isoCmd.AddCommand(iso_downloadCmd)
}
