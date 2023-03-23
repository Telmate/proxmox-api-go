package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ConfigGroup struct {
	Name    GroupName `json:"name"`
	Comment string    `json:"comment,omitempty"`
	// Setting the Members will update the group membership to only include the specified members.
	Members *[]UserID `json:"members,omitempty"`
}

// Creates the specified group
func (config *ConfigGroup) Create(client *Client) error {
	config.Validate(true)
	params := config.mapToApiValues(true)
	err := client.Post(params, "/access/groups")
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Group: %v, (params: %v)", err, string(params))
	}
	if config.Members != nil {
		return config.Name.SetMembers(config.Members, client)
	}
	return nil
}

// Maps the struct to the API values proxmox understands
func (config *ConfigGroup) mapToApiValues(create bool) (params map[string]interface{}) {
	params = map[string]interface{}{
		"comment": config.Comment,
	}
	if create {
		params["groupid"] = string(config.Name)
	}
	return
}

func (config ConfigGroup) mapToStruct(params map[string]interface{}) *ConfigGroup {
	if _, isSet := params["groupid"]; isSet {
		config.Name = GroupName(params["groupid"].(string))
	}
	if _, isSet := params["comment"]; isSet {
		config.Comment = params["comment"].(string)
	}
	if _, isSet := params["members"]; isSet {
		config.Members = UserID{}.mapToArray(params["members"].([]interface{}))
	}
	return &config
}

// Custom error for when the *ConfigGroup is nil
func (config *ConfigGroup) nilCheck() error {
	if config == nil {
		return errors.New("pointer for (ConfigGroup) is nil")
	}
	return nil
}

// Created or updates the specified group
func (config *ConfigGroup) Set(client *Client) (err error) {
	err = config.nilCheck()
	if err != nil {
		return
	}
	existence, err := config.Name.CheckExistence(client)
	if err != nil {
		return
	}
	if existence {
		return config.Update(client)
	}
	return config.Create(client)
}

// Updates the specified group
func (config *ConfigGroup) Update(client *Client) error {
	config.Validate(false)
	params := config.mapToApiValues(true)
	err := client.Put(params, "/access/groups/"+string(config.Name))
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating Group: %v, (params: %v)", err, string(params))
	}
	if config.Members != nil {
		return config.Name.SetMembers(config.Members, client)
	}
	return nil
}

// Validates all items and sub items of the ConfigGroup
func (config *ConfigGroup) Validate(create bool) (err error) {
	err = config.nilCheck()
	if err != nil {
		return
	}
	if create {
		err = config.Name.Validate()
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

// GroupName may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_
type GroupName string

// Add users to the specified group
func (group GroupName) AddUsersToGroup(members *[]UserID, client *Client) error {
	users, err := listUsersFull(client)
	if err != nil {
		return err
	}
	return configUserShort{}.updateUsersMembership(group.usersToAddToGroup(users, members), client)
}

// Convert a array of GroupName to a comma separated string
func (GroupName) arrayToCsv(groupList *[]GroupName) (groups string) {
	if groupList == nil {
		return
	}
	switch len(*groupList) {
	case 0:
		return
	case 1:
		return string((*groupList)[0])
	}
	for i, e := range *groupList {
		if i > 0 {
			groups = groups + ","
		}
		groups = groups + string(e)
	}
	return
}

// Check if the specified group exists.
func (group GroupName) CheckExistence(client *Client) (bool, error) {
	err := group.Validate()
	if err != nil {
		return false, nil
	}
	paramArray, err := listGroups(client)
	if err != nil {
		return false, nil
	}
	return ItemInKeyOfArray(paramArray, "groupid", string(group)), nil
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

// Deletes the specified group
func (group GroupName) Delete(client *Client) (err error) {
	err = group.Validate()
	if err != nil {
		return
	}
	return client.Delete("/access/groups/" + string(group))
}

func (group GroupName) inArray(groups []GroupName) bool {
	if group == "" || groups == nil {
		return false
	}
	for _, e := range groups {
		if e == group {
			return true
		}
	}
	return false
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

// Recursively remove all users from the specified group
func (group GroupName) RemoveAllUsersFromGroup(client *Client) (err error) {
	users, err := listUsersFull(client)
	if err != nil {
		return
	}
	return configUserShort{}.updateUsersMembership(group.removeAllUsersFromGroup(users), client)
}

// Generate a array of users with their updated group memberships.
// This list only includes users who where a member of the specified GroupName.
func (group GroupName) removeAllUsersFromGroup(allUsers []interface{}) *[]configUserShort {
	usersToUpdate := []configUserShort{}
	for _, e := range allUsers {
		params := e.(map[string]interface{})
		if _, isSet := params["userid"]; !isSet {
			continue
		}
		if _, isSet := params["groups"]; !isSet {
			continue
		}
		groups := GroupName("").csvToArray(params["groups"].(string))
		if group.inArray(groups) {
			groups = group.removeFromArray(groups)
			usersToUpdate = append(usersToUpdate, configUserShort{User: UserID{}.mapToStruct(params["userid"].(string)), Groups: &groups})
		}
	}
	return &usersToUpdate
}

func (group GroupName) removeAllUsersFromGroupExcept(allUsers []interface{}, members *[]UserID) *[]configUserShort {
	if group == "" {
		return nil
	}
	if members == nil {
		return group.removeAllUsersFromGroup(allUsers)
	}
	if len(*members) == 0 {
		return group.removeAllUsersFromGroup(allUsers)
	}
	usersToUpdate := []configUserShort{}
	for _, e := range allUsers {
		params := e.(map[string]interface{})
		if _, isSet := params["userid"]; !isSet {
			continue
		}
		var userInMembers bool
		for _, ee := range *members {
			if params["userid"] == ee.ToString() {
				userInMembers = true
				break
			}
		}
		if !userInMembers {
			var groups []GroupName
			if _, isSet := params["groups"]; isSet {
				groups = GroupName("").csvToArray(params["groups"].(string))
			}
			if group.inArray(groups) {
				groups = group.removeFromArray(groups)
				usersToUpdate = append(usersToUpdate, configUserShort{User: UserID{}.mapToStruct(params["userid"].(string)), Groups: &groups})
			}
		}
	}
	return &usersToUpdate
}

// Remove the specified GroupName from the array of GroupName
func (group GroupName) removeFromArray(groups []GroupName) []GroupName {
	newGroups := []GroupName{}
	for _, e := range groups {
		if e != group {
			newGroups = append(newGroups, e)
		}
	}
	return newGroups
}

// Remove users from the specified group
func (group GroupName) RemoveUsersFromGroup(members *[]UserID, client *Client) (err error) {
	users, err := listUsersFull(client)
	if err != nil {
		return err
	}
	return configUserShort{}.updateUsersMembership(group.usersToRemoveFromGroup(users, members), client)
}

// Recursively add and remove users from the specified group so only the the specified users will be members of the group
func (group GroupName) SetMembers(members *[]UserID, client *Client) (err error) {
	users, err := listUsersFull(client)
	if err != nil {
		return
	}
	err = configUserShort{}.updateUsersMembership(group.removeAllUsersFromGroupExcept(users, members), client)
	if err != nil {
		return
	}
	return configUserShort{}.updateUsersMembership(group.usersToAddToGroup(users, members), client)
}

func (group GroupName) usersToAddToGroup(allUsers []interface{}, members *[]UserID) *[]configUserShort {
	if group == "" || members == nil {
		return nil
	}
	usersToUpdate := []configUserShort{}
	for _, e := range allUsers {
		params := e.(map[string]interface{})
		if _, isSet := params["userid"]; !isSet {
			continue
		}
		for ii, ee := range *members {
			if params["userid"] == ee.ToString() {
				var groups []GroupName
				if _, isSet := params["groups"]; isSet {
					groups = GroupName("").csvToArray(params["groups"].(string))
				}
				if !group.inArray(groups) {
					groups = append(groups, group)
					usersToUpdate = append(usersToUpdate, configUserShort{User: (*members)[ii], Groups: &groups})
				}
			}
		}
	}
	return &usersToUpdate
}

func (group GroupName) usersToRemoveFromGroup(allUsers []interface{}, members *[]UserID) *[]configUserShort {
	if group == "" || members == nil {
		return nil
	}
	usersToUpdate := []configUserShort{}
	for _, e := range allUsers {
		params := e.(map[string]interface{})
		if _, isSet := params["userid"]; !isSet {
			continue
		}
		for ii, ee := range *members {
			if params["userid"] == ee.ToString() {
				var groups []GroupName
				if _, isSet := params["groups"]; isSet {
					groups = GroupName("").csvToArray(params["groups"].(string))
				}
				if group.inArray(groups) {
					groups = group.removeFromArray(groups)
					usersToUpdate = append(usersToUpdate, configUserShort{User: (*members)[ii], Groups: &groups})
				}
			}
		}
	}
	return &usersToUpdate
}

// Check if a groupname is valid.
func (group GroupName) Validate() error {
	if group == "" {
		return errors.New("variable of type (GroupName) may not be empty")
	}
	// proxmox does not seem to enforce any limit on the length of a group name. When going over thousands of charters the ui kinda breaks.
	if len([]rune(group)) > 1000 {
		return errors.New("variable of type (GroupName) may not be more tha 1000 characters long")
	}
	regex, _ := regexp.Compile(`^([a-z]|[A-Z]|[0-9]|_|-)*$`)
	if regex.Match([]byte(group)) {
		return nil
	}
	return errors.New("")
}

// Returns a list of all existing groups
func ListGroups(client *Client) (*[]ConfigGroup, error) {
	paramArray, err := listGroups(client)
	if err != nil {
		return nil, err
	}
	groups := make([]ConfigGroup, len(paramArray))
	for i, e := range paramArray {
		groups[i] = *ConfigGroup{}.mapToStruct(e.(map[string]interface{}))
	}
	return &groups, nil
}

// list all groups directly from the api without any extra formatting
func listGroups(client *Client) ([]interface{}, error) {
	return client.GetItemListInterfaceArray("/access/groups")
}

func NewConfigGroupFromApi(groupId GroupName, client *Client) (*ConfigGroup, error) {
	err := groupId.Validate()
	if err != nil {
		return nil, err
	}
	config, err := client.GetItemConfigMapStringInterface("/access/groups/"+string(groupId), "group", "CONFIG")
	if err != nil {
		return nil, err
	}
	return ConfigGroup{Name: groupId}.mapToStruct(config), nil
}
