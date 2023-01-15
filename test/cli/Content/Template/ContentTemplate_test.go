package content_template_test

import (
	"encoding/json"
	"testing"
	"time"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	"github.com/perimeter-81/proxmox-api-go/proxmox"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const storage string = "local"

func checkIfTemplateDoesNotExist(t *testing.T, template, node, storage string) {
	Test := cliTest.Test{
		NotExpected: template,
		NotContains: true,
		Args:        []string{"-i", "list", "files", cliTest.FirstNode, storage, string(proxmox.ContentType_Template)},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Download_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		Args: []string{"-i", "delete", "file", cliTest.FirstNode, storage, string(proxmox.ContentType_Template), cliTest.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Existence_Removed_0(t *testing.T) {
	checkIfTemplateDoesNotExist(t, cliTest.DownloadedLXCTemplate, cliTest.FirstNode, storage)
}

func Test_ContentTemplate_Download(t *testing.T) {
	Test := cliTest.Test{
		Expected: "(" + cliTest.DownloadedLXCTemplate + ")",
		Contains: true,
		Args:     []string{"-i", "content", "template", "download", cliTest.FirstNode, storage, cliTest.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_List(t *testing.T) {
	Test := cliTest.Test{
		Return: true,
		Args:   []string{"-i", "list", "files", cliTest.FirstNode, storage, string(proxmox.ContentType_Template)},
	}
	var data []*proxmox.Content_FileProperties
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
	assert.Equal(t, cliTest.DownloadedLXCTemplate, data[0].Name)
	assert.NotEqual(t, "", data[0].Format)
	assert.Greater(t, data[0].Size, uint(0))
	assert.Greater(t, data[0].CreationTime, time.UnixMilli(0))
}

func Test_ContentTemplate_Download_Delete(t *testing.T) {
	Test := cliTest.Test{
		Expected: cliTest.DownloadedLXCTemplate,
		Contains: true,
		Args:     []string{"-i", "delete", "file", cliTest.FirstNode, storage, string(proxmox.ContentType_Template), cliTest.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Existence_Removed_1(t *testing.T) {
	checkIfTemplateDoesNotExist(t, cliTest.DownloadedLXCTemplate, cliTest.FirstNode, storage)
}
