package proxmox

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"unicode"
)

type ConfigSnapshot struct {
	Name        SnapshotName `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	VmState     bool         `json:"ram,omitempty"`
}

// TODO write tests for this
func (config ConfigSnapshot) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"snapname":    config.Name,
		"description": config.Description,
		"vmstate":     config.VmState,
	}
}

func (config ConfigSnapshot) CreateSnapshot(c *Client, vmr *VmRef) (err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	err = config.Validate()
	if err != nil {
		return
	}
	params := config.mapToApiValues()
	_, err = c.PostWithTask(params, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/")
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Snapshot: %v, (params: %v)", err, string(params))
	}
	return
}

func (config ConfigSnapshot) Validate() error {
	return config.Name.Validate()
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
func UpdateSnapshotDescription(c *Client, vmr *VmRef, snapshot SnapshotName, description string) (err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	err = snapshot.Validate()
	if err != nil {
		return
	}
	return c.Put(map[string]interface{}{"description": description}, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/"+string(snapshot)+"/config")
}

func DeleteSnapshot(c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	err = snapshot.Validate()
	if err != nil {
		return
	}
	return c.DeleteWithTask("/nodes/" + vmr.node + "/" + vmr.vmType + "/" + strconv.Itoa(vmr.vmId) + "/snapshot/" + string(snapshot))
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
	Name        SnapshotName `json:"name"`
	SnapTime    uint         `json:"time,omitempty"`
	Description string       `json:"description,omitempty"`
	VmState     bool         `json:"ram,omitempty"`
	Children    []*Snapshot  `json:"children,omitempty"`
	Parent      SnapshotName `json:"parent,omitempty"`
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
			list[i].Name = SnapshotName(e.(map[string]interface{})["name"].(string))
		}
		if _, isSet := e.(map[string]interface{})["parent"]; isSet {
			list[i].Parent = SnapshotName(e.(map[string]interface{})["parent"].(string))
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

// Minimum length of 3 characters
// Maximum length of 40 characters
// First character must be a letter
// Must only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_
type SnapshotName string

const (
	SnapshotName_Error_IllegalCharacters string = "SnapshotName must only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	SnapshotName_Error_MaxLength         string = "SnapshotName must be at most 40 characters long"
	SnapshotName_Error_MinLength         string = "SnapshotName must be at least 3 characters long"
	SnapshotName_Error_StartNoLetter     string = "SnapshotName must start with a letter"
)

func (name SnapshotName) Validate() error {
	regex, _ := regexp.Compile(`^([a-zA-Z])([a-z]|[A-Z]|[0-9]|_|-){2,39}$`)
	if !regex.Match([]byte(name)) {
		if len(name) < 3 {
			return fmt.Errorf(SnapshotName_Error_MinLength)
		}
		if len(name) > 40 {
			return fmt.Errorf(SnapshotName_Error_MaxLength)
		}
		if !unicode.IsLetter(rune(name[0])) {
			return fmt.Errorf(SnapshotName_Error_StartNoLetter)
		}
		return fmt.Errorf(SnapshotName_Error_IllegalCharacters)
	}
	return nil
}
