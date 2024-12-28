package proxmox

import (
	"context"
	"encoding/json"
	"errors"
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

// Creates a snapshot and validates the input
func (config ConfigSnapshot) Create(ctx context.Context, c *Client, vmr *VmRef) (err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	if err = config.Validate(); err != nil {
		return
	}
	return config.Create_Unsafe(ctx, c, vmr)
}

// Create a snapshot without validating the input, use ConfigSnapshot.Create() to validate the input.
func (config ConfigSnapshot) Create_Unsafe(ctx context.Context, c *Client, vmr *VmRef) error {
	params := config.mapToApiValues()
	_, err := c.PostWithTask(ctx, params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/")
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Snapshot: %v, (params: %v)", err, string(params))
	}
	return nil
}

// deprecated use ConfigSnapshot.Create() instead
func (config ConfigSnapshot) CreateSnapshot(ctx context.Context, c *Client, vmr *VmRef) error {
	return config.Create(ctx, c, vmr)
}

func (config ConfigSnapshot) Validate() error {
	return config.Name.Validate()
}

type rawSnapshots []interface{}

func ListSnapshots(ctx context.Context, c *Client, vmr *VmRef) (rawSnapshots, error) {
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return nil, err
	}
	return c.GetItemConfigInterfaceArray(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/", "Guest", "SNAPSHOTS")
}

// Updates the description of the specified snapshot, same as SnapshotName.UpdateDescription()
func UpdateSnapshotDescription(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName, description string) (err error) {
	return snapshot.UpdateDescription(ctx, c, vmr, description)
}

// Deletes a snapshot, same as SnapshotName.Delete()
func DeleteSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Delete(ctx, c, vmr)
}

// Rollback to a snapshot, same as SnapshotName.Rollback()
func RollbackSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Rollback(ctx, c, vmr)
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

// Deletes the specified snapshot, validates the input
func (snap SnapshotName) Delete(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	if err = snap.Validate(); err != nil {
		return
	}
	// TODO check if snapshot exists
	return snap.Delete_Unsafe(ctx, c, vmr)
}

// Deletes the specified snapshot without validating the input, use SnapshotName.Delete() to validate the input.
func (snap SnapshotName) Delete_Unsafe(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	return c.DeleteWithTask(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/"+string(snap))
}

// Rollback to the specified snapshot, validates the input
func (snap SnapshotName) Rollback(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	if err = snap.Validate(); err != nil {
		return
	}
	// TODO check if snapshot exists
	return snap.Rollback_Unsafe(ctx, c, vmr)
}

// Rollback to the specified snapshot without validating the input, use SnapshotName.Rollback() to validate the input.
func (snap SnapshotName) Rollback_Unsafe(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	return c.PostWithTask(ctx, nil, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.FormatInt(int64(vmr.vmId), 10)+"/snapshot/"+string(snap)+"/rollback")
}

// Updates the description of the specified snapshot, validates the input
func (snap SnapshotName) UpdateDescription(ctx context.Context, c *Client, vmr *VmRef, description string) (err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	if err = snap.Validate(); err != nil {
		return
	}
	// TODO check if snapshot exists
	return snap.UpdateDescription_Unsafe(ctx, c, vmr, description)
}

// Updates the description of the specified snapshot without validating the input, use SnapshotName.UpdateDescription() to validate the input.
func (snap SnapshotName) UpdateDescription_Unsafe(ctx context.Context, c *Client, vmr *VmRef, description string) error {
	return c.Put(ctx, map[string]interface{}{"description": description}, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/snapshot/"+string(snap)+"/config")
}

func (name SnapshotName) Validate() error {
	regex, _ := regexp.Compile(`^([a-zA-Z])([a-z]|[A-Z]|[0-9]|_|-){2,39}$`)
	if !regex.Match([]byte(name)) {
		if len(name) < 3 {
			return errors.New(SnapshotName_Error_MinLength)
		}
		if len(name) > 40 {
			return errors.New(SnapshotName_Error_MaxLength)
		}
		if !unicode.IsLetter(rune(name[0])) {
			return errors.New(SnapshotName_Error_StartNoLetter)
		}
		return errors.New(SnapshotName_Error_IllegalCharacters)
	}
	return nil
}
