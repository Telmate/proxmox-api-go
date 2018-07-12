package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func main() {
	var insecure *bool
	insecure = flag.Bool("insecure", false, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	fvmid := flag.Int("vmid", -1, "custom vmid (instead of auto)")
	flag.Parse()
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !*insecure {
		tlsconf = nil
	}
	c, _ := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, tlsconf)
	err := c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"))
	if err != nil {
		log.Fatal(err)
	}
	vmid := *fvmid
	if vmid < 0 {
		if len(flag.Args()) > 1 {
			vmid, err = strconv.Atoi(flag.Args()[1])
			if err != nil {
				vmid = 0
			}
		} else if flag.Args()[0] == "idstatus" {
			vmid = 0
		}
	}

	var jbody interface{}
	var vmr *proxmox.VmRef
	switch flag.Args()[0] {
	case "start":
		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StartVm(vmr)

	case "stop":

		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StopVm(vmr)

	case "destroy":
		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StopVm(vmr)
		failError(err)
		jbody, _ = c.DeleteVm(vmr)

	case "getConfig":
		vmr = proxmox.NewVmRef(vmid)
		config, err := proxmox.NewConfigQemuFromApi(vmr, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		log.Println(string(cj))

	case "createQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		failError(err)
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		failError(config.CreateVm(vmr, c))
		log.Println("Complete")

	case "installQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		failError(err)
		if vmid > 0 {
			vmr = proxmox.NewVmRef(vmid)
		} else {
			nextid, err := c.GetNextID(0)
			failError(err)
			vmr = proxmox.NewVmRef(nextid)
		}
		vmr.SetNode(flag.Args()[1])
		log.Print("Creating node: ")
		log.Println(vmr)
		failError(config.CreateVm(vmr, c))
		_, err = c.StartVm(vmr)
		failError(err)
		sshPort, err := proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("Waiting for CDRom install shutdown (at least 5 minutes)")
		failError(proxmox.WaitForShutdown(vmr, c))
		log.Println("Restarting")
		_, err = c.StartVm(vmr)
		failError(err)
		sshPort, err = proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("SSH Portforward on:" + sshPort)
		log.Println("Complete")

	case "idstatus":
		maxid, err := proxmox.MaxVmId(c)
		failError(err)
		nextid, err := c.GetNextID(vmid)
		failError(err)
		log.Println("---")
		log.Printf("MaxID: %d\n", maxid)
		log.Printf("NextID: %d\n", nextid)
		log.Println("---")

	case "cloneQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		failError(err)
		log.Println("Looking for template: " + flag.Args()[1])
		sourceVmr, err := c.GetVmRefByName(flag.Args()[1])
		failError(err)
		if sourceVmr == nil {
			log.Fatal("Can't find template")
			return
		}
		nextid, err := c.GetNextID(0)
		vmr = proxmox.NewVmRef(nextid)
		vmr.SetNode(flag.Args()[2])
		log.Print("Creating node: ")
		log.Println(vmr)
		failError(config.CloneVm(sourceVmr, vmr, c))
		log.Println("Complete")

	case "rollbackQemu":
		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.RollbackQemuVm(vmr, flag.Args()[2])
		failError(err)

	case "sshforward":
		vmr = proxmox.NewVmRef(vmid)
		sshPort, err := proxmox.SshForwardUsernet(vmr, c)
		failError(err)
		log.Println("SSH Portforward on:" + sshPort)

	case "sshbackward":
		vmr = proxmox.NewVmRef(vmid)
		err = proxmox.RemoveSshForwardUsernet(vmr, c)
		failError(err)
		log.Println("SSH Portforward off")

	case "sendstring":
		vmr = proxmox.NewVmRef(vmid)
		err = proxmox.SendKeysString(vmr, c, flag.Args()[2])
		failError(err)
		log.Println("Keys sent")

	default:
		fmt.Printf("unknown action, try start|stop vmid")
	}
	if jbody != nil {
		log.Println(jbody)
	}
	//log.Println(vmr)
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
	return
}
