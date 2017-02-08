package main

import (
	"flag"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"log"
	"os"
	"strconv"
)

func main() {
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	flag.Parse()

	c, _ := proxmox.NewClient(os.Getenv("PM_API_URL"), nil, nil)
	err := c.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"))
	if err != nil {
		log.Fatal(err)
	}

	vmid, _ := strconv.Atoi(flag.Args()[1])

	var jbody interface{}
	var vmr *proxmox.VmRef
	switch flag.Args()[0] {
	case "start":
		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StartVm(vmr)
	case "stop":
		vmr = proxmox.NewVmRef(vmid)
		jbody, _ = c.StopVm(vmr)
	case "createQemu":
		config, err := proxmox.NewConfigQemuFromJson(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		err = config.CreateVm(vmr, c)
		if err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf("unknown action, try start|stop vmid")
	}

	log.Println(jbody)
	log.Println(vmr)
}
