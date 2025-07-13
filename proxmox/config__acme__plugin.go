package proxmox

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
)

// Acme Plugin options for the Proxmox API
type ConfigAcmePlugin struct {
	ID              string   `json:"pluginid"`
	API             string   `json:"api"`
	Data            string   `json:"data,omitempty"`
	Enable          bool     `json:"enable"`
	Nodes           []string `json:"nodes,omitempty"`
	ValidationDelay int      `json:"validation-delay"`
}

func (config ConfigAcmePlugin) mapToApiValues() (params map[string]interface{}) {
	params = map[string]interface{}{
		"api":              config.API,
		"data":             base64.StdEncoding.EncodeToString([]byte(config.Data)),
		"disable":          BoolInvert(config.Enable),
		"nodes":            ArrayToCSV(config.Nodes),
		"validation-delay": config.ValidationDelay,
	}
	return
}

func (config ConfigAcmePlugin) SetAcmePlugin(ctx context.Context, pluginId string, client *Client) (err error) {
	err = ValidateIntInRange(0, 172800, config.ValidationDelay, "validation-delay")
	if err != nil {
		return
	}

	config.ID = pluginId

	pluginExists, err := client.CheckAcmePluginExistence(ctx, pluginId)
	if err != nil {
		return
	}

	if pluginExists {
		err = config.UpdateAcmePlugin(ctx, client)
	} else {
		err = config.CreateAcmePlugin(ctx, client)
	}
	return
}

func (config ConfigAcmePlugin) CreateAcmePlugin(ctx context.Context, client *Client) (err error) {
	params := config.mapToApiValues()
	params["id"] = config.ID
	params["type"] = "dns"
	err = client.CreateAcmePlugin(ctx, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Acme plugin: %v, (params: %v)", err, string(params))
	}
	return
}

func (config ConfigAcmePlugin) UpdateAcmePlugin(ctx context.Context, client *Client) (err error) {
	params := config.mapToApiValues()
	err = client.UpdateAcmePlugin(ctx, config.ID, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating Acme plugin: %v, (params: %v)", err, string(params))
	}
	return
}

func NewConfigAcmePluginFromApi(ctx context.Context, id string, client *Client) (config *ConfigAcmePlugin, err error) {
	// prepare json map to receive the information from the api
	var rawConfig map[string]interface{}
	rawConfig, err = client.GetAcmePluginConfig(ctx, id)
	if err != nil {
		return nil, err
	}

	config = new(ConfigAcmePlugin)

	config.ID = id
	config.API = rawConfig["api"].(string)

	if _, isSet := rawConfig["data"]; isSet {
		config.Data = rawConfig["data"].(string)
	}
	if _, isSet := rawConfig["disable"]; isSet {
		config.Enable = BoolInvert(Itob(int(rawConfig["disable"].(float64))))
	} else {
		config.Enable = true
	}
	if _, isSet := rawConfig["validation-delay"]; isSet {
		config.ValidationDelay = int(rawConfig["validation-delay"].(float64))
	} else {
		config.ValidationDelay = 30
	}

	return
}

func NewConfigAcmePluginFromJson(input []byte) (config *ConfigAcmePlugin, err error) {
	config = &ConfigAcmePlugin{}
	err = json.Unmarshal([]byte(input), config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
