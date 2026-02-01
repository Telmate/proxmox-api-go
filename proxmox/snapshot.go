package proxmox

import (
	"context"
	"errors"
	"iter"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	SnapshotInterface interface {
		CreateLxc(ctx context.Context, guest VmRef, name SnapshotName, description string) error
		CreateLxcNoCheck(ctx context.Context, guest VmRef, name SnapshotName, description string) error
		CreateQemu(ctx context.Context, guest VmRef, name SnapshotName, description string, vmState bool) error
		CreateQemuNoCheck(ctx context.Context, guest VmRef, name SnapshotName, description string, vmState bool) error

		// Delete deletes the specified snapshot, validates the input
		//
		// Returns true if the snapshot was deleted, false if it did not exist.
		Delete(context.Context, VmRef, SnapshotName) (bool, error)
		DeleteNoCheck(context.Context, VmRef, SnapshotName) (bool, error)

		List(context.Context, VmRef) (RawSnapshots, error)
		ListNoCheck(context.Context, VmRef) (RawSnapshots, error)

		// ReadLxc reads the configuration of the specified LXC snapshot.
		//
		// The description is the description of the snapshot.
		ReadLxc(context.Context, VmRef, SnapshotName) (RawConfigLXC, error)
		ReadLxcNoCheck(context.Context, VmRef, SnapshotName) (RawConfigLXC, error)

		// ReadQemu reads the configuration of the specified QEMU snapshot.
		//
		// The description is the description of the snapshot.
		ReadQemu(context.Context, VmRef, SnapshotName) (RawConfigQemu, error)
		ReadQemuNoCheck(context.Context, VmRef, SnapshotName) (RawConfigQemu, error)

		// Rollback to the specified snapshot, validates the input
		//
		// If start is true, the VM will be started after the rollback.
		// When the snapshot does not contain a VM state, the VM will always be started.
		Rollback(ctx context.Context, guest VmRef, name SnapshotName, start bool) error
		RollbackNoCheck(ctx context.Context, guest VmRef, name SnapshotName, start bool) error

		Update(ctx context.Context, guest VmRef, name SnapshotName, description string) error
		UpdateNoCheck(ctx context.Context, guest VmRef, name SnapshotName, description string) error
	}

	snapshotClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ SnapshotInterface = (*snapshotClient)(nil)

func (c *snapshotClient) CreateLxc(ctx context.Context, guest VmRef, name SnapshotName, description string) error {
	if err := name.Validate(); err != nil {
		return err
	}
	return c.CreateLxcNoCheck(ctx, guest, name, description)
}

func (c *snapshotClient) CreateLxcNoCheck(ctx context.Context, ref VmRef, name SnapshotName, description string) error {
	if ref.node == "" {
		if _, err := c.oldClient.GetVmInfo(ctx, &ref); err != nil {
			return err
		}
	}
	ref.vmType = GuestLxc
	return name.create(ctx, c.api, ref, description, false)
}

func (c *snapshotClient) CreateQemu(ctx context.Context, ref VmRef, name SnapshotName, description string, vmState bool) error {
	if err := name.Validate(); err != nil {
		return err
	}
	return c.CreateQemuNoCheck(ctx, ref, name, description, vmState)
}

func (c *snapshotClient) CreateQemuNoCheck(ctx context.Context, ref VmRef, name SnapshotName, description string, vmState bool) error {
	if ref.node == "" {
		if _, err := c.oldClient.GetVmInfo(ctx, &ref); err != nil {
			return err
		}
	}
	ref.vmType = GuestQemu
	return name.create(ctx, c.api, ref, description, vmState)
}

func (c *snapshotClient) Delete(ctx context.Context, ref VmRef, name SnapshotName) (bool, error) {
	if err := name.Validate(); err != nil {
		return false, err
	}
	return c.DeleteNoCheck(ctx, ref, name)
}

func (c *snapshotClient) DeleteNoCheck(ctx context.Context, ref VmRef, name SnapshotName) (bool, error) {
	if err := c.oldClient.CheckVmRef(ctx, &ref); err != nil {
		return false, err
	}
	return name.delete(ctx, c.api, ref)
}

func (c *snapshotClient) List(ctx context.Context, ref VmRef) (RawSnapshots, error) {
	return c.ListNoCheck(ctx, ref)
}

func (c *snapshotClient) ListNoCheck(ctx context.Context, ref VmRef) (RawSnapshots, error) {
	if err := c.oldClient.CheckVmRef(ctx, &ref); err != nil {
		return nil, err
	}
	return snapshotList(ctx, c.api, ref)
}

func (c *snapshotClient) ReadLxc(ctx context.Context, ref VmRef, name SnapshotName) (RawConfigLXC, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	return c.ReadLxcNoCheck(ctx, ref, name)
}

func (c *snapshotClient) ReadLxcNoCheck(ctx context.Context, guest VmRef, name SnapshotName) (RawConfigLXC, error) {
	if guest.node == "" {
		if _, err := c.oldClient.GetVmInfo(ctx, &guest); err != nil {
			return nil, err
		}
	}
	params, err := name.readConfig(ctx, c.api, guest)
	if err != nil {
		return nil, err
	}
	return &rawConfigLXC{a: params}, nil
}

func (c *snapshotClient) ReadQemu(ctx context.Context, guest VmRef, name SnapshotName) (RawConfigQemu, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	return c.ReadQemuNoCheck(ctx, guest, name)
}

func (c *snapshotClient) ReadQemuNoCheck(ctx context.Context, guest VmRef, name SnapshotName) (RawConfigQemu, error) {
	if guest.node == "" {
		if _, err := c.oldClient.GetVmInfo(ctx, &guest); err != nil {
			return nil, err
		}
	}
	params, err := name.readConfig(ctx, c.api, guest)
	if err != nil {
		return nil, err
	}
	return &rawConfigQemu{a: params}, nil
}

func (c *snapshotClient) Rollback(ctx context.Context, guest VmRef, name SnapshotName, start bool) error {
	if err := name.Validate(); err != nil {
		return err
	}
	return c.RollbackNoCheck(ctx, guest, name, start)
}

func (c *snapshotClient) RollbackNoCheck(ctx context.Context, ref VmRef, name SnapshotName, start bool) error {
	if err := c.oldClient.CheckVmRef(ctx, &ref); err != nil {
		return err
	}
	return name.rollback(ctx, c.api, ref, start)
}

func (c *snapshotClient) Update(ctx context.Context, ref VmRef, name SnapshotName, description string) error {
	if err := name.Validate(); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, ref, name, description)
}

func (c *snapshotClient) UpdateNoCheck(ctx context.Context, ref VmRef, name SnapshotName, description string) error {
	if err := c.oldClient.CheckVmRef(ctx, &ref); err != nil {
		return err
	}
	return name.update(ctx, c.api, ref, description)
}

// Deprecated
type ConfigSnapshot struct {
	Name        SnapshotName `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	VmState     bool         `json:"ram,omitempty"`
}

// Deprecated use SnapshotInterface.CreateLxc() or SnapshotInterface.CreateQemu() instead
// Creates a snapshot and validates the input
func (config ConfigSnapshot) Create(ctx context.Context, c *Client, vmr *VmRef) (err error) {
	if err = config.Validate(); err != nil {
		return
	}
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	switch vmr.vmType {
	case GuestQemu:
		return c.New().Snapshot.CreateQemu(ctx, *vmr, config.Name, config.Description, config.VmState)
	case GuestLxc:
		return c.New().Snapshot.CreateLxc(ctx, *vmr, config.Name, config.Description)
	}
	return nil
}

// Deprecated use SnapshotInterface.CreateLxcNoCheck() or SnapshotInterface.CreateQemuNoCheck() instead
// Create a snapshot without validating the input, use ConfigSnapshot.Create() to validate the input.
func (config ConfigSnapshot) CreateNoCheck(ctx context.Context, c *Client, vmr *VmRef) error {
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return err
	}
	switch vmr.vmType {
	case GuestQemu:
		return c.New().Snapshot.CreateQemuNoCheck(ctx, *vmr, config.Name, config.Description, config.VmState)
	case GuestLxc:
		return c.New().Snapshot.CreateLxcNoCheck(ctx, *vmr, config.Name, config.Description)
	}
	return nil
}

func (config ConfigSnapshot) Validate() error {
	return config.Name.Validate()
}

type (
	RawSnapshots interface {
		AsArray() []RawSnapshotInfo
		AsMap() map[SnapshotName]RawSnapshotInfo
		Iter() iter.Seq[RawSnapshotInfo]
		Len() int
		SelectSnapshot(SnapshotName) (RawSnapshotInfo, bool)

		// Tree builds a tree structure of snapshots.
		//
		// This operation is quite expensive, prefer using AsArray() or AsMap() when possible.
		//
		// Each call to Tree returns a new RawSnapshotTree and rebuilds the snapshot hierarchy.
		//
		// The tree and its nodes must not be modified, doing so results in undefined behavior.
		Tree() RawSnapshotTree
	}
	rawSnapshots struct {
		a []any
	}
)

var _ RawSnapshots = (*rawSnapshots)(nil)

func (r *rawSnapshots) AsArray() []RawSnapshotInfo {
	snapshots := make([]RawSnapshotInfo, len(r.a))
	for i := range r.a {
		snapshots[i] = &rawSnapshotInfo{a: r.a[i].(map[string]any)}
	}
	return snapshots
}

func (r *rawSnapshots) AsMap() map[SnapshotName]RawSnapshotInfo {
	snapshots := make(map[SnapshotName]RawSnapshotInfo, len(r.a))
	for i := range r.a {
		snapshotObj := &rawSnapshotInfo{a: r.a[i].(map[string]any)}
		name := SnapshotName(r.a[i].(map[string]any)[snapshotListApiKeyName].(string))
		snapshotObj.name = &name
		snapshots[name] = snapshotObj
	}
	return snapshots
}

func (r *rawSnapshots) Iter() iter.Seq[RawSnapshotInfo] {
	return func(yield func(RawSnapshotInfo) bool) {
		for i := range r.a {
			if !yield(&rawSnapshotInfo{a: r.a[i].(map[string]any)}) {
				return
			}
		}
	}
}

func (r *rawSnapshots) Len() int { return len(r.a) }

func (r *rawSnapshots) SelectSnapshot(name SnapshotName) (RawSnapshotInfo, bool) {
	for i := range r.a {
		params := r.a[i].(map[string]any)
		if params[snapshotListApiKeyName] == name.String() {
			return &rawSnapshotInfo{a: params}, true
		}
	}
	return nil, false
}

func (r *rawSnapshots) Tree() RawSnapshotTree {
	const maxTime = 68043243391 // Package time, Const maxWall
	snapshots := make(map[SnapshotName]snapshot, len(r.a))

	var rootSnapshots [](*Snapshot)

	for i := range r.a { // First pass: populate snapshot information
		params := r.a[i].(map[string]any)
		var snapshotObj Snapshot
		snapshotCapsule := snapshot{snapshot: &snapshotObj}

		if v, isSet := params[snapshotListApiKeyName]; isSet {
			snapshotObj.Name = SnapshotName(v.(string))
		}
		if v, isSet := params[snapshotListApiKeyDescription]; isSet {
			snapshotObj.Description = v.(string)
		}
		if v, isSet := params[snapshotListApiKeyTime]; isSet {
			snapshotObj.Time = util.Pointer(time.Unix(int64(v.(float64)), 0))
		} else {
			snapshotObj.Time = util.Pointer(time.Unix(maxTime, 0)) // Set to max time to indicate current snapshot
		}
		if v, isSet := params[snapshotListApiKeyVmState]; isSet {
			snapshotObj.VmState = util.Pointer(int(v.(float64)) == 1)
		}
		if v, isSet := params[snapshotListApiKeyParent]; isSet {
			snapshotCapsule.parent = SnapshotName(v.(string))
		} else { // No parent, we are a root snapshot
			rootSnapshots = append(rootSnapshots, &snapshotObj)
		}

		snapshots[snapshotObj.Name] = snapshotCapsule
	}

	for _, child := range snapshots { // Second pass: build parent-child relationships
		if child.parent == "" {
			continue
		}
		parent := snapshots[child.parent]
		parent.snapshot.Children = append(parent.snapshot.Children, child.snapshot)
		child.snapshot.Parent = parent.snapshot
	}

	for _, s := range snapshots { // Third pass: sort children by time in ascending order
		if len(s.snapshot.Children) < 2 {
			continue
		}
		slices.SortFunc(s.snapshot.Children, snapshotSortTime)
	}

	const currentSnapshotName SnapshotName = "current"
	currentSnapshot := snapshots[currentSnapshotName].snapshot

	if len(rootSnapshots) > 1 { // Fourth pass: sort root snapshots by time in ascending order
		slices.SortFunc(rootSnapshots, snapshotSortTime)
	}

	currentSnapshot.Time = nil // Set nil as we set it to skip nil checks earlier

	return &rawSnapshotTree{
		current: currentSnapshot,
		root:    rootSnapshots}
}

func snapshotSortTime(a, b *Snapshot) int {
	aTime := a.Time.Unix() // Unix does get inlined
	bTime := b.Time.Unix()
	if aTime < bTime {
		return -1
	}
	if aTime > bTime {
		return 1
	}
	// handles the edge-case where two snapshots have the same timestamp.
	// This keeps the sorting idempotent, instead of relying on the random map iteration order.
	if a.Name < b.Name {
		return -1
	}
	return 1
}

type (
	RawSnapshotInfo interface {
		Get() SnapshotInfo
		GetName() SnapshotName
		GetTime() *time.Time
		GetDescription() string
		GetVmState() *bool
		GetParent() *SnapshotName
	}
	rawSnapshotInfo struct {
		a    map[string]any
		name *SnapshotName // Set by AsMap()
	}
)

var _ RawSnapshotInfo = (*rawSnapshotInfo)(nil)

func (r *rawSnapshotInfo) Get() SnapshotInfo {
	return SnapshotInfo{
		Name:        r.GetName(),
		Time:        r.GetTime(),
		Description: r.GetDescription(),
		VmState:     r.GetVmState(),
		Parent:      r.GetParent()}
}

func (r *rawSnapshotInfo) GetName() SnapshotName {
	if r.name != nil {
		return *r.name
	}
	if v, isSet := r.a[snapshotListApiKeyName]; isSet {
		return SnapshotName(v.(string))
	}
	return ""
}

func (r *rawSnapshotInfo) GetTime() *time.Time {
	if v, isSet := r.a[snapshotListApiKeyTime]; isSet {
		return util.Pointer(time.Unix(int64(v.(float64)), 0))
	}
	return nil
}

func (r *rawSnapshotInfo) GetDescription() string {
	if v, isSet := r.a[snapshotListApiKeyDescription]; isSet {
		return v.(string)
	}
	return ""
}

func (r *rawSnapshotInfo) GetVmState() *bool {
	if v, isSet := r.a[snapshotListApiKeyVmState]; isSet {
		return util.Pointer(int(v.(float64)) == 1)
	}
	return nil
}

func (r *rawSnapshotInfo) GetParent() *SnapshotName {
	if v, isSet := r.a[snapshotListApiKeyParent]; isSet {
		return util.Pointer(SnapshotName(v.(string)))
	}
	return nil
}

type (
	RawSnapshotTree interface {

		// Current gives the current snapshot (the one the VM is currently at)
		Current() *Snapshot

		// Root gives the root snapshots (the ones without a parent).
		//
		// On some filesystems it is possible to to create multiple root snapshots by deleting a shared parent snapshot.
		Root() []*Snapshot

		// Walk traverses the snapshot tree in depth-first order, visiting children from oldest to newest.
		//
		// The provided function fn is called for each snapshot.
		// If fn returns false, the walk stops early.
		Walk(fn func(s *Snapshot) bool)
	}
	rawSnapshotTree struct {
		root    []*Snapshot
		current *Snapshot
	}
)

var _ RawSnapshotTree = (*rawSnapshotTree)(nil)

func (r *rawSnapshotTree) Current() *Snapshot { return r.current }

func (r *rawSnapshotTree) Root() []*Snapshot { return r.root }

func (r *rawSnapshotTree) Walk(fn func(s *Snapshot) bool) {
	for _, root := range r.root {
		if !rawSnapshotTreeWalk(root, fn) {
			return
		}
	}
}

func rawSnapshotTreeWalk(s *Snapshot, fn func(s *Snapshot) bool) bool {
	if !fn(s) {
		return false
	}
	for _, child := range s.Children {
		if !rawSnapshotTreeWalk(child, fn) {
			return false
		}
	}
	return true
}

type snapshot struct {
	parent   SnapshotName
	snapshot *Snapshot
}

// Deprecated use SnapshotInterface.List() instead
func ListSnapshots(ctx context.Context, c *Client, vmr *VmRef) (RawSnapshots, error) {
	return c.New().Snapshot.List(ctx, *vmr)
}

// Deprecated use SnapshotInterface.Update() instead
// Updates the description of the specified snapshot, same as SnapshotName.UpdateDescription()
func UpdateSnapshotDescription(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName, description string) (err error) {
	return snapshot.UpdateDescription(ctx, c, vmr, description)
}

// Deprecated use SnapshotInterface.Delete() instead
// Deletes a snapshot, same as SnapshotName.Delete()
func DeleteSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Delete(ctx, c, vmr)
}

// Deprecated use SnapshotInterface.Rollback() instead
// Rollback to a snapshot, same as SnapshotName.Rollback()
func RollbackSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Rollback(ctx, c, vmr)
}

// Used for formatting the output when retrieving snapshots
type Snapshot struct {
	Name        SnapshotName
	Time        *time.Time // Nil for current snapshot
	Description string
	VmState     *bool // Nil for LXC snapshots
	Children    []*Snapshot
	Parent      *Snapshot // Nil if root snapshot
}

type SnapshotInfo struct {
	Name        SnapshotName
	Time        *time.Time // Nil for current snapshot
	Description string
	VmState     *bool         // Nil for LXC snapshots and the running snapshot.
	Parent      *SnapshotName // Nil if root snapshot
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

func (snap SnapshotName) create(ctx context.Context, c *clientAPI, vmr VmRef, description string, vmState bool) error {
	builder := strings.Builder{}
	builder.WriteString(snapshotApiKeyName + "=")
	builder.WriteString(snap.String())
	if description != "" {
		builder.WriteString("&" + snapshotApiKeyDescription + "=")
		builder.WriteString(body.Escape(description))
	}
	if vmState {
		builder.WriteString("&" + snapshotApiKeyVmState + "=1")
	}
	return c.postRawTask(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/snapshot", util.Pointer([]byte(builder.String())))
}

func (snap SnapshotName) delete(ctx context.Context, c *clientAPI, vmr VmRef) (bool, error) {
	if err := c.deleteTask(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/snapshot/"+snap.String()); err != nil {
		var taskErr TaskError
		if errors.As(err, &taskErr) {
			if strings.HasPrefix(taskErr.Message, `snapshot '`+snap.String()+`' does not exist`) {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// Deprecated use SnapshotName.Delete() instead
// Deletes the specified snapshot, validates the input
func (snap SnapshotName) Delete(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	if err = snap.Validate(); err != nil {
		return
	}
	// TODO check if snapshot exists
	return snap.DeleteNoCheck(ctx, c, vmr)
}

// Deprecated use SnapshotInterface.DeleteNoCheck() instead
// Deletes the specified snapshot without validating the input, use SnapshotName.Delete() to validate the input.
func (snap SnapshotName) DeleteNoCheck(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	_, err = c.New().Snapshot.DeleteNoCheck(ctx, *vmr, snap)
	return
}

func (snap SnapshotName) readConfig(ctx context.Context, c *clientAPI, vmr VmRef) (map[string]any, error) {
	return c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/snapshot/"+snap.String()+"/config", "Guest", "SNAPSHOT")
}

func (snap SnapshotName) rollback(ctx context.Context, c *clientAPI, vmr VmRef, start bool) error {
	var body *[]byte
	if start {
		body = util.Pointer([]byte("start=1"))
	}
	return c.postRawRetry(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+strconv.FormatInt(int64(vmr.vmId), 10)+"/snapshot/"+string(snap)+"/rollback", body, 3)
}

// Deprecated use SnapshotInterface.Rollback() instead
// Rollback to the specified snapshot, validates the input
func (snap SnapshotName) Rollback(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	return "", c.New().Snapshot.Rollback(ctx, *vmr, snap, false)
}

// Deprecated use SnapshotInterface.RollbackNoCheck() instead
// Rollback to the specified snapshot without validating the input, use SnapshotName.Rollback() to validate the input.
func (snap SnapshotName) RollbackNoCheck(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	return "", c.New().Snapshot.RollbackNoCheck(ctx, *vmr, snap, false)
}

func (snap SnapshotName) String() string { return string(snap) } // for fmt.Stringer

func (snap SnapshotName) update(ctx context.Context, c *clientAPI, vmr VmRef, description string) error {
	return c.putRawRetry(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/snapshot/"+string(snap)+"/config", util.Pointer([]byte(snapshotApiKeyDescription+"="+body.Escape(description))), 3)
}

// Deprecated use SnapshotInterface.Update() instead
// Updates the description of the specified snapshot, validates the input
func (snap SnapshotName) UpdateDescription(ctx context.Context, c *Client, vmr *VmRef, description string) (err error) {
	return c.New().Snapshot.Update(ctx, *vmr, snap, description)
}

// Deprecated use SnapshotInterface.UpdateNoCheck() instead
// Updates the description of the specified snapshot without validating the input, use SnapshotName.UpdateDescription() to validate the input.
func (snap SnapshotName) UpdateDescriptionNoCheck(ctx context.Context, c *Client, vmr *VmRef, description string) error {
	return c.New().Snapshot.UpdateNoCheck(ctx, *vmr, snap, description)
}

var snapShotNameRegex = regexp.MustCompile(`^([a-zA-Z])([a-z]|[A-Z]|[0-9]|_|-){2,39}$`)

func (name SnapshotName) Validate() error {
	if !snapShotNameRegex.Match([]byte(name)) {
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

func snapshotList(ctx context.Context, c *clientAPI, vmr VmRef) (*rawSnapshots, error) {
	raw, err := c.getList(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/snapshot", "Guest", "SNAPSHOTS")
	if err != nil {
		return nil, err
	}
	return &rawSnapshots{a: raw}, nil
}

const (
	snapshotApiKeyName        string = "snapname"
	snapshotApiKeyDescription string = "description"
	snapshotApiKeyVmState     string = "vmstate"

	snapshotListApiKeyName        string = "name"
	snapshotListApiKeyDescription string = "description"
	snapshotListApiKeyParent      string = "parent"
	snapshotListApiKeyTime        string = "snaptime"
	snapshotListApiKeyVmState     string = "vmstate"
)
