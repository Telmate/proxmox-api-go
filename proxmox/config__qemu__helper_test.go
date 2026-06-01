package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type qemuTests struct {
	Invalid qemuInvalid
	Valid   qemuValid
}

type qemuInvalid struct {
	Create       []qemuInvalidCreate
	CreateUpdate []qemuInvalidUpdate
	Update       []qemuInvalidUpdate
}

type qemuInvalidCreate struct {
	config ConfigQemu
	output map[string]string

	// When nil we only check that we have an error instead of ckecking the exact error
	err     error
	version Version
	name    string
}

var _ qemuInvalidInterfaceCreate = (*qemuInvalidCreate)(nil)

func (c *qemuInvalidCreate) Config() ConfigQemu        { return c.config }
func (c *qemuInvalidCreate) Err() error                { return c.err }
func (c *qemuInvalidCreate) Name() string              { return c.name }
func (c *qemuInvalidCreate) Output() map[string]string { return c.output }
func (u *qemuInvalidCreate) Version() Version          { return u.version }

type qemuInvalidUpdate struct {
	config  ConfigQemu
	current ConfigQemu
	output  map[string]string

	// When nil we only check that we have an error instead of ckecking the exact error
	err     error
	version Version
	name    string
}

var _ qemuInvalidInterfaceCreate = (*qemuInvalidUpdate)(nil)

func (u *qemuInvalidUpdate) Config() ConfigQemu        { return u.config }
func (u *qemuInvalidUpdate) Err() error                { return u.err }
func (u *qemuInvalidUpdate) Name() string              { return u.name }
func (u *qemuInvalidUpdate) Output() map[string]string { return u.output }
func (u *qemuInvalidUpdate) Version() Version          { return u.version }

type qemuValid struct {
	Create       []qemuValidCreate
	CreateUpdate []qemuValidUpdate
	Update       []qemuValidUpdate
}

type qemuInvalidInterfaceCreate interface {
	Config() ConfigQemu
	Err() error
	Name() string
	Output() map[string]string
	Version() Version
}

type qemuValidCreate struct {
	output  map[string]string
	config  ConfigQemu
	version Version
	name    string
}

var _ qemuValidInterfaceCreate = (*qemuValidCreate)(nil)

func (c *qemuValidCreate) Name() string              { return c.name }
func (c *qemuValidCreate) Config() ConfigQemu        { return c.config }
func (c *qemuValidCreate) Output() map[string]string { return c.output }
func (c *qemuValidCreate) Version() Version          { return c.version }

type qemuValidUpdate struct {
	output   map[string]string
	config   ConfigQemu
	current  ConfigQemu
	current2 configQemuUpdate
	version  Version
	name     string
}

var _ qemuValidInterfaceCreate = (*qemuValidUpdate)(nil)

func (u *qemuValidUpdate) Name() string              { return u.name }
func (u *qemuValidUpdate) Config() ConfigQemu        { return u.config }
func (u *qemuValidUpdate) Output() map[string]string { return u.output }
func (u *qemuValidUpdate) Version() Version          { return u.version }

type qemuValidInterfaceCreate interface {
	Config() ConfigQemu
	Name() string
	Output() map[string]string
	Version() Version
}

// runs:
// Validate
// validateCreate
// validateUpdate
// mapToApiCreate
// mapToApiUpdate
func qemuTestHelper(t *testing.T, testData func() qemuTests) {
	t.Helper()
	testInvalidCreate, testInvalidUpdate, testValidCreate, testValidUpdate := qemuFormatTests(testData)
	refrenceInvalidCreate, refrenceInvalidUpdate, refrenceValidCreate, refrenceValidUpdate := qemuFormatTests(testData)

	for i := range testInvalidCreate {
		t.Run("invalid/create/"+testInvalidCreate[i].Name()+"/validate", func(*testing.T) {
			config := testInvalidCreate[i].Config()
			err := config.validateCreate()
			require.Error(t, err)
			if testInvalidCreate[i].Err() != nil {
				require.Equal(t, testInvalidCreate[i].Err(), err)
			}
			err = config.Validate(nil, testInvalidCreate[i].Version())
			require.Error(t, err)
			if testInvalidCreate[i].Err() != nil {
				require.Equal(t, testInvalidCreate[i].Err(), err)
			}
			require.Equal(t, refrenceInvalidCreate[i].Config(), config, "mutated input config")
		})
		t.Run("invalid/create/"+testInvalidCreate[i].Name()+"/mapToApi", func(*testing.T) {
			config := testInvalidCreate[i].Config()
			_, body := config.mapToApiCreate(testInvalidCreate[i].Version())
			if body == nil {
				var emptyMap map[string]string
				require.Equal(t, testInvalidCreate[i].Output(), emptyMap)
			} else {
				testParamsEqualRaw(t, testInvalidCreate[i].Output(), body)
			}
			require.Equal(t, refrenceInvalidCreate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testInvalidUpdate {
		t.Run("invalid/update/"+testInvalidUpdate[i].Name()+"/validate", func(*testing.T) {
			config := testInvalidUpdate[i].Config()
			err := config.validateUpdate(&testInvalidUpdate[i].current)
			require.Error(t, err)
			if testInvalidUpdate[i].Err() != nil {
				require.Equal(t, testInvalidUpdate[i].Err(), err)
			}
			require.Error(t, err)
			err = config.Validate(&testInvalidUpdate[i].current, testInvalidUpdate[i].Version())
			require.Error(t, err)
			if testInvalidUpdate[i].Err() != nil {
				require.Equal(t, testInvalidUpdate[i].Err(), err)
			}
			require.Equal(t, refrenceInvalidUpdate[i].Config(), config, "mutated input config")
		})
		t.Run("invalid/update/"+testInvalidUpdate[i].Name()+"/mapToApi", func(*testing.T) {
			config := testInvalidUpdate[i].Config()
			require.NotPanics(t, func() {
				config.mapToApiUpdate(&testInvalidUpdate[i].current, configQemuUpdate{}, testInvalidUpdate[i].Version())
			})
			require.Equal(t, refrenceInvalidUpdate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testValidCreate {
		t.Run("valid/create/"+testValidCreate[i].Name()+"/validate", func(*testing.T) {
			config := testValidCreate[i].Config()
			err := config.validateCreate()
			require.NoError(t, err)
			err = config.Validate(nil, testValidUpdate[i].Version())
			require.NoError(t, err)
			require.Equal(t, refrenceValidCreate[i].Config(), config, "mutated input config")
		})
		t.Run("valid/create/"+testValidCreate[i].Name()+"/mapToApi", func(*testing.T) {
			config := testValidCreate[i].Config()
			_, body := config.mapToApiCreate(testValidCreate[i].Version())
			if body == nil {
				var emptyMap map[string]string
				require.Equal(t, testValidCreate[i].Output(), emptyMap)
			} else {
				testParamsEqualRaw(t, testValidCreate[i].Output(), body)
			}
			require.Equal(t, refrenceValidCreate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testValidUpdate {
		t.Run("valid/update/"+testValidUpdate[i].Name()+"/validate", func(*testing.T) {
			config := testValidUpdate[i].Config()
			err := config.validateUpdate(&testValidUpdate[i].current)
			require.NoError(t, err)
			err = config.Validate(&testValidUpdate[i].current, testValidUpdate[i].Version())
			require.NoError(t, err)
			require.Equal(t, refrenceValidUpdate[i].Config(), config, "mutated input config")
		})
		t.Run("valid/update/"+testValidUpdate[i].Name()+"/mapToApi", func(*testing.T) {
			config := testValidUpdate[i].Config()
			_, body := config.mapToApiUpdate(&testValidUpdate[i].current, testValidUpdate[i].current2, testValidUpdate[i].Version())
			if body == nil {
				var emptyMap map[string]string
				require.Equal(t, testValidUpdate[i].Output(), emptyMap)
			} else {
				testParamsEqualRaw(t, testValidUpdate[i].Output(), body)
			}
			require.Equal(t, refrenceValidUpdate[i].Config(), config, "mutated input config")
		})
	}
}

func qemuTestInjected(
	t *testing.T, testData func() qemuTests,
	testFunc func(
		t *testing.T,
		config ConfigQemu,
		current *ConfigQemu,
		version Version,
		output map[string]string,
		expectedErr error,
		valid bool,
	),
) {
	t.Helper()
	testInvalidCreate, testInvalidUpdate, testValidCreate, testValidUpdate := qemuFormatTests(testData)
	refrenceInvalidCreate, refrenceInvalidUpdate, refrenceValidCreate, refrenceValidUpdate := qemuFormatTests(testData)

	for i := range testInvalidCreate {
		t.Run("invalid/create/"+testInvalidCreate[i].Name(), func(*testing.T) {
			config := testInvalidCreate[i].Config()
			testFunc(t, config, nil, testInvalidCreate[i].Version(), nil, testInvalidCreate[i].Err(), false)
			require.Equal(t, refrenceInvalidCreate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testInvalidUpdate {
		t.Run("invalid/update/"+testInvalidUpdate[i].Name()+"/validate", func(*testing.T) {
			config := testInvalidUpdate[i].Config()
			testFunc(t, config, &testInvalidUpdate[i].current, testInvalidUpdate[i].Version(), nil, testInvalidUpdate[i].Err(), false)
			require.Equal(t, refrenceInvalidUpdate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testValidCreate {
		t.Run("valid/create/"+testValidCreate[i].Name()+"/validate", func(*testing.T) {
			config := testValidCreate[i].Config()
			testFunc(t, config, nil, testValidCreate[i].Version(), testValidCreate[i].Output(), nil, true)
			require.Equal(t, refrenceValidCreate[i].Config(), config, "mutated input config")
		})
	}
	for i := range testValidUpdate {
		t.Run("valid/update/"+testValidUpdate[i].Name()+"/validate", func(*testing.T) {
			config := testValidUpdate[i].Config()
			testFunc(t, config, &testValidUpdate[i].current, testValidUpdate[i].Version(), testValidUpdate[i].Output(), nil, true)
			require.Equal(t, refrenceValidUpdate[i].Config(), config, "mutated input config")
		})
	}
}

func qemuFormatTests(tests func() qemuTests) ([]qemuInvalidInterfaceCreate, []qemuInvalidUpdate, []qemuValidInterfaceCreate, []qemuValidUpdate) {
	testData := tests()
	invalidCreate := make([]qemuInvalidInterfaceCreate, len(testData.Invalid.Create)+len(testData.Invalid.CreateUpdate))
	var index int
	if len(testData.Invalid.Create) > 0 {
		for index = range testData.Invalid.Create {
			invalidCreate[index] = qemuInvalidInterfaceCreate(&testData.Invalid.Create[index])
		}
		index++
	}
	for i := range testData.Invalid.CreateUpdate {
		invalidCreate[index+i] = qemuInvalidInterfaceCreate(&testData.Invalid.CreateUpdate[i])
	}
	invalidUpdate := append(testData.Invalid.Update, testData.Invalid.CreateUpdate...)
	index = 0
	validCreate := make([]qemuValidInterfaceCreate, len(testData.Valid.Create)+len(testData.Valid.CreateUpdate))
	if len(testData.Valid.Create) > 0 {
		for index = range testData.Valid.Create {
			validCreate[index] = qemuValidInterfaceCreate(&testData.Valid.Create[index])
		}
		index++
	}
	for i := range testData.Valid.CreateUpdate {
		validCreate[index+i] = qemuValidInterfaceCreate(&testData.Valid.CreateUpdate[i])
	}
	validUpdate := append(testData.Valid.Update, testData.Valid.CreateUpdate...)
	return invalidCreate, invalidUpdate, validCreate, validUpdate
}
