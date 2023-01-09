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

// Maps the API values from proxmox to a struct
func mapToStruct(userId string, params map[string]interface{}) *ConfigUser {
	config := ConfigUser{UserID: userId}
	if _, isSet := params["comment"]; isSet {
		config.Comment = params["comment"].(string)
	}
	if _, isSet := params["email"]; isSet {
		config.Email = params["email"].(string)
	}
	if _, isSet := params["enable"]; isSet {
		config.Enable = Itob(int(params["enable"].(float64)))
	}
	if _, isSet := params["expire"]; isSet {
		config.Expire = int(params["expire"].(float64))
	}
	if _, isSet := params["firstname"]; isSet {
		config.FirstName = params["firstname"].(string)
	}
	if _, isSet := params["keys"]; isSet {
		config.Keys = params["keys"].(string)
	}
	if _, isSet := params["lastname"]; isSet {
		config.LastName = params["lastname"].(string)
	}
	if _, isSet := params["groups"]; isSet {
		config.Groups = ArrayToStringType(params["groups"].([]interface{}))
	}
	return &config
}

func NewConfigUserFromApi(userId string, client *Client) (config *ConfigUser, err error) {
	userConfig, err := client.GetItemConfigMapStringInterface("/access/users/"+userId, "user", "CONFIG")
	if err != nil {
		return
	}
	config = mapToStruct(userId, userConfig)
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
