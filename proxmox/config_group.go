package proxmox

import (
	"errors"
	"regexp"
)

type ConfigGroup struct {
	Name    GroupName `json:"name"`
	Comment string    `json:"comment,omitempty"`
	// Setting the Members will update the group membership to only include the specified members.
	Members *[]UserID `json:"members,omitempty"`
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

// GroupName may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_
type GroupName string

// Deletes the specified group
func (group GroupName) Delete(client *Client) (err error) {
	err = group.Validate()
	if err != nil {
		return
	}
	return client.Delete("/access/groups/" + string(group))
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
