package proxmox

import (
	"encoding/json"
	"fmt"
)

type ConfigSnapshot struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	VmState     bool   `json:"ram,omitempty"`
}

func (config *ConfigSnapshot) mapValues() map[string]interface{} {
	return map[string]interface{}{
		"snapname":    config.Name,
		"description": config.Description,
		"vmstate":     config.VmState,
	}
}

func (config *ConfigSnapshot) CreateSnapshot(guestId uint, c *Client) (err error) {
	params := config.mapValues()
	_, err = c.CreateSnapshot(NewVmRef(int(guestId)), params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Snapshot: %v, (params: %v)", err, string(params))
	}
	return
}

// Used for formatting the output when retrieving snapshots
type Snapshot struct {
	Name        string      `json:"name"`
	SnapTime    uint        `json:"time,omitempty"`
	Description string      `json:"description,omitempty"`
	VmState     bool        `json:"ram,omitempty"`
	Children    []*Snapshot `json:"children,omitempty"`
	Parent      string      `json:"parent,omitempty"`
}

// Formats the taskResponse as a list of snapshots
func FormatSnapshotsList(taskResponse []interface{}) (list []*Snapshot) {
	list = make([]*Snapshot, len(taskResponse))
	for i, e := range taskResponse {
		list[i] = &Snapshot{}
		if _, isSet := e.(map[string]interface{})["description"]; isSet {
			list[i].Description = e.(map[string]interface{})["description"].(string)
		}
		if _, isSet := e.(map[string]interface{})["name"]; isSet {
			list[i].Name = e.(map[string]interface{})["name"].(string)
		}
		if _, isSet := e.(map[string]interface{})["parent"]; isSet {
			list[i].Parent = e.(map[string]interface{})["parent"].(string)
		}
		if _, isSet := e.(map[string]interface{})["snaptime"]; isSet {
			list[i].SnapTime = uint(e.(map[string]interface{})["snaptime"].(float64))
		}
		if _, isSet := e.(map[string]interface{})["vmstate"]; isSet {
			list[i].VmState = Itob(int(e.(map[string]interface{})["vmstate"].(float64)))
		}
	}
	return
}

// Formats a list of snapshots as a tree of snapshots
func FormatSnapshotsTree(taskResponse []interface{}) (tree []*Snapshot) {
	list := FormatSnapshotsList(taskResponse)
	for _, e := range list {
		for _, ee := range list {
			if e.Parent == ee.Name {
				ee.Children = append(ee.Children, e)
				break
			}
		}
	}
	for _, e := range list {
		if e.Parent == "" {
			tree = append(tree, e)
		}
		e.Parent = ""
	}
	return
}
