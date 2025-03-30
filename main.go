package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/Telmate/proxmox-api-go/cli"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/proxmox"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	APIURL      string
	HTTPHeaders string
	User        string
	Password    string
	OTP         string
	NewCLI      bool
}

func loadAppConfig() AppConfig {
	newCLI := os.Getenv("NEW_CLI") == "true"

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	return AppConfig{
		APIURL:      os.Getenv("PM_API_URL"),
		HTTPHeaders: os.Getenv("PM_HTTP_HEADERS"),
		User:        os.Getenv("PM_USER"),
		Password:    os.Getenv("PM_PASS"),
		OTP:         os.Getenv("PM_OTP"),
		NewCLI:      newCLI,
	}
}

func initializeProxmoxClient(ctx context.Context, config AppConfig, insecure bool, proxyURL string, taskTimeout int) (*proxmox.Client, error) {
	tlsconf := &tls.Config{InsecureSkipVerify: insecure}
	if !insecure {
		tlsconf = nil
	}

	client, err := proxmox.NewClient(config.APIURL, nil, config.HTTPHeaders, tlsconf, proxyURL, taskTimeout)
	if err != nil {
		return nil, err
	}

	if userRequiresAPIToken(config.User) {
		client.SetAPIToken(config.User, config.Password)
		_, err := client.GetVersion(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		err = client.Login(ctx, config.User, config.Password, config.OTP)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func main() {

	config := loadAppConfig()

	if config.NewCLI {
		err := cli.Execute()
		if err != nil {
			failError(err)
		}
		os.Exit(0)
	}

	// Command-line flags
	insecure := flag.Bool("insecure", false, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	fConfigFile := flag.String("file", "", "file to get the config from")
	taskTimeout := flag.Int("timeout", 300, "API task timeout in seconds")
	proxyURL := flag.String("proxy", "", "proxy URL to connect to")
	fvmid := flag.Int("vmid", -1, "custom VMID (instead of auto)")
	flag.Parse()

	ctx := context.Background()

	// Initialize Proxmox client
	c, err := initializeProxmoxClient(ctx, config, *insecure, *proxyURL, *taskTimeout)
	if err != nil {
		log.Fatalf("Failed to initialize Proxmox client: %v", err)
	}

	tmpID := *fvmid
	var vmid proxmox.GuestID
	if tmpID < 0 {
		if len(flag.Args()) > 1 {
			tmpID, err = strconv.Atoi(flag.Args()[1])
			if err != nil {
				vmid = 0
			}
			vmid = proxmox.GuestID(tmpID)
		} else if len(flag.Args()) == 0 || (flag.Args()[0] == "idstatus") {
			vmid = 0
		}
	} else {
		vmid = proxmox.GuestID(tmpID)
	}

	var jbody interface{}
	var vmr *proxmox.VmRef

	if len(flag.Args()) == 0 {
		fmt.Printf("Missing action, try start|stop vmid\n")
		os.Exit(0)
	}

	switch flag.Args()[0] {

	// TODO make testUserPermissions in new cli
	case "testUserPermissions":
		// testuserpermission [user] [path]
		// ex: testuserpermission root@pam(default) /(default)
		var testpath string
		var testUser proxmox.UserID
		if len(flag.Args()) < 2 {
			testUser, err = proxmox.NewUserID(os.Getenv("PM_USER"))
		} else {
			testUser, err = proxmox.NewUserID(flag.Args()[1])
		}
		failError(err)
		if len(flag.Args()) < 3 {
			testpath = ""
		} else {
			testpath = flag.Args()[2]
		}
		permissions, err := c.GetUserPermissions(ctx, testUser, testpath)
		failError(err)
		sort.Strings(permissions)
		log.Println(permissions)

	case "start":
		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StartVm(ctx, vmr)
		failError(err)

	case "stop":

		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StopVm(ctx, vmr)
		failError(err)

	case "destroy":
		vmr = proxmox.NewVmRef(vmid)
		jbody, err = c.StopVm(ctx, vmr)
		failError(err)
		jbody, err = c.DeleteVm(ctx, vmr)
		failError(err)

	case "getConfig":
		vmr = proxmox.NewVmRef(vmid)
		err := c.CheckVmRef(ctx, vmr)
		failError(err)
		vmType := vmr.GetVmType()
		var config interface{}
		switch vmType {
		case "qemu":
			config, err = proxmox.NewConfigQemuFromApi(ctx, vmr, c)
		case "lxc":
			config, err = proxmox.NewConfigLxcFromApi(ctx, vmr, c)
		}
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))
		// TODO make getNetworkInterfaces in new cli
	case "getNetworkInterfaces":
		vmr = proxmox.NewVmRef(vmid)
		err := c.CheckVmRef(ctx, vmr)
		failError(err)
		networkInterfaces, err := c.GetVmAgentNetworkInterfaces(ctx, vmr)
		failError(err)

		networkInterfaceJSON, err := json.Marshal(networkInterfaces)
		failError(err)
		fmt.Println(string(networkInterfaceJSON))

	case "createQemu":
		config, err := proxmox.NewConfigQemuFromJson(GetConfig(*fConfigFile))
		failError(err)
		config.ID = &vmid
		config.Node = util.Pointer(proxmox.NodeName(flag.Args()[2]))
		_, err = config.Create(ctx, c)
		failError(err)
		log.Println("Complete")

	case "createLxc":
		config, err := proxmox.NewConfigLxcFromJson(GetConfig(*fConfigFile))
		failError(err)
		vmr = proxmox.NewVmRef(vmid)
		vmr.SetNode(flag.Args()[2])
		failError(config.CreateLxc(ctx, vmr, c))
		log.Println("Complete")
		// TODO make installQemu in new cli
	case "installQemu":
		config, err := proxmox.NewConfigQemuFromJson(GetConfig(*fConfigFile))
		var mode string
		if config.QemuIso != "" {
			mode = "(ISO boot mode)"
		} else if config.QemuPxe {
			mode = "(PXE boot mode)"
		}
		failError(err)
		if vmid > 0 {
			vmr = proxmox.NewVmRef(vmid)
		} else {
			nextid, err := c.GetNextID(ctx, nil)
			failError(err)
			vmr = proxmox.NewVmRef(nextid)
		}
		vmr.SetNode(flag.Args()[1])
		config.ID = &vmid
		config.Node = util.Pointer(vmr.Node())
		log.Printf("Creating node %s: \n", mode)
		log.Println(vmr)

		vmr, err = config.Create(ctx, c)
		failError(err)
		_, err = c.StartVm(ctx, vmr)
		failError(err)

		// ISO mode waits for the VM to reboot to exit
		// while PXE mode just launches the VM and is done
		if config.QemuIso != "" {
			_, err := proxmox.SshForwardUsernet(ctx, vmr, c)
			failError(err)
			log.Println("Waiting for CDRom install shutdown (at least 5 minutes)")
			failError(proxmox.WaitForShutdown(ctx, vmr, c))
			log.Println("Restarting")
			_, err = c.StartVm(ctx, vmr)
			failError(err)
			_, err = proxmox.SshForwardUsernet(ctx, vmr, c)
			failError(err)
			//log.Println("SSH Portforward on:" + sshPort)
		}

		log.Println("Complete")

	case "idstatus":
		maxid, err := proxmox.MaxVmId(ctx, c)
		failError(err)
		nextid, err := c.GetNextID(ctx, &vmid)
		failError(err)
		log.Println("---")
		log.Printf("MaxID: %d\n", maxid)
		log.Printf("NextID: %d\n", nextid)
		log.Println("---")
		// TODO make cloneQemu in new cli
	case "cloneQemu":
		var config *proxmox.CloneQemuTarget
		failError(json.Unmarshal(GetConfig(*fConfigFile), &config))
		fmt.Println("Parsed conf: ", config)
		log.Println("Looking for template: " + flag.Args()[1])
		sourceVmrs, err := c.GetVmRefsByName(ctx, flag.Args()[1])
		failError(err)
		if sourceVmrs == nil {
			log.Fatal("Can't find template")
			return
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

		if vmid != 0 {
			if config.Full != nil {
				config.Full.ID = &vmid
			} else if config.Linked != nil {
				config.Linked.ID = &vmid
			}
		}

		vmr, task, err := sourceVmr.CloneQemu(ctx, *config, c)
		failError(err)
		failError(task.WaitForCompletion())
		log.Println("Created guest with ID: " + vmr.VmId().String())

	case "createQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(ctx, flag.Args()[1])
		failError(err)
		jbody, err = c.CreateQemuSnapshot(sourceVmr, flag.Args()[2])
		failError(err)

	case "deleteQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(ctx, flag.Args()[1])
		failError(err)
		jbody, err = c.DeleteQemuSnapshot(sourceVmr, flag.Args()[2])
		failError(err)

	case "listQemuSnapshot":
		sourceVmr, err := c.GetVmRefByName(ctx, flag.Args()[1])
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
		sourceVmrs, err := c.GetVmRefsByName(ctx, flag.Args()[1])
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
		sourceVmr, err := c.GetVmRefByName(ctx, flag.Args()[1])
		failError(err)
		jbody, err = c.RollbackQemuVm(sourceVmr, flag.Args()[2])
		failError(err)
		// TODO make sshforward in new cli
	case "sshforward":
		vmr = proxmox.NewVmRef(vmid)
		sshPort, err := proxmox.SshForwardUsernet(ctx, vmr, c)
		failError(err)
		log.Println("SSH Portforward on:" + sshPort)
		// TODO make sshbackward in new cli
	case "sshbackward":
		vmr = proxmox.NewVmRef(vmid)
		err = proxmox.RemoveSshForwardUsernet(ctx, vmr, c)
		failError(err)
		log.Println("SSH Portforward off")
		// TODO make sendstring in new cli
	case "sendstring":
		vmr = proxmox.NewVmRef(vmid)
		err = proxmox.SendKeysString(ctx, vmr, c, flag.Args()[2])
		failError(err)
		log.Println("Keys sent")

	case "nextid":
		id, err := c.GetNextID(ctx, nil)
		failError(err)
		log.Printf("Getting Next Free ID: %d\n", id)

	case "checkid":
		if len(flag.Args()) < 2 {
			fmt.Printf("Missing vmid\n")
			os.Exit(1)
		}
		i, err := strconv.Atoi(flag.Args()[1])
		failError(err)
		exists, err := c.VMIdExists(ctx, proxmox.GuestID(i))
		failError(err)
		if exists {
			log.Printf("Selected ID is in use: %d\n", i)
		} else {
			log.Printf("Selected ID is free: %d\n", i)
		}
		// TODO make migrate in new cli
	case "migrate":
		vmr := proxmox.NewVmRef(vmid)
		c.GetVmInfo(ctx, vmr)
		args := flag.Args()
		if len(args) <= 1 {
			fmt.Printf("Missing target node\n")
			os.Exit(1)
		}
		_, err := c.MigrateNode(ctx, vmr, proxmox.NodeName(args[2]), true)

		if err != nil {
			log.Printf("Error to move %+v\n", err)
			os.Exit(1)
		}
		log.Printf("VM %d is moved on %s\n", vmid, args[1])

	case "getNodeList":
		nodes, err := c.GetNodeList(ctx)
		if err != nil {
			log.Printf("Error listing Nodes %+v\n", err)
			os.Exit(1)
		}
		nodeList, err := json.Marshal(nodes)
		failError(err)
		fmt.Println(string(nodeList))

	// only returns enabled resources
	// TODO make getResourceList in new cli
	case "getResourceList":
		resource, err := c.GetResourceList(ctx, "")
		if err != nil {
			log.Printf("Error listing resources %+v\n", err)
			os.Exit(1)
		}
		rsList, err := json.Marshal(resource)
		failError(err)
		fmt.Println(string(rsList))

	case "getVmList":
		vms, err := c.GetVmList(ctx)
		if err != nil {
			log.Printf("Error listing VMs %+v\n", err)
			os.Exit(1)
		}
		vmList, err := json.Marshal(vms)
		failError(err)
		fmt.Println(string(vmList))

	case "getVmInfo":
		if len(flag.Args()) < 2 {
			fmt.Printf("Missing vmid\n")
			os.Exit(1)
		}
		i, err := strconv.Atoi(flag.Args()[1])
		failError(err)
		vmr := proxmox.NewVmRef(proxmox.GuestID(i))
		config, err := proxmox.NewConfigQemuFromApi(ctx, vmr, c)
		failError(err)
		fmt.Println(config)
		// TODO make getVmInfo in new cli
	case "getVersion":
		versionInfo, err := c.GetVersion(ctx)
		failError(err)
		version, err := json.Marshal(versionInfo)
		failError(err)
		fmt.Println(string(version))

	//Pool
	case "getPoolList":
		pools, err := proxmox.ListPoolsWithComments(ctx, c)
		if err != nil {
			log.Printf("Error listing pools %+v\n", err)
			os.Exit(1)
		}
		poolList, err := json.Marshal(pools)
		failError(err)
		fmt.Println(string(poolList))

	case "getPoolInfo":
		if len(flag.Args()) < 2 {
			log.Printf("Error poolid required")
			os.Exit(1)
		}
		poolid := flag.Args()[1]
		poolinfo, err := c.GetPoolInfo(ctx, poolid)
		if err != nil {
			log.Printf("Error getting pool info %+v\n", err)
			os.Exit(1)
		}
		poolList, err := json.Marshal(poolinfo)
		failError(err)
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

		failError(proxmox.ConfigPool{
			Name:    proxmox.PoolName(poolid),
			Comment: &comment,
		}.Create(ctx, c))
		fmt.Printf("Pool %s created\n", poolid)

	case "deletePool":
		if len(flag.Args()) < 2 {
			log.Printf("Error: poolid required")
			os.Exit(1)
		}
		poolid := flag.Args()[1]

		failError(proxmox.PoolName(poolid).Delete(ctx, c))
		fmt.Printf("Pool %s removed\n", poolid)

	case "updatePoolComment":
		if len(flag.Args()) < 3 {
			log.Printf("Error: poolid and comment required")
			os.Exit(1)
		}

		poolid := flag.Args()[1]
		comment := flag.Args()[2]

		failError(proxmox.ConfigPool{
			Name:    proxmox.PoolName(poolid),
			Comment: &comment,
		}.Update(ctx, c))
		fmt.Printf("Pool %s updated\n", poolid)

	//Users
	case "getUser":
		var config interface{}
		userId, err := proxmox.NewUserID(flag.Args()[1])
		failError(err)
		config, err = proxmox.NewConfigUserFromApi(ctx, userId, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))

	case "getUserList":
		users, err := proxmox.ListUsers(ctx, c, true)
		if err != nil {
			log.Printf("Error listing users %+v\n", err)
			os.Exit(1)
		}
		userList, err := json.Marshal(users)
		failError(err)
		fmt.Println(string(userList))

	case "updateUserPassword":
		if len(flag.Args()) < 3 {
			log.Printf("Error: Userid and Password required")
			os.Exit(1)
		}
		userId, err := proxmox.NewUserID(flag.Args()[1])
		failError(err)
		err = proxmox.ConfigUser{
			Password: proxmox.UserPassword(flag.Args()[2]),
			User:     userId,
		}.UpdateUserPassword(ctx, c)
		failError(err)
		fmt.Printf("Password of User %s updated\n", userId.String())

	case "setUser":
		var password proxmox.UserPassword
		config, err := proxmox.NewConfigUserFromJson(GetConfig(*fConfigFile))
		failError(err)
		userId, err := proxmox.NewUserID(flag.Args()[1])
		failError(err)
		if len(flag.Args()) > 2 {
			password = proxmox.UserPassword(flag.Args()[2])
		}
		failError(config.SetUser(ctx, userId, password, c))
		log.Printf("User %s has been configured\n", userId.String())

	case "deleteUser":
		if len(flag.Args()) < 2 {
			log.Printf("Error: userId required")
			os.Exit(1)
		}
		userId, err := proxmox.NewUserID(flag.Args()[1])
		failError(err)
		err = proxmox.ConfigUser{User: userId}.DeleteUser(ctx, c)
		failError(err)
		fmt.Printf("User %s removed\n", userId.String())

	//ACME Account
	case "getAcmeAccountList":
		accounts, err := c.GetAcmeAccountList(ctx)
		if err != nil {
			log.Printf("Error listing Acme accounts %+v\n", err)
			os.Exit(1)
		}
		accountList, err := json.Marshal(accounts)
		failError(err)
		fmt.Println(string(accountList))

	case "getAcmeAccount":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Acme account name required")
			os.Exit(1)
		}
		var config interface{}
		acmeid := flag.Args()[1]
		config, err := proxmox.NewConfigAcmeAccountFromApi(ctx, acmeid, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))

	case "createAcmeAccount":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Acme account name required")
			os.Exit(1)
		}
		config, err := proxmox.NewConfigAcmeAccountFromJson(GetConfig(*fConfigFile))
		failError(err)
		acmeid := flag.Args()[1]
		failError(config.CreateAcmeAccount(ctx, acmeid, c))
		log.Printf("Acme account %s has been created\n", acmeid)
		// TODO make updateAcmeAccountEmail in new cli
	case "updateAcmeAccountEmail":
		if len(flag.Args()) < 3 {
			log.Printf("Error: acme name and email(s) required")
			os.Exit(1)
		}
		acmeid := flag.Args()[1]
		_, err := c.UpdateAcmeAccountEmails(ctx, acmeid, flag.Args()[2])
		failError(err)
		fmt.Printf("Acme account %s has been updated\n", acmeid)

	case "deleteAcmeAccount":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Acme account name required")
			os.Exit(1)
		}
		acmeid := flag.Args()[1]
		_, err := c.DeleteAcmeAccount(ctx, acmeid)
		failError(err)
		fmt.Printf("Acme account %s removed\n", acmeid)

	//ACME Plugin
	case "getAcmePluginList":
		plugins, err := c.GetAcmePluginList(ctx)
		if err != nil {
			log.Printf("Error listing Acme plugins %+v\n", err)
			os.Exit(1)
		}
		pluginList, err := json.Marshal(plugins)
		failError(err)
		fmt.Println(string(pluginList))
		// TODO make getAcmePlugin in new cli
	case "getAcmePlugin":
		var config interface{}
		pluginid := flag.Args()[1]
		config, err := proxmox.NewConfigAcmePluginFromApi(ctx, pluginid, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))
		// TODO make setAcmePlugin in new cli
	case "setAcmePlugin":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Acme plugin name required")
			os.Exit(1)
		}
		config, err := proxmox.NewConfigAcmePluginFromJson(GetConfig(*fConfigFile))
		failError(err)
		pluginid := flag.Args()[1]
		failError(config.SetAcmePlugin(ctx, pluginid, c))
		log.Printf("Acme plugin %s has been configured\n", pluginid)
		// TODO make deleteAcmePlugin in new cli
	case "deleteAcmePlugin":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Acme plugin name required")
			os.Exit(1)
		}
		pluginid := flag.Args()[1]
		err := c.DeleteAcmePlugin(ctx, pluginid)
		failError(err)
		fmt.Printf("Acme plugin %s removed\n", pluginid)

	//Metrics
	case "getMetricsServer":
		var config interface{}
		metricsid := flag.Args()[1]
		config, err := proxmox.NewConfigMetricsFromApi(ctx, metricsid, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))

	case "getMetricsServerList":
		metrics, err := c.GetMetricsServerList(ctx)
		if err != nil {
			log.Printf("Error listing Metrics Servers %+v\n", err)
			os.Exit(1)
		}
		metricList, err := json.Marshal(metrics)
		failError(err)
		fmt.Println(string(metricList))

	case "setMetricsServer":
		config, err := proxmox.NewConfigMetricsFromJson(GetConfig(*fConfigFile))
		failError(err)
		meticsid := flag.Args()[1]
		failError(config.SetMetrics(ctx, meticsid, c))
		log.Printf("Merics Server %s has been configured\n", meticsid)

	case "deleteMetricsServer":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Metrics Server name required")
			os.Exit(1)
		}
		metricsid := flag.Args()[1]
		err := c.DeleteMetricServer(ctx, metricsid)
		failError(err)
		fmt.Printf("Metrics Server %s removed\n", metricsid)

	//Storage
	case "getStorageList":
		storage, err := c.GetStorageList(ctx)
		if err != nil {
			log.Printf("Error listing Storages %+v\n", err)
			os.Exit(1)
		}
		storageList, err := json.Marshal(storage)
		failError(err)
		fmt.Println(string(storageList))

	case "getStorage":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Storage id required")
			os.Exit(1)
		}
		var config interface{}
		storageid := flag.Args()[1]
		config, err := proxmox.NewConfigStorageFromApi(ctx, storageid, c)
		failError(err)
		cj, err := json.MarshalIndent(config, "", "  ")
		failError(err)
		log.Println(string(cj))

	case "createStorage":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Storage id required")
			os.Exit(1)
		}
		config, err := proxmox.NewConfigStorageFromJson(GetConfig(*fConfigFile))
		failError(err)
		storageid := flag.Args()[1]
		failError(config.CreateWithValidate(ctx, storageid, c))
		log.Printf("Storage %s has been created\n", storageid)

	case "updateStorage":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Storage id required")
			os.Exit(1)
		}
		config, err := proxmox.NewConfigStorageFromJson(GetConfig(*fConfigFile))
		failError(err)
		storageid := flag.Args()[1]
		failError(config.UpdateWithValidate(ctx, storageid, c))
		log.Printf("Storage %s has been updated\n", storageid)

	case "deleteStorage":
		if len(flag.Args()) < 2 {
			log.Printf("Error: Storage id required")
			os.Exit(1)
		}
		storageid := flag.Args()[1]
		err := c.DeleteStorage(ctx, storageid)
		failError(err)
		fmt.Printf("Storage %s removed\n", storageid)

	// Network
	case "getNetworkList":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: Proxmox node name required"))
		}
		node := flag.Args()[1]
		typeFilter := ""
		if len(flag.Args()) == 3 {
			typeFilter = flag.Args()[2]
		}
		exitStatus, err := c.GetNetworkList(ctx, node, typeFilter)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("List of current network configuration: %s", exitStatus)

	case "getNetworkInterface":
		if len(flag.Args()) < 3 {
			failError(fmt.Errorf("error: Proxmox node name and network interface name required"))
		}
		node := flag.Args()[1]
		iface := flag.Args()[2]
		exitStatus, err := c.GetNetworkInterface(ctx, node, iface)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("Network interface %s configuration: %s", iface, exitStatus)

	case "createNetwork":
		config, err := proxmox.NewConfigNetworkFromJSON(GetConfig(*fConfigFile))
		failError(err)
		failError(config.CreateNetwork(ctx, c))
		log.Printf("Network %s has been created\n", config.Iface)

	case "updateNetwork":
		config, err := proxmox.NewConfigNetworkFromJSON(GetConfig(*fConfigFile))
		failError(err)
		failError(config.UpdateNetwork(ctx, c))
		log.Printf("Network %s has been updated\n", config.Iface)

	case "deleteNetwork":
		if len(flag.Args()) < 3 {
			failError(fmt.Errorf("error: Proxmox node name and network interface name required"))
		}
		node := flag.Args()[1]
		iface := flag.Args()[2]
		exitStatus, err := c.DeleteNetwork(ctx, node, iface)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("Network interface %s deleted", iface)

	case "applyNetwork":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: Proxmox node name required"))
		}
		node := flag.Args()[1]
		exitStatus, err := c.ApplyNetwork(ctx, node)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("Network configuration on node %s has been applied\n", node)

	case "revertNetwork":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: Proxmox node name required"))
		}
		node := flag.Args()[1]
		exitStatus, err := c.RevertNetwork(ctx, node)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("Network configuration on node %s has been reverted\n", node)

	//SDN
	case "applySDN":
		exitStatus, err := c.ApplySDN(ctx)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("SDN configuration has been applied\n")

	case "getZonesList":
		zones, err := c.GetSDNZones(ctx, true, "")
		if err != nil {
			log.Printf("Error listing SDN zones %+v\n", err)
			os.Exit(1)
		}
		zonesList, err := json.Marshal(zones)
		failError(err)
		fmt.Println(string(zonesList))

	case "getZone":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: Zone name is needed"))
		}
		zoneName := flag.Args()[1]
		zone, err := c.GetSDNZone(ctx, zoneName)
		if err != nil {
			log.Printf("Error listing SDN zones %+v\n", err)
			os.Exit(1)
		}
		zoneList, err := json.Marshal(zone)
		failError(err)
		fmt.Println(string(zoneList))

	case "createZone":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: Zone name is needed"))
		}
		zoneName := flag.Args()[1]
		config, err := proxmox.NewConfigSDNZoneFromJson(GetConfig(*fConfigFile))
		failError(err)
		failError(config.CreateWithValidate(ctx, zoneName, c))
		log.Printf("Zone %s has been created\n", zoneName)

	case "deleteZone":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: zone name required"))
		}
		zoneName := flag.Args()[1]
		err := c.DeleteSDNZone(ctx, zoneName)
		failError(err)

	case "updateZone":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: zone name required"))
		}
		zoneName := flag.Args()[1]
		config, err := proxmox.NewConfigSDNZoneFromJson(GetConfig(*fConfigFile))
		failError(err)
		failError(config.UpdateWithValidate(ctx, zoneName, c))
		log.Printf("Zone %s has been updated\n", zoneName)

	case "getVNetsList":
		zones, err := c.GetSDNVNets(ctx, true)
		if err != nil {
			log.Printf("Error listing SDN zones %+v\n", err)
			os.Exit(1)
		}
		vnetsList, err := json.Marshal(zones)
		failError(err)
		fmt.Println(string(vnetsList))

	case "getVNet":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: VNet name is needed"))
		}
		vnetName := flag.Args()[1]
		vnet, err := c.GetSDNVNet(ctx, vnetName)
		if err != nil {
			log.Printf("Error listing SDN VNets %+v\n", err)
			os.Exit(1)
		}
		vnetsList, err := json.Marshal(vnet)
		failError(err)
		fmt.Println(string(vnetsList))

	case "createVNet":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: VNet name is needed"))
		}
		vnetName := flag.Args()[1]
		config, err := proxmox.NewConfigSDNVNetFromJson(GetConfig(*fConfigFile))
		failError(err)
		failError(config.CreateWithValidate(ctx, vnetName, c))
		log.Printf("VNet %s has been created\n", vnetName)

	case "deleteVNet":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: VNet name required"))
		}
		vnetName := flag.Args()[1]
		err := c.DeleteSDNVNet(ctx, vnetName)
		failError(err)

	case "updateVNet":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: zone name required"))
		}
		vnetName := flag.Args()[1]
		config, err := proxmox.NewConfigSDNVNetFromJson(GetConfig(*fConfigFile))
		failError(err)
		failError(config.UpdateWithValidate(ctx, vnetName, c))
		log.Printf("VNet %s has been updated\n", vnetName)

	case "getDNSList":
		dns, err := c.GetSDNDNSs(ctx, "")
		if err != nil {
			log.Printf("Error listing SDN DNS entries %+v\n", err)
			os.Exit(1)
		}
		dnsList, err := json.Marshal(dns)
		failError(err)
		fmt.Println(string(dnsList))

	case "getDNS":
		if len(flag.Args()) < 2 {
			failError(fmt.Errorf("error: DNS name is needed"))
		}
		name := flag.Args()[1]
		dns, err := c.GetSDNDNS(ctx, name)
		if err != nil {
			log.Printf("Error listing SDN DNS %+v\n", err)
			os.Exit(1)
		}
		dnsList, err := json.Marshal(dns)
		failError(err)
		fmt.Println(string(dnsList))

	case "unlink":
		if len(flag.Args()) < 4 {
			failError(fmt.Errorf("error: invoke with <vmID> <node> <diskID [<forceRemoval: false|true>]"))
		}

		vmIdUnparsed := flag.Args()[1]
		node := flag.Args()[2]
		vmId, err := strconv.Atoi(vmIdUnparsed)
		if err != nil {
			failError(fmt.Errorf("failed to convert vmId: %s to a string, error: %+v", vmIdUnparsed, err))
		}

		disks := flag.Args()[3]
		forceRemoval := false
		if len(flag.Args()) > 4 {
			forceRemovalUnparsed := flag.Args()[4]
			forceRemoval, err = strconv.ParseBool(forceRemovalUnparsed)
			if err != nil {
				failError(fmt.Errorf("failed to convert <forceRemoval>: %s to a bool, error: %+v", forceRemovalUnparsed, err))
			}
		}

		exitStatus, err := c.Unlink(ctx, node, proxmox.GuestID(vmId), disks, forceRemoval)
		if err != nil {
			failError(fmt.Errorf("error: %+v\n api error: %s", err, exitStatus))
		}
		log.Printf("Unlinked disks: %s from vmId: %d. Disks removed: %t", disks, vmId, forceRemoval)

	default:
		fmt.Printf("unknown action, try start|stop vmid\n")
	}
	if jbody != nil {
		log.Println(jbody)
	}
	//log.Println(vmr)
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}

// GetConfig get config from file
func GetConfig(configFile string) (configSource []byte) {
	var err error
	if configFile != "" {
		configSource, err = os.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		configSource, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
