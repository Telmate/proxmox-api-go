package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_Architecture_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `aarch64`,
				input:  rawConfigQemu{a: map[string]any{"arch": string("aarch64")}},
				output: testQemuBaseConfig_get(ConfigQemu{Architecture: new(QemuCpuArchitectureArm64)})},
			{name: `x86_64`,
				input:  rawConfigQemu{a: map[string]any{"arch": string("x86_64")}},
				output: testQemuBaseConfig_get(ConfigQemu{Architecture: new(QemuCpuArchitectureAmd64)})},
		}
	}
}

func Test_RawConfigQemu_GetArchitecture(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Architecture_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.Architecture, raw.GetArchitecture())
		})
}

func Test_RawConfigQemu_Architecture_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_Architecture_Get().Test(t)
}
