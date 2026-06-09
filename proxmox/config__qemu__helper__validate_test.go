package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type qemuTestTypeValidateFunc func() (qemuTestTypeInvalid, qemuTestTypeValid)

func (tests qemuTestTypeValidateFunc) Test(t *testing.T) {
	t.Helper()
	test := tests.format()
	refrence := tests.format()
	for i := range test.invalidCreate {
		t.Run("invalid/create/"+test.invalidCreate[i].name, func(t *testing.T) {
			config := test.invalidCreate[i].input
			err := config.validateCreate()
			require.Error(t, err)
			if test.invalidCreate[i].err != nil {
				require.Equal(t, test.invalidCreate[i].err, err)
			}
			err = config.Validate(test.invalidCreate[i].current, test.invalidCreate[i].version)
			require.Error(t, err)
			if test.invalidCreate[i].err != nil {
				require.Equal(t, test.invalidCreate[i].err, err)
			}
			require.Equal(t, refrence.invalidCreate[i].input, config, "mutated input config")
		})
	}
	for i := range test.invalidUpdate {
		t.Run("invalid/update/"+test.invalidUpdate[i].name, func(t *testing.T) {
			config := test.invalidUpdate[i].input
			var current ConfigQemu
			if v := test.invalidUpdate[i].current; v != nil {
				current = *v
			}
			err := config.validateUpdate(&current)
			require.Error(t, err)
			if test.invalidUpdate[i].err != nil {
				require.Equal(t, test.invalidUpdate[i].err, err)
			}
			err = config.Validate(&current, test.invalidUpdate[i].version)
			require.Error(t, err)
			if test.invalidUpdate[i].err != nil {
				require.Equal(t, test.invalidUpdate[i].err, err)
			}
			require.Equal(t, refrence.invalidUpdate[i].input, config, "mutated input config")
		})
	}
	for i := range test.validCreate {
		t.Run("valid/create/"+test.validCreate[i].name, func(t *testing.T) {
			config := test.validCreate[i].input
			err := config.validateCreate()
			require.NoError(t, err)
			err = config.Validate(test.validCreate[i].current, test.validCreate[i].version)
			require.NoError(t, err)
			require.Equal(t, refrence.validCreate[i].input, config, "mutated input config")
		})
	}
	for i := range test.validUpdate {
		t.Run("valid/update/"+test.validUpdate[i].name, func(t *testing.T) {
			config := test.validUpdate[i].input
			var current ConfigQemu
			if v := test.validUpdate[i].current; v != nil {
				current = *v
			}
			err := config.validateUpdate(&current)
			require.NoError(t, err)
			err = config.Validate(&current, test.validUpdate[i].version)
			require.NoError(t, err)
			require.Equal(t, refrence.validUpdate[i].input, config, "mutated input config")
		})
	}
}

func (tests qemuTestTypeValidateFunc) Inject(t *testing.T,
	testFunc func(
		t *testing.T,
		config ConfigQemu,
		current *ConfigQemu,
		version Version,
		expectedErr error,
		valid bool,
	),
) {
	t.Helper()
	test := tests.format()
	refrence := tests.format()
	if testFunc != nil {
		for i := range test.invalidCreate {
			t.Run("create/invalid/"+test.invalidCreate[i].name, func(t *testing.T) {
				config := test.invalidCreate[i].input
				testFunc(t, config, test.invalidCreate[i].current, test.invalidCreate[i].version, test.invalidCreate[i].err, false)
				require.Equal(t, refrence.invalidCreate[i].input, config, "mutated input config")
			})
		}
		for i := range test.invalidUpdate {
			t.Run("update/invalid/"+test.invalidUpdate[i].name, func(t *testing.T) {
				config := test.invalidUpdate[i].input
				testFunc(t, config, test.invalidUpdate[i].current, test.invalidUpdate[i].version, test.invalidUpdate[i].err, false)
				require.Equal(t, refrence.invalidUpdate[i].input, config, "mutated input config")
			})
		}
		for i := range test.validCreate {
			t.Run("create/valid/"+test.validCreate[i].name, func(t *testing.T) {
				config := test.validCreate[i].input
				testFunc(t, config, test.validCreate[i].current, test.validCreate[i].version, nil, true)
				require.Equal(t, refrence.validCreate[i].input, config, "mutated input config")
			})
		}
		for i := range test.validUpdate {
			t.Run("update/valid/"+test.validUpdate[i].name, func(t *testing.T) {
				config := test.validUpdate[i].input
				testFunc(t, config, test.validUpdate[i].current, test.validUpdate[i].version, nil, true)
				require.Equal(t, refrence.validUpdate[i].input, config, "mutated input config")
			})
		}
	}
}

func (tests qemuTestTypeValidateFunc) format() struct {
	invalidCreate []qemuTestCaseInvalid
	invalidUpdate []qemuTestCaseInvalid
	validCreate   []qemuTestCaseValid
	validUpdate   []qemuTestCaseValid
} {
	invalid, valid := tests()
	return struct {
		invalidCreate []qemuTestCaseInvalid
		invalidUpdate []qemuTestCaseInvalid
		validCreate   []qemuTestCaseValid
		validUpdate   []qemuTestCaseValid
	}{
		invalidCreate: append(invalid.create, invalid.createUpdate...),
		invalidUpdate: append(invalid.update, invalid.createUpdate...),
		validCreate:   append(valid.create, valid.createUpdate...),
		validUpdate:   append(valid.update, valid.createUpdate...),
	}
}
