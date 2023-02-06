package proxmox

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
