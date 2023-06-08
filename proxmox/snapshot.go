package proxmox

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ConfigSnapshot struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	VmState     bool   `json:"ram,omitempty"`
}

func (config ConfigSnapshot) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"snapname":    config.Name,
		"description": config.Description,
		"vmstate":     config.VmState,
	}
}

func (config ConfigSnapshot) CreateSnapshot(c *Client, vmr *VmRef) (err error) {
	params := config.mapToApiValues()
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	_, err = c.PostWithTask(params, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/")
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Snapshot: %v, (params: %v)", err, string(params))
	}
	return
}

type rawSnapshots []interface{}

func ListSnapshots(c *Client, vmr *VmRef) (rawSnapshots, error) {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	return c.GetItemConfigInterfaceArray("/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/", "Guest", "SNAPSHOTS")
}

// Can only be used to update the description of an already existing snapshot
func UpdateSnapshotDescription(c *Client, vmr *VmRef, snapshot, description string) (err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	return c.Put(map[string]interface{}{"description": description}, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/"+snapshot+"/config")
}

func DeleteSnapshot(c *Client, vmr *VmRef, snapshot string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	return c.DeleteWithTask("/nodes/" + vmr.node + "/" + vmr.vmType + "/" + strconv.Itoa(vmr.vmId) + "/snapshot/" + snapshot)
}

func RollbackSnapshot(c *Client, vmr *VmRef, snapshot string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	return c.PostWithTask(nil, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/"+snapshot+"/rollback")
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
func (raw rawSnapshots) FormatSnapshotsList() (list []*Snapshot) {
	list = make([]*Snapshot, len(raw))
	for i, e := range raw {
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
func (raw rawSnapshots) FormatSnapshotsTree() (tree []*Snapshot) {
	list := raw.FormatSnapshotsList()
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
