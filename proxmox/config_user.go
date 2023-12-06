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
	Groups    *[]GroupName `json:"groups,omitempty"`
	Keys      string       `json:"keys,omitempty"`
	LastName  string       `json:"lastname,omitempty"`
	// Password is always empty when getting information from Proxmox
	Password UserPassword `json:"-"`
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
		"groups":    GroupName("").arrayToCsv(config.Groups),
		"keys":      config.Keys,
		"lastname":  config.LastName,
	}
	if create {
		params["password"] = config.Password
		params["userid"] = config.User.ToString()
	}
	return
}

func (ConfigUser) mapToArray(params []interface{}) *[]ConfigUser {
	users := make([]ConfigUser, len(params))
	for i, e := range params {
		users[i] = *ConfigUser{}.mapToStruct(e.(map[string]interface{}))
	}
	return &users
}

// Maps the API values from proxmox to a struct
func (config ConfigUser) mapToStruct(params map[string]interface{}) *ConfigUser {
	if _, isSet := params["userid"]; isSet {
		config.User = UserID{}.mapToStruct(params["userid"].(string))
	}
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
		config.Groups = GroupName("").mapToArray(params["groups"])
	}
	return &config
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

type ApiToken struct {
	TokenId string `json:"tokenid"`
	Comment string `json:"comment,omitempty"`
	Expire  int64  `json:"expire"`
	Privsep bool   `json:"privsep"`
}
type ApiTokenCreateResult struct {
	Info  map[string]interface{} `json:"info"`
	Value string                 `json:"value"`
}
type ApiTokenCreateResultWrapper struct {
	Data ApiTokenCreateResult `json:"data"`
}

// Maps the API values from proxmox to a struct
func (tokens ApiToken) mapToStruct(params map[string]interface{}) *ApiToken {
	if _, isSet := params["tokenid"]; isSet {
		tokens.TokenId = params["tokenid"].(string)
	}
	if _, isSet := params["comment"]; isSet {
		tokens.Comment = params["comment"].(string)
	}
	if _, isSet := params["expire"]; isSet {
		tokens.Expire = int64(params["expire"].(float64))
	}
	if _, isSet := params["privsep"]; isSet {
		tokens.Privsep = false
		if params["privsep"] == 1 {
			tokens.Privsep = true
		}
	}
	return &tokens
}

func (ApiToken) mapToArray(params []interface{}) *[]ApiToken {
	tokens := make([]ApiToken, len(params))
	for i, e := range params {
		tokens[i] = *ApiToken{}.mapToStruct(e.(map[string]interface{}))
	}
	return &tokens
}

func (config ConfigUser) CreateApiToken(client *Client, token ApiToken) (value string, err error) {
	status, err := client.CreateItemReturnStatus(map[string]interface{}{
		"comment": token.Comment,
		"expire":  token.Expire,
		"privsep": token.Privsep,
	}, "/access/users/"+config.User.ToString()+"/token/"+token.TokenId)
	if err != nil {
		return
	}
	var resultWrapper *ApiTokenCreateResultWrapper
	err = json.Unmarshal([]byte(status), &resultWrapper)
	value = resultWrapper.Data.Value
	return
}

func (config ConfigUser) UpdateApiToken(client *Client, token ApiToken) (err error) {
	err = client.Put(map[string]interface{}{
		"comment": token.Comment,
		"expire":  token.Expire,
		"privsep": token.Privsep,
	}, "/access/users/"+config.User.ToString()+"/token/"+token.TokenId)
	return
}

func (config ConfigUser) ListApiTokens(client *Client) (tokens *[]ApiToken, err error) {
	status, err := client.GetItemListInterfaceArray("/access/users/" + config.User.ToString() + "/token")
	if err != nil {
		return
	}
	tokens = ApiToken{}.mapToArray(status)
	return
}

func (config ConfigUser) DeleteApiToken(client *Client, token ApiToken) (err error) {
	err = client.Delete("/access/users/" + config.User.ToString() + "/token/" + token.TokenId)
	return
}

// Validates all items and sub items in the ConfigUser struct
func (config ConfigUser) Validate() (err error) {
	err = config.User.Validate()
	if err != nil {
		return
	}
	if config.Groups != nil {
		if len(*config.Groups) != 0 {
			for _, e := range *config.Groups {
				err = e.Validate()
				if err != nil {
					return
				}
			}
		}
	}
	return config.Password.Validate()
}

// user config used when only the group group membership needs updating.
type configUserShort struct {
	User   UserID
	Groups *[]GroupName
}

func (config configUserShort) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"groups": GroupName("").arrayToCsv(config.Groups),
	}
}

func (config configUserShort) updateUserMembership(client *Client) (err error) {
	err = updateUser(config.User, config.mapToApiValues(), client)
	if err != nil {
		return fmt.Errorf("error updating User: %v", err)
	}
	return
}

func (configUserShort) updateUsersMembership(users *[]configUserShort, client *Client) (err error) {
	if users == nil {
		return
	}
	for _, e := range *users {
		err = e.updateUserMembership(client)
		if err != nil {
			return
		}
	}
	return
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

// TODO improve when Name and Realm have their own types
func (id UserID) Validate() error {
	if id.Name == "" {
		return errors.New("no username is specified")
	}
	if id.Realm == "" {
		return errors.New("no realm is specified")
	}
	return nil
}

// May be empty or should be at least be 5 characters long.
type UserPassword string

func (password UserPassword) Validate() error {
	if utf8.RuneCountInString(string(password)) >= 5 || password == "" {
		return nil
	}
	return errors.New("the minimum password length is 5")
}

// Check if the user already exists in proxmox.
func CheckUserExistence(userId UserID, client *Client) (existence bool, err error) {
	list, err := listUsersFull(client)
	if err != nil {
		return
	}
	// This should be the case where you have an API Token with privilege separation but no permissions attached
	if len(list) == 0 {
		return false, fmt.Errorf("user %s has valid credentials but cannot retrieve user list, check privilege separation of api token", userId.ToString())
	}
	existence = ItemInKeyOfArray(list, "userid", userId.ToString())
	return
}

// List all users that exist in proxmox
// Setting full to TRUE the output wil include group information.
// Depending on the number of existing groups it take substantially longer to parse
func ListUsers(client *Client, full bool) (*[]ConfigUser, error) {
	var err error
	var userList []interface{}
	if full {
		userList, err = listUsersFull(client)
	} else {
		userList, err = listUsersPartial(client)
	}
	if err != nil {
		return nil, err
	}
	return ConfigUser{}.mapToArray(userList), nil
}

// Returns users without group information
func listUsersPartial(client *Client) ([]interface{}, error) {
	return client.GetItemListInterfaceArray("/access/users")
}

// Returns users with group information
func listUsersFull(client *Client) ([]interface{}, error) {
	return client.GetItemListInterfaceArray("/access/users?full=1")
}

func NewConfigUserFromApi(userId UserID, client *Client) (*ConfigUser, error) {
	userConfig, err := client.GetItemConfigMapStringInterface("/access/users/"+userId.ToString(), "user", "CONFIG")
	if err != nil {
		return nil, err
	}
	return ConfigUser{User: userId}.mapToStruct(userConfig), nil
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

// Converts an comma separated list of "username@realm" to a array of UserID objects
func NewUserIDs(userIds string) (*[]UserID, error) {
	if userIds == "" {
		return &[]UserID{}, nil
	}
	tmpUserIds := strings.Split(userIds, ",")
	users := make([]UserID, len(tmpUserIds))
	for i, e := range tmpUserIds {
		var err error
		users[i], err = NewUserID(e)
		if err != nil {
			return nil, err
		}
	}
	return &users, nil
}

// URL for updating users
func updateUser(user UserID, params map[string]interface{}, client *Client) error {
	return client.Put(params, "/access/users/"+user.ToString())
}
