package cli_metricservers_test

import (
	"testing"

	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
)

func Test_MetricServer_InfluxDB_0_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: true,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_Set_Full(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 8087,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1600,
	"timeout": 10,
	"influxdb": {
		"api-path-prefix": "pathprefix",
		"bucket": "test-bucket",
		"protocol": "https",
		"max-body-size": 1,
		"organization": "test-organization",
		"verify-certificate": false
	}
}`,
		Contains: []string{"(test-metricserver0)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_List(t *testing.T) {
	Test := cliTest.Test{
		Contains: []string{`"id":"test-metricserver0"`},
		Args:     []string{"-i", "list", "metricservers"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver0",
	"port": 8087,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1600,
	"timeout": 10,
	"influxdb": {
		"api-path-prefix": "pathprefix",
		"bucket": "test-bucket",
		"protocol": "https",
		"max-body-size": 1,
		"organization": "test-organization",
		"verify-certificate": false
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_Set_Empty(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 8087,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"influxdb": {
		"protocol": "https",
		"verify-certificate": false
	}
}`,
		Contains: []string{"(test-metricserver0)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver0",
	"port": 8087,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1500,
	"timeout": 1,
	"influxdb": {
		"bucket": "proxmox",
		"protocol": "https",
		"max-body-size": 25000000,
		"organization": "proxmox",
		"verify-certificate": false
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_0_Delete(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: false,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver0"},
	}
	Test.StandardTest(t)
}
