package proxmox

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/array"
	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	GroupInterface interface {
		AddMembers(context.Context, []GroupName, []UserID) error
		AddMembersNoCheck(context.Context, []GroupName, []UserID) error

		Create(context.Context, ConfigGroup) error
		CreateNoCheck(context.Context, ConfigGroup) error

		Delete(context.Context, GroupName) (bool, error)
		DeleteNoCheck(context.Context, GroupName) (bool, error)

		Exists(context.Context, GroupName) (bool, error)
		ExistsNoCheck(context.Context, GroupName) (bool, error)

		// List all groups.
		List(ctx context.Context) (RawGroups, error)
		ListNoCheck(ctx context.Context) (RawGroups, error)

		Read(context.Context, GroupName) (RawGroupConfig, error)
		ReadNoCheck(context.Context, GroupName) (RawGroupConfig, error)

		RemoveMembers(context.Context, []GroupName, []UserID) error
		RemoveMembersNoCheck(context.Context, []GroupName, []UserID) error

		Set(context.Context, ConfigGroup) error
		SetNoCheck(context.Context, ConfigGroup) error

		Update(context.Context, ConfigGroup) error
		UpdateNoCheck(context.Context, ConfigGroup) error
	}

	groupClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ GroupInterface = (*groupClient)(nil)

func (c *groupClient) AddMembers(ctx context.Context, groups []GroupName, members []UserID) error {
	var err error
	for i := range groups {
		if err = groups[i].Validate(); err != nil {
			return err
		}
	}
	for i := range members {
		if err := members[i].Validate(); err != nil {
			return err
		}
	}
	return c.AddMembersNoCheck(ctx, groups, members)
}

func (c *groupClient) AddMembersNoCheck(ctx context.Context, groups []GroupName, members []UserID) error {
	var err error
	for i := range members {
		if err = members[i].addGroups(ctx, &groups, c.api); err != nil {
			return err
		}
	}
	return nil
}

func (c *groupClient) Create(ctx context.Context, config ConfigGroup) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.CreateNoCheck(ctx, config)
}

func (c *groupClient) CreateNoCheck(ctx context.Context, config ConfigGroup) error {
	return config.create(ctx, c.api)
}

func (c *groupClient) Delete(ctx context.Context, name GroupName) (bool, error) {
	if err := name.Validate(); err != nil {
		return false, err
	}
	return c.DeleteNoCheck(ctx, name)
}

func (c *groupClient) DeleteNoCheck(ctx context.Context, name GroupName) (bool, error) {
	return name.delete(ctx, c.api)
}

func (c *groupClient) Exists(ctx context.Context, name GroupName) (bool, error) {
	if err := name.Validate(); err != nil {
		return false, err
	}
	return c.ExistsNoCheck(ctx, name)
}

func (c *groupClient) ExistsNoCheck(ctx context.Context, name GroupName) (bool, error) {
	return name.exists(ctx, c.api)
}

func (c *groupClient) List(ctx context.Context) (RawGroups, error) {
	return c.ListNoCheck(ctx)
}

func (c *groupClient) ListNoCheck(ctx context.Context) (RawGroups, error) {
	return groupsList(ctx, c.api)
}

func (c *groupClient) Read(ctx context.Context, name GroupName) (RawGroupConfig, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	return c.ReadNoCheck(ctx, name)
}

func (c *groupClient) ReadNoCheck(ctx context.Context, name GroupName) (RawGroupConfig, error) {
	raw, exists, err := name.read(ctx, c.api)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("group does not exist")
	}
	return raw, err
}

func (c *groupClient) RemoveMembers(ctx context.Context, groups []GroupName, members []UserID) error {
	for i := range groups {
		if err := groups[i].Validate(); err != nil {
			return err
		}
	}
	for i := range members {
		if err := members[i].Validate(); err != nil {
			return err
		}
	}
	return c.RemoveMembersNoCheck(ctx, groups, members)
}

func (c *groupClient) RemoveMembersNoCheck(ctx context.Context, groups []GroupName, members []UserID) error {
	var err error
	for i := range members {
		if err = members[i].removeGroups(ctx, &groups, c.api); err != nil {
			return err
		}
	}
	return nil
}

func (c *groupClient) Set(ctx context.Context, config ConfigGroup) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.SetNoCheck(ctx, config)
}

func (c *groupClient) SetNoCheck(ctx context.Context, config ConfigGroup) error {
	return config.set(ctx, c.api)
}

func (c *groupClient) Update(ctx context.Context, config ConfigGroup) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, config)
}

func (c *groupClient) UpdateNoCheck(ctx context.Context, config ConfigGroup) error {
	var raw *rawGroupConfig
	if config.Members != nil { // We need the current members to be able to update them.
		var exists bool
		var err error
		raw, exists, err = config.Name.read(ctx, c.api)
		if err != nil {
			return err
		}
		if !exists {
			return errors.New("group does not exist")
		}
	}
	return config.update(ctx, raw, c.api)
}

type ConfigGroup struct {
	Name    GroupName `json:"name"`
	Comment *string   `json:"comment,omitempty"` // Never nil when returned.
	// Setting the Members will update the group membership to only include the specified members.
	Members *[]UserID `json:"members,omitempty"`
}

// Deprecated: use GroupInterface.Create() instead.
// Creates the specified group
func (config ConfigGroup) Create(ctx context.Context, client *Client) error {
	return client.New().Group.Create(ctx, config)
}

func (config ConfigGroup) create(ctx context.Context, c *clientAPI) error {
	if err := c.postRawRetry(ctx, "/access/groups", config.mapToApiCreate(), 3); err != nil {
		return err
	}
	if config.Members != nil {
		return config.Name.addMembers(ctx, config.Members, c)
	}
	return nil
}

// Maps the struct to the API values proxmox understands
func (config ConfigGroup) mapToApiCreate() *[]byte {
	if config.Comment != nil && *config.Comment != "" {
		return util.Pointer([]byte(groupApiKeyName + "=" + config.Name.String() + "&" + groupApiKeyComment + "=" + body.Escape(*config.Comment)))
	}
	return util.Pointer([]byte(groupApiKeyName + "=" + config.Name.String()))
}

// Maps the struct to the API values proxmox understands
func (config ConfigGroup) mapToApiUpdate(current *rawGroupConfig) *[]byte {
	if config.Comment == nil {
		return nil
	}
	if current != nil && *config.Comment == current.GetComment() {
		return nil
	}
	b := []byte(groupApiKeyComment + "=" + body.Escape(*config.Comment))
	return &b
}

// Deprecated: use GroupInterface.Set() instead.
// Created or updates the specified group
func (config ConfigGroup) Set(ctx context.Context, client *Client) error {
	return client.New().Group.Set(ctx, config)
}

func (config ConfigGroup) set(ctx context.Context, c *clientAPI) error {
	raw, exists, err := config.Name.read(ctx, c)
	if err != nil {
		return err
	}
	if !exists { // Create
		return config.create(ctx, c)
	}
	return config.update(ctx, raw, c)
}

// Deprecated: use GroupInterface.Update() instead.
// Updates the specified group
func (config ConfigGroup) Update(ctx context.Context, client *Client) error {
	return client.New().Group.Update(ctx, config)
}

// Updates the specified group
// current is required when we want to update the members
func (config ConfigGroup) update(ctx context.Context, current *rawGroupConfig, c *clientAPI) error {
	if b := config.mapToApiUpdate(current); b != nil {
		if err := c.putRawRetry(ctx, "/access/groups/"+config.Name.String(), b, 3); err != nil {
			return err
		}
	}
	if config.Members != nil {
		currentMembers := current.GetMembers()
		return config.Name.setMembers(ctx, &currentMembers, config.Members, c)
	}
	return nil
}

// Validates all items and sub items of the ConfigGroup
func (config ConfigGroup) Validate() (err error) {
	if err = config.Name.Validate(); err != nil {
		return
	}
	if config.Members != nil {
		for _, e := range *config.Members {
			err = e.Validate()
			if err != nil {
				return
			}
		}
	}
	return
}

type (
	RawGroups interface {
		FormatArray() []RawGroupConfig
		FormatMap() map[GroupName]RawGroupConfig
		Len() int
		SelectName(GroupName) (RawGroupConfig, bool)
	}
	rawGroups struct{ a []any }
)

var _ RawGroups = (*rawGroups)(nil)

func (r *rawGroups) FormatArray() []RawGroupConfig {
	groups := make([]RawGroupConfig, len(r.a))
	for i := range r.a {
		groups[i] = &rawGroupConfig{a: r.a[i].(map[string]any)}
	}
	return groups
}

func (r *rawGroups) FormatMap() map[GroupName]RawGroupConfig {
	groups := make(map[GroupName]RawGroupConfig, len(r.a))
	for i := range r.a {
		raw := rawGroupConfig{a: r.a[i].(map[string]any)}
		group := GroupName(raw.a[groupApiKeyName].(string))
		raw.group = &group
		groups[group] = &raw
	}
	return groups
}

func (r *rawGroups) Len() int { return len(r.a) }

func (r *rawGroups) SelectName(name GroupName) (RawGroupConfig, bool) {
	for i := range r.a {
		tmpMap := r.a[i].(map[string]any)
		if v, ok := tmpMap[groupApiKeyName]; ok {
			if v.(string) == name.String() {
				return &rawGroupConfig{a: tmpMap, group: util.Pointer(name)}, true
			}
		}
	}
	return nil, false
}

type (
	RawGroupConfig interface {
		Get() ConfigGroup
		GetName() GroupName
		GetComment() string
		GetMembers() []UserID
	}
	rawGroupConfig struct {
		a     map[string]any
		group *GroupName // Not always set.
	}
)

var _ RawGroupConfig = (*rawGroupConfig)(nil)

func (r *rawGroupConfig) Get() ConfigGroup {
	return ConfigGroup{
		Name:    r.GetName(),
		Comment: util.Pointer(r.GetComment()),
		Members: util.Pointer(r.GetMembers())}
}

func (r *rawGroupConfig) GetName() GroupName {
	if r.group != nil {
		return *r.group
	}
	if v, ok := r.a[groupApiKeyName]; ok {
		return GroupName(v.(string))
	}
	return ""
}

func (r *rawGroupConfig) GetComment() string {
	if v, ok := r.a[groupApiKeyComment]; ok {
		return v.(string)
	}
	return ""
}

func (r *rawGroupConfig) GetMembers() []UserID {
	if v, ok := r.a[groupApiKeyMembers]; ok {
		rawMembers := v.([]any)
		members := make([]UserID, len(rawMembers))
		for i := range rawMembers {
			var user UserID
			_ = user.Parse(rawMembers[i].(string))
			members[i] = user
		}
		return members
	}
	return []UserID{}
}

// GroupName may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_
type GroupName string

const (
	GroupName_Error_Invalid   string = "variable of type (GroupName) may only contain the following characters: -_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	GroupName_Error_Empty     string = "variable of type (GroupName) may not be empty"
	GroupName_Error_MaxLength string = "variable of type (GroupName) may not be more than 1000 characters long"
)

// Deprecated: use GroupInterface.AddMembers() instead.
// Add users to the specified group
func (group GroupName) AddUsersToGroup(ctx context.Context, members *[]UserID, client *Client) error {
	return client.New().Group.AddMembers(ctx, []GroupName{group}, *members)
}

func (group GroupName) addMembers(ctx context.Context, members *[]UserID, c *clientAPI) error {
	var err error
	for i := range *members {
		if err = (*members)[i].addGroups(ctx, &[]GroupName{group}, c); err != nil {
			return err
		}
	}
	return nil
}

// Deprecated: use GroupInterface.Exists() instead.
// Check if the specified group exists.
func (group GroupName) CheckExistence(ctx context.Context, client *Client) (bool, error) {
	return client.New().Group.Exists(ctx, group)
}

// Convert a comma separated string to an array of GroupName
func (GroupName) csvToArray(csv string) []GroupName {
	if csv == "" {
		return []GroupName{}
	}
	tmpArray := strings.Split(csv, ",")
	groups := make([]GroupName, len(tmpArray))
	for i := range tmpArray {
		groups[i] = GroupName(tmpArray[i])
	}
	return groups
}

// Deprecated: use GroupInterface.Delete() instead.
// Deletes the specified group
func (group GroupName) Delete(ctx context.Context, client *Client) error {
	exists, err := client.New().Group.Delete(ctx, group)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("group does not exist")
	}
	return nil
}

func (group GroupName) delete(ctx context.Context, c *clientAPI) (bool, error) {
	if err := c.deleteRetry(ctx, "/access/groups/"+group.String(), 3); err != nil {
		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(apiErr.Message, "delete group failed: group '"+group.String()+"' does not exist") {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (group GroupName) exists(ctx context.Context, c *clientAPI) (bool, error) {
	_, exists, err := group.read(ctx, c)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// params can only be of type []interface{} or string
func (GroupName) mapToArray(params any) *[]GroupName {
	groupList := []GroupName{}
	switch tmpParams := params.(type) {
	case []interface{}:
		groups := ArrayToStringType(tmpParams)
		if len(groups) == 1 {
			if groups[0] == "" {
				return &groupList
			}
		}
		groupList = make([]GroupName, len(groups))
		for i := range groups {
			groupList[i] = GroupName(groups[i])
		}
	case string:
		groupList = GroupName("").csvToArray(tmpParams)
	}
	return &groupList
}

func (group GroupName) read(ctx context.Context, c *clientAPI) (*rawGroupConfig, bool, error) {
	raw, err := c.getMap(ctx, "/access/groups/"+group.String(), "group", "CONFIG")
	if err != nil {
		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(apiErr.Message, "group '"+group.String()+"' does not exist") {
				return nil, false, nil
			}
		}
		return nil, false, err
	}
	return &rawGroupConfig{a: raw, group: &group}, true, nil
}

// Deprecated:
// Recursively remove all users from the specified group
func (group GroupName) RemoveAllUsersFromGroup(ctx context.Context, client *Client) error {
	raw, err := client.New().Group.Read(ctx, group)
	if err != nil {
		return err
	}
	users := raw.GetMembers()
	for i := range users {
		if err = users[i].removeGroups(ctx, &[]GroupName{group}, client.new().apiRaw()); err != nil {
			return err
		}
	}
	return nil
}

// Deprecated: use GroupInterface.RemoveMembers() instead.
// Remove users from the specified group
func (group GroupName) RemoveUsersFromGroup(ctx context.Context, members *[]UserID, client *Client) error {
	return client.New().Group.RemoveMembers(ctx, []GroupName{group}, *members)
}

// Deprecated: use GroupInterface.Set() instead.
// Recursively add and remove users from the specified group so only the the specified users will be members of the group
func (group GroupName) SetMembers(ctx context.Context, members *[]UserID, client *Client) error {
	return client.New().Group.Set(ctx, ConfigGroup{
		Name:    group,
		Members: members})
}

func (group GroupName) setMembers(ctx context.Context, current, members *[]UserID, c *clientAPI) error {
	membersToRemove := array.RemoveItems(*current, *members)
	membersToAdd := array.RemoveItems(*members, *current)
	var err error
	for i := range membersToRemove {
		if err = membersToRemove[i].removeGroups(ctx, &[]GroupName{group}, c); err != nil {
			return err
		}
	}
	for i := range membersToAdd {
		if err = membersToAdd[i].addGroups(ctx, &[]GroupName{group}, c); err != nil {
			return err
		}
	}
	return nil
}

func (group GroupName) String() string { return string(group) } // For fmt.Stringer interface.

// Check if a groupName is valid.
func (group GroupName) Validate() error {
	if group == "" {
		return errors.New(GroupName_Error_Empty)
	}
	// proxmox does not seem to enforce any limit on the length of a group name. When going over thousands of characters the ui kinda breaks.
	if len([]rune(group)) > 1000 {
		return errors.New(GroupName_Error_MaxLength)
	}
	regex, _ := regexp.Compile(`^([a-z]|[A-Z]|[0-9]|_|-)*$`)
	if !regex.Match([]byte(group)) {
		return errors.New(GroupName_Error_Invalid)
	}
	return nil
}

// Deprecated: use GroupInterface.List() instead.
// Returns a list of all existing groups
func ListGroups(ctx context.Context, client *Client) (*[]ConfigGroup, error) {
	raw, err := client.New().Group.List(ctx)
	if err != nil {
		return nil, err
	}
	rawGroups := raw.FormatArray()
	groups := make([]ConfigGroup, len(rawGroups))
	for i := range rawGroups {
		groups[i] = rawGroups[i].Get()
	}
	return &groups, nil
}

func groupsList(ctx context.Context, c *clientAPI) (*rawGroups, error) {
	params, err := c.getList(ctx, "/access/groups", "List", "Groups")
	if err != nil {
		return nil, err
	}
	return &rawGroups{a: params}, nil
}

// Deprecated: use GroupInterface.Read() instead.
func NewConfigGroupFromApi(ctx context.Context, groupId GroupName, client *Client) (*ConfigGroup, error) {
	raw, err := client.New().Group.Read(ctx, groupId)
	if err != nil {
		return nil, err
	}
	return util.Pointer(raw.Get()), nil
}

const (
	groupApiKeyComment string = "comment"
	groupApiKeyMembers string = "members"
	groupApiKeyName    string = "groupid"
)
