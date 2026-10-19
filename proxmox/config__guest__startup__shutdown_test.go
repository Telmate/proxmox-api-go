package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_RawConfigQemu_StartupShutdown_Get() qemuTestGetFunc {
	return func() []qemuTestCaseGet {
		return []qemuTestCaseGet{
			{name: `Order`,
				input: rawConfigQemu{a: map[string]any{
					"startup": string("order=0")}},
				output: testQemuBaseConfig_get(ConfigQemu{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(0)),
						ShutdownTimeout: new(TimeDuration(-1)),
						StartupDelay:    new(TimeDuration(-1))}})},
			{name: `ShutdownTimeout`,
				input: rawConfigQemu{a: map[string]any{
					"startup": string("down=503")}},
				output: testQemuBaseConfig_get(ConfigQemu{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(-1)),
						ShutdownTimeout: new(TimeDuration(503)),
						StartupDelay:    new(TimeDuration(-1))}})},
			{name: `StartupDelay`,
				input: rawConfigQemu{a: map[string]any{
					"startup": string("up=7")}},
				output: testQemuBaseConfig_get(ConfigQemu{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(-1)),
						ShutdownTimeout: new(TimeDuration(-1)),
						StartupDelay:    new(TimeDuration(7))}})},
			{name: `all`,
				input: rawConfigQemu{a: map[string]any{
					"startup": string("up=8454,down=43742,order=75")}},
				output: testQemuBaseConfig_get(ConfigQemu{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(75)),
						ShutdownTimeout: new(TimeDuration(43742)),
						StartupDelay:    new(TimeDuration(8454))}})}}
	}
}

func testData_RawConfigLXC_StartupShutdown_Get() lxcTestGetFunc {
	return func() []lxcTestCaseGet {
		return []lxcTestCaseGet{
			{name: `Order`,
				input: rawConfigLXC{a: map[string]any{"startup": string("order=0")}},
				output: lxcGetBaseConfig(ConfigLXC{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(0)),
						ShutdownTimeout: new(TimeDuration(-1)),
						StartupDelay:    new(TimeDuration(-1))}})},
			{name: `ShutdownTimeout`,
				input: rawConfigLXC{a: map[string]any{"startup": string("down=503")}},
				output: lxcGetBaseConfig(ConfigLXC{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(-1)),
						ShutdownTimeout: new(TimeDuration(503)),
						StartupDelay:    new(TimeDuration(-1))}})},
			{name: `StartupDelay`,
				input: rawConfigLXC{a: map[string]any{"startup": string("up=7")}},
				output: lxcGetBaseConfig(ConfigLXC{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(-1)),
						ShutdownTimeout: new(TimeDuration(-1)),
						StartupDelay:    new(TimeDuration(7))}})},
			{name: `all`,
				input: rawConfigLXC{a: map[string]any{
					"startup": string("up=8454,down=43742,order=75")}},
				output: lxcGetBaseConfig(ConfigLXC{
					StartupShutdown: &StartupAndShutdown{
						Order:           new(GuestStartupOrder(75)),
						ShutdownTimeout: new(TimeDuration(43742)),
						StartupDelay:    new(TimeDuration(8454))}})}}
	}
}

func Test_RawConfigQemu_GetStartupShutdown(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_StartupShutdown_Get().Inject(t,
		func(t *testing.T, raw RawConfigQemu, output *ConfigQemu, _ error) {
			require.Equal(t, output.StartupShutdown, raw.GetStartupShutdown())
		})
}

func Test_RawConfigQemu_StartupShutdown_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigQemu_StartupShutdown_Get().Test(t)
}

func Test_RawConfigLXC_GetStartupShutdown(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_StartupShutdown_Get().Inject(t,
		func(t *testing.T, raw RawConfigLXC, output *ConfigLXC) {
			require.Equal(t, output.StartupShutdown, raw.GetStartupShutdown())
		})
}

func Test_RawConfigLXC_StartupShutdown_Get(t *testing.T) {
	t.Parallel()
	testData_RawConfigLXC_StartupShutdown_Get().Test(t)
}
