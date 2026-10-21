package proxmox

type QemuRef struct {
	Node NodeName
	ID   GuestID
}

var _ GuestRef = (*QemuRef)(nil)

func (ref QemuRef) GetID() GuestID     { return ref.ID }
func (ref QemuRef) GetNode() NodeName  { return ref.Node }
func (ref QemuRef) GetType() GuestType { return GuestQemu }

func (ref QemuRef) generalize() guestRef {
	return guestRef{
		ID:   ref.GetID(),
		Node: ref.GetNode(),
		Type: ref.GetType()}
}
