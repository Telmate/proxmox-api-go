package proxmox

import (
	"context"
)

type LxcRef struct {
	ID   GuestID
	Node NodeName
}

var _ GuestRef = (*LxcRef)(nil)

func (ref LxcRef) GetID() GuestID     { return ref.ID }
func (ref LxcRef) GetNode() NodeName  { return ref.Node }
func (ref LxcRef) GetType() GuestType { return GuestLxc }

func (ref LxcRef) generalize() guestRef {
	return guestRef{
		ID:   ref.GetID(),
		Node: ref.GetNode(),
		Type: ref.GetType(),
	}
}

func (ref LxcRef) legacy() *VmRef {
	return &VmRef{
		node:   ref.GetNode(),
		vmId:   ref.GetID(),
		vmType: ref.GetType(),
	}
}

func (ref LxcRef) read(ctx context.Context, c *clientAPI, version Version) (*rawConfigLXC, error) {
	rawConfig, err := c.getGuestConfig(ctx, &VmRef{node: ref.Node, vmId: ref.ID, vmType: GuestLxc})
	if err != nil {
		return nil, err
	}
	return &rawConfigLXC{
		a:       rawConfig,
		guestID: ref.ID,
		node:    ref.Node,
		version: version,
	}, nil
}

func (ref LxcRef) readInterfaceStatus(ctx context.Context, c *clientAPI) (*rawLxcInfoNetworkInterfaces, error) {
	raw, err := c.getList(ctx, "/nodes/"+ref.Node.String()+"/lxc/"+ref.ID.String()+"/interfaces", "", "")
	if err != nil {
		return nil, err
	}
	return &rawLxcInfoNetworkInterfaces{a: raw}, nil
}
