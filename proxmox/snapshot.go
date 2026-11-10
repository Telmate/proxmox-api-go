package proxmox

import (
	"context"
	"iter"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/noCopy"
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

func (c *snapshotClient) CreateLxc(ctx context.Context, vmr VmRef, name SnapshotName, description string) error {
	if err := name.Validate(); err != nil {
		return err
	}
	ref := LxcRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	if _, err := ref.checkLxc(ctx, c.api); err != nil {
		return err
	}
	return c.CreateLxcNoCheck(ctx, ref.legacy(), name, description)
}

func (c *snapshotClient) CreateLxcNoCheck(ctx context.Context, vmr VmRef, name SnapshotName, description string) error {
	ref := LxcRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	return name.create(ctx, c.api, ref, description, false)
}

func (c *snapshotClient) CreateQemu(ctx context.Context, vmr VmRef, name SnapshotName, description string, vmState bool) error {
	if err := name.Validate(); err != nil {
		return err
	}
	ref := QemuRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	if _, err := ref.checkQemu(ctx, c.api); err != nil {
		return err
	}
	return c.CreateQemuNoCheck(ctx, ref.legacy(), name, description, vmState)
}

func (c *snapshotClient) CreateQemuNoCheck(ctx context.Context, vmr VmRef, name SnapshotName, description string, vmState bool) error {
	ref := QemuRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	return name.create(ctx, c.api, ref, description, vmState)
}

func (c *snapshotClient) Delete(ctx context.Context, vmr VmRef, name SnapshotName) (bool, error) {
	if err := name.Validate(); err != nil {
		return false, err
	}
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return false, err
	}
	return c.DeleteNoCheck(ctx, ref.legacy(), name)
}

func (c *snapshotClient) DeleteNoCheck(ctx context.Context, vmr VmRef, name SnapshotName) (bool, error) {
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	return name.delete(ctx, c.api, ref)
}

func (c *snapshotClient) List(ctx context.Context, vmr VmRef) (RawSnapshots, error) {
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return nil, err
	}
	return c.ListNoCheck(ctx, ref.legacy())
}

func (c *snapshotClient) ListNoCheck(ctx context.Context, vmr VmRef) (RawSnapshots, error) {
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	return snapshotList(ctx, c.api, ref)
}

func (c *snapshotClient) ReadLxc(ctx context.Context, vmr VmRef, name SnapshotName) (RawConfigLXC, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	ref := LxcRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	if _, err := ref.checkLxc(ctx, c.api); err != nil {
		return nil, err
	}
	return c.ReadLxcNoCheck(ctx, ref.legacy(), name)
}

func (c *snapshotClient) ReadLxcNoCheck(ctx context.Context, vmr VmRef, name SnapshotName) (RawConfigLXC, error) {
	ref := LxcRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	params, err := name.readConfig(ctx, c.api, ref)
	if err != nil {
		return nil, err
	}
	return &rawConfigLXC{
		a:       params,
		guestID: ref.ID,
		node:    ref.Node}, nil
}

func (c *snapshotClient) ReadQemu(ctx context.Context, vmr VmRef, name SnapshotName) (RawConfigQemu, error) {
	if err := name.Validate(); err != nil {
		return nil, err
	}
	ref := QemuRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	if _, err := ref.checkQemu(ctx, c.api); err != nil {
		return nil, err
	}
	return c.ReadQemuNoCheck(ctx, ref.legacy(), name)
}

func (c *snapshotClient) ReadQemuNoCheck(ctx context.Context, vmr VmRef, name SnapshotName) (RawConfigQemu, error) {
	ref := QemuRef{Node: vmr.node, ID: vmr.vmId}.generalize()
	params, err := name.readConfig(ctx, c.api, ref)
	if err != nil {
		return nil, err
	}
	return &rawConfigQemu{a: params, id: ref.ID, node: ref.Node}, nil
}

func (c *snapshotClient) Rollback(ctx context.Context, vmr VmRef, name SnapshotName, start bool) error {
	if err := name.Validate(); err != nil {
		return err
	}
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return err
	}
	return c.RollbackNoCheck(ctx, ref.legacy(), name, start)
}

func (c *snapshotClient) RollbackNoCheck(ctx context.Context, vmr VmRef, name SnapshotName, start bool) error {
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	return name.rollback(ctx, c.api, ref, start)
}

func (c *snapshotClient) Update(ctx context.Context, vmr VmRef, name SnapshotName, description string) error {
	if err := name.Validate(); err != nil {
		return err
	}
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, ref.legacy(), name, description)
}

func (c *snapshotClient) UpdateNoCheck(ctx context.Context, vmr VmRef, name SnapshotName, description string) error {
	ref := guestRef{Node: vmr.node, ID: vmr.vmId, Type: vmr.vmType}
	return name.update(ctx, c.api, ref, description)
}

// Deprecated
type ConfigSnapshot struct {
	Name        SnapshotName `json:"name,omitempty"`
	Description string       `json:"description,omitempty"`
	VmState     bool         `json:"ram,omitempty"`
}

// Deprecated: use SnapshotInterface.CreateLxc() or SnapshotInterface.CreateQemu() instead
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

// Deprecated: use SnapshotInterface.CreateLxcNoCheck() or SnapshotInterface.CreateQemuNoCheck() instead
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
		_ noCopy.True
		a []responseSnapshotInfo
	}
)

var _ RawSnapshots = (*rawSnapshots)(nil)

func (r *rawSnapshots) AsArray() []RawSnapshotInfo {
	snapshots := make([]RawSnapshotInfo, len(r.a))
	for i := range r.a {
		snapshots[i] = &rawSnapshotInfo{a: &r.a[i]}
	}
	return snapshots
}

func (r *rawSnapshots) AsMap() map[SnapshotName]RawSnapshotInfo {
	snapshots := make(map[SnapshotName]RawSnapshotInfo, len(r.a))
	for i := range r.a {
		snapshotObj := &rawSnapshotInfo{a: &r.a[i]}
		name := SnapshotName(r.a[i].Name)
		snapshotObj.name = &name
		snapshots[name] = snapshotObj
	}
	return snapshots
}

func (r *rawSnapshots) Iter() iter.Seq[RawSnapshotInfo] {
	return func(yield func(RawSnapshotInfo) bool) {
		for i := range r.a {
			if !yield(&rawSnapshotInfo{a: &r.a[i]}) {
				return
			}
		}
	}
}

func (r *rawSnapshots) Len() int { return len(r.a) }

func (r *rawSnapshots) SelectSnapshot(name SnapshotName) (RawSnapshotInfo, bool) {
	for i := range r.a {
		if r.a[i].Name == name.String() {
			return &rawSnapshotInfo{a: &r.a[i]}, true
		}
	}
	return nil, false
}

func (r *rawSnapshots) Tree() RawSnapshotTree {
	const maxTime = 68043243391 // Package time, Const maxWall
	snapshots := make(map[SnapshotName]snapshot, len(r.a))

	var rootSnapshots [](*Snapshot)

	for i := range r.a { // First pass: populate snapshot information
		var snapshotObj Snapshot
		snapshotCapsule := snapshot{snapshot: &snapshotObj}

		snapshotObj.Name = SnapshotName(r.a[i].Name)
		snapshotObj.Description = r.a[i].Description
		if r.a[i].Time != nil {
			snapshotObj.Time = new(time.Unix(*r.a[i].Time, 0))
		} else {
			snapshotObj.Time = new(time.Unix(maxTime, 0)) // Set to max time to indicate current snapshot
		}
		if r.a[i].State != nil {
			snapshotObj.VmState = new(*r.a[i].State == 1)
		}
		if r.a[i].Parent != nil {
			snapshotCapsule.parent = SnapshotName(*r.a[i].Parent)
		} else {
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
		_    noCopy.True
		a    *responseSnapshotInfo // Never Nil
		name *SnapshotName         // Set by AsMap()
	}
	responseSnapshotInfo struct {
		_           noCopy.True
		Name        string  `json:"name"`
		Time        *int64  `json:"snaptime,omitempty"`
		Description string  `json:"description"`
		State       *int8   `json:"vmstate,omitempty"`
		Parent      *string `json:"parent,omitempty"`
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
	return SnapshotName(r.a.Name)
}

func (r *rawSnapshotInfo) GetTime() *time.Time {
	if r.a.Time != nil {
		return new(time.Unix(*r.a.Time, 0))
	}
	return nil
}

func (r *rawSnapshotInfo) GetDescription() string { return strings.TrimSpace(r.a.Description) }

func (r *rawSnapshotInfo) GetVmState() *bool {
	if r.a.State != nil {
		return new(*r.a.State == 1)
	}
	return nil
}

func (r *rawSnapshotInfo) GetParent() *SnapshotName {
	if r.a.Parent != nil {
		return new(SnapshotName(*r.a.Parent))
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
		current *Snapshot
		root    []*Snapshot
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

// Deprecated: use SnapshotInterface.List() instead
func ListSnapshots(ctx context.Context, c *Client, vmr *VmRef) (RawSnapshots, error) {
	return c.New().Snapshot.List(ctx, *vmr)
}

// Deprecated: use SnapshotInterface.Update() instead
// Updates the description of the specified snapshot, same as SnapshotName.UpdateDescription()
func UpdateSnapshotDescription(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName, description string) (err error) {
	return snapshot.UpdateDescription(ctx, c, vmr, description)
}

// Deprecated: use SnapshotInterface.Delete() instead
// Deletes a snapshot, same as SnapshotName.Delete()
func DeleteSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Delete(ctx, c, vmr)
}

// Deprecated: use SnapshotInterface.Rollback() instead
// Rollback to a snapshot, same as SnapshotName.Rollback()
func RollbackSnapshot(ctx context.Context, c *Client, vmr *VmRef, snapshot SnapshotName) (exitStatus string, err error) {
	return snapshot.Rollback(ctx, c, vmr)
}

// Used for formatting the output when retrieving snapshots
type Snapshot struct {
	Parent      *Snapshot  // Nil if root snapshot
	Time        *time.Time // Nil for current snapshot
	VmState     *bool      // Nil for LXC snapshots
	Name        SnapshotName
	Description string
	Children    []*Snapshot
}

type SnapshotInfo struct {
	Parent      *SnapshotName // Nil if root snapshot
	Time        *time.Time    // Nil for current snapshot
	VmState     *bool         // Nil for LXC snapshots and the running snapshot.
	Name        SnapshotName
	Description string
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

func (snap SnapshotName) create(ctx context.Context, c *clientAPI, ref guestRef, description string, vmState bool) error {
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
	return c.postTask(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/snapshot", new([]byte(builder.String())))
}

func (snap SnapshotName) delete(ctx context.Context, c *clientAPI, ref guestRef) (bool, error) {
	if err := c.deleteTask(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/snapshot/"+snap.String()); err != nil {
		if taskErr, ok := err.(*TaskError); ok {
			if strings.HasPrefix(taskErr.Message, `snapshot '`+snap.String()+`' does not exist`) {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// Deprecated: use SnapshotName.Delete() instead
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

// Deprecated: use SnapshotInterface.DeleteNoCheck() instead
// Deletes the specified snapshot without validating the input, use SnapshotName.Delete() to validate the input.
func (snap SnapshotName) DeleteNoCheck(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	_, err = c.New().Snapshot.DeleteNoCheck(ctx, *vmr, snap)
	return
}

func (snap SnapshotName) readConfig(ctx context.Context, c *clientAPI, ref guestRef) (map[string]any, error) {

	return c.getMap(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/snapshot/"+snap.String()+"/config", "Guest", "SNAPSHOT")
}

func (snap SnapshotName) rollback(ctx context.Context, c *clientAPI, ref guestRef, start bool) error {
	var body *[]byte
	if start {
		body = new([]byte("start=1"))
	}
	return c.post(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+strconv.FormatInt(int64(ref.ID), 10)+"/snapshot/"+string(snap)+"/rollback", body)
}

// Deprecated: use SnapshotInterface.Rollback() instead
// Rollback to the specified snapshot, validates the input
func (snap SnapshotName) Rollback(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	if err = c.CheckVmRef(ctx, vmr); err != nil {
		return
	}
	return "", c.New().Snapshot.Rollback(ctx, *vmr, snap, false)
}

// Deprecated: use SnapshotInterface.RollbackNoCheck() instead
// Rollback to the specified snapshot without validating the input, use SnapshotName.Rollback() to validate the input.
func (snap SnapshotName) RollbackNoCheck(ctx context.Context, c *Client, vmr *VmRef) (exitStatus string, err error) {
	return "", c.New().Snapshot.RollbackNoCheck(ctx, *vmr, snap, false)
}

func (snap SnapshotName) String() string { return string(snap) } // for fmt.Stringer

func (snap SnapshotName) update(ctx context.Context, c *clientAPI, ref guestRef, description string) error {
	return c.putRawRetry(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/snapshot/"+string(snap)+"/config", new([]byte(snapshotApiKeyDescription+"="+body.Escape(description))), RequestRetryCount)
}

// Deprecated: use SnapshotInterface.Update() instead
// Updates the description of the specified snapshot, validates the input
func (snap SnapshotName) UpdateDescription(ctx context.Context, c *Client, vmr *VmRef, description string) (err error) {
	return c.New().Snapshot.Update(ctx, *vmr, snap, description)
}

// Deprecated: use SnapshotInterface.UpdateNoCheck() instead
// Updates the description of the specified snapshot without validating the input, use SnapshotName.UpdateDescription() to validate the input.
func (snap SnapshotName) UpdateDescriptionNoCheck(ctx context.Context, c *Client, vmr *VmRef, description string) error {
	return c.New().Snapshot.UpdateNoCheck(ctx, *vmr, snap, description)
}

var snapshotNameRegex = regexp.MustCompile(`^([a-zA-Z])([a-z]|[A-Z]|[0-9]|_|-){2,39}$`)

func (name SnapshotName) Validate() error {
	if !snapshotNameRegex.Match([]byte(name)) {
		if len(name) < 3 {
			return newValidateError(SnapshotName_Error_MinLength)
		}
		if len(name) > 40 {
			return newValidateError(SnapshotName_Error_MaxLength)
		}
		if !unicode.IsLetter(rune(name[0])) {
			return newValidateError(SnapshotName_Error_StartNoLetter)
		}
		return newValidateError(SnapshotName_Error_IllegalCharacters)
	}
	return nil
}

func snapshotList(ctx context.Context, c *clientAPI, ref guestRef) (*rawSnapshots, error) {
	var data responseData[[]responseSnapshotInfo]
	if err := c.getData(ctx, &data, "guest snapshots", "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/snapshot"); err != nil {
		return nil, err
	}
	return &rawSnapshots{a: *data.Data}, nil
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
