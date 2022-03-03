package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
)

func main() {
	var insecure *bool
	insecure = flag.Bool("insecure", false, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	taskTimeout := flag.Int("timeout", 300, "api task timeout in seconds")
	proxyUrl := flag.String("proxy", "", "proxy url to connect to")
	fvmid := flag.Int("vmid", -1, "custom vmid (instead of auto)")
	flag.Parse()
	tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !*insecure {
		tlsconf = nil
	}
	c, err := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, tlsconf, *proxyUrl, *taskTimeout)
	if userRequiresAPIToken(os.Getenv("PM_USER")) {
		c.SetAPIToken(os.Getenv("PM_USER"), os.Getenv("PM_PASS"))
		// As test, get the version of the server
		_, err := c.GetVersion()
		if err != nil {
			log.Fatalf("login error: %s", err)
		}
	} else {
		err = c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"), os.Getenv("PM_OTP"))
		if err != nil {
			log.Fatal(err)
		}
	}

	vmid := *fvmid
	if vmid < 0 {
		if len(flag.Args()) > 1 {
			vmid, err = strconv.Atoi(flag.Args()[len(flag.Args())-1])
			if err != nil {
				vmid = 0
			}
		} else if len(flag.Args()) == 0 || (flag.Args()[0] == "idstatus") {
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
		jbody, err = c.StartVm(vmr)
		failError(err)

	case "stop":

		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StopVm(vmr)
		failError(err)

	case "destroy":
		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StopVm(vmr)
		failError(err)
		jbody, err = c.DeleteVm(vmr)
		failError(err)

	case "getConfig":
		vmr = proxmox.NewVmRef(vmid)
		err := c.CheckVmRef(vmr)
		failError(err)
		vmType := vmr.GetVmType()
		var config interface{}
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
		err := c.CheckVmRef(vmr)
		failError(err)
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
		var mode string
		if config.QemuIso != "" {
			mode = "(ISO boot mode)"
		} else if config.QemuPxe == true {
			mode = "(PXE boot mode)"
		}
		failError(err)
		if vmid > 0 {
			vmr = proxmox.NewVmRef(vmid)
		} else {
			nextid, err := c.GetNextID(0)
			failError(err)
			vmr = proxmox.NewVmRef(nextid)
		}
		vmr.SetNode(flag.Args()[1])
		log.Printf("Creating node %s: \n", mode)
		log.Println(vmr)
		failError(config.CreateVm(vmr, c))
		_, err = c.StartVm(vmr)
		failError(err)

		// ISO mode waits for the VM to reboot to exit
		// while PXE mode just launches the VM and is done
		if config.QemuIso != "" {
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
		}

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
		sourceVmrs, err := c.GetVmRefsByName(flag.Args()[1])
		failError(err)
		if sourceVmrs == nil {
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
		// prefer source Vm located on same node
		sourceVmr := sourceVmrs[0]
		for _, candVmr := range sourceVmrs {
			if candVmr.Node() == vmr.Node() {
				sourceVmr = candVmr
			}
		}

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
		if err == nil {
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
		}
		failError(err)

	case "listQemuSnapshot2":
		sourceVmrs, err := c.GetVmRefsByName(flag.Args()[1])
		if err == nil {
			for _, sourceVmr := range sourceVmrs {
				jbody, _, err = c.ListQemuSnapshot(sourceVmr)
				if rec, ok := jbody.(map[string]interface{}); ok {
					temp := rec["data"].([]interface{})
					for _, val := range temp {
						snapshotName := val.(map[string]interface{})
						if snapshotName["name"] != "current" {
							fmt.Printf("%d@%s:%s\n", sourceVmr.VmId(), sourceVmr.Node(), snapshotName["name"])
						}
					}
				} else {
					fmt.Printf("record not a map[string]interface{}: %v\n", jbody)
				}
			}
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
		if len(flag.Args()) < 2 {
			fmt.Printf("Missing vmid\n")
			os.Exit(1)
		}
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

	case "getNodeList":
		nodes, err := c.GetNodeList()
		if err != nil {
			log.Printf("Error listing Nodes %+v\n", err)
			os.Exit(1)
		}
		nodeList, err := json.Marshal(nodes)
		fmt.Println(string(nodeList))

	case "getVmList":
		vms, err := c.GetVmList()
		if err != nil {
			log.Printf("Error listing VMs %+v\n", err)
			os.Exit(1)
		}
		vmList, err := json.Marshal(vms)
		fmt.Println(string(vmList))

	case "getVersion":
		versionInfo, err := c.GetVersion()
		failError(err)
		version, err := json.Marshal(versionInfo)
		failError(err)
		fmt.Println(string(version))

	case "getPoolList":
		pools, err := c.GetPoolList()
		if err != nil {
			log.Printf("Error listing pools %+v\n", err)
			os.Exit(1)
		}
		poolList, err := json.Marshal(pools)
		fmt.Println(string(poolList))

	case "getPoolInfo":
		if len(flag.Args()) < 2 {
			log.Printf("Error poolid required")
			os.Exit(1)
		}
		poolid := flag.Args()[1]
		poolinfo, err := c.GetPoolInfo(poolid)
		if err != nil {
			log.Printf("Error getting pool info %+v\n", err)
			os.Exit(1)
		}
		poolList, err := json.Marshal(poolinfo)
		fmt.Println(string(poolList))

	case "createPool":
		if len(flag.Args()) < 2 {
			log.Printf("Error: poolid required")
			os.Exit(1)
		}
		poolid := flag.Args()[1]

		comment := ""
		if len(flag.Args()) == 3 {
			comment = flag.Args()[2]
		}

		err := c.CreatePool(poolid, comment)
		failError(err)
		fmt.Printf("Pool %s created\n", poolid)

	case "deletePool":
		if len(flag.Args()) < 2 {
			log.Printf("Error: poolid required")
			os.Exit(1)
		}
		poolid := flag.Args()[1]

		err := c.DeletePool(poolid)
		failError(err)
		fmt.Printf("Pool %s removed\n", poolid)

	case "updatePoolComment":
		if len(flag.Args()) < 3 {
			log.Printf("Error: poolid and comment required")
			os.Exit(1)
		}

		poolid := flag.Args()[1]
		comment := flag.Args()[2]

		err := c.UpdatePoolComment(poolid, comment)
		failError(err)
		fmt.Printf("Pool %s updated\n", poolid)

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

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}
