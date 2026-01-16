package proxmox

type ClientNew struct {
	ApiToken ApiTokenInterface
	Group    GroupInterface
	Pool     PoolInterface
	User     UserInterface
}
