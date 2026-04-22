package proxmox

type ClientNew struct {
	ApiToken  ApiTokenInterface
	Group     GroupInterface
	Guest     GuestInterface
	Pool      PoolInterface
	QemuGuest QemuGuestInterface
	Snapshot  SnapshotInterface
	User      UserInterface
}
