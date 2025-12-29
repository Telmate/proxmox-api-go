package proxmox

type ClientNew struct {
	ApiToken ApiTokenInterface
	Group    GroupInterface
	User     UserInterface
}
