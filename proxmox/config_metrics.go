package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
)

type ConfigMetricsGraphite struct {
	Protocol string `json:"protocol"`
	Path     string `json:"path,omitempty"`
}

type ConfigMetricsInfluxDB struct {
	ApiPathPrefix     string `json:"api-path-prefix,omitempty"`
	Bucket            string `json:"bucket,omitempty"`
	Protocol          string `json:"protocol"`
	MaxBodySize       int    `json:"max-body-size,omitempty"`
	Organization      string `json:"organization,omitempty"`
	Token             string `json:"token,omitempty"` //token key is never returned from api
	VerifyCertificate bool   `json:"verify-certificate"`
}

// Metrics options for the Proxmox API
type ConfigMetrics struct {
	Name     string                 `json:"name"`
	Port     int                    `json:"port"`
	Server   string                 `json:"server"`
	Type     string                 `json:"type"` //type key is only used on create
	Enable   bool                   `json:"enable"`
	MTU      int                    `json:"mtu"`
	Timeout  int                    `json:"timeout,omitempty"`
	Graphite *ConfigMetricsGraphite `json:"graphite,omitempty"`
	InfluxDB *ConfigMetricsInfluxDB `json:"influxdb,omitempty"`
}

func (config *ConfigMetrics) mapToApiValues(create bool) (params map[string]interface{}) {
	var deletions string
	params = map[string]interface{}{
		"port":    config.Port,
		"server":  config.Server,
		"disable": BoolInvert(config.Enable),
		"mtu":     config.MTU,
		"timeout": config.Timeout,
	}
	if create {
		params["type"] = config.Type
	}
	if config.Graphite != nil {
		params["path"] = config.Graphite.Path
		params["proto"] = config.Graphite.Protocol
	}
	if config.InfluxDB != nil {
		if config.InfluxDB.ApiPathPrefix != "" {
			params["api-path-prefix"] = config.InfluxDB.ApiPathPrefix
		} else {
			deletions = AddToList(deletions, "api-path-prefix")
		}
		params["bucket"] = config.InfluxDB.Bucket
		params["max-body-size"] = config.InfluxDB.MaxBodySize
		params["organization"] = config.InfluxDB.Organization
		params["influxdbproto"] = config.InfluxDB.Protocol
		if config.InfluxDB.Token != "" {
			params["token"] = config.InfluxDB.Token
		}
		params["verify-certificate"] = config.InfluxDB.VerifyCertificate
	}

	if !create && deletions != "" {
		params["delete"] = deletions
	}
	return
}

func (config *ConfigMetrics) RemoveMetricsNestedStructs() {
	if config.Type != "graphite" {
		config.Graphite = nil
	} else {
		config.InfluxDB = nil
	}
}

func (config *ConfigMetrics) ValidateMetrics() (err error) {
	err = ValidateStringInArray([]string{"graphite", "influxdb"}, config.Type, "type")
	if err != nil {
		return
	}
	err = ValidateStringNotEmpty(config.Server, "server")
	if err != nil {
		return
	}
	err = ValidateStringInArray([]string{"udp", "tcp"}, config.Graphite.Protocol, "graphite:{ protocol }")
	if err != nil {
		return
	}
	err = ValidateStringInArray([]string{"udp", "http", "https"}, config.InfluxDB.Protocol, "influxdb:{ protocol }")
	if err != nil {
		return
	}
	err = ValidateIntInRange(1, 65536, config.Port, "port")
	if err != nil {
		return
	}
	err = ValidateIntGreaterOrEquals(1, config.InfluxDB.MaxBodySize, "influxdb:{ max-body-size }")
	if err != nil {
		return
	}
	err = ValidateIntInRange(512, 65536, config.MTU, "mtu")
	if err != nil {
		return
	}
	err = ValidateIntGreaterOrEquals(0, config.Timeout, "timeout")
	return
}

func (config *ConfigMetrics) SetMetrics(metricsId string, client *Client) (err error) {
	err = config.ValidateMetrics()
	if err != nil {
		return
	}

	config.Name = metricsId

	metricsExists, err := client.CheckMetricServerExistence(metricsId)
	if err != nil {
		return err
	}

	if metricsExists {
		err = config.UpdateMetrics(client)
	} else {
		err = config.CreateMetrics(client)
	}
	return
}

func (config *ConfigMetrics) CreateMetrics(client *Client) (err error) {
	config.RemoveMetricsNestedStructs()
	params := config.mapToApiValues(true)
	err = client.CreateMetricServer(config.Name, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Metrics Server: %v, (params: %v)", err, string(params))
	}
	return
}

func (config *ConfigMetrics) UpdateMetrics(client *Client) (err error) {
	config.RemoveMetricsNestedStructs()
	params := config.mapToApiValues(false)
	err = client.UpdateMetricServer(config.Name, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating Metrics Server: %v, (params: %v)", err, string(params))
	}
	return
}

func InstantiateConfigMetrics() *ConfigMetrics {
	graphite := ConfigMetricsGraphite{
		Path:     "proxmox",
		Protocol: "udp",
	}
	influxdb := ConfigMetricsInfluxDB{
		Bucket:            "proxmox",
		MaxBodySize:       25000000,
		Organization:      "proxmox",
		Protocol:          "udp",
		VerifyCertificate: true,
	}
	return &ConfigMetrics{
		Enable:   true,
		MTU:      1500,
		Timeout:  1,
		Graphite: &graphite,
		InfluxDB: &influxdb,
	}
}

func NewConfigMetricsFromApi(metricsId string, client *Client) (config *ConfigMetrics, err error) {
	// prepare json map to receive the information from the api
	var rawConfig map[string]interface{}
	rawConfig, err = client.GetMetricServerConfig(metricsId)
	if err != nil {
		return nil, err
	}
	config = InstantiateConfigMetrics()

	config.Name = metricsId
	config.Port = int(rawConfig["port"].(float64))
	config.Server = rawConfig["server"].(string)
	config.Type = rawConfig["type"].(string)

	if _, isSet := rawConfig["disable"]; isSet {
		config.Enable = BoolInvert(Itob(int(rawConfig["disable"].(float64))))
	}
	if _, isSet := rawConfig["mtu"]; isSet {
		config.MTU = int(rawConfig["mtu"].(float64))
	}
	if _, isSet := rawConfig["timeout"]; isSet {
		config.Timeout = int(rawConfig["timeout"].(float64))
	}

	config.RemoveMetricsNestedStructs()

	if config.Graphite != nil {
		if _, isSet := rawConfig["path"]; isSet {
			config.Graphite.Path = rawConfig["path"].(string)
		}
		if _, isSet := rawConfig["proto"]; isSet {
			config.Graphite.Protocol = rawConfig["proto"].(string)
		}
	}
	if config.InfluxDB != nil {
		if _, isSet := rawConfig["api-path-prefix"]; isSet {
			config.InfluxDB.ApiPathPrefix = rawConfig["api-path-prefix"].(string)
		}
		if _, isSet := rawConfig["bucket"]; isSet {
			config.InfluxDB.Bucket = rawConfig["bucket"].(string)
		}
		if _, isSet := rawConfig["influxdbproto"]; isSet {
			config.InfluxDB.Protocol = rawConfig["influxdbproto"].(string)
		}
		if _, isSet := rawConfig["max-body-size"]; isSet {
			config.InfluxDB.MaxBodySize = int(rawConfig["max-body-size"].(float64))
		}
		if _, isSet := rawConfig["organization"]; isSet {
			config.InfluxDB.Organization = rawConfig["organization"].(string)
		}
		if _, isSet := rawConfig["token"]; isSet {
			config.InfluxDB.Token = rawConfig["token"].(string)
		}
		if _, isSet := rawConfig["verify-certificate"]; isSet {
			config.InfluxDB.VerifyCertificate = Itob(int(rawConfig["verify-certificate"].(float64)))
		}
	}
	return
}

func NewConfigMetricsFromJson(input []byte) (config *ConfigMetrics, err error) {
	config = InstantiateConfigMetrics()
	err = json.Unmarshal([]byte(input), &config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
