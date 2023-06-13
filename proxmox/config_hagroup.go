package proxmox

import (
	"errors"
	"strings"
)

type HAGroup struct {
	Comment    string   // Description.
	Group      string   // The HA group identifier.
	Nodes      []string // List of cluster node names with optional priority. LIKE: <node>[:<pri>]{,<node>[:<pri>]}*
	NoFailback bool     // The CRM tries to run services on the node with the highest priority. If a node with higher priority comes online, the CRM migrates the service to that node. Enabling nofailback prevents that behavior.
	Restricted bool     // Resources bound to restricted groups may only run on nodes defined by the group.
	Type       string   // Group type
}

func (c *Client) GetHAGroupList() (haGroups []HAGroup, err error) {
	list, err := c.GetItemList("/cluster/ha/groups")

	if err != nil {
		return nil, err
	}

	haGroups = []HAGroup{}

	for _, item := range list["data"].([]interface{}) {
		itemMap := item.(map[string]interface{})

		haGroups = append(haGroups, HAGroup{
			Comment:    itemMap["comment"].(string),
			Group:      itemMap["group"].(string),
			Nodes:      strings.Split(itemMap["nodes"].(string), ","),
			NoFailback: itemMap["nofailback"].(float64) == 1,
			Restricted: itemMap["restricted"].(float64) == 1,
			Type:       itemMap["type"].(string),
		})
	}

	return haGroups, nil
}

func (c *Client) GetHAGroupByName(GroupName string) (*HAGroup, error) {
	groups, err := c.GetHAGroupList()

	if err != nil {
		return nil, err
	}

	for _, item := range groups {
		if item.Group == GroupName {
			return &item, nil
		}
	}

	return nil, errors.New("cannot find HaGroup by name " + GroupName)
}
