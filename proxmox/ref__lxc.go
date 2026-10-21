package proxmox

type LxcRef struct {
	Node NodeName
	ID   GuestID
}

var _ GuestRef = (*LxcRef)(nil)

func (ref LxcRef) GetID() GuestID     { return ref.ID }
func (ref LxcRef) GetNode() NodeName  { return ref.Node }
func (ref LxcRef) GetType() GuestType { return GuestLxc }

func (ref LxcRef) generalize() guestRef {
	return guestRef{
		ID:   ref.GetID(),
		Node: ref.GetNode(),
		Type: ref.GetType()}
}
