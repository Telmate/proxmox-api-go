package cli_metricservers_test

import (
	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"testing"
)

func Test_MetricServer_Graphite_0_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: true,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_0_Set_Full(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 1,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": true,
	"mtu": 512,
	"timeout": 10,
	"graphite": {
		"protocol": "udp",
		"path": "test-path"
	}
}`,
		Expected: "(test-metricserver-g0)",
		Contains: true,
		Args:     []string{"-i", "set", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_0_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver-g0",
	"port": 1,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": true,
	"mtu": 512,
	"timeout": 10,
	"graphite": {
		"protocol": "udp",
		"path": "test-path"
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_0_Set_Empty(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 65536,
	"server": "192.168.67.4",
	"type": "graphite",
	"graphite": {
		"protocol": "udp"
	}
}`,
		Expected: "(test-metricserver-g0)",
		Contains: true,
		Args:     []string{"-i", "set", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_0_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver-g0",
	"port": 65536,
	"server": "192.168.67.4",
	"type": "graphite",
	"enable": true,
	"mtu": 1500,
	"timeout": 1,
	"graphite": {
		"protocol": "udp",
		"path": "proxmox"
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Graphite_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: false,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver-g0"},
	}
	Test.StandardTest(t)
}
