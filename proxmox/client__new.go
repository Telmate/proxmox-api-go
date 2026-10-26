package proxmox

type ClientNew struct {
	ApiToken  ApiTokenInterface
	Group     GroupInterface
	Guest     GuestInterface
	LxcGuest  LxcGuestInterface
	Node      NodeInterface
	Pool      PoolInterface
	QemuAgent QemuAgentInterface
	QemuGuest QemuGuestInterface
	Snapshot  SnapshotInterface
	User      UserInterface
}
