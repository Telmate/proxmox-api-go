package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type qemuTestsApiFunc func() qemuTestsAPI

func (tests qemuTestsApiFunc) Test(t *testing.T) {
	t.Helper()
	test := tests.format()
	refrence := tests.format()
	for i := range test.create {
		t.Run("create/"+test.create[i].name, func(*testing.T) {
			config := test.create[i].config
			_, output := config.mapToApiCreate(test.create[i].version)
			testParamsEqualRaw(t, test.create[i].body, output)
			require.Equal(t, refrence.create[i].config, config, "mutated input config")
		})
	}
	for i := range test.update {
		t.Run("update/"+test.update[i].name, func(*testing.T) {
			config := test.update[i].config
			_, output := config.mapToApiUpdate(&test.update[i].currentLegacy, test.update[i].currentUpdate, test.update[i].version)
			testParamsEqualRaw(t, test.update[i].body, output)
			require.Equal(t, refrence.update[i].config, config, "mutated input config")
		})
	}
}

func (tests qemuTestsApiFunc) format() struct {
	create []qemuTestCaseAPI
	update []qemuTestCaseAPI
} {
	data := tests()
	return struct {
		create []qemuTestCaseAPI
		update []qemuTestCaseAPI
	}{
		create: append(data.create, data.createUpdate...),
		update: append(data.update, data.createUpdate...),
	}
}
