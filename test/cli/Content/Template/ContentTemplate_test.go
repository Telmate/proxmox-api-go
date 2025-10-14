package content_template_test

import (
	"encoding/json"
	"testing"
	"time"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkIfTemplateDoesNotExist(t *testing.T, template, node, storage string) {
	Test := cliTest.Test{
		NotContains: []string{template},
		Args:        []string{"-i", "list", "files", test.FirstNode, storage, string(proxmox.ContentType_Template)},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Download_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		Args: []string{"-i", "delete", "file", test.FirstNode, test.CtStorage, string(proxmox.ContentType_Template), test.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Existence_Removed_0(t *testing.T) {
	checkIfTemplateDoesNotExist(t, test.DownloadedLXCTemplate, test.FirstNode, test.CtStorage)
}

func Test_ContentTemplate_Download(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{"(" + test.DownloadedLXCTemplate + ")"},
		Args:     []string{"-i", "content", "template", "download", test.FirstNode, test.CtStorage, test.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_List(t *testing.T) {
	Test := cliTest.Test{
		Return: true,
		Args:   []string{"-i", "list", "files", test.FirstNode, test.CtStorage, string(proxmox.ContentType_Template)},
	}
	var data []*proxmox.Content_FileProperties
	require.NoError(t, json.Unmarshal(Test.StandardTest(t), &data))
	assert.Equal(t, test.DownloadedLXCTemplate, data[0].Name)
	assert.NotEqual(t, "", data[0].Format)
	assert.Greater(t, data[0].Size, uint(0))
	assert.Greater(t, data[0].CreationTime, time.UnixMilli(0))
}

func Test_ContentTemplate_Download_Delete(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{test.DownloadedLXCTemplate},
		Args:     []string{"-i", "delete", "file", test.FirstNode, test.CtStorage, string(proxmox.ContentType_Template), test.DownloadedLXCTemplate},
	}
	Test.StandardTest(t)
}

func Test_ContentTemplate_Existence_Removed_1(t *testing.T) {
	checkIfTemplateDoesNotExist(t, test.DownloadedLXCTemplate, test.FirstNode, test.CtStorage)
}
