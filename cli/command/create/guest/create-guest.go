package guest

import (
	"context"

	"github.com/Telmate/proxmox-api-go/cli"
	"github.com/Telmate/proxmox-api-go/cli/command/create"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/spf13/cobra"
)

var guestCmd = &cobra.Command{
	Use:   "guest",
	Short: "With this command you can create new Lxc containers and Qemu virtual machines in proxmox",
}

func init() {
	create.CreateCmd.AddCommand(guestCmd)
}

func createGuest(ctx context.Context, args []string, IDtype string) (err error) {
	id := cli.ValidateGuestIDset(args, IDtype+"ID")
	node := cli.RequiredIDset(args, 1, "NodeID")
	vmr := proxmox.NewVmRef(id)
	vmr.SetNode(node)
	c := cli.NewClient()
	switch IDtype {
	case "LxcGuest":
		var config proxmox.ConfigLxc
		config, err = proxmox.NewConfigLxcFromJson(cli.NewConfig())
		if err != nil {
			return
		}
		err = config.CreateLxc(ctx, vmr, c)
	case "QemuGuest":
		// var config *proxmox.ConfigQemu
		// config, err = proxmox.NewConfigQemuFromJson(cli.NewConfig())
		// if err != nil {
		// return
		// }

		_, err = proxmox.ConfigQemu{
			CPU: &proxmox.QemuCPU{
				Affinity: util.Pointer([]uint{0, 1, 2}),
				Cores:    util.Pointer(proxmox.QemuCpuCores(4)),
				// Flags: &proxmox.CpuFlags{
				// 	AES: util.Pointer(proxmox.TriBoolFalse),
				// },
				Limit:        util.Pointer(proxmox.CpuLimit(65)),
				Numa:         util.Pointer(bool(true)),
				Sockets:      util.Pointer(proxmox.QemuCpuSockets(1)),
				Type:         util.Pointer(proxmox.CpuType("athlon")),
				Units:        util.Pointer(proxmox.CpuUnits(1024)),
				VirtualCores: util.Pointer(proxmox.CpuVirtualCores(2)),
			},
		}.Update(ctx, true, vmr, c)
		// err = config.Create(vmr, c)
	}
	if err != nil {
		return
	}
	cli.PrintItemCreated(guestCmd.OutOrStdout(), id.String(), IDtype)
	return
}
