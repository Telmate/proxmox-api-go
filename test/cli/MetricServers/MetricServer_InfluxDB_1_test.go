package cli_metricservers_test

import (
	"testing"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	cliTest "github.com/Telmate/proxmox-api-go/test/cli"
)

func Test_MetricServer_InfluxDB_1_Cleanup(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: true,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_1_Set_Empty(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 8088,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"influxdb": {
		"protocol": "http",
		"verify-certificate": false
	}
}`,
		Contains: []string{"(test-metricserver1)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_1_Get_Empty(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver1",
	"port": 8088,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1500,
	"timeout": 1,
	"influxdb": {
		"bucket": "proxmox",
		"protocol": "http",
		"max-body-size": 25000000,
		"organization": "proxmox",
		"verify-certificate": false
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_1_Set_Full(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"port": 8088,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1600,
	"timeout": 10,
	"influxdb": {
		"api-path-prefix": "pathprefix",
		"bucket": "test-bucket",
		"protocol": "http",
		"max-body-size": 25000001,
		"organization": "test-organization",
		"verify-certificate": false
	}
}`,
		Contains: []string{"(test-metricserver1)"},
		Args:     []string{"-i", "set", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_1_Get_Full(t *testing.T) {
	cliTest.SetEnvironmentVariables()
	Test := cliTest.Test{
		OutputJson: `
{
	"name": "test-metricserver1",
	"port": 8088,
	"server": "192.168.67.3",
	"type": "influxdb",
	"enable": false,
	"mtu": 1600,
	"timeout": 10,
	"influxdb": {
		"api-path-prefix": "pathprefix",
		"bucket": "test-bucket",
		"protocol": "http",
		"max-body-size": 25000001,
		"organization": "test-organization",
		"verify-certificate": false
	}
}`,
		Args: []string{"-i", "get", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_InfluxDB_1_Delete(t *testing.T) {
	Test := cliTest.Test{
		ReqErr: false,
		Args:   []string{"-i", "delete", "metricserver", "test-metricserver1"},
	}
	Test.StandardTest(t)
}
