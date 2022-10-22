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

type snapshot struct {
	Name        string      `json:"name"`
	SnapTime    uint        `json:"time,omitempty"`
	Description string      `json:"description,omitempty"`
	VmState     bool        `json:"ram,omitempty"`
	Children    []*snapshot `json:"children,omitempty"`
	Parent      string      `json:"-"`
}

func FormatSnapshotsTree(taskResponse []interface{}) (tree *snapshot) {
	snapshotList := make([]*snapshot, len(taskResponse))
	for i, e := range taskResponse {
		snapshotList[i] = &snapshot{}
		if _, isSet := e.(map[string]interface{})["description"]; isSet {
			snapshotList[i].Description = e.(map[string]interface{})["description"].(string)
		}
		if _, isSet := e.(map[string]interface{})["name"]; isSet {
			snapshotList[i].Name = e.(map[string]interface{})["name"].(string)
		}
		if _, isSet := e.(map[string]interface{})["parent"]; isSet {
			snapshotList[i].Parent = e.(map[string]interface{})["parent"].(string)
		}
		if _, isSet := e.(map[string]interface{})["snaptime"]; isSet {
			snapshotList[i].SnapTime = uint(e.(map[string]interface{})["snaptime"].(float64))
		}
		if _, isSet := e.(map[string]interface{})["vmstate"]; isSet {
			snapshotList[i].VmState = Itob(int(e.(map[string]interface{})["vmstate"].(float64)))
		}
	}
	for _, e := range snapshotList {
		for _, ee := range snapshotList {
			if e.Parent == ee.Name {
				ee.Children = append(ee.Children, e)
				break
			}
		}
	}
	for _, e := range snapshotList {
		if e.Parent == "" {
			tree = e
			break
		}
	}
	return
}
