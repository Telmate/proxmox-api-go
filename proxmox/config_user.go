package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const Error_NewUserID string = "no username or realm specified, syntax is \"username@realm\""

// User options for the Proxmox API
type ConfigUser struct {
	User      UserID       `json:"user"`
	Comment   string       `json:"comment,omitempty"`
	Email     string       `json:"email,omitempty"`
	Enable    bool         `json:"enable"`
	Expire    uint         `json:"expire"`
	FirstName string       `json:"firstname,omitempty"`
	Groups    []string     `json:"groups,omitempty"`
	Keys      string       `json:"keys,omitempty"`
	LastName  string       `json:"lastname,omitempty"`
	Password  UserPassword `json:"-"`
}

func (config ConfigUser) CreateUser(client *Client) (err error) {
	err = config.Validate()
	if err != nil {
		return
	}
	params := config.mapToApiValues(true)
	err = client.Post(params, "/access/users")
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating User: %v, (params: %v)", err, string(params))
	}
	return
}

func (config ConfigUser) DeleteUser(client *Client) (err error) {
	existence, err := CheckUserExistence(config.User, client)
	if err != nil {
		return
	}
	if !existence {
		return fmt.Errorf("user (%s) could not be deleted, the user does not exist", config.User.ToString())
	}
	// Proxmox silently fails a user delete if the users does not exist
	return client.Delete("/access/users/" + config.User.ToString())
}

// Maps the struct to the API values proxmox understands
func (config ConfigUser) mapToApiValues(create bool) (params map[string]interface{}) {
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
		params["userid"] = config.User.ToString()
	}
	return
}

// Create or update the user depending on if the user already exists or not.
// "userId" and "password" overwrite what is specified in "*ConfigUser".
func (config *ConfigUser) SetUser(userId UserID, password UserPassword, client *Client) (err error) {
	if config != nil {
		config.User = userId
		config.Password = password
	}

	userExists, err := CheckUserExistence(userId, client)
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
			User:     userId,
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
	params := config.mapToApiValues(false)
	err = client.Put(params, "/access/users/"+config.User.ToString())
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
	err = config.Password.Validate()
	if err != nil {
		return err
	}
	return client.Put(map[string]interface{}{
		"userid":   config.User.ToString(),
		"password": config.Password,
	}, "/access/password")
}

func (config ConfigUser) Validate() error {
	return config.Password.Validate()
}

type UserID struct {
	// TODO create custom type for Name.
	// the name only seems to allows some characters, and using the string type would imply that all characters are allowed.
	// https://bugzilla.proxmox.com/show_bug.cgi?id=4461
	Name string `json:"name"`
	// TODO create custom type for Realm.
	// the realm only allows some characters, and using the string type would imply that all characters are allowed.
	// https://bugzilla.proxmox.com/show_bug.cgi?id=4462
	Realm string `json:"realm"`
}

// Map the params to an array of UserID objects
func (UserID) mapToArray(params []interface{}) *[]UserID {
	members := ArrayToStringType(params)
	UserList := make([]UserID, len(members))
	for i := range members {
		UserList[i] = UserID{}.mapToStruct(members[i])
	}
	return &UserList
}

// transforms the  username@realm to a UserID object
func (UserID) mapToStruct(userId string) UserID {
	user, _ := NewUserID(userId)
	return user
}

// Converts the userID to "username@realm"
// Returns an empty string when either the Name or Realm is empty
func (id UserID) ToString() string {
	if id.Name == "" || id.Realm == "" {
		return ""
	}
	return id.Name + "@" + id.Realm
}

// TODO add func (id UserID) Validate()

// May be empty or should be at least be 5 characters long.
type UserPassword string

func (password UserPassword) Validate() error {
	if utf8.RuneCountInString(string(password)) >= 5 || password == "" {
		return nil
	}
	return errors.New("the minimum password length is 5")
}

// List all users that exist in proxmox
func ListUsers(client *Client) (users []interface{}, err error) {
	return client.GetItemListInterfaceArray("/access/users?full=1")
}

// Check if the user already exists in proxmox.
func CheckUserExistence(userId UserID, client *Client) (existence bool, err error) {
	list, err := ListUsers(client)
	if err != nil {
		return
	}
	existence = ItemInKeyOfArray(list, "userid", userId.ToString())
	return
}

// Maps the API values from proxmox to a struct
func mapToStructConfigUser(userId UserID, params map[string]interface{}) *ConfigUser {
	config := ConfigUser{User: userId}
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
		config.Expire = uint(params["expire"].(float64))
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

func NewConfigUserFromApi(userId UserID, client *Client) (config *ConfigUser, err error) {
	userConfig, err := client.GetItemConfigMapStringInterface("/access/users/"+userId.ToString(), "user", "CONFIG")
	if err != nil {
		return
	}
	config = mapToStructConfigUser(userId, userConfig)
	return
}

func NewConfigUserFromJson(input []byte) (config *ConfigUser, err error) {
	if len(input) != 0 {
		config = &ConfigUser{}
		err = json.Unmarshal([]byte(input), config)
	}
	return
}

// Converts "username@realm" to a UserID object
func NewUserID(userId string) (id UserID, err error) {
	tmpList := strings.Split(userId, "@")
	if len(tmpList) == 2 {
		if tmpList[0] != "" && tmpList[1] != "" {
			return UserID{
				Name:  tmpList[0],
				Realm: tmpList[1],
			}, nil
		}
	}
	return UserID{}, errors.New(Error_NewUserID)
}
