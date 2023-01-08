package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

// User options for the Proxmox API
type ConfigUser struct {
	UserID    string   `json:"userid"`
	Comment   string   `json:"comment,omitempty"`
	Email     string   `json:"email,omitempty"`
	Enable    bool     `json:"enable"`
	Expire    int      `json:"expire"`
	FirstName string   `json:"firstname,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Keys      string   `json:"keys,omitempty"`
	LastName  string   `json:"lastname,omitempty"`
	Password  string   `json:"-"`
}

func (config ConfigUser) CreateUser(client *Client) (err error) {
	params := config.mapToAPI(true)
	err = client.CreateUser(params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating User: %v, (params: %v)", err, string(params))
	}
	return
}

// Maps the struct to the API values proxmox understands
func (config ConfigUser) mapToAPI(create bool) (params map[string]interface{}) {
	params = map[string]interface{}{
		"comment":   config.Comment,
		"email":     config.Email,
		"enable":    config.Enable,
		"expire":    config.Expire,
		"firstname": config.FirstName,
		"groups":    ArrayToCSV(config.Groups),
		"keys":      config.Keys,
		"lastname":  config.LastName,
	}
	if create {
		params["password"] = config.Password
		params["userid"] = config.UserID
	}
	return
}

// Create or update the user depending on if the user already exists or not.
// "userId" and "password" overwrite what is specified in "*ConfigUser".
func (config *ConfigUser) SetUser(userId, password string, client *Client) (err error) {
	err = ValidateUserPassword(password)
	if err != nil {
		return err
	}

	if config != nil {
		config.UserID = userId
		config.Password = password
	}

	userExists, err := client.CheckUserExistance(userId)
	if err != nil {
		return err
	}

	if config != nil {
		if userExists {
			err = config.UpdateUser(client)
			if err != nil {
				return err
			}
		} else {
			err = config.CreateUser(client)
		}
	} else {
		config = &ConfigUser{
			Password: password,
			UserID:   userId,
		}
		if userExists {
			if config.Password != "" {
				err = config.UpdateUserPassword(client)
			}
		} else {
			err = config.CreateUser(client)
		}
	}
	return
}

func (config *ConfigUser) UpdateUser(client *Client) (err error) {
	params := config.mapToAPI(false)
	err = client.UpdateUser(config.UserID, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating User: %v, (params: %v)", err, string(params))
	}
	if config.Password != "" {
		err = config.UpdateUserPassword(client)
	}
	return
}

func (config ConfigUser) UpdateUserPassword(client *Client) (err error) {
	err = ValidateUserPassword(config.Password)
	if err != nil {
		return err
	}
	return client.Put(map[string]interface{}{
		"userid":   config.UserID,
		"password": config.Password,
	}, "/access/password")
}

func NewConfigUserFromApi(userId string, client *Client) (config *ConfigUser, err error) {
	// prepare json map to receive the information from the api
	var userConfig map[string]interface{}
	userConfig, err = client.GetUserConfig(userId)
	if err != nil {
		return nil, err
	}
	config = new(ConfigUser)

	config.UserID = userId

	if _, isSet := userConfig["comment"]; isSet {
		config.Comment = userConfig["comment"].(string)
	}
	if _, isSet := userConfig["email"]; isSet {
		config.Email = userConfig["email"].(string)
	}
	if _, isSet := userConfig["enable"]; isSet {
		config.Enable = Itob(int(userConfig["enable"].(float64)))
	}
	if _, isSet := userConfig["expire"]; isSet {
		config.Expire = int(userConfig["expire"].(float64))
	}
	if _, isSet := userConfig["firstname"]; isSet {
		config.FirstName = userConfig["firstname"].(string)
	}
	if _, isSet := userConfig["keys"]; isSet {
		config.Keys = userConfig["keys"].(string)
	}
	if _, isSet := userConfig["lastname"]; isSet {
		config.LastName = userConfig["lastname"].(string)
	}
	if _, isSet := userConfig["groups"]; isSet {
		config.Groups = ArrayToStringType(userConfig["groups"].([]interface{}))
	}

	return
}

func NewConfigUserFromJson(input []byte) (config *ConfigUser, err error) {
	if len(input) != 0 {
		config = &ConfigUser{}
		err = json.Unmarshal([]byte(input), config)
	}
	return
}

func ValidateUserPassword(password string) error {
	if utf8.RuneCountInString(password) >= 5 || password == "" {
		return nil
	}
	return errors.New("error updating User: the minimum password length is 5")
}
