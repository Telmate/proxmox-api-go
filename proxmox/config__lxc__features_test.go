package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigLXC_Features_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `CreateDeviceNodes Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("mknod=1")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(true),
					FUSE:              new(false),
					NFS:               new(false),
					Nesting:           new(false),
					SMB:               new(false)}}})},
			{name: `CreateDeviceNodes Unprivileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("mknod=1"),
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: new(true),
						FUSE:              new(false),
						KeyCtl:            new(false),
						Nesting:           new(false)}},
					Privileged: new(false)})},
			{name: `FUSE Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("fuse=1")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(true),
					NFS:               new(false),
					Nesting:           new(false),
					SMB:               new(false)}}})},
			{name: `FUSE Unprivileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("fuse=1"),
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: new(false),
						FUSE:              new(true),
						KeyCtl:            new(false),
						Nesting:           new(false)}},
					Privileged: new(false)})},
			{name: `KeyCtl Unprivileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("keyctl=1"),
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: new(false),
						FUSE:              new(false),
						KeyCtl:            new(true),
						Nesting:           new(false)}},
					Privileged: new(false)})},
			{name: `NFS Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("mount=nfs")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(false),
					NFS:               new(true),
					Nesting:           new(false),
					SMB:               new(false)}}})},
			{name: `NFS and SMB Privileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("mount=nfs;cifs"),
					"unprivileged": float64(0)}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(false),
					NFS:               new(true),
					Nesting:           new(false),
					SMB:               new(true)}}})},
			{name: `Nesting Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("nesting=1")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(false),
					NFS:               new(false),
					Nesting:           new(true),
					SMB:               new(false)}}})},
			{name: `Nesting Unprivileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("nesting=1"),
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Features: &LxcFeatures{Unprivileged: &UnprivilegedFeatures{
						CreateDeviceNodes: new(false),
						FUSE:              new(false),
						KeyCtl:            new(false),
						Nesting:           new(true)}},
					Privileged: new(false)})},
			{name: `SMB Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("mount=cifs")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(false),
					NFS:               new(false),
					Nesting:           new(false),
					SMB:               new(true)}}})},
			{name: `SMB and NFS Privileged`,
				input: rawConfigLXC{a: map[string]any{"features": string("mount=cifs;nfs")}},
				output: lxcGetBaseConfig(ConfigLXC{Features: &LxcFeatures{Privileged: &PrivilegedFeatures{
					CreateDeviceNodes: new(false),
					FUSE:              new(false),
					NFS:               new(true),
					Nesting:           new(false),
					SMB:               new(true)}}})},
			{name: `multiple unprivileged`,
				input: rawConfigLXC{a: map[string]any{
					"features":     string("keyctl=0,mknod=1,nesting=1,fuse=0"),
					"unprivileged": float64(1)}},
				output: lxcGetBaseConfig(ConfigLXC{
					Privileged: new(false),
					Features: &LxcFeatures{
						Unprivileged: &UnprivilegedFeatures{
							CreateDeviceNodes: new(true),
							FUSE:              new(false),
							KeyCtl:            new(false),
							Nesting:           new(true)}}})}}
	}
}

func Test_RawConfigLXC_GetFeatures(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Features_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.Features, raw.GetFeatures())
		})
}

func Test_RawConfigLXC_Features_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_Features_Get().Test(t)
}
