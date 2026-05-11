package api

import (
	"context"
	"testing"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/stretchr/testify/require"
)

func GetConfig(t *testing.T, ctx context.Context, c *pveSDK.Client, guestID pveSDK.GuestID) *pveSDK.ConfigQemu {
	vmr := pveSDK.NewVmRef(pveSDK.GuestID(guestID))
	raw, err := pveSDK.NewRawConfigQemuFromApi(ctx, vmr, c)
	require.NoError(t, err)
	require.NotNil(t, raw)
	config, err := raw.Get(*vmr)
	require.NoError(t, err)
	require.NotNil(t, config)
	config.Smbios1 = ""                           // This field is unique and cannot be predicted, so we ignore it for the comparison.
	config.QemuUnusedDisks = pveSDK.QemuDevices{} // TODO include this field in test once reworked
	return config
}

func CheckConfig(t *testing.T, ctx context.Context, c *pveSDK.Client, guestID pveSDK.GuestID, expected *pveSDK.ConfigQemu) {
	config := GetConfig(t, ctx, c, guestID)
	require.EqualExportedValues(t, expected, config)
}

// ClearDIskID removes the disk ID of each created disk.
// During update the order disks are created in depends on the order on the update body.
// This order is currently random due to a map[string]any
func ClearDiskID(expected *pveSDK.ConfigQemu) {
	if expected.Disks != nil {
		if expected.Disks.Ide != nil {
			if expected.Disks.Ide.Disk_0 != nil && expected.Disks.Ide.Disk_0.Disk != nil {
				expected.Disks.Ide.Disk_0.Disk.Id = 0
			}
			if expected.Disks.Ide.Disk_1 != nil && expected.Disks.Ide.Disk_1.Disk != nil {
				expected.Disks.Ide.Disk_1.Disk.Id = 0
			}
		}
		if expected.Disks.Sata != nil {
			if expected.Disks.Sata.Disk_0 != nil && expected.Disks.Sata.Disk_0.Disk != nil {
				expected.Disks.Sata.Disk_0.Disk.Id = 0
			}
			if expected.Disks.Sata.Disk_1 != nil && expected.Disks.Sata.Disk_1.Disk != nil {
				expected.Disks.Sata.Disk_1.Disk.Id = 0
			}
		}
		if expected.Disks.Scsi != nil {
			if expected.Disks.Scsi.Disk_0 != nil && expected.Disks.Scsi.Disk_0.Disk != nil {
				expected.Disks.Scsi.Disk_0.Disk.Id = 0
			}
			if expected.Disks.Scsi.Disk_1 != nil && expected.Disks.Scsi.Disk_1.Disk != nil {
				expected.Disks.Scsi.Disk_1.Disk.Id = 0
			}
		}
		if expected.Disks.VirtIO != nil {
			if expected.Disks.VirtIO.Disk_0 != nil && expected.Disks.VirtIO.Disk_0.Disk != nil {
				expected.Disks.VirtIO.Disk_0.Disk.Id = 0
			}
			if expected.Disks.VirtIO.Disk_1 != nil && expected.Disks.VirtIO.Disk_1.Disk != nil {
				expected.Disks.VirtIO.Disk_1.Disk.Id = 0
			}
		}
	}
}
