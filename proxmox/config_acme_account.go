package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Acme Account options for the Proxmox API
type ConfigAcmeAccount struct {
	Name      string   `json:"name"`
	Contact   []string `json:"contact"`
	Directory string   `json:"directory"`
	Tos       bool     `json:"tos,omitempty"`
}

func (config ConfigAcmeAccount) CreateAcmeAccount(acmeId string, client *Client) (err error) {
	params := map[string]interface{}{
		"name":    acmeId,
		"contact": ArrayToCSV(config.Contact),
	}
	if !config.Tos {
		return errors.New("error creating Acme account: the terms of service must be accepted")
	}

	acmeDirectories, err := client.GetAcmeDirectoriesUrl()
	if err != nil {
		return err
	}

	var tos string
	if inArray(acmeDirectories, config.Directory) {
		tos, err = client.GetAcmeTosUrl()
		if err != nil {
			return err
		}
	} else {
		tos = "true"
	}
	params["tos_url"] = tos
	params["directory"] = config.Directory

	exitStatus, err := client.CreateAcmeAccount(params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Acme profile: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}
	return
}

func NewConfigAcmeAccountFromApi(id string, client *Client) (config *ConfigAcmeAccount, err error) {
	// prepare json map to receive the information from the api
	var acmeConfig map[string]interface{}
	acmeConfig, err = client.GetAcmeAccountConfig(id)
	if err != nil {
		return nil, err
	}

	config = new(ConfigAcmeAccount)

	config.Name = id

	config.Directory = acmeConfig["directory"].(string)

	// Using the proxmox cli you can make the "tos" key empty
	if acmeConfig["tos"] != nil {
		if acmeConfig["tos"].(string) != "" {
			config.Tos = true
		}
	}

	contactArray := ArrayToStringType(acmeConfig["account"].(map[string]interface{})["contact"].([]interface{}))
	config.Contact = make([]string, len(contactArray))
	for i, element := range contactArray {
		config.Contact[i] = strings.TrimPrefix(element, "mailto:")
	}

	return
}

func NewConfigAcmeAccountFromJson(input []byte) (config *ConfigAcmeAccount, err error) {
	config = &ConfigAcmeAccount{}
	err = json.Unmarshal([]byte(input), config)
	if err != nil {
		log.Fatal(err)
	}
	return
}
