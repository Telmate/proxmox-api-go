package proxmox

type ConfigQemu struct {
	name         string
	description  string
	memory       string
	diskSize     string
	storage      string
	qemuOs       string
	qemuCores    int
	qemuSockets  int
	qemuIso      string
	qemuNicModel string
	qemuBrige    string
	qemuVlanTag  int
}

func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	network := config.qemuNicModel + ",bridge=" + config.qemuBrige
	if config.qemuVlanTag > 0 {
		network = network + ",tag=" + string(config.qemuVlanTag)
	}
	params := map[string]string{
		"vmid":        string(vmr.vmId),
		"name":        config.name,
		"ide2":        config.qemuIso + ",media=cdrom",
		"ostype":      config.qemuOs,
		"virtio0":     config.storage + ":" + config.diskSize,
		"sockets":     string(config.qemuSockets),
		"cores":       string(config.qemuCores),
		"cpu":         "host",
		"memory":      config.memory,
		"net0":        network,
		"pool":        "default",
		"description": config.description}

	_, err = client.CreateQemuVm(vmr.node, params)
	return
}
