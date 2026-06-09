package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func testData_ConfigQemu_Watchdog_Api() qemuTestsApiFunc {
	return qemuTestsApiFunc(func() qemuTestsAPI {
		return qemuTestsAPI{
			createUpdate: []qemuTestCaseAPI{
				{name: `delete no effect`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Delete: true}}},
				{name: `full`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionPause)}},
					body: map[string]string{"watchdog": "model%3Di6300esb%2Caction%3Dpause"}}, // "model=i6300esb,action=pause"
				{name: `only Action`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogActionNone)}},
					body: map[string]string{"watchdog": "%2Caction%3Dnone"}}, // ",action=none"
				{name: `only Model`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Model: new(WatchdogModelI6300esb)}},
					body: map[string]string{"watchdog": "model%3Di6300esb"}}}, // "model=i6300esb"
			update: []qemuTestCaseAPI{
				{name: `delete`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Delete: true}},
					currentLegacy: ConfigQemu{Watchdog: &Watchdog{}},
					body:          map[string]string{"delete": "watchdog"}},
				{name: `change Action`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogActionDebug)}},
					currentLegacy: ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionShutdown)}},
					body: map[string]string{"watchdog": "model%3Di6300esb%2Caction%3Ddebug"}}, // "model=i6300esb,action=debug"
				{name: `change Model`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Model: new(WatchdogModelIb700)}},
					currentLegacy: ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionShutdown)}},
					body: map[string]string{"watchdog": "model%3Dib700%2Caction%3Dshutdown"}}, // "model=ib700,action=shutdown"
				{name: `change all`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelIb700),
						Action: new(WatchdogActionNone)}},
					currentLegacy: ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionShutdown)}},
					body: map[string]string{"watchdog": "model%3Dib700%2Caction%3Dnone"}}, // "model=ib700,action=none"
				{name: `change same no effect`,
					config: &ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionShutdown)}},
					currentLegacy: ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModelI6300esb),
						Action: new(WatchdogActionShutdown)}}}}}
	})
}

func testData_ConfigQemu_Watchdog_Validate() qemuTestTypeValidateFunc {
	return qemuTestTypeValidateFunc(func() (qemuTestTypeInvalid, qemuTestTypeValid) {
		invalid := qemuTestTypeInvalid{
			createUpdate: []qemuTestCaseInvalid{
				{name: `errors.New(Watchdog_Error_ModelRequired)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{}}),
					err:   errors.New(Watchdog_Error_ModelRequired)},
				{name: `errors.New(Watchdog_Error_ActionRequired)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Model: new(WatchdogModel("ib700"))}}),
					err: errors.New(Watchdog_Error_ActionRequired)},
				{name: `errors.New(WatchdogAction_Error)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModel("ib700")),
						Action: new(WatchdogAction("invalid")),
					}}),
					err: errors.New(WatchdogAction_Error)},
				{name: `create errors.New(WatchdogModel_Error)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Model:  new(WatchdogModel("invalid")),
						Action: new(WatchdogAction("none")),
					}}),
					err: errors.New(WatchdogModel_Error)}},
			update: []qemuTestCaseInvalid{
				{name: `errors.New(WatchdogAction_Error)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("invalid")),
						Model:  new(WatchdogModel("ib700"))}}),
					current: &ConfigQemu{Watchdog: &Watchdog{}},
					err:     errors.New(WatchdogAction_Error)},
				{name: `errors.New(WatchdogModel_Error)`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("none")),
						Model:  new(WatchdogModel("invalid"))}}),
					current: &ConfigQemu{Watchdog: &Watchdog{}},
					err:     errors.New(WatchdogModel_Error)}}}
		valid := qemuTestTypeValid{
			createUpdate: []qemuTestCaseValid{
				{name: `create minimal`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("reset")),
						Model:  new(WatchdogModel("ib700"))}})},
				{name: `delete no effect`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{Delete: true}})},
			},
			update: []qemuTestCaseValid{
				{name: `delete overwrites invalid`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("invalid")),
						Delete: true,
						Model:  new(WatchdogModel("invalid"))}}),
					current: &ConfigQemu{Watchdog: &Watchdog{}}},
				{name: `minimal`,
					input: testQemuBaseConfig_Validate(ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("poweroff")),
						Model:  new(WatchdogModel("ib700"))}}),
					current: &ConfigQemu{Watchdog: &Watchdog{
						Action: new(WatchdogAction("none")),
						Model:  new(WatchdogModel("i6300esb"))}}}}}
		return invalid, valid
	})
}

func Test_ConfigQemu_Watchdog_Api(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_Watchdog_Api().Test(t)
}

func Test_ConfigQemu_Watchdog_Validate(t *testing.T) {
	t.Parallel()
	testData_ConfigQemu_Watchdog_Validate().Test(t)
}

func Test_Watchdog_Validate(t *testing.T) {
	t.Parallel()
	validate := func(t *testing.T, config ConfigQemu, current *ConfigQemu, version Version, expectedErr error, valid bool) {
		t.Helper()
		var currentWatchdog *Watchdog
		if current != nil {
			currentWatchdog = current.Watchdog
		}
		err := config.Watchdog.Validate(currentWatchdog)
		if valid {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			if expectedErr != nil {
				require.Equal(t, expectedErr, err)
			}
		}
	}
	testData_ConfigQemu_Watchdog_Validate().Inject(t, validate)
}

func testData_ConfigQemu_Watchdog_Get() []qemuTestCaseGet {
	return []qemuTestCaseGet{
		{name: `invalid whitespace`,
			input:  map[string]any{"watchdog": string(" ")},
			output: testQemuBaseConfig_get(ConfigQemu{})},
		{name: ``,
			input: map[string]any{"watchdog": string("model=i6300esb,action=shutdown")},
			output: testQemuBaseConfig_get(ConfigQemu{Watchdog: &Watchdog{
				Model:  new(WatchdogModelI6300esb),
				Action: new(WatchdogActionShutdown),
			}}),
		},
	}
}
