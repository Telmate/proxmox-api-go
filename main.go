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
	taskTimeout := flag.Int("timeout", 300, "api task timeout in seconds")
	fvmid := flag.Int("vmid", -1, "custom vmid (instead of auto)")
	flag.Parse()
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !*insecure {
		tlsconf = nil
	}
	c, _ := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, tlsconf, *taskTimeout)
	err := c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"), os.Getenv("PM_OTP"))
	if err != nil {
		log.Fatal(err)
	}
	vmid := *fvmid
	if vmid < 0 {
		if len(flag.Args()) > 1 {
			vmid, err = strconv.Atoi(flag.Args()[len(flag.Args())-1])
			if err != nil {
				vmid = 0
			}
		} else if flag.Args()[0] == "idstatus" {
			vmid = 0
		}
	}

	var jbody interface{}
	var vmr *proxmox.VmRef

	if len(flag.Args()) == 0 {
		fmt.Printf("Missing action, try start|stop vmid\n")
		os.Exit(0)
	}

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
		c.CheckVmRef(vmr)
		vmType := vmr.GetVmType()
		var config interface{}
		var err error
		if vmType == "qemu" {
			config, err = proxmox.NewConfigQemuFromApi(vmr, c)
		} else if vmType == "lxc" {
			config, err = proxmox.NewConfigLxcFromApi(vmr, c)
		}
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		log.Println(string(cj))

	case "getNetworkInterfaces":
		vmr = proxmox.NewVmRef(vmid)
		c.CheckVmRef(vmr)
		networkInterfaces, err := c.GetVmAgentNetworkInterfaces(vmr)
		failError(err)

		networkInterfaceJson, err := json.Marshal(networkInterfaces)
		fmt.Println(string(networkInterfaceJson))

	case "createQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		failError(err)
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		failError(config.CreateVm(vmr, c))
		log.Println("Complete")

	case "createLxc":
		config, err := proxmox.NewConfigLxcFromJson(os.Stdin)
		failError(err)
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		failError(config.CreateLxc(vmr, c))
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
		if vmid == 0 {
			vmid, err = c.GetNextID(0)
		}
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		log.Print("Creating node: ")
		log.Println(vmr)
		failError(config.CloneVm(sourceVmr, vmr, c))
		failError(config.UpdateConfig(vmr, c))
		log.Println("Complete")

	case "createQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(flag.Args()[1])
		jbody, err = c.CreateQemuSnapshot(sourceVmr, flag.Args()[2])
		failError(err)

	case "deleteQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(flag.Args()[1])
		jbody, err = c.DeleteQemuSnapshot(sourceVmr, flag.Args()[2])
		failError(err)

	case "listQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(flag.Args()[1])
		jbody, _, err = c.ListQemuSnapshot(sourceVmr)
		if rec, ok := jbody.(map[string]interface{}); ok {
			temp := rec["data"].([]interface{})
			for _, val := range temp {
				snapshotName := val.(map[string]interface{})
				if snapshotName["name"] != "current" {
					fmt.Println(snapshotName["name"])
				}
			}
		} else {
			fmt.Printf("record not a map[string]interface{}: %v\n", jbody)
		}
		failError(err)

	case "rollbackQemu":
		sourceVmr, err := c.GetVmRefByName(flag.Args()[1])
		jbody, err = c.RollbackQemuVm(sourceVmr, flag.Args()[2])
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

	case "nextid":
		id, err := c.GetNextID(0)
		failError(err)
		log.Printf("Getting Next Free ID: %d\n", id)

	case "checkid":
		i, err := strconv.Atoi(flag.Args()[1])
		failError(err)
		id, err := c.VMIdExists(i)
		failError(err)
		log.Printf("Selected ID is free: %d\n", id)

	case "migrate":
		vmr := proxmox.NewVmRef(vmid)
		c.GetVmInfo(vmr)
		args := flag.Args()
		if len(args) <= 1 {
			fmt.Printf("Missing target node\n")
			os.Exit(1)
		}
		_, err := c.MigrateNode(vmr, args[1], true)

		if err != nil {
			log.Printf("Error to move %+v\n", err)
			os.Exit(1)
		}
		log.Printf("VM %d is moved on %s\n", vmid, args[1])

	default:
		fmt.Printf("unknown action, try start|stop vmid\n")
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
