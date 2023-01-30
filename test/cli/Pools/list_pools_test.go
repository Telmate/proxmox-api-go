package cli_pool_test

import (
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"testing"
)

// Test0
func Test_Pool_0_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   true,
		Args:     []string{"-i", "delete", "pool", "test-pool0"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Create_Without_Comment(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "create", "pool", "test-pool0"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_List(t *testing.T) {
	Test := cliTest.Test{
		Expected: `"test-pool0"`,
		ReqErr:   false,
		Contains: true,
		Args:     []string{"-i", "list", "pools"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Get_Without_Comment(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: `"comment"`,
		ReqErr:      false,
		NotContains: true,
		Args:        []string{"-i", "get", "pool", "test-pool0"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Update_Comment(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:   false,
		Contains: true,
		Expected: "(test-pool0)",
		Args:     []string{"-i", "update", "poolcomment", "test-pool0", "this is a comment"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Get_With_Comment(t *testing.T) {
	Test := cliTest.Test{
		Expected: `"this is a comment"`,
		ReqErr:   false,
		Contains: true,
		Args:     []string{"-i", "get", "pool", "test-pool0"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "pool", "test-pool0"},
	}
	Test.StandardTest(t)
}

func Test_Pool_0_Removed(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: `"test-pool0"`,
		ReqErr:      false,
		NotContains: true,
		Args:        []string{"-i", "list", "pools"},
	}
	Test.StandardTest(t)
}

// Test1
func Test_Pool_1_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   true,
		Args:     []string{"-i", "delete", "pool", "test-pool1"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Create_With_Comment(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "create", "pool", "test-pool1", "This is a comment"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Get_With_Comment(t *testing.T) {
	Test := cliTest.Test{
		Expected: `"This is a comment"`,
		ReqErr:   false,
		Contains: true,
		Args:     []string{"-i", "get", "pool", "test-pool1"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Update_Comment(t *testing.T) {
	Test := cliTest.Test{
		ReqErr:   false,
		Contains: true,
		Expected: "(test-pool1)",
		Args:     []string{"-i", "update", "poolcomment", "test-pool1"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Get_Without_Comment(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: `"comment"`,
		ReqErr:      false,
		NotContains: true,
		Args:        []string{"-i", "get", "pool", "test-pool1"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: "",
		ReqErr:   false,
		Args:     []string{"-i", "delete", "pool", "test-pool1"},
	}
	Test.StandardTest(t)
}

func Test_Pool_1_Removed(t *testing.T) {
	Test := cliTest.Test{
		NotExpected: `"test-pool1"`,
		ReqErr:      false,
		NotContains: true,
		Args:        []string{"-i", "list", "pools"},
	}
	Test.StandardTest(t)
}
