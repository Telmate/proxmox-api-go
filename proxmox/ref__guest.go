package proxmox

import (
	"context"
	"errors"
	"strings"
)

type GuestRef interface {
	GetID() GuestID
	GetNode() NodeName
	GetType() GuestType
}

const (
	GuestRef_Error_IDnotSet = "guest reference id not set"
)

// This will be the future replacement of VmRef
type guestRef struct {
	Node NodeName
	ID   GuestID
	Type GuestType
}

// nil, nil is a possibility
func (ref *guestRef) check(ctx context.Context, kind *GuestType, c *clientAPI) (*rawGuestResource, error) {
	err := ref.checkID()
	if err != nil {
		return nil, err
	}
	var raw *rawGuestResource
	if ref.Node != "" && ref.Type != guestUnknown {
		return nil, ref.Node.Validate()
	}
	var raws *rawGuestResources
	raws, err = c.listGuestResources(ctx)
	if err != nil {
		return nil, err
	}
	var exists bool
	raw, exists = raws.selectID(ref.ID)
	if !exists {
		return nil, errorMsg{}.guestDoesNotExist(ref.ID)
	}
	tmpType := raw.GetType()
	if kind != nil && tmpType != *kind {
		return nil, errors.New("guest is not of type: " + kind.String())
	}
	ref.Type = tmpType // this ensures whe don't compare with itself if kind is a ptr to ref.Node
	ref.Node = raw.GetNode()
	return raw, nil
}

// nil, nil is a possibility
func (ref *guestRef) checkLxc(ctx context.Context, c *clientAPI) (*rawGuestResource, error) {
	return ref.check(ctx, new(GuestLxc), c)
}

// nil, nil is a possibility
func (ref *guestRef) checkQemu(ctx context.Context, c *clientAPI) (*rawGuestResource, error) {
	return ref.check(ctx, new(GuestQemu), c)
}

func (ref guestRef) checkID() error {
	if ref.ID == 0 {
		return errors.New(GuestRef_Error_IDnotSet)
	}
	return nil
}

func (ref guestRef) cloneLxc(ctx context.Context, settings CloneLxcTarget, c *Client, ca *clientAPI) (*VmRef, error) {
	id, node, pool, body := settings.mapToAPI()
	var err error
	url := "/nodes/" + ref.Node.String() + "/lxc/" + ref.ID.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, nil, body, c, ca)
	} else {
		err = ca.postTask(ctx, url, body)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestLxc}, nil
}

func (ref guestRef) cloneQemu(ctx context.Context, settings CloneQemuTarget, c *Client, ca *clientAPI) (*VmRef, error) {
	id, node, pool, body := settings.mapToAPI()
	var err error
	url := "/nodes/" + ref.Node.String() + "/qemu/" + ref.ID.String() + "/clone"
	if id == 0 {
		id, err = guestCreateLoop_Unsafe(ctx, "newid", url, nil, body, c, ca)
	} else {
		err = ca.postTask(ctx, url, body)
	}
	if err != nil {
		return nil, err
	}
	return &VmRef{
		vmId:   id,
		node:   node,
		pool:   pool,
		vmType: GuestQemu}, nil
}

func (ref guestRef) delete(ctx context.Context, purge bool, c *clientAPI) (bool, error) {
	url := "/nodes/" + ref.Node.String() + "/" + ref.Type.String() + "/" + ref.ID.String()
	if purge {
		url += "?purge=1"
	}
	if err := c.deleteTask(ctx, url); err != nil {
		if taskErr, ok := err.(*TaskError); ok {
			if strings.HasPrefix(taskErr.Message, "Configuration file '") { // task id: UPID:pve-9l:00120E24:08C58A66:6A2F4143:qmdestroy:105:root@pam: | message: Configuration file 'nodes/pve-9l/qemu-server/105.conf' does not exist
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (ref guestRef) legacy() VmRef {
	return VmRef{
		node:   ref.Node,
		vmId:   ref.ID,
		vmType: ref.Type}
}

func (ref guestRef) migrate(ctx context.Context, c *clientAPI, newNode NodeName, live bool) error {
	var b strings.Builder
	b.WriteString("target=")
	b.WriteString(newNode.String())
	b.WriteString("&with-local-disks=1")
	if live {
		b.WriteString("&online=1")
	}
	return c.postTask(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/migrate", new([]byte(b.String())))
}

func (ref guestRef) pendingChanges(ctx context.Context, c *clientAPI) (bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, ref)
	if err != nil {
		return false, err
	}
	return parsePendingChanges(changes), nil
}

func (ref guestRef) pendingConfig(ctx context.Context, c *clientAPI) (map[string]any, bool, error) {
	changes, err := c.getGuestPendingChanges(ctx, ref)
	if err != nil {
		return nil, false, err
	}
	settings, ok := parsePendingConfig(changes)
	return settings, ok, nil
}

func (ref guestRef) reboot(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "reboot", nil)
}

func (ref guestRef) shutdown(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "shutdown", nil)

}
func (ref guestRef) shutdownForce(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "shutdown", new([]byte(`force=1`)))
}

func (ref guestRef) start(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "start", nil)
}

func (ref guestRef) status(ctx context.Context, c *clientAPI) (*rawGuestStatus, error) {
	raw, err := c.getMap(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/status/current", "guest", "STATE")
	if err != nil {
		return nil, err
	}
	return &rawGuestStatus{a: raw}, nil
}

// We should rarely need this function, use: guestRef.stopOverruleOpertunistic_Unsafe() instead.
func (ref guestRef) stop(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "stop", nil)
}

// We should rarely need this function, use: guestRef.stopOverruleOpertunistic_Unsafe() instead.
func (ref guestRef) stopOverrule(ctx context.Context, c *clientAPI) error {
	return c.updateGuestStatus(ctx, ref, "stop", new([]byte("overrule-shutdown=1")))
}

// Doesn't see the difference between a stopped and absent guest
func (ref guestRef) stopOverruleOpertunistic(ctx context.Context, c *clientAPI, version Version) error {
	if version.Major >= 8 {
		return ref.stopOverrule(ctx, c)
	}
	return ref.stop(ctx, c)
}

const (
	pendingApiKeyKey   string = "key"
	pendingApiValueKey string = "value"
)

func parsePendingChanges(changes []any) bool {
	for i := range changes {
		m := changes[i].(map[string]any)
		// we always have the key `key`
		if _, ok := m[pendingApiValueKey]; ok {
			if len(m) > 2 {
				return true
			}
		} else if len(m) > 1 {
			return true
		}
	}
	return false
}

func parsePendingConfig(changes []any) (map[string]any, bool) {
	var pending bool
	config := make(map[string]any, len(changes))
	for i := range changes {
		m := changes[i].(map[string]any)
		// we always have the key `key`
		if v, ok := m[pendingApiValueKey]; ok {
			config[m[pendingApiKeyKey].(string)] = v
			if len(m) > 2 {
				pending = true
			}
		} else if len(m) > 1 {
			pending = true
		}
	}
	return config, pending
}
