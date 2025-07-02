package proxmox

// Enum
// TODO need custom json marshaller/unmarshaller to handle this type
type PowerState uint8

const (
	PowerStateUnknown PowerState = 0
	PowerStateStopped PowerState = 1
	PowerStateRunning PowerState = 2
)

func (new *PowerState) combine(current PowerState) PowerState {
	if new != nil {
		return *new
	}
	return current
}

func (PowerState) parse(state string) PowerState {
	switch state {
	case "stopped":
		return PowerStateStopped
	case "running":
		return PowerStateRunning
	default:
		return PowerStateUnknown
	}
}

func (state PowerState) String() string { // String is for fmt.Stringer.
	switch state {
	case PowerStateStopped:
		return "stopped"
	case PowerStateRunning:
		return "running"
	default:
		return ""
	}
}
