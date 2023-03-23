package cli_metricservers_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_MetricServer_Graphite_1_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: true,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_1_Set_Empty(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 35466,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": false, 
	"graphite": {
		"protocol": "tcp"
	}
}`,
		Contains: []string{"(test-metricserver-g1)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_1_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver-g1",
	"port": 35466,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": false,
	"mtu": 1500,
	"timeout": 1,
	"graphite": {
		"protocol": "tcp",
		"path": "proxmox"
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_1_Set_Full(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 35465,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": false,
	"mtu": 1600,
	"timeout": 0,
	"graphite": {
		"protocol": "tcp",
		"path": "test-path"
	}
}`,
		Contains: []string{"(test-metricserver-g1)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_1_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver-g1",
	"port": 35465,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": false,
	"mtu": 1600,
	"graphite": {
		"protocol": "tcp",
		"path": "test-path"
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_1_Delete(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: false,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver-g1"},
	}
	Test.StandardTest(t)
}
