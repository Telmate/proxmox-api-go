package proxmox

// inspired by https://github.com/Telmate/vagrant-proxmox/blob/master/lib/vagrant-proxmox/proxmox/connection.rb

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TaskStatusCheckInterval - time between async checks in seconds
const TaskStatusCheckInterval = 2

const exitStatusSuccess = "OK"

// Client - URL, user and password to specific Proxmox node
type Client struct {
	session     *Session
	ApiUrl      string
	Username    string
	Password    string
	Otp         string
	TaskTimeout int
}

// VmRef - virtual machine ref parts
// map[type:qemu node:proxmox1-xx id:qemu/132 diskread:5.57424738e+08 disk:0 netin:5.9297450593e+10 mem:3.3235968e+09 uptime:1.4567097e+07 vmid:132 template:0 maxcpu:2 netout:6.053310416e+09 maxdisk:3.4359738368e+10 maxmem:8.592031744e+09 diskwrite:1.49663619584e+12 status:running cpu:0.00386980694947209 name:appt-app1-dev.xxx.xx]
type VmRef struct {
	vmId    int
	node    string
	pool    string
	vmType  string
	haState string
	haGroup string
}

func (vmr *VmRef) SetNode(node string) {
	vmr.node = node
}

func (vmr *VmRef) SetPool(pool string) {
	vmr.pool = pool
}

func (vmr *VmRef) SetVmType(vmType string) {
	vmr.vmType = vmType
}

func (vmr *VmRef) GetVmType() string {
	return vmr.vmType
}

func (vmr *VmRef) VmId() int {
	return vmr.vmId
}

func (vmr *VmRef) Node() string {
	return vmr.node
}

func (vmr *VmRef) Pool() string {
	return vmr.pool
}

func (vmr *VmRef) HaState() string {
	return vmr.haState
}

func (vmr *VmRef) HaGroup() string {
	return vmr.haGroup
}

func (vmr *VmRef) nilCheck() error {
	if vmr == nil {
		return errors.New("vm reference may not be nil")
	}
	return nil
}

func NewVmRef(vmId int) (vmr *VmRef) {
	vmr = &VmRef{vmId: vmId, node: "", vmType: ""}
	return
}

func NewClient(apiUrl string, hclient *http.Client, http_headers string, tls *tls.Config, proxyString string, taskTimeout int) (client *Client, err error) {
	var sess *Session
	sess, err_s := NewSession(apiUrl, hclient, proxyString, tls)
	sess, err = createHeaderList(http_headers, sess)
	if err != nil {
		return nil, err
	}
	if err_s == nil {
		client = &Client{session: sess, ApiUrl: apiUrl, TaskTimeout: taskTimeout}
	}

	return client, err_s
}

// SetAPIToken specifies a pair of user identifier and token UUID to use
// for authenticating API calls.
// If this is set, a ticket from calling `Login` will not be used.
//
// - `userID` is expected to be in the form `USER@REALM!TOKENID`
// - `token` is just the UUID you get when initially creating the token
//
// See https://pve.proxmox.com/wiki/User_Management#pveum_tokens
func (c *Client) SetAPIToken(userID, token string) {
	c.session.SetAPIToken(userID, token)
}

func (c *Client) Login(username string, password string, otp string) (err error) {
	c.Username = username
	c.Password = password
	c.Otp = otp
	return c.session.Login(username, password, otp)
}

func (c *Client) GetVersion() (data map[string]interface{}, err error) {
	resp, err := c.session.Get("/version", nil, nil)
	if err != nil {
		return nil, err
	}

	return ResponseJSON(resp)
}

func (c *Client) GetJsonRetryable(url string, data *map[string]interface{}, tries int) error {
	var statErr error
	for ii := 0; ii < tries; ii++ {
		_, statErr = c.session.GetJSON(url, nil, nil, data)
		if statErr == nil {
			return nil
		}
		if strings.Contains(statErr.Error(), "500 no such resource") {
			return statErr
		}
		// fmt.Printf("[DEBUG][GetJsonRetryable] Sleeping for %d seconds before asking url %s", ii+1, url)
		time.Sleep(time.Duration(ii+1) * time.Second)
	}
	return statErr
}

func (c *Client) GetNodeList() (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable("/nodes", &list, 3)
	return
}

const resourceListGuest string = "vm"

// GetResourceList returns a list of all enabled proxmox resources.
// For resource types that can be in a disabled state, disabled resources
// will not be returned
// TODO this func should not be exported
func (c *Client) GetResourceList(resourceType string) (list []interface{}, err error) {
	url := "/cluster/resources"
	if resourceType != "" {
		url = url + "?type=" + resourceType
	}
	return c.GetItemListInterfaceArray(url)
}

// TODO deprecate once nothing uses this anymore, use ListGuests() instead
func (c *Client) GetVmList() (map[string]interface{}, error) {
	list, err := c.GetResourceList(resourceListGuest)
	return map[string]interface{}{"data": list}, err
}

func (c *Client) CheckVmRef(vmr *VmRef) (err error) {
	if vmr.node == "" || vmr.vmType == "" {
		_, err = c.GetVmInfo(vmr)
	}
	return
}

func (c *Client) GetVmInfo(vmr *VmRef) (vmInfo map[string]interface{}, err error) {
	vms, err := c.GetResourceList(resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if int(vm["vmid"].(float64)) == vmr.vmId {
			vmInfo = vm
			vmr.node = vmInfo["node"].(string)
			vmr.vmType = vmInfo["type"].(string)
			vmr.pool = ""
			if vmInfo["pool"] != nil {
				vmr.pool = vmInfo["pool"].(string)
			}
			if vmInfo["hastate"] != nil {
				vmr.haState = vmInfo["hastate"].(string)
			}
			return
		}
	}
	return nil, fmt.Errorf("vm '%d' not found", vmr.vmId)
}

func (c *Client) GetVmRefByName(vmName string) (vmr *VmRef, err error) {
	vmrs, err := c.GetVmRefsByName(vmName)
	if err != nil {
		return nil, err
	}

	return vmrs[0], nil
}

func (c *Client) GetVmRefsByName(vmName string) (vmrs []*VmRef, err error) {
	vms, err := c.GetResourceList(resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if vm["name"] != nil && vm["name"].(string) == vmName {
			vmr := NewVmRef(int(vm["vmid"].(float64)))
			vmr.node = vm["node"].(string)
			vmr.vmType = vm["type"].(string)
			vmr.pool = ""
			if vm["pool"] != nil {
				vmr.pool = vm["pool"].(string)
			}
			if vm["hastate"] != nil {
				vmr.haState = vm["hastate"].(string)
			}
			vmrs = append(vmrs, vmr)
		}
	}
	if len(vmrs) == 0 {
		return nil, fmt.Errorf("vm '%s' not found", vmName)
	} else {
		return
	}
}

func (c *Client) GetVmState(vmr *VmRef) (vmState map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	return c.GetItemConfigMapStringInterface("/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/status/current", "vm", "STATE")
}

func (c *Client) GetVmConfig(vmr *VmRef) (vmConfig map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	return c.GetItemConfigMapStringInterface("/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/config", "vm", "CONFIG")
}

func (c *Client) GetStorageStatus(vmr *VmRef, storageName string) (storageStatus map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/storage/%s/status", vmr.node, storageName)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("storage STATUS not readable")
	}
	storageStatus = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetStorageContent(vmr *VmRef, storageName string) (data map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", vmr.node, storageName)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("storage Content not readable")
	}
	return
}

func (c *Client) GetVmSpiceProxy(vmr *VmRef) (vmSpiceProxy map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/spiceproxy", vmr.node, vmr.vmType, vmr.vmId)
	_, err = c.session.PostJSON(url, nil, nil, nil, &data)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("vm SpiceProxy not readable")
	}
	vmSpiceProxy = data["data"].(map[string]interface{})
	return
}

func (a *AgentNetworkInterface) UnmarshalJSON(b []byte) error {
	var intermediate struct {
		HardwareAddress string `json:"hardware-address"`
		IPAddresses     []struct {
			IPAddress     string `json:"ip-address"`
			IPAddressType string `json:"ip-address-type"`
			Prefix        int    `json:"prefix"`
		} `json:"ip-addresses"`
		Name       string           `json:"name"`
		Statistics map[string]int64 `json:"statistics"`
	}
	err := json.Unmarshal(b, &intermediate)
	if err != nil {
		return err
	}

	a.IPAddresses = make([]net.IP, len(intermediate.IPAddresses))
	for idx, ip := range intermediate.IPAddresses {
		a.IPAddresses[idx] = net.ParseIP((strings.Split(ip.IPAddress, "%"))[0])
		if a.IPAddresses[idx] == nil {
			return fmt.Errorf("could not parse %s as IP", ip.IPAddress)
		}
	}
	a.MACAddress = intermediate.HardwareAddress
	a.Name = intermediate.Name
	a.Statistics = intermediate.Statistics
	return nil
}

func (c *Client) GetVmAgentNetworkInterfaces(vmr *VmRef) ([]AgentNetworkInterface, error) {
	var ifs []AgentNetworkInterface
	err := c.doAgentGet(vmr, "network-get-interfaces", &ifs)
	return ifs, err
}

func (c *Client) doAgentGet(vmr *VmRef, command string, output interface{}) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/agent/%s", vmr.node, vmr.vmType, vmr.vmId, command)
	resp, err := c.session.Get(url, nil, nil)
	if err != nil {
		return err
	}

	return TypedResponse(resp, output)
}

func (c *Client) CreateTemplate(vmr *VmRef) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/template", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, nil)
	if err != nil {
		return err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return err
	}

	exitStatus, err := c.WaitForCompletion(taskResponse)
	if err != nil {
		return err
	}

	// Specifically ignore empty exit status for LXCs, since they don't return a task ID
	// when creating templates in the first place (but still successfully create them).
	if exitStatus != "OK" && vmr.vmType != "lxc" {
		return errors.New("Can't convert Vm to template:" + exitStatus)
	}

	return nil
}

func (c *Client) MonitorCmd(vmr *VmRef, command string) (monitorRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(map[string]interface{}{"command": command})
	url := fmt.Sprintf("/nodes/%s/%s/%d/monitor", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		return nil, err
	}
	monitorRes, err = ResponseJSON(resp)
	return
}

func (c *Client) Sendkey(vmr *VmRef, qmKey string) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}
	reqbody := ParamsToBody(map[string]interface{}{"key": qmKey})
	url := fmt.Sprintf("/nodes/%s/%s/%d/sendkey", vmr.node, vmr.vmType, vmr.vmId)
	// No return, even for errors: https://bugzilla.proxmox.com/show_bug.cgi?id=2275
	_, err = c.session.Put(url, nil, nil, &reqbody)

	return err
}

// WaitForCompletion - poll the API for task completion
func (c *Client) WaitForCompletion(taskResponse map[string]interface{}) (waitExitStatus string, err error) {
	if taskResponse["errors"] != nil {
		errJSON, _ := json.MarshalIndent(taskResponse["errors"], "", "  ")
		return string(errJSON), fmt.Errorf("error response")
	}
	if taskResponse["data"] == nil {
		return "", nil
	}
	waited := 0
	taskUpid := taskResponse["data"].(string)
	for waited < c.TaskTimeout {
		exitStatus, statErr := c.GetTaskExitstatus(taskUpid)
		if statErr != nil {
			if statErr != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
				return "", statErr
			}
		}
		if exitStatus != nil {
			waitExitStatus = exitStatus.(string)
			return
		}
		time.Sleep(TaskStatusCheckInterval * time.Second)
		waited = waited + TaskStatusCheckInterval
	}
	return "", fmt.Errorf("Wait timeout for:" + taskUpid)
}

var (
	rxTaskNode          = regexp.MustCompile("UPID:(.*?):")
	rxExitStatusSuccess = regexp.MustCompile(`^(OK|WARNINGS)`)
)

func (c *Client) GetTaskExitstatus(taskUpid string) (exitStatus interface{}, err error) {
	node := rxTaskNode.FindStringSubmatch(taskUpid)[1]
	url := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskUpid)
	var data map[string]interface{}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		exitStatus = data["data"].(map[string]interface{})["exitstatus"]
	}
	if exitStatus != nil && rxExitStatusSuccess.FindString(exitStatus.(string)) == "" {
		err = fmt.Errorf(exitStatus.(string))
	}
	return
}

func (c *Client) StatusChangeVm(vmr *VmRef, params map[string]interface{}, setStatus string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/status/%s", vmr.node, vmr.vmType, vmr.vmId, setStatus)
	for i := 0; i < 3; i++ {
		exitStatus, err = c.PostWithTask(params, url)
		if err != nil {
			time.Sleep(TaskStatusCheckInterval * time.Second)
		} else {
			return
		}
	}
	return
}

func (c *Client) StartVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "start")
}

func (c *Client) StopVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "stop")
}

func (c *Client) ShutdownVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "shutdown")
}

func (c *Client) ResetVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "reset")
}

func (c *Client) PauseVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "suspend")
}

func (c *Client) HibernateVm(vmr *VmRef) (exitStatus string, err error) {
	params := map[string]interface{}{
		"todisk": true,
	}
	return c.StatusChangeVm(vmr, params, "suspend")
}

func (c *Client) ResumeVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, nil, "resume")
}

func (c *Client) DeleteVm(vmr *VmRef) (exitStatus string, err error) {
	return c.DeleteVmParams(vmr, nil)
}

func (c *Client) DeleteVmParams(vmr *VmRef, params map[string]interface{}) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}

	// Remove HA if required
	if vmr.haState != "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return "", err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
			if err != nil {
				return "", err
			}
		}
	}

	values := ParamsToValues(params)
	url := fmt.Sprintf("/nodes/%s/%s/%d", vmr.node, vmr.vmType, vmr.vmId)
	var taskResponse map[string]interface{}
	if len(values) != 0 {
		_, err = c.session.RequestJSON("DELETE", url, &values, nil, nil, &taskResponse)
	} else {
		_, err = c.session.RequestJSON("DELETE", url, nil, nil, nil, &taskResponse)
	}
	if err != nil {
		return
	}
	exitStatus, err = c.WaitForCompletion(taskResponse)
	return
}

func (c *Client) CreateQemuVm(node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	// Create VM disks first to ensure disks names.
	createdDisks, createdDisksErr := c.createVMDisks(node, vmParams)
	if createdDisksErr != nil {
		return "", createdDisksErr
	}

	// Then create the VM itself.
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu", node)
	var resp *http.Response
	resp, err = c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		// Only attempt to read the body if it is available.
		if resp != nil && resp.Body != nil {
			defer resp.Body.Close()
			// This might not work if we never got a body. We'll ignore errors in trying to read,
			// but extract the body if possible to give any error information back in the exitStatus
			b, _ := io.ReadAll(resp.Body)
			exitStatus = string(b)

			return exitStatus, err
		}

		return "", err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	exitStatus, err = c.WaitForCompletion(taskResponse)
	// Delete VM disks if the VM didn't create.
	if exitStatus != "OK" {
		deleteDisksErr := c.DeleteVMDisks(node, createdDisks)
		if deleteDisksErr != nil {
			return "", deleteDisksErr
		}
	}

	return
}

func (c *Client) CreateLxcContainer(node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/lxc", node)
	var resp *http.Response
	resp, err = c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		defer resp.Body.Close()
		// This might not work if we never got a body. We'll ignore errors in trying to read,
		// but extract the body if possible to give any error information back in the exitStatus
		b, _ := io.ReadAll(resp.Body)
		exitStatus = string(b)
		return exitStatus, err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	exitStatus, err = c.WaitForCompletion(taskResponse)

	return
}

func (c *Client) CloneLxcContainer(vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/lxc/%s/clone", vmr.node, vmParams["vmid"])
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

func (c *Client) CloneQemuVm(vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/clone", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED superseded by CreateSnapshot()
func (c *Client) CreateQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	snapshotParams := map[string]interface{}{
		"snapname": snapshotName,
	}
	reqbody := ParamsToBody(snapshotParams)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED superseded by DeleteSnapshot()
func (c *Client) DeleteQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	return DeleteSnapshot(c, vmr, snapshotName)
}

// DEPRECATED superseded by ListSnapshots()
func (c *Client) ListQemuSnapshot(vmr *VmRef) (taskResponse map[string]interface{}, exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, "", err
		}
		return taskResponse, "", nil
	}
	return
}

// DEPRECATED superseded by RollbackSnapshot()
func (c *Client) RollbackQemuVm(vmr *VmRef, snapshot string) (exitStatus string, err error) {
	return RollbackSnapshot(c, vmr, snapshot)
}

// DEPRECATED SetVmConfig - send config options
func (c *Client) SetVmConfig(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	return c.PostWithTask(params, "/nodes/"+vmr.node+"/"+vmr.vmType+"/"+strconv.Itoa(vmr.vmId)+"/config")
}

// SetLxcConfig - send config options
func (c *Client) SetLxcConfig(vmr *VmRef, vmParams map[string]interface{}) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// MigrateNode - Migrate a VM
func (c *Client) MigrateNode(vmr *VmRef, newTargetNode string, online bool) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"target": newTargetNode, "online": online, "with-local-disks": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/migrate", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		return exitStatus, err
	}
	return nil, err
}

// ResizeQemuDisk allows the caller to increase the size of a disk by the indicated number of gigabytes
// TODO Deprecate once LXC is able to resize disk by itself (qemu can already do this)
func (c *Client) ResizeQemuDisk(vmr *VmRef, disk string, moreSizeGB int) (exitStatus interface{}, err error) {
	size := fmt.Sprintf("+%dG", moreSizeGB)
	return c.ResizeQemuDiskRaw(vmr, disk, size)
}

// ResizeQemuDiskRaw allows the caller to provide the raw resize string to be send to proxmox.
// See the proxmox API documentation for full information, but the short version is if you prefix
// your desired size with a '+' character it will ADD size to the disk.  If you just specify the size by
// itself it will do an absolute resizing to the specified size. Permitted suffixes are K, M, G, T
// to indicate order of magnitude (kilobyte, megabyte, etc). Decrease of disk size is not permitted.
// TODO Deprecate once LXC is able to resize disk by itself (qemu can already do this)
func (c *Client) ResizeQemuDiskRaw(vmr *VmRef, disk string, size string) (exitStatus interface{}, err error) {
	// PUT
	//disk:virtio0
	//size:+2G
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "size": size})
	url := fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

func (c *Client) MoveLxcDisk(vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_volume", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED use MoveQemuDisk() instead.
// MoveQemuDisk - Move a disk from one storage to another
func (c *Client) MoveQemuDisk(vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// MoveQemuDiskToVM - Move a disk to a different VM, using the same storage
func (c *Client) MoveQemuDiskToVM(vmrSource *VmRef, disk string, vmrTarget *VmRef) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "target-vmid": vmrTarget.vmId, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmrSource.node, vmrSource.vmType, vmrSource.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// GetNextID - Get next free VMID
func (c *Client) GetNextID(currentID int) (nextID int, err error) {
	var data map[string]interface{}
	var url string
	if currentID >= 100 {
		url = fmt.Sprintf("/cluster/nextid?vmid=%d", currentID)
	} else {
		url = "/cluster/nextid"
	}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		if data["errors"] != nil {
			if currentID >= 100 {
				return c.GetNextID(currentID + 1)
			} else {
				return -1, fmt.Errorf("error using /cluster/nextid")
			}
		}
		nextID, err = strconv.Atoi(data["data"].(string))
	} else if strings.HasPrefix(err.Error(), "400 ") {
		return c.GetNextID(currentID + 1)
	}
	return
}

// VMIdExists - If you pass an VMID that exists it will return true, otherwise it wil return false
func (c *Client) VMIdExists(vmID int) (exists bool, err error) {
	vms, err := c.GetResourceList(resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if vmID == int(vm["vmid"].(float64)) {
			return true, err
		}
	}
	return
}

// CreateVMDisk - Create single disk for VM on host node.
func (c *Client) CreateVMDisk(
	nodeName string,
	storageName string,
	fullDiskName string,
	diskParams map[string]interface{},
) error {
	reqbody := ParamsToBody(diskParams)
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", nodeName, storageName)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return err
		}
		if diskName, containsData := taskResponse["data"]; !containsData || diskName != fullDiskName {
			return fmt.Errorf("cannot create VM disk %s - %s", fullDiskName, diskName)
		}
	} else {
		return err
	}

	return nil
}

var rxStorageModels = regexp.MustCompile(`(ide|sata|scsi|virtio)\d+`)

// createVMDisks - Make disks parameters and create all VM disks on host node.
func (c *Client) createVMDisks(
	node string,
	vmParams map[string]interface{},
) (disks []string, err error) {
	var createdDisks []string
	vmID := vmParams["vmid"].(int)
	for deviceName, deviceConf := range vmParams {
		if matched := rxStorageModels.MatchString(deviceName); matched {
			deviceConfMap := ParsePMConf(deviceConf.(string), "")
			// This if condition to differentiate between `disk` and `cdrom`.
			if media, containsFile := deviceConfMap["media"]; containsFile && media == "disk" {
				fullDiskName := deviceConfMap["file"].(string)
				storageName, volumeName := getStorageAndVolumeName(fullDiskName, ":")
				diskParams := map[string]interface{}{
					"vmid":     vmID,
					"filename": volumeName,
					"size":     deviceConfMap["size"],
				}
				err := c.CreateVMDisk(node, storageName, fullDiskName, diskParams)
				if err != nil {
					return createdDisks, err
				} else {
					createdDisks = append(createdDisks, fullDiskName)
				}
			}
		}
	}

	return createdDisks, nil
}

// CreateNewDisk - This method allows simpler disk creation for direct client users
// It should work for any existing container and virtual machine
func (c *Client) CreateNewDisk(vmr *VmRef, disk string, volume string) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{disk: volume})
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DeleteVMDisks - Delete VM disks from host node.
// By default the VM disks are deleted when the VM is deleted,
// so mainly this is used to delete the disks in case VM creation didn't complete.
func (c *Client) DeleteVMDisks(
	node string,
	disks []string,
) error {
	for _, fullDiskName := range disks {
		storageName, volumeName := getStorageAndVolumeName(fullDiskName, ":")
		url := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storageName, volumeName)
		_, err := c.session.Post(url, nil, nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// VzDump - Create backup
func (c *Client) VzDump(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/vzdump", vmr.node)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DeleteVolume - Delete volume
func (c *Client) DeleteVolume(vmr *VmRef, storageName string, volumeName string) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", vmr.node, storageName, volumeName)
	resp, err := c.session.Delete(url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// CreateVNCProxy - Creates a TCP VNC proxy connections
func (c *Client) CreateVNCProxy(vmr *VmRef, params map[string]interface{}) (vncProxyRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		return nil, err
	}
	vncProxyRes, err = ResponseJSON(resp)
	if err != nil {
		return nil, err
	}
	if vncProxyRes["data"] == nil {
		return nil, fmt.Errorf("VNC Proxy not readable")
	}
	vncProxyRes = vncProxyRes["data"].(map[string]interface{})
	return
}

// QemuAgentPing - Execute ping.
func (c *Client) QemuAgentPing(vmr *VmRef) (pingRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/ping", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		if taskResponse["data"] == nil {
			return nil, fmt.Errorf("qemu agent ping not readable")
		}
		pingRes = taskResponse["data"].(map[string]interface{})
	}
	return
}

// QemuAgentFileWrite - Writes the given file via guest agent.
func (c *Client) QemuAgentFileWrite(vmr *VmRef, params map[string]interface{}) (err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-write", vmr.node, vmr.vmId)
	_, err = c.session.Post(url, nil, nil, &reqbody)
	return
}

// QemuAgentSetUserPassword - Sets the password for the given user to the given password.
func (c *Client) QemuAgentSetUserPassword(vmr *VmRef, params map[string]interface{}) (result map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		if taskResponse["data"] == nil {
			return nil, fmt.Errorf("qemu agent set user password not readable")
		}
		result = taskResponse["data"].(map[string]interface{})
	}
	return
}

// QemuAgentExec - Executes the given command in the vm via the guest-agent and returns an object with the pid.
func (c *Client) QemuAgentExec(vmr *VmRef, params map[string]interface{}) (result map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		if taskResponse["data"] == nil {
			return nil, fmt.Errorf("qemu agent exec not readable")
		}
		result = taskResponse["data"].(map[string]interface{})
	}
	return
}

// GetExecStatus - Gets the status of the given pid started by the guest-agent
func (c *Client) GetExecStatus(vmr *VmRef, pid string) (status map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	err = c.GetJsonRetryable(fmt.Sprintf("/nodes/%s/%s/%d/agent/exec-status?pid=%s", vmr.node, vmr.vmType, vmr.vmId, pid), &status, 3)
	if err == nil {
		status = status["data"].(map[string]interface{})
	}
	return
}

// SetQemuFirewallOptions - Set Firewall options.
func (c *Client) SetQemuFirewallOptions(vmr *VmRef, fwOptions map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(fwOptions)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// GetQemuFirewallOptions - Get VM firewall options.
func (c *Client) GetQemuFirewallOptions(vmr *VmRef) (firewallOptions map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		firewallOptions, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		return firewallOptions, nil
	}
	return
}

// CreateQemuIPSet - Create new IPSet
func (c *Client) CreateQemuIPSet(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// AddQemuIPSet - Add IP or Network to IPSet.
func (c *Client) AddQemuIPSet(vmr *VmRef, name string, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", vmr.node, vmr.vmId, name)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// GetQemuIPSet - List IPSets
func (c *Client) GetQemuIPSet(vmr *VmRef) (ipsets map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		ipsets, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		return ipsets, nil
	}
	return
}

// DeleteQemuIPSet - Delete IPSet
func (c *Client) DeleteQemuIPSet(vmr *VmRef, IPSetName string) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", vmr.node, vmr.vmId, IPSetName)
	resp, err := c.session.Delete(url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DeleteQemuIPSetNetwork - Remove IP or Network from IPSet.
func (c *Client) DeleteQemuIPSetNetwork(vmr *VmRef, IPSetName string, network string, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	values := ParamsToValues(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s/%s", vmr.node, vmr.vmId, IPSetName, network)
	resp, err := c.session.Delete(url, &values, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

func (c *Client) Upload(node string, storage string, contentType string, filename string, file io.Reader) error {
	var doStreamingIO bool
	var fileSize int64
	var contentLength int64

	if f, ok := file.(*os.File); ok {
		doStreamingIO = true
		fileInfo, err := f.Stat()
		if err != nil {
			return err
		}
		fileSize = fileInfo.Size()
	}

	var body io.Reader
	var mimetype string
	var err error

	if doStreamingIO {
		body, mimetype, contentLength, err = createStreamedUploadBody(contentType, filename, fileSize, file)
	} else {
		body, mimetype, err = createUploadBody(contentType, filename, file)
	}
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/nodes/%s/storage/%s/upload", c.session.ApiUrl, node, storage)
	headers := c.session.Headers.Clone()
	headers.Add("Content-Type", mimetype)
	headers.Add("Accept", "application/json")
	req, err := c.session.NewRequest(http.MethodPost, url, &headers, body)
	if err != nil {
		return err
	}

	if doStreamingIO {
		req.ContentLength = contentLength
	}

	resp, err := c.session.Do(req)
	if err != nil {
		return err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return err
	}
	exitStatus, err := c.WaitForCompletion(taskResponse)
	if err != nil {
		return err
	}
	if exitStatus != exitStatusSuccess {
		return fmt.Errorf("moving file to destination failed: %v", exitStatus)
	}
	return nil
}

func (c *Client) UploadLargeFile(node string, storage string, contentType string, filename string, filesize int64, file io.Reader) error {
	var contentLength int64

	var body io.Reader
	var mimetype string
	var err error
	body, mimetype, contentLength, err = createStreamedUploadBody(contentType, filename, filesize, file)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/nodes/%s/storage/%s/upload", c.session.ApiUrl, node, storage)
	headers := c.session.Headers.Clone()
	headers.Add("Content-Type", mimetype)
	headers.Add("Accept", "application/json")
	req, err := c.session.NewRequest(http.MethodPost, url, &headers, body)
	if err != nil {
		return err
	}

	req.ContentLength = contentLength

	resp, err := c.session.Do(req)
	if err != nil {
		return err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return err
	}
	exitStatus, err := c.WaitForCompletion(taskResponse)
	if err != nil {
		return err
	}
	if exitStatus != exitStatusSuccess {
		return fmt.Errorf("moving file to destination failed: %v", exitStatus)
	}
	return nil
}

func createUploadBody(contentType string, filename string, r io.Reader) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	err := w.WriteField("content", contentType)
	if err != nil {
		return nil, "", err
	}

	fw, err := w.CreateFormFile("filename", filename)
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(fw, r)
	if err != nil {
		return nil, "", err
	}

	err = w.Close()
	if err != nil {
		return nil, "", err
	}

	return &buf, w.FormDataContentType(), nil
}

// createStreamedUploadBody - Use MultiReader to create the multipart body from the file reader,
// avoiding allocation of large files in memory before upload (useful e.g. for Windows ISOs).
func createStreamedUploadBody(contentType string, filename string, fileSize int64, r io.Reader) (io.Reader, string, int64, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	err := w.WriteField("content", contentType)
	if err != nil {
		return nil, "", 0, err
	}

	_, err = w.CreateFormFile("filename", filename)
	if err != nil {
		return nil, "", 0, err
	}

	headerSize := buf.Len()

	err = w.Close()
	if err != nil {
		return nil, "", 0, err
	}

	mr := io.MultiReader(bytes.NewReader(buf.Bytes()[:headerSize]),
		r,
		bytes.NewReader(buf.Bytes()[headerSize:]))

	contentLength := int64(buf.Len()) + fileSize

	return mr, w.FormDataContentType(), contentLength, nil
}

// getStorageAndVolumeName - Extract disk storage and disk volume, since disk name is saved
// in Proxmox with its storage.
func getStorageAndVolumeName(
	fullDiskName string,
	separator string,
) (storageName string, diskName string) {
	storageAndVolumeName := strings.Split(fullDiskName, separator)
	storageName, volumeName := storageAndVolumeName[0], storageAndVolumeName[1]

	// when disk type is dir, volumeName is `file=local:100/vm-100-disk-0.raw`
	re := regexp.MustCompile(`\d+/(?P<filename>\S+.\S+)`)
	match := re.FindStringSubmatch(volumeName)
	if len(match) == 2 {
		volumeName = match[1]
	}

	return storageName, volumeName
}

func (c *Client) UpdateVMPool(vmr *VmRef, pool string) (exitStatus interface{}, err error) {
	// Same pool
	if vmr.pool == pool {
		return
	}

	// Remove from old pool
	if vmr.pool != "" {
		paramMap := map[string]interface{}{
			"vms":    vmr.vmId,
			"delete": 1,
		}
		reqbody := ParamsToBody(paramMap)
		url := fmt.Sprintf("/pools/%s", vmr.pool)
		resp, err := c.session.Put(url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)

			if err != nil {
				return nil, err
			}
		}
	}
	// Add to the new pool
	if pool != "" {
		paramMap := map[string]interface{}{
			"vms": vmr.vmId,
		}
		reqbody := ParamsToBody(paramMap)
		url := fmt.Sprintf("/pools/%s", pool)
		resp, err := c.session.Put(url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return
}

func (c *Client) ReadVMHA(vmr *VmRef) (err error) {
	var list map[string]interface{}
	url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
	err = c.GetJsonRetryable(url, &list, 3)
	if err == nil {
		list = list["data"].(map[string]interface{})
		for elem, value := range list {
			if elem == "group" {
				vmr.haGroup = value.(string)
			}
			if elem == "state" {
				vmr.haState = value.(string)
			}
		}
	}
	return
}

func (c *Client) UpdateVMHA(vmr *VmRef, haState string, haGroup string) (exitStatus interface{}, err error) {
	// Same hastate & hagroup
	if vmr.haState == haState && vmr.haGroup == haGroup {
		return
	}

	// Remove HA
	if haState == "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	// Activate HA
	if vmr.haState == "" {
		paramMap := map[string]interface{}{
			"sid": vmr.vmId,
		}
		if haGroup != "" {
			paramMap["group"] = haGroup
		}
		reqbody := ParamsToBody(paramMap)
		resp, err := c.session.Post("/cluster/ha/resources", nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)

			if err != nil {
				return nil, err
			}
		}
	}

	// Set wanted state
	paramMap := map[string]interface{}{
		"state": haState,
		"group": haGroup,
	}
	reqbody := ParamsToBody(paramMap)
	url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if err != nil {
			return nil, err
		}
	}

	return
}

func (c *Client) GetPoolList() (pools map[string]interface{}, err error) {
	return c.GetItemList("/pools")
}

func (c *Client) GetPoolInfo(poolid string) (poolInfo map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface("/pools/"+poolid, "pool", "CONFIG")
}

func (c *Client) CreatePool(poolid string, comment string) error {
	return c.Post(map[string]interface{}{
		"poolid":  poolid,
		"comment": comment,
	}, "/pools")
}

func (c *Client) UpdatePoolComment(poolid string, comment string) error {
	return c.Put(map[string]interface{}{
		"poolid":  poolid,
		"comment": comment,
	}, "/pools/"+poolid)
}

func (c *Client) DeletePool(poolid string) error {
	return c.Delete("/pools/" + poolid)
}

// permissions check
func (c *Client) GetUserPermissions(id UserID, path string) (permissions []string, err error) {
	existence, err := CheckUserExistence(id, c)
	if err != nil {
		return nil, err
	}
	if !existence {
		return nil, fmt.Errorf("cannot get user (%s) permissions, the user does not exist", id)
	}
	permlist, err := c.GetItemList("/access/permissions?userid=" + id.ToString() + "&path=" + path)
	failError(err)
	data := permlist["data"].(map[string]interface{})
	for pth, prm := range data {
		// ignoring other paths than "/" for now!
		if pth == "/" {
			for k := range prm.(map[string]interface{}) {
				permissions = append(permissions, k)
			}
		}
	}
	return
}

// ACME
func (c *Client) GetAcmeDirectoriesUrl() (url []string, err error) {
	config, err := c.GetItemConfigInterfaceArray("/cluster/acme/directories", "Acme directories", "CONFIG")
	url = make([]string, len(config))
	for i, element := range config {
		url[i] = element.(map[string]interface{})["url"].(string)
	}
	return
}

func (c *Client) GetAcmeTosUrl() (url string, err error) {
	return c.GetItemConfigString("/cluster/acme/tos", "Acme T.O.S.", "CONFIG")
}

// ACME Account
func (c *Client) GetAcmeAccountList() (accounts map[string]interface{}, err error) {
	return c.GetItemList("/cluster/acme/account")
}

func (c *Client) GetAcmeAccountConfig(id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface("/cluster/acme/account/"+id, "acme", "CONFIG")
}

func (c *Client) CreateAcmeAccount(params map[string]interface{}) (exitStatus string, err error) {
	return c.PostWithTask(params, "/cluster/acme/account/")
}

func (c *Client) UpdateAcmeAccountEmails(id, emails string) (exitStatus string, err error) {
	params := map[string]interface{}{
		"contact": emails,
	}
	return c.PutWithTask(params, "/cluster/acme/account/"+id)
}

func (c *Client) DeleteAcmeAccount(id string) (exitStatus string, err error) {
	return c.DeleteWithTask("/cluster/acme/account/" + id)
}

// ACME Plugin
func (c *Client) GetAcmePluginList() (accounts map[string]interface{}, err error) {
	return c.GetItemList("/cluster/acme/plugins")
}

func (c *Client) GetAcmePluginConfig(id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface("/cluster/acme/plugins/"+id, "acme plugin", "CONFIG")
}

func (c *Client) CreateAcmePlugin(params map[string]interface{}) error {
	return c.Post(params, "/cluster/acme/plugins/")
}

func (c *Client) UpdateAcmePlugin(id string, params map[string]interface{}) error {
	return c.Put(params, "/cluster/acme/plugins/"+id)
}

func (c *Client) CheckAcmePluginExistence(id string) (existance bool, err error) {
	list, err := c.GetAcmePluginList()
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "plugin", id)
	return
}

func (c *Client) DeleteAcmePlugin(id string) (err error) {
	return c.Delete("/cluster/acme/plugins/" + id)
}

// Metrics
func (c *Client) GetMetricServerConfig(id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface("/cluster/metrics/server/"+id, "metrics server", "CONFIG")
}

func (c *Client) GetMetricsServerList() (metricServers map[string]interface{}, err error) {
	return c.GetItemList("/cluster/metrics/server")
}

func (c *Client) CreateMetricServer(id string, params map[string]interface{}) error {
	return c.Post(params, "/cluster/metrics/server/"+id)
}

func (c *Client) UpdateMetricServer(id string, params map[string]interface{}) error {
	return c.Put(params, "/cluster/metrics/server/"+id)
}

func (c *Client) CheckMetricServerExistence(id string) (existance bool, err error) {
	list, err := c.GetMetricsServerList()
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "id", id)
	return
}

func (c *Client) DeleteMetricServer(id string) error {
	return c.Delete("/cluster/metrics/server/" + id)
}

// storage
func (c *Client) EnableStorage(id string) error {
	return c.Put(map[string]interface{}{
		"disable": false,
	}, "/storage/"+id)
}

func (c *Client) GetStorageList() (metricServers map[string]interface{}, err error) {
	return c.GetItemList("/storage")
}

func (c *Client) GetStorageConfig(id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface("/storage/"+id, "storage", "CONFIG")
}

func (c *Client) CreateStorage(params map[string]interface{}) error {
	return c.Post(params, "/storage")
}

func (c *Client) CheckStorageExistance(id string) (existance bool, err error) {
	list, err := c.GetStorageList()
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "storage", id)
	return
}

func (c *Client) UpdateStorage(id string, params map[string]interface{}) error {
	return c.Put(params, "/storage/"+id)
}

func (c *Client) DeleteStorage(id string) error {
	return c.Delete("/storage/" + id)
}

// Network

// GetNetworkList gets a json encoded list of currently configured network interfaces on the
// passed in node. The typeFilter parameter can be used to filter by interface type. Pass in
// the empty string "" for typeFilter to list all network interfaces on the node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) GetNetworkList(node string, typeFilter string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	if typeFilter != "" {
		url += fmt.Sprintf("?type=%s", typeFilter)
	}
	resp, err := c.session.Get(url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// GetNetworkInterface gets a json encoded object containing the configuration of the network
// interface with the name passed in as iface from the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) GetNetworkInterface(node string, iface string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	resp, err := c.session.Get(url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// CreateNetwork creates a network with the configuration of the passed in parameters
// on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) CreateNetwork(node string, params map[string]interface{}) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	return c.CreateItemReturnStatus(params, url)
}

// UpdateNetwork updates the network corresponding to the passed in interface name on the passed
// in node with the configuration in the passed in parameters.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) UpdateNetwork(node string, iface string, params map[string]interface{}) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	return c.UpdateItemReturnStatus(params, url)
}

// DeleteNetwork deletes the network with the passed in iface name on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) DeleteNetwork(node string, iface string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	resp, err := c.session.Delete(url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// ApplyNetwork applies the pending network configuration on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c Client) ApplyNetwork(node string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	return c.PutWithTask(nil, url)
}

// RevertNetwork reverts the pending network configuration on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) RevertNetwork(node string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	return c.DeleteWithTask(url)
}

// SDN

func (c *Client) ApplySDN() (string, error) {
	return c.PutWithTask(nil, "/cluster/sdn")
}

// GetSDNVNets returns a list of all VNet definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNVNets(pending bool) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets?pending=%d", Btoi(pending))
	err = c.GetJsonRetryable(url, &list, 3)
	return
}

// CheckSDNVNetExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNVNetExistance(id string) (existance bool, err error) {
	list, err := c.GetSDNVNets(true)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "vnet", id)
	return
}

// GetSDNVNet returns details about the DNS entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNVNet struct.
func (c *Client) GetSDNVNet(name string) (dns map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s", name)
	err = c.GetJsonRetryable(url, &dns, 3)
	return
}

// CreateSDNVNet creates a new SDN DNS in the cluster
func (c *Client) CreateSDNVNet(params map[string]interface{}) error {
	return c.Post(params, "/cluster/sdn/vnets")
}

// DeleteSDNVNet deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNVNet(name string) error {
	return c.Delete(fmt.Sprintf("/cluster/sdn/vnets/%s", name))
}

// UpdateSDNVNet updates the given DNS with the provided parameters
func (c *Client) UpdateSDNVNet(id string, params map[string]interface{}) error {
	return c.Put(params, "/cluster/sdn/vnets/"+id)
}

// GetSDNSubnets returns a list of all Subnet definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNSubnets(vnet string) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", vnet)
	err = c.GetJsonRetryable(url, &list, 3)
	return
}

// CheckSDNSubnetExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNSubnetExistance(vnet, id string) (existance bool, err error) {
	list, err := c.GetSDNSubnets(vnet)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "subnet", id)
	return
}

// GetSDNSubnet returns details about the Subnet entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned map["data"] section can be unmarshalled into a ConfigSDNSubnet struct.
func (c *Client) GetSDNSubnet(vnet, name string) (subnet map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, name)
	err = c.GetJsonRetryable(url, &subnet, 3)
	return
}

// CreateSDNSubnet creates a new SDN DNS in the cluster
func (c *Client) CreateSDNSubnet(vnet string, params map[string]interface{}) error {
	return c.Post(params, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", vnet))
}

// DeleteSDNSubnet deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNSubnet(vnet, name string) error {
	return c.Delete(fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, name))
}

// UpdateSDNSubnet updates the given DNS with the provided parameters
func (c *Client) UpdateSDNSubnet(vnet, id string, params map[string]interface{}) error {
	return c.Put(params, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, id))
}

// GetSDNDNSs returns a list of all DNS definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNDNSs(typeFilter string) (list map[string]interface{}, err error) {
	url := "/cluster/sdn/dns"
	if typeFilter != "" {
		url += fmt.Sprintf("&type=%s", typeFilter)
	}
	err = c.GetJsonRetryable(url, &list, 3)
	return
}

// CheckSDNDNSExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNDNSExistance(id string) (existance bool, err error) {
	list, err := c.GetSDNDNSs("")
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "dns", id)
	return
}

// GetSDNDNS returns details about the DNS entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNDNS struct.
func (c *Client) GetSDNDNS(name string) (dns map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/dns/%s", name)
	err = c.GetJsonRetryable(url, &dns, 3)
	return
}

// CreateSDNDNS creates a new SDN DNS in the cluster
func (c *Client) CreateSDNDNS(params map[string]interface{}) error {
	return c.Post(params, "/cluster/sdn/dns")
}

// DeleteSDNDNS deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNDNS(name string) error {
	return c.Delete(fmt.Sprintf("/cluster/sdn/dns/%s", name))
}

// UpdateSDNDNS updates the given DNS with the provided parameters
func (c *Client) UpdateSDNDNS(id string, params map[string]interface{}) error {
	return c.Put(params, "/cluster/sdn/dns/"+id)
}

// GetSDNZones returns a list of all the SDN zones defined in the cluster.
func (c *Client) GetSDNZones(pending bool, typeFilter string) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/zones?pending=%d", Btoi(pending))
	if typeFilter != "" {
		url += fmt.Sprintf("&type=%s", typeFilter)
	}
	err = c.GetJsonRetryable(url, &list, 3)
	return
}

// CheckSDNZoneExistance returns true if a zone with the provided ID exists, false otherwise.
func (c *Client) CheckSDNZoneExistance(id string) (existance bool, err error) {
	list, err := c.GetSDNZones(true, "")
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "zone", id)
	return
}

// GetSDNZone returns details about the zone whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNZone struct.
func (c *Client) GetSDNZone(zoneName string) (zone map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/zones/%s", zoneName)
	err = c.GetJsonRetryable(url, &zone, 3)
	return
}

// CreateSDNZone creates a new SDN zone in the cluster
func (c *Client) CreateSDNZone(params map[string]interface{}) error {
	return c.Post(params, "/cluster/sdn/zones")
}

// DeleteSDNZone deletes an existing SDN zone in the cluster
func (c *Client) DeleteSDNZone(zoneName string) error {
	return c.Delete(fmt.Sprintf("/cluster/sdn/zones/%s", zoneName))
}

// UpdateSDNZone updates the given zone with the provided parameters
func (c *Client) UpdateSDNZone(id string, params map[string]interface{}) error {
	return c.Put(params, "/cluster/sdn/zones/"+id)
}

// Shared
func (c *Client) GetItemConfigMapStringInterface(url, text, message string) (map[string]interface{}, error) {
	data, err := c.GetItemConfig(url, text, message)
	if err != nil {
		return nil, err
	}
	return data["data"].(map[string]interface{}), err
}

func (c *Client) GetItemConfigString(url, text, message string) (string, error) {
	data, err := c.GetItemConfig(url, text, message)
	if err != nil {
		return "", err
	}
	return data["data"].(string), err
}

func (c *Client) GetItemConfigInterfaceArray(url, text, message string) ([]interface{}, error) {
	data, err := c.GetItemConfig(url, text, message)
	if err != nil {
		return nil, err
	}
	return data["data"].([]interface{}), err
}

func (c *Client) GetItemConfig(url, text, message string) (config map[string]interface{}, err error) {
	err = c.GetJsonRetryable(url, &config, 3)
	if err != nil {
		return nil, err
	}
	if config["data"] == nil {
		return nil, fmt.Errorf(text + " " + message + " not readable")
	}
	return
}

// Makes a POST request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *Client) Post(Params map[string]interface{}, url string) (err error) {
	reqbody := ParamsToBody(Params)
	_, err = c.session.Post(url, nil, nil, &reqbody)
	return
}

// CreateItemReturnStatus creates an item on the Proxmox API.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *Client) CreateItemReturnStatus(params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(params)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	exitStatus = c.HandleTaskError(resp)
	return
}

// Makes a POST request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
func (c *Client) PostWithTask(Params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(Params)
	var resp *http.Response
	resp, err = c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(resp)
}

// Makes a PUT request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *Client) Put(Params map[string]interface{}, url string) (err error) {
	reqbody := ParamsToBodyWithAllEmpty(Params)
	_, err = c.session.Put(url, nil, nil, &reqbody)
	return
}

// UpdateItemReturnStatus updates an item on the Proxmox API.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *Client) UpdateItemReturnStatus(params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(params)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	exitStatus = c.HandleTaskError(resp)
	return
}

// Makes a PUT request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
func (c *Client) PutWithTask(Params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBodyWithAllEmpty(Params)
	var resp *http.Response
	resp, err = c.session.Put(url, nil, nil, &reqbody)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(resp)
}

// Makes a DELETE request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *Client) Delete(url string) (err error) {
	_, err = c.session.Delete(url, nil, nil)
	return
}

// Makes a DELETE request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
func (c *Client) DeleteWithTask(url string) (exitStatus string, err error) {
	var resp *http.Response
	resp, err = c.session.Delete(url, nil, nil)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(resp)
}

func (c *Client) GetItemListInterfaceArray(url string) ([]interface{}, error) {
	list, err := c.GetItemList(url)
	if err != nil {
		return nil, err
	}
	data, ok := list["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast response to list, resp: %v", list)
	}
	return data, nil
}

func (c *Client) GetItemList(url string) (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable(url, &list, 3)
	return
}

// HandleTaskError reads the body from the passed in HTTP response and closes it.
// It returns the body of the passed in HTTP response.
func (c *Client) HandleTaskError(resp *http.Response) (exitStatus string) {
	// Only attempt to read the body if it is available.
	if resp == nil || resp.Body == nil {
		return "no body available for HTTP response"
	}

	defer resp.Body.Close()
	// This might not work if we never got a body. We'll ignore errors in trying to read,
	// but extract the body if possible to give any error information back in the exitStatus
	b, _ := io.ReadAll(resp.Body)

	return string(b)
}

// CheckTask polls the API to check if the Proxmox task has been completed.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *Client) CheckTask(resp *http.Response) (exitStatus string, err error) {
	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	return c.WaitForCompletion(taskResponse)
}
