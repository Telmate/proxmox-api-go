package cli_metricservers_test

import (
	_ "github.com/perimeter-81/proxmox-api-go/cli/command/commands"
	cliTest "github.com/perimeter-81/proxmox-api-go/test/cli"
	"testing"
)

func Test_MetricServer_Errors_Type(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "this gives an error"
}`,
		ReqErr:      true,
		ErrContains: "(type)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_Server(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": ""
}`,
		ReqErr:      true,
		ErrContains: "(server)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_Port_Lower(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 0
}`,
		ReqErr:      true,
		ErrContains: "(port)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_Port_Upper(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 65537
}`,
		ReqErr:      true,
		ErrContains: "(port)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_MTU_Lower(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 65536,
	"mtu": 511
}`,
		ReqErr:      true,
		ErrContains: "(mtu)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_MTU_Upper(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 1,
	"mtu": 65537
}`,
		ReqErr:      true,
		ErrContains: "(mtu)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_Timeout(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 1,
	"mtu": 512,
	"timeout": -1
}`,
		ReqErr:      true,
		ErrContains: "(timeout)",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

// Graphite
func Test_MetricServer_Errors_Graphite_Protocol(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "graphite",
	"server": "192.168.67.3",
	"graphite": {
		"protocol": "notvalid"
	}
}`,
		ReqErr:      true,
		ErrContains: "(graphite:{ protocol })",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

// InfluxDB
func Test_MetricServer_Errors_InfluxDB_Protocol(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"influxdb": {
		"protocol": "notvalid"
	}
}`,
		ReqErr:      true,
		ErrContains: "(influxdb:{ protocol })",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}

func Test_MetricServer_Errors_InfluxDB_MaxBodySize(t *testing.T) {
	Test := cliTest.Test{
		InputJson: `
{
	"type": "influxdb",
	"server": "192.168.67.3",
	"port": 8089,
	"influxdb": {
		"max-body-size": 0
	}
}`,
		ReqErr:      true,
		ErrContains: "(influxdb:{ max-body-size })",
		Args:        []string{"-i", "set", "metricserver", "test-metricserver00"},
	}
	Test.StandardTest(t)
}
