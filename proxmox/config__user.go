package proxmox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Telmate/proxmox-api-go/internal/array"
	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	UserInterface interface {
		Create(context.Context, ConfigUser) error
		CreateNoCheck(context.Context, ConfigUser) error

		Delete(context.Context, UserID) error
		DeleteNoCheck(context.Context, UserID) error

		Exists(context.Context, UserID) (bool, error)
		ExistsNoCheck(context.Context, UserID) (bool, error)

		// List all users.
		List(ctx context.Context) (RawUsersInfo, error)
		ListNoCheck(ctx context.Context) (RawUsersInfo, error)

		// List all users without their group membership and API tokens.
		ListPartial(ctx context.Context) (RawUsersInfo, error)
		ListPartialNoCheck(ctx context.Context) (RawUsersInfo, error)

		// Read the user configuration for the specified userID.
		Read(context.Context, UserID) (RawConfigUser, error)
		ReadNoCheck(context.Context, UserID) (RawConfigUser, error)

		Set(context.Context, ConfigUser) error
		SetNoCheck(context.Context, ConfigUser) error

		Update(context.Context, ConfigUser) error
		UpdateNoCheck(context.Context, ConfigUser) error
	}

	userClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

// User options for the Proxmox API
type ConfigUser struct {
	Comment   *string       `json:"comment,omitempty"`   // Never nil when returned.
	Email     *string       `json:"email,omitempty"`     // Never nil when returned.
	Enable    *bool         `json:"enable"`              // Never nil when returned.
	Expire    *uint         `json:"expire"`              // Never nil when returned.
	FirstName *string       `json:"firstname,omitempty"` // Never nil when returned.
	Groups    *[]GroupName  `json:"groups,omitempty"`    // nil when we did not request group info.
	Keys      *string       `json:"keys,omitempty"`
	LastName  *string       `json:"lastname,omitempty"` // Never nil when returned.
	Password  *UserPassword `json:"password,omitempty"` // Never returned.
	User      UserID        `json:"user"`
}

type UserInfo struct {
	Config ConfigUser
	Tokens *[]ApiTokenConfig
}

var _ UserInterface = (*userClient)(nil)

func (c *userClient) Create(ctx context.Context, config ConfigUser) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.CreateNoCheck(ctx, config)
}

func (c *userClient) CreateNoCheck(ctx context.Context, config ConfigUser) error {
	return config.create(ctx, c.api)
}

func (c *userClient) Delete(ctx context.Context, id UserID) error {
	if err := id.Validate(); err != nil {
		return err
	}
	return c.DeleteNoCheck(ctx, id)
}

func (c *userClient) DeleteNoCheck(ctx context.Context, id UserID) error {
	return id.delete(ctx, c.api)
}

func (c *userClient) Exists(ctx context.Context, id UserID) (bool, error) {
	if err := id.Validate(); err != nil {
		return false, err
	}
	return c.ExistsNoCheck(ctx, id)
}

func (c *userClient) ExistsNoCheck(ctx context.Context, id UserID) (bool, error) {
	return id.exists(ctx, c.api)
}

func (c *userClient) List(ctx context.Context) (RawUsersInfo, error) {
	return c.ListNoCheck(ctx)
}

func (c *userClient) ListNoCheck(ctx context.Context) (RawUsersInfo, error) {
	return userListFull(ctx, c.api)
}

func (c *userClient) ListPartial(ctx context.Context) (RawUsersInfo, error) {
	return c.ListPartialNoCheck(ctx)
}

func (c *userClient) ListPartialNoCheck(ctx context.Context) (RawUsersInfo, error) {
	return userListPartial(ctx, c.api)
}

func (c *userClient) Read(ctx context.Context, id UserID) (RawConfigUser, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	return c.ReadNoCheck(ctx, id)
}

func (c *userClient) ReadNoCheck(ctx context.Context, id UserID) (RawConfigUser, error) {
	raw, exists, err := id.read(ctx, c.api)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("user " + id.String() + " does not exist")
	}
	return raw, nil
}

func (c *userClient) Set(ctx context.Context, config ConfigUser) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.SetNoCheck(ctx, config)
}

func (c *userClient) SetNoCheck(ctx context.Context, config ConfigUser) error {
	exists, err := config.User.exists(ctx, c.api)
	if err != nil {
		return err
	}
	if exists {
		return config.update(ctx, c.api)
	}
	return config.create(ctx, c.api)
}

func (c *userClient) Update(ctx context.Context, config ConfigUser) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, config)
}

func (c *userClient) UpdateNoCheck(ctx context.Context, config ConfigUser) error {
	return config.update(ctx, c.api)
}

const Error_NewUserID string = "no username or realm specified, syntax is \"username@realm\""

// Deprecated: use UserInterface.Create() instead.
func (config ConfigUser) CreateUser(ctx context.Context, client *Client) error {
	return client.New().User.Create(ctx, config)
}

func (config ConfigUser) create(ctx context.Context, c *clientAPI) error {
	if err := c.postRawRetry(ctx, "/access/users", config.mapToApiCreate(), 3); err != nil {
		return errors.New("error creating User: " + err.Error())
	}
	if config.Password != nil {
		return config.User.setPassword(ctx, *config.Password, c)
	}
	return nil
}

// Deprecated: use UserInterface.Delete() instead.
func (config ConfigUser) DeleteUser(ctx context.Context, client *Client) error {
	return client.New().User.Delete(ctx, config.User)
}

func (config ConfigUser) mapToApiCreate() *[]byte {
	builder := strings.Builder{}
	builder.WriteString(userApiKeyUserID + "=" + config.User.String())
	if config.Comment != nil && *config.Comment != "" {
		builder.WriteString("&" + userApiKeyComment + "=" + body.Escape(*config.Comment))
	}
	if config.Email != nil && *config.Email != "" {
		builder.WriteString("&" + userApiKeyEmail + "=" + body.Escape(*config.Email))
	}
	if config.Enable != nil && !*config.Enable { // defaults to enabled when unset
		builder.WriteString("&" + userApiKeyEnable + "=0")
	}
	if config.Expire != nil && *config.Expire != 0 {
		builder.WriteString("&" + userApiKeyExpire + "=" + strconv.FormatUint(uint64(*config.Expire), 10))
	}
	if config.FirstName != nil && *config.FirstName != "" {
		builder.WriteString("&" + userApiKeyFirstName + "=" + body.Escape(*config.FirstName))
	}
	if config.Groups != nil && len(*config.Groups) != 0 {
		builder.WriteString("&" + userApiKeyGroups + "=" + body.Escape(array.CSV(*config.Groups)))
	}
	if config.Keys != nil && *config.Keys != "" {
		builder.WriteString("&" + userApiKeyKeys + "=" + body.Escape(*config.Keys))
	}
	if config.LastName != nil && *config.LastName != "" {
		builder.WriteString("&" + userApiKeyLastName + "=" + body.Escape(*config.LastName))
	}
	b := []byte(builder.String())
	return &b
}

func (config ConfigUser) mapToApiUpdate() *[]byte {
	builder := strings.Builder{}
	if config.Comment != nil {
		builder.WriteString("&" + userApiKeyComment + "=" + body.Escape(*config.Comment))
	}
	if config.Email != nil {
		builder.WriteString("&" + userApiKeyEmail + "=" + body.Escape(*config.Email))
	}
	if config.Enable != nil {
		builder.WriteString("&" + userApiKeyEnable + "=")
		if *config.Enable {
			builder.WriteString("1")
		} else {
			builder.WriteString("0")
		}
	}
	if config.Expire != nil {
		builder.WriteString("&" + userApiKeyExpire + "=" + strconv.FormatUint(uint64(*config.Expire), 10))
	}
	if config.FirstName != nil {
		builder.WriteString("&" + userApiKeyFirstName + "=" + body.Escape(*config.FirstName))
	}
	if config.Groups != nil {
		builder.WriteString("&" + userApiKeyGroups + "=" + body.Escape(array.CSV(*config.Groups)))
	}
	if config.Keys != nil {
		builder.WriteString("&" + userApiKeyKeys + "=" + body.Escape(*config.Keys))
	}
	if config.LastName != nil {
		builder.WriteString("&" + userApiKeyLastName + "=" + body.Escape(*config.LastName))
	}
	if builder.Len() > 0 {
		b := bytes.NewBufferString(builder.String()[1:]).Bytes()
		return &b
	}
	return nil
}

func (ConfigUser) mapToArray(params []any) *[]ConfigUser {
	users := make([]ConfigUser, len(params))
	for i, e := range params {
		users[i] = *(&rawConfigUser{a: e.(map[string]any)}).Get()
	}
	return &users
}

// Deprecated: use UserInterface.Set() instead.
// Create or update the user depending on if the user already exists or not.
// "userId" and "password" overwrite what is specified in "*ConfigUser".
func (config *ConfigUser) SetUser(ctx context.Context, userId UserID, password UserPassword, client *Client) (err error) {
	if config != nil {
		config.User = userId
		config.Password = &password
	}

	userExists, err := CheckUserExistence(ctx, userId, client)
	if err != nil {
		return err
	}

	if config != nil {
		if userExists {
			err = config.UpdateUser(ctx, client)
			if err != nil {
				return err
			}
		} else {
			err = config.CreateUser(ctx, client)
		}
	} else {
		config = &ConfigUser{
			Password: &password,
			User:     userId,
		}
		if userExists {
			if config.Password != nil && *config.Password != "" {
				err = config.UpdateUserPassword(ctx, client)
			}
		} else {
			err = config.CreateUser(ctx, client)
		}
	}
	return
}

// Deprecated: use UserInterface.Update() instead.
func (config ConfigUser) UpdateUser(ctx context.Context, client *Client) (err error) {
	return client.New().User.Update(ctx, config)
}

// Deprecated: use UserInterface.Update() instead.
func (config ConfigUser) UpdateUserPassword(ctx context.Context, client *Client) (err error) {
	err = config.Password.Validate()
	if err != nil {
		return err
	}
	return client.Put(ctx, map[string]interface{}{
		"userid":   config.User.String(),
		"password": config.Password,
	}, "/access/password")
}

func (config ConfigUser) update(ctx context.Context, c *clientAPI) error {
	if body := config.mapToApiUpdate(); body != nil {
		if err := c.updateUser(ctx, config.User, body); err != nil {
			return err
		}
	}
	if config.Password != nil {
		return config.User.setPassword(ctx, *config.Password, c)
	}
	return nil
}

// Deprecated: remove when ConfigUser.CreateApiToken() is removed.
type ApiTokenCreateResult struct {
	Info  map[string]interface{} `json:"info"`
	Value string                 `json:"value"`
}

// Deprecated: remove when ConfigUser.CreateApiToken() is removed.
type ApiTokenCreateResultWrapper struct {
	Data ApiTokenCreateResult `json:"data"`
}

// Deprecated: remove when ConfigUser.ListApiTokens() is removed.
// Maps the API values from proxmox to a struct
func (tokens ApiTokenConfig) mapToStruct(params map[string]interface{}) *ApiTokenConfig {
	if _, isSet := params["tokenid"]; isSet {
		tokens.Name = ApiTokenName(params["tokenid"].(string))
	}
	if _, isSet := params["comment"]; isSet {
		tokens.Comment = util.Pointer(params["comment"].(string))
	}
	if _, isSet := params["expire"]; isSet {
		tokens.Expiration = util.Pointer(uint(params["expire"].(float64)))
	}
	if _, isSet := params["privsep"]; isSet {
		tokens.PrivilegeSeparation = util.Pointer(false)
		if params["privsep"] == 1 {
			tokens.PrivilegeSeparation = util.Pointer(true)
		}
	}
	return &tokens
}

// Deprecated: remove when ConfigUser.ListApiTokens() is removed.
func (ApiTokenConfig) mapToArray(params []interface{}) *[]ApiTokenConfig {
	tokens := make([]ApiTokenConfig, len(params))
	for i, e := range params {
		tokens[i] = *ApiTokenConfig{}.mapToStruct(e.(map[string]interface{}))
	}
	return &tokens
}

// Deprecated: use UserInterface.CreateApiToken() instead.
func (config ConfigUser) CreateApiToken(ctx context.Context, client *Client, token ApiTokenConfig) (value string, err error) {
	status, err := client.CreateItemReturnStatus(ctx, map[string]interface{}{
		"comment": token.Comment,
		"expire":  token.Expiration,
		"privsep": token.PrivilegeSeparation,
	}, "/access/users/"+config.User.String()+"/token/"+token.Name.String())
	if err != nil {
		return
	}
	var resultWrapper *ApiTokenCreateResultWrapper
	err = json.Unmarshal([]byte(status), &resultWrapper)
	value = resultWrapper.Data.Value
	return
}

// Deprecated: use ApiTokenInterface.Update() instead.
func (config ConfigUser) UpdateApiToken(ctx context.Context, client *Client, token ApiTokenConfig) (err error) {
	err = client.Put(ctx, map[string]interface{}{
		"comment": token.Comment,
		"expire":  token.Expiration,
		"privsep": token.PrivilegeSeparation,
	}, "/access/users/"+config.User.String()+"/token/"+token.Name.String())
	return
}

// Deprecated: use ApiTokenInterface.List() instead.
func (config ConfigUser) ListApiTokens(ctx context.Context, client *Client) (tokens *[]ApiTokenConfig, err error) {
	status, err := client.GetItemListInterfaceArray(ctx, "/access/users/"+config.User.String()+"/token")
	if err != nil {
		return
	}
	tokens = ApiTokenConfig{}.mapToArray(status)
	return
}

// Deprecated: use ApiTokenInterface.Delete() instead.
func (config ConfigUser) DeleteApiToken(ctx context.Context, client *Client, token ApiTokenConfig) (err error) {
	err = client.Delete(ctx, "/access/users/"+config.User.String()+"/token/"+token.Name.String())
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
			for i := range *config.Groups {
				err = (*config.Groups)[i].Validate()
				if err != nil {
					return
				}
			}
		}
	}
	if config.Password != nil {
		err = config.Password.Validate()
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

func (id UserID) addGroups(ctx context.Context, groups *[]GroupName, c *clientAPI) error {
	raw, exists, err := id.read(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user " + id.String() + " does not exist")
	}
	newGroups := array.Combine(*(raw.GetGroups()), *groups)
	return id.setGroups(ctx, &newGroups, c)
}

func (id UserID) delete(ctx context.Context, client *clientAPI) error {
	return client.deleteRetry(ctx, "/access/users/"+id.String(), 3)
}

func (id UserID) exists(ctx context.Context, c clientApiInterface) (bool, error) {
	_, exists, err := c.getUserConfig(ctx, id)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Parses "username@realm" to a UserID object
func (id *UserID) Parse(userID string) error {
	index := strings.IndexRune(userID, '@')
	if index == -1 || index == 0 || index == len(userID)-1 {
		return errors.New(Error_NewUserID)
	}
	id.Name = userID[:index]
	id.Realm = userID[index+1:]
	return nil
}

func (id UserID) listApiTokens(ctx context.Context, c *clientAPI) (RawApiTokens, error) {
	params, err := c.getList(ctx, "/access/users/"+id.String()+"/token", "List", "API Tokens")
	if err != nil {
		return nil, err
	}
	return &rawApiTokens{a: params}, nil
}

func (id UserID) read(ctx context.Context, c *clientAPI) (*rawConfigUser, bool, error) {
	userConfig, exists, err := c.getUserConfig(ctx, id)
	if err != nil {
		return nil, false, err
	}
	return &rawConfigUser{a: userConfig, user: &id}, exists, nil
}

func (id UserID) removeGroups(ctx context.Context, groups *[]GroupName, c *clientAPI) error {
	raw, exists, err := id.read(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("user " + id.String() + " does not exist")
	}
	newGroups := array.Subtract(*(raw.GetGroups()), *groups)
	return id.setGroups(ctx, &newGroups, c)
}

func (id UserID) setGroups(ctx context.Context, groups *[]GroupName, c *clientAPI) error {
	b := []byte(userApiKeyGroups + "=" + array.CSV(*groups))
	return c.updateUser(ctx, id, &b)
}

func (id UserID) setPassword(ctx context.Context, password UserPassword, c *clientAPI) error {
	body := []byte(userApiKeyUserID + "=" + url.QueryEscape(id.String()) + "&" + userApiKeyPassword + "=" + url.QueryEscape(password.String()))
	err := c.putRawRetry(ctx, "/access/password", &body, 3)
	if err != nil {
		return errors.New("error setting password: " + err.Error())
	}
	return nil
}

// Converts the userID to "username@realm"
// Returns an empty string when either the Name or Realm is empty
func (id UserID) String() string { // String is for fmt.Stringer.
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
	if utf8.RuneCountInString(string(password)) >= 8 || password == "" {
		return nil
	}
	return errors.New("the minimum password length is 8 characters")
}

func (password UserPassword) String() string { return string(password) } // String is for fmt.Stringer.

type (
	RawConfigUser interface {
		Get() *ConfigUser
		GetComment() string
		GetEmail() string
		GetEnable() bool
		GetExpire() uint
		GetFirstName() string
		GetGroups() *[]GroupName
		GetKeys() *string
		GetLastName() string
		GetUser() UserID
	}

	rawConfigUser struct {
		a    map[string]any
		user *UserID
	}
)

var _ RawConfigUser = (*rawConfigUser)(nil)

func (r *rawConfigUser) Get() *ConfigUser {
	return &ConfigUser{
		Comment:   util.Pointer(r.GetComment()),
		Email:     util.Pointer(r.GetEmail()),
		Enable:    util.Pointer(r.GetEnable()),
		Expire:    util.Pointer(r.GetExpire()),
		FirstName: util.Pointer(r.GetFirstName()),
		Groups:    r.GetGroups(),
		Keys:      r.GetKeys(),
		LastName:  util.Pointer(r.GetLastName()),
		User:      r.GetUser()}
}

func (r *rawConfigUser) GetComment() string { return userGetComment(r.a) }

func (r *rawConfigUser) GetEmail() string { return userGetEmail(r.a) }

func (r *rawConfigUser) GetEnable() bool { return userGetEnable(r.a) }

func (r *rawConfigUser) GetExpire() uint { return userGetExpire(r.a) }

func (r *rawConfigUser) GetFirstName() string { return userGetFirstName(r.a) }

func (r *rawConfigUser) GetGroups() *[]GroupName { return userGetGroups(r.a) }

func (r *rawConfigUser) GetKeys() *string { return userGetKeys(r.a) }

func (r *rawConfigUser) GetLastName() string { return userGetLastName(r.a) }

func (r *rawConfigUser) GetUser() UserID { return userGetUser(r.a, r.user) }

type (
	RawUsersInfo interface {
		AsArray() []RawUserInfo
		AsMap() map[UserID]RawUserInfo
		Len() int
		SelectUser(UserID) (RawUserInfo, bool)
	}

	rawUsersInfo struct {
		a    []any
		full bool
	}
)

func (r *rawUsersInfo) AsArray() []RawUserInfo {
	raw := make([]RawUserInfo, len(r.a))
	for i := range r.a {
		raw[i] = &rawUserInfo{a: r.a[i].(map[string]any), full: r.full}
	}
	return raw
}

func (r *rawUsersInfo) AsMap() map[UserID]RawUserInfo {
	raw := make(map[UserID]RawUserInfo, len(r.a))
	for i := range r.a {
		tmpMap := r.a[i].(map[string]any)
		var id UserID
		_ = id.Parse(tmpMap[userApiKeyUserID].(string))
		raw[id] = &rawUserInfo{a: tmpMap, full: r.full, user: &id}
	}
	return raw
}

func (r *rawUsersInfo) Len() int { return len(r.a) }

func (r *rawUsersInfo) SelectUser(user UserID) (RawUserInfo, bool) {
	for i := range r.a {
		raw := r.a[i].(map[string]any)
		if vv, ok := raw[userApiKeyUserID]; ok && vv == user.String() {
			return &rawUserInfo{a: raw, full: r.full, user: &user}, true
		}
	}
	return nil, false
}

var _ RawUsersInfo = (*rawUsersInfo)(nil)

type (
	RawUserInfo interface {
		Get() UserInfo
		GetConfig() ConfigUser
		GetConfigComment() string
		GetConfigEmail() string
		GetConfigEnable() bool
		GetConfigExpire() uint
		GetConfigFirstName() string
		GetConfigGroups() *[]GroupName
		GetConfigKeys() *string
		GetConfigLastName() string
		GetConfigUser() UserID
		GetTokens() *[]ApiTokenConfig
	}

	rawUserInfo struct {
		a    map[string]any
		full bool
		user *UserID
	}
)

var _ RawUserInfo = (*rawUserInfo)(nil)

func (r *rawUserInfo) Get() UserInfo {
	return UserInfo{
		Config: r.GetConfig(),
		Tokens: r.GetTokens()}
}

func (r *rawUserInfo) GetConfig() ConfigUser {
	return ConfigUser{
		Comment:   util.Pointer(r.GetConfigComment()),
		Email:     util.Pointer(r.GetConfigEmail()),
		Enable:    util.Pointer(r.GetConfigEnable()),
		Expire:    util.Pointer(r.GetConfigExpire()),
		FirstName: util.Pointer(r.GetConfigFirstName()),
		Groups:    r.GetConfigGroups(),
		Keys:      r.GetConfigKeys(),
		LastName:  util.Pointer(r.GetConfigLastName()),
		User:      r.GetConfigUser()}
}

func (r *rawUserInfo) GetConfigComment() string { return userGetComment(r.a) }

func (r *rawUserInfo) GetConfigEmail() string { return userGetEmail(r.a) }

func (r *rawUserInfo) GetConfigEnable() bool { return userGetEnable(r.a) }

func (r *rawUserInfo) GetConfigExpire() uint { return userGetExpire(r.a) }

func (r *rawUserInfo) GetConfigFirstName() string { return userGetFirstName(r.a) }

func (r *rawUserInfo) GetConfigGroups() *[]GroupName {
	if r.full {
		return userGetGroups(r.a)
	}
	return nil
}

func (r *rawUserInfo) GetConfigKeys() *string { return userGetKeys(r.a) }

func (r *rawUserInfo) GetConfigLastName() string { return userGetLastName(r.a) }

func (r *rawUserInfo) GetConfigUser() UserID { return userGetUser(r.a, r.user) }

func (r *rawUserInfo) GetTokens() *[]ApiTokenConfig {
	if !r.full {
		return nil
	}
	if v, isSet := r.a[userApiKeyTokens]; isSet && v != nil {
		tmpMaps := v.([]any)
		results := make([]ApiTokenConfig, len(tmpMaps))
		for i := range tmpMaps {
			tmpMap := tmpMaps[i].(map[string]any)
			results[i] = ApiTokenConfig{
				Comment:             util.Pointer(apiTokenGetComment(tmpMap)),
				Expiration:          util.Pointer(apiTokenGetExpiration(tmpMap)),
				Name:                apiTokenGetName(tmpMap, nil),
				PrivilegeSeparation: util.Pointer(apiTokenGetPrivilegeSeparation(tmpMap))}
		}
		return &results
	}
	return &[]ApiTokenConfig{}
}

// Deprecated: use UserInterface.Exists() instead.
// Check if the user already exists in proxmox.
func CheckUserExistence(ctx context.Context, userId UserID, client *Client) (existence bool, err error) {
	list, err := listUsersFull(ctx, client)
	if err != nil {
		return
	}
	// This should be the case where you have an API Token with privilege separation but no permissions attached
	if len(list) == 0 {
		return false, fmt.Errorf("user %s has valid credentials but cannot retrieve user list, check privilege separation of api token", userId.String())
	}
	existence = ItemInKeyOfArray(list, "userid", userId.String())
	return
}

// Deprecated: use UserInterface.List() instead.
func ListUsers(ctx context.Context, client *Client, full bool) (*[]ConfigUser, error) {
	var err error
	var userList []interface{}
	if full {
		userList, err = listUsersFull(ctx, client)
	} else {
		userList, err = listUsersPartial(ctx, client)
	}
	if err != nil {
		return nil, err
	}
	return ConfigUser{}.mapToArray(userList), nil
}

// Deprecated: remove with ListUsers().
func listUsersPartial(ctx context.Context, client *Client) ([]interface{}, error) {
	return client.GetItemListInterfaceArray(ctx, "/access/users")
}

// Deprecated: remove with ListUsers().
func listUsersFull(ctx context.Context, client *Client) ([]interface{}, error) {
	return client.GetItemListInterfaceArray(ctx, "/access/users?full=1")
}

// Returns users without group information
func userListFull(ctx context.Context, c *clientAPI) (*rawUsersInfo, error) {
	params, err := c.getList(ctx, "/access/users?full=1", "List", "Users")
	if err != nil {
		return nil, err
	}
	return &rawUsersInfo{a: params, full: true}, nil
}

// Returns users with group information
func userListPartial(ctx context.Context, c *clientAPI) (*rawUsersInfo, error) {
	params, err := c.getList(ctx, "/access/users", "List", "Users")
	if err != nil {
		return nil, err
	}
	return &rawUsersInfo{a: params, full: false}, nil
}

// Deprecated: use UserInterface.Read() instead.
func NewRawConfigUserFromApi(ctx context.Context, userID UserID, c *Client) (RawConfigUser, error) {
	return c.new().userGetRawConfig(ctx, userID)
}

// Deprecated: remove when NewRawConfigUserFromApi is removed.
func (c *clientNewTest) userGetRawConfig(ctx context.Context, userID UserID) (RawConfigUser, error) {
	if err := userID.Validate(); err != nil {
		return nil, err
	}
	raw, _, err := userID.read(ctx, c.apiRaw())
	return raw, err
}

// Deprecated: use json.Unmarshal() instead.
func NewConfigUserFromJson(input []byte) (config *ConfigUser, err error) {
	if len(input) != 0 {
		config = &ConfigUser{}
		err = json.Unmarshal([]byte(input), config)
	}
	return
}

// Deprecated: use UserID.Parse() instead
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

// Deprecated:
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

func userGetComment(params map[string]any) string {
	if v, isSet := params[userApiKeyComment]; isSet {
		return v.(string)
	}
	return ""
}

func userGetEmail(params map[string]any) string {
	if v, isSet := params[userApiKeyEmail]; isSet {
		return v.(string)
	}
	return ""
}

func userGetEnable(params map[string]any) bool {
	if v, isSet := params[userApiKeyEnable]; isSet {
		return Itob(int(v.(float64)))
	}
	return false
}

func userGetExpire(params map[string]any) uint {
	if v, isSet := params[userApiKeyExpire]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func userGetFirstName(params map[string]any) string {
	if v, isSet := params[userApiKeyFirstName]; isSet {
		return v.(string)
	}
	return ""
}

func userGetGroups(params map[string]any) *[]GroupName {
	if v, isSet := params[userApiKeyGroups]; isSet {
		return GroupName("").mapToArray(v)
	}
	return &[]GroupName{}
}

func userGetKeys(params map[string]any) *string {
	if v, isSet := params[userApiKeyKeys]; isSet {
		return util.Pointer(v.(string))
	}
	return nil
}

func userGetLastName(params map[string]any) string {
	if v, isSet := params[userApiKeyLastName]; isSet {
		return v.(string)
	}
	return ""
}

func userGetUser(params map[string]any, id *UserID) UserID {
	if id != nil {
		return *id
	}
	var user UserID
	if v, isSet := params[userApiKeyUserID]; isSet {
		_ = user.Parse(v.(string))
	}
	return user
}

const (
	userApiKeyComment   string = "comment"
	userApiKeyTokens    string = "tokens"
	userApiKeyEmail     string = "email"
	userApiKeyEnable    string = "enable"
	userApiKeyExpire    string = "expire"
	userApiKeyFirstName string = "firstname"
	userApiKeyGroups    string = "groups"
	userApiKeyKeys      string = "keys"
	userApiKeyLastName  string = "lastname"
	userApiKeyPassword  string = "password"
	userApiKeyUserID    string = "userid"
)
