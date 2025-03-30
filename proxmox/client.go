package proxmox

// inspired by https://github.com/Telmate/vagrant-proxmox/blob/master/lib/vagrant-proxmox/proxmox/connection.rb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TaskStatusCheckInterval - time between async checks in seconds
const TaskStatusCheckInterval = 2

const exitStatusSuccess = "OK"

// Client - URL, user and password to specific Proxmox node
type Client struct {
	session            *Session
	ApiUrl             string
	Username           string
	Password           string
	Otp                string
	TaskTimeout        int
	permissionMutex    sync.Mutex
	permissions        map[permissionPath]privileges
	version            *Version
	versionMutex       sync.Mutex
	guestCreationMutex sync.Mutex
	Features           *FeatureFlags
}

type FeatureFlags struct {
	AsyncTask bool
}

type clientInterface interface {
	delete(ctx context.Context, url string) error
	getItemConfig(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error)
	getItemList(ctx context.Context, url string) (map[string]interface{}, error)
}

type client struct {
	c *Client
}

func (c *client) delete(ctx context.Context, url string) (err error) {
	_, err = c.c.session.Delete(ctx, url, nil, nil)
	return
}

func (c *client) getItemConfig(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error) {
	return c.c.GetItemConfig(ctx, url, text, message, errorStrings...)
}

func (c *client) getItemList(ctx context.Context, url string) (map[string]interface{}, error) {
	return c.c.GetItemList(ctx, url)
}

type mocClient struct {
	deleteFunc        func(ctx context.Context, url string) error
	getItemConfigFunc func(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error)
	getItemListFunc   func(ctx context.Context, url string) (map[string]interface{}, error)
}

func (c *mocClient) delete(ctx context.Context, url string) (err error) {
	return c.deleteFunc(ctx, url)
}

func (c *mocClient) getItemConfig(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error) {
	return c.getItemConfigFunc(ctx, url, text, message, errorStrings)
}

func (c *mocClient) getItemList(ctx context.Context, url string) (map[string]interface{}, error) {
	return c.getItemListFunc(ctx, url)
}

const (
	Client_Error_Nil            string = "client may not be nil"
	Client_Error_NotInitialized string = "client not initialized"
)

// Checks if the client is initialized and returns an error if not
func (c *Client) checkInitialized() error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if c.session == nil {
		return errors.New(Client_Error_NotInitialized)
	}
	return nil
}

// provides a fake client to bypass *Client.checkInitialized() during testing
func fakeClient() *Client {
	return &Client{session: &Session{}}
}

const (
	VmRef_Error_Nil string = "vm reference may not be nil"
)

// VmRef - virtual machine ref parts
// map[type:qemu node:proxmox1-xx id:qemu/132 diskread:5.57424738e+08 disk:0 netin:5.9297450593e+10 mem:3.3235968e+09 uptime:1.4567097e+07 vmid:132 template:0 maxcpu:2 netout:6.053310416e+09 maxdisk:3.4359738368e+10 maxmem:8.592031744e+09 diskwrite:1.49663619584e+12 status:running cpu:0.00386980694947209 name:appt-app1-dev.xxx.xx]
type VmRef struct {
	vmId    GuestID
	node    NodeName
	pool    PoolName
	vmType  string
	haState string
	haGroup string
}

func (vmr *VmRef) SetNode(node string) {
	vmr.node = NodeName(node)
}

func (vmr *VmRef) SetPool(pool string) {
	vmr.pool = PoolName(pool)
}

func (vmr *VmRef) SetVmType(vmType string) {
	vmr.vmType = vmType
}

func (vmr *VmRef) GetVmType() string {
	return vmr.vmType
}

func (vmr *VmRef) VmId() GuestID {
	return vmr.vmId
}

func (vmr *VmRef) Node() NodeName {
	return vmr.node
}

func (vmr *VmRef) Pool() PoolName {
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

func NewVmRef(vmId GuestID) (vmr *VmRef) {
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
		client = &Client{session: sess, ApiUrl: apiUrl, TaskTimeout: taskTimeout, permissions: make(map[permissionPath]privileges)}
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

// SetTicket let's set directly ticket and csrfPreventionToken obtained in
// a different way, for example using OIDC identity provider
//
// Parameters:
// - `ticket`
// - `csrfPreventionToken`
//
// Docs: https://pve.proxmox.com/wiki/Proxmox_VE_API#Authentication
func (c *Client) SetTicket(ticket, csrfPreventionToken string) {
	c.session.setTicket(ticket, csrfPreventionToken)
}

func (c *Client) Login(ctx context.Context, username string, password string, otp string) (err error) {
	c.Username = username
	c.Password = password
	c.Otp = otp
	return c.session.Login(ctx, username, password, otp)
}

// Updates the client's cached version information and returns it.
func (c *Client) GetVersion(ctx context.Context) (version Version, err error) {
	if c == nil {
		return Version{}, errors.New(Client_Error_Nil)
	}
	params, err := c.GetItemConfigMapStringInterface(ctx, "/version", "version", "data")
	version = version.mapToSDK(params)
	cachedVersion := Version{ // clones the struct
		Major: version.Major,
		Minor: version.Minor,
		Patch: version.Patch,
	}
	c.versionMutex.Lock()
	c.version = &cachedVersion
	c.versionMutex.Unlock()
	return
}

func (c *Client) GetJsonRetryable(ctx context.Context, url string, data *map[string]interface{}, tries int, errorString ...string) error {
	var statErr error
	for ii := 0; ii < tries; ii++ {
		_, statErr = c.session.GetJSON(ctx, url, nil, nil, data)
		if statErr == nil {
			return nil
		}
		// TODO can probable check for `500` status code instead of providing a list of error strings to check for
		if strings.Contains(statErr.Error(), "500 no such resource") {
			return statErr
		}
		for _, e := range errorString {
			if strings.Contains(statErr.Error(), e) {
				return statErr
			}
		}
		// fmt.Printf("[DEBUG][GetJsonRetryable] Sleeping for %d seconds before asking url %s", ii+1, url)
		time.Sleep(time.Duration(ii+1) * time.Second)
	}
	return statErr
}

func (c *Client) GetNodeList(ctx context.Context) (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable(ctx, "/nodes", &list, 3)
	return
}

const resourceListGuest string = "vm"

// GetResourceList returns a list of all enabled proxmox resources.
// For resource types that can be in a disabled state, disabled resources
// will not be returned
// TODO this func should not be exported
func (c *Client) GetResourceList(ctx context.Context, resourceType string) (list []interface{}, err error) {
	url := "/cluster/resources"
	if resourceType != "" {
		url = url + "?type=" + resourceType
	}
	return c.GetItemListInterfaceArray(ctx, url)
}

// TODO deprecate once nothing uses this anymore, use ListGuests() instead
func (c *Client) GetVmList(ctx context.Context) (map[string]interface{}, error) {
	list, err := c.GetResourceList(ctx, resourceListGuest)
	return map[string]interface{}{"data": list}, err
}

func (c *Client) CheckVmRef(ctx context.Context, vmr *VmRef) (err error) {
	if vmr == nil {
		return errors.New(VmRef_Error_Nil)
	}
	if vmr.node == "" || vmr.vmType == "" {
		_, err = c.GetVmInfo(ctx, vmr)
	}
	return
}

func (c *Client) GetVmInfo(ctx context.Context, vmr *VmRef) (vmInfo map[string]interface{}, err error) {
	vms, err := c.GetResourceList(ctx, resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if GuestID(vm["vmid"].(float64)) == vmr.vmId {
			vmInfo = vm
			vmr.node = NodeName(vmInfo["node"].(string))
			vmr.vmType = vmInfo["type"].(string)
			vmr.pool = ""
			if vmInfo["pool"] != nil {
				vmr.pool = PoolName(vmInfo["pool"].(string))
			}
			if vmInfo["hastate"] != nil {
				vmr.haState = vmInfo["hastate"].(string)
			}
			return
		}
	}
	return nil, fmt.Errorf("vm '%d' not found", vmr.vmId)
}

func (c *Client) GetVmRefByName(ctx context.Context, vmName string) (vmr *VmRef, err error) {
	vmrs, err := c.GetVmRefsByName(ctx, vmName)
	if err != nil {
		return nil, err
	}

	return vmrs[0], nil
}

func (c *Client) GetVmRefsByName(ctx context.Context, vmName string) (vmrs []*VmRef, err error) {
	vms, err := c.GetResourceList(ctx, resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if vm["name"] != nil && vm["name"].(string) == vmName {
			vmr := NewVmRef(GuestID(vm["vmid"].(float64)))
			vmr.node = NodeName(vm["node"].(string))
			vmr.vmType = vm["type"].(string)
			vmr.pool = ""
			if vm["pool"] != nil {
				vmr.pool = PoolName(vm["pool"].(string))
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

func (c *Client) GetVmRefById(ctx context.Context, ID GuestID) (vmr *VmRef, err error) {
	var exist bool = false
	vms, err := c.GetResourceList(ctx, resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if int(vm["vmid"].(float64)) != 0 && GuestID(vm["vmid"].(float64)) == ID {
			vmr = NewVmRef(GuestID(vm["vmid"].(float64)))
			vmr.node = NodeName(vm["node"].(string))
			vmr.vmType = vm["type"].(string)
			vmr.pool = ""
			if vm["pool"] != nil {
				vmr.pool = PoolName(vm["pool"].(string))
			}
			if vm["hastate"] != nil {
				vmr.haState = vm["hastate"].(string)
			}
			return
		}
	}
	if !exist {
		return nil, fmt.Errorf("vm 'id-%d' not found", ID)
	} else {
		return
	}
}

func (c *Client) GetVmState(ctx context.Context, vmr *VmRef) (vmState map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return c.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/status/current", "vm", "STATE")
}

func (c *Client) GetVmConfig(ctx context.Context, vmr *VmRef) (vmConfig map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	return c.GetItemConfigMapStringInterface(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/config", "vm", "CONFIG")
}

func (c *Client) GetStorageStatus(ctx context.Context, vmr *VmRef, storageName string) (storageStatus map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/storage/%s/status", vmr.node, storageName)
	err = c.GetJsonRetryable(ctx, url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("storage STATUS not readable")
	}
	storageStatus = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetStorageContent(ctx context.Context, vmr *VmRef, storageName string) (data map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", vmr.node, storageName)
	err = c.GetJsonRetryable(ctx, url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("storage Content not readable")
	}
	return
}

func (c *Client) GetVmSpiceProxy(ctx context.Context, vmr *VmRef) (vmSpiceProxy map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/spiceproxy", vmr.node, vmr.vmType, vmr.vmId)
	_, err = c.session.PostJSON(ctx, url, nil, nil, nil, &data)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, fmt.Errorf("vm SpiceProxy not readable")
	}
	vmSpiceProxy = data["data"].(map[string]interface{})
	return
}

// deprecated use *VmRef.GetAgentInformation() instead
func (c *Client) GetVmAgentNetworkInterfaces(ctx context.Context, vmr *VmRef) ([]AgentNetworkInterface, error) {
	return vmr.GetAgentInformation(ctx, c, true)
}

func (c *Client) CreateTemplate(ctx context.Context, vmr *VmRef) error {
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/template", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, nil)
	if err != nil {
		return err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return err
	}

	exitStatus, err := c.WaitForCompletion(ctx, taskResponse)
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

func (c *Client) MonitorCmd(ctx context.Context, vmr *VmRef, command string) (monitorRes map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(map[string]interface{}{"command": command})
	url := fmt.Sprintf("/nodes/%s/%s/%d/monitor", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err != nil {
		return nil, err
	}
	monitorRes, err = ResponseJSON(resp)
	return
}

func (c *Client) Sendkey(ctx context.Context, vmr *VmRef, qmKey string) error {
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return err
	}
	reqbody := ParamsToBody(map[string]interface{}{"key": qmKey})
	url := fmt.Sprintf("/nodes/%s/%s/%d/sendkey", vmr.node, vmr.vmType, vmr.vmId)
	// No return, even for errors: https://bugzilla.proxmox.com/show_bug.cgi?id=2275
	_, err = c.session.Put(ctx, url, nil, nil, &reqbody)

	return err
}

// WaitForCompletion - poll the API for task completion
// TODO this should be removed in favor of the `Task` interface. But it's only used in things we will deprecate anyway.
func (c *Client) WaitForCompletion(ctx context.Context, taskResponse map[string]interface{}) (waitExitStatus string, err error) {
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
		exitStatus, statErr := c.GetTaskExitstatus(ctx, taskUpid)
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

func (c *Client) GetTaskExitstatus(ctx context.Context, taskUpid string) (exitStatus interface{}, err error) {
	node := rxTaskNode.FindStringSubmatch(taskUpid)[1]
	url := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskUpid)
	var data map[string]interface{}
	_, err = c.session.GetJSON(ctx, url, nil, nil, &data)
	if err == nil {
		exitStatus = data["data"].(map[string]interface{})["exitstatus"]
	}
	if exitStatus != nil && rxExitStatusSuccess.FindString(exitStatus.(string)) == "" {
		err = fmt.Errorf(exitStatus.(string))
	}
	return
}

func (c *Client) StatusChangeVm(ctx context.Context, vmr *VmRef, params map[string]interface{}, setStatus string) (Task, error) {
	err := c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/status/%s", vmr.node, vmr.vmType, vmr.vmId, setStatus)
	return c.postWithTask(ctx, params, url)
}

func (c *Client) StartVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "start")
}

func (c *Client) StopVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "stop")
}

func (c *Client) ShutdownVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "shutdown")
}

func (c *Client) ResetVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "reset")
}

func (c *Client) RebootVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "reboot")
}

func (c *Client) PauseVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "suspend")
}

func (c *Client) HibernateVm(ctx context.Context, vmr *VmRef) (Task, error) {
	params := map[string]interface{}{
		"todisk": true,
	}
	return c.StatusChangeVm(ctx, vmr, params, "suspend")
}

func (c *Client) ResumeVm(ctx context.Context, vmr *VmRef) (Task, error) {
	return c.StatusChangeVm(ctx, vmr, nil, "resume")
}

func (c *Client) DeleteVm(ctx context.Context, vmr *VmRef) (exitStatus string, err error) {
	return c.DeleteVmParams(ctx, vmr, nil)
}

func (c *Client) DeleteVmParams(ctx context.Context, vmr *VmRef, params map[string]interface{}) (exitStatus string, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return "", err
	}

	// Remove HA if required
	if vmr.haState != "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(ctx, url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return "", err
			}
			exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
			if err != nil {
				return "", err
			}
		}
	}

	values := ParamsToValues(params)
	url := fmt.Sprintf("/nodes/%s/%s/%d", vmr.node, vmr.vmType, vmr.vmId)
	var taskResponse map[string]interface{}
	if len(values) != 0 {
		_, err = c.session.RequestJSON(ctx, "DELETE", url, &values, nil, nil, &taskResponse)
	} else {
		_, err = c.session.RequestJSON(ctx, "DELETE", url, nil, nil, nil, &taskResponse)
	}
	if err != nil {
		return
	}
	exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
	return
}

// Deprecated use ConfigQemu.Create() instead
func (c *Client) CreateQemuVm(ctx context.Context, node NodeName, vmParams map[string]interface{}) (exitStatus string, err error) {
	// Create VM disks first to ensure disks names.
	createdDisks, createdDisksErr := c.createVMDisks(ctx, node, vmParams)
	if createdDisksErr != nil {
		return "", createdDisksErr
	}

	// Then create the VM itself.
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu", node)
	var resp *http.Response
	resp, err = c.session.Post(ctx, url, nil, nil, &reqbody)
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
	exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
	// Delete VM disks if the VM didn't create.
	if exitStatus != "OK" {
		deleteDisksErr := c.DeleteVMDisks(ctx, node, createdDisks)
		if deleteDisksErr != nil {
			return "", deleteDisksErr
		}
	}

	return
}

func (c *Client) CreateLxcContainer(ctx context.Context, node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/lxc", node)
	var resp *http.Response
	resp, err = c.session.Post(ctx, url, nil, nil, &reqbody)
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
	exitStatus, err = c.WaitForCompletion(ctx, taskResponse)

	return
}

// Deprecated: use VmRef.CloneLxc() instead
func (c *Client) CloneLxcContainer(ctx context.Context, vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/lxc/%s/clone", vmr.node, vmParams["vmid"])
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// Deprecated: use VmRef.CloneQemu() instead
func (c *Client) CloneQemuVm(ctx context.Context, vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/clone", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED superseded by CreateSnapshot()
func (c *Client) CreateQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	ctx := context.Background()
	err = c.CheckVmRef(ctx, vmr)
	snapshotParams := map[string]interface{}{
		"snapname": snapshotName,
	}
	reqbody := ParamsToBody(snapshotParams)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED superseded by DeleteSnapshot()
func (c *Client) DeleteQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	task, err := DeleteSnapshot(context.Background(), c, vmr, SnapshotName(snapshotName))
	if err != nil {
		return "", err
	}
	err = task.WaitForCompletion()
	return task.ExitStatus(), err
}

// DEPRECATED superseded by ListSnapshots()
func (c *Client) ListQemuSnapshot(vmr *VmRef) (taskResponse map[string]interface{}, exitStatus string, err error) {
	ctx := context.Background()
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Get(ctx, url, nil, nil)
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
	task, err := RollbackSnapshot(context.Background(), c, vmr, SnapshotName(snapshot))
	if err != nil {
		return "", err
	}
	err = task.WaitForCompletion()
	return task.ExitStatus(), err
}

// DEPRECATED SetVmConfig - send config options
func (c *Client) SetVmConfig(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	task, err := c.postWithTask(context.Background(), params, "/nodes/"+vmr.node.String()+"/"+vmr.vmType+"/"+vmr.vmId.String()+"/config")
	if err != nil {
		return nil, err
	}
	err = task.WaitForCompletion()
	return task.ExitStatus(), err
}

// SetLxcConfig - send config options
func (c *Client) SetLxcConfig(ctx context.Context, vmr *VmRef, vmParams map[string]interface{}) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// MigrateNode - Migrate a VM
// Deprecated: use VmRef.Migrate() instead
func (c *Client) MigrateNode(ctx context.Context, vmr *VmRef, newTargetNode NodeName, online bool) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"target": newTargetNode, "online": online, "with-local-disks": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/migrate", vmr.node.String(), vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		return exitStatus, err
	}
	return nil, err
}

// ResizeQemuDisk allows the caller to increase the size of a disk by the indicated number of gigabytes
// TODO Deprecate once LXC is able to resize disk by itself (qemu can already do this)
func (c *Client) ResizeQemuDisk(ctx context.Context, vmr *VmRef, disk string, moreSizeGB int) (exitStatus interface{}, err error) {
	size := fmt.Sprintf("+%dG", moreSizeGB)
	return c.ResizeQemuDiskRaw(ctx, vmr, disk, size)
}

// ResizeQemuDiskRaw allows the caller to provide the raw resize string to be send to proxmox.
// See the proxmox API documentation for full information, but the short version is if you prefix
// your desired size with a '+' character it will ADD size to the disk.  If you just specify the size by
// itself it will do an absolute resizing to the specified size. Permitted suffixes are K, M, G, T
// to indicate order of magnitude (kilobyte, megabyte, etc). Decrease of disk size is not permitted.
// TODO Deprecate once LXC is able to resize disk by itself (qemu can already do this)
func (c *Client) ResizeQemuDiskRaw(ctx context.Context, vmr *VmRef, disk string, size string) (exitStatus interface{}, err error) {
	// PUT
	//disk:virtio0
	// size:+2G
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "size": size})
	url := fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

func (c *Client) MoveLxcDisk(ctx context.Context, vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_volume", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DEPRECATED use MoveQemuDisk() instead.
// MoveQemuDisk - Move a disk from one storage to another
func (c *Client) MoveQemuDisk(vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	ctx := context.Background()
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// MoveQemuDiskToVM - Move a disk to a different VM, using the same storage
func (c *Client) MoveQemuDiskToVM(ctx context.Context, vmrSource *VmRef, disk string, vmrTarget *VmRef) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "target-vmid": vmrTarget.vmId, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmrSource.node, vmrSource.vmType, vmrSource.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// Unlink - Unlink (detach) a set of disks from a VM.
// Reference: https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/qemu/{vmid}/unlink
func (c *Client) Unlink(ctx context.Context, node string, ID GuestID, diskIds string, forceRemoval bool) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/qemu/%d/unlink", node, ID)
	data := ParamsToBody(map[string]interface{}{
		"idlist": diskIds,
		"force":  forceRemoval,
	})
	resp, err := c.session.Put(ctx, url, nil, nil, &data)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	json, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	return c.WaitForCompletion(ctx, json)
}

// GetNextID - Get next free GuestID
func (c *Client) GetNextID(ctx context.Context, currentID *GuestID) (GuestID, error) {
	if currentID != nil {
		if err := currentID.Validate(); err != nil {
			return 0, err
		}
	}
	return c.GetNextIdNoCheck(ctx, currentID)
}

// GetNextIdNoCheck - Get next free GuestID without validating the input
func (c *Client) GetNextIdNoCheck(ctx context.Context, startID *GuestID) (GuestID, error) {
	var url string
	if startID != nil {
		url = "/cluster/nextid?vmid=" + startID.String()
	} else {
		url = "/cluster/nextid"
	}
	tmpID, err := c.GetItemConfigString(ctx, url, "API", "cluster/nextid")
	if err != nil {
		return 0, err
	}
	var id int
	id, err = strconv.Atoi(tmpID)
	return GuestID(id), err
}

// VMIdExists - If you pass an VMID that exists it will return true, otherwise it wil return false
func (c *Client) VMIdExists(ctx context.Context, guestID GuestID) (exists bool, err error) {
	vms, err := c.GetResourceList(ctx, resourceListGuest)
	if err != nil {
		return
	}
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if guestID == GuestID(vm["vmid"].(float64)) {
			return true, err
		}
	}
	return
}

// CreateVMDisk - Create single disk for VM on host node.
func (c *Client) CreateVMDisk(
	ctx context.Context,
	nodeName NodeName,
	storageName string,
	fullDiskName string,
	diskParams map[string]interface{},
) error {
	reqbody := ParamsToBody(diskParams)
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", nodeName, storageName)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
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
	ctx context.Context,
	node NodeName,
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
				err := c.CreateVMDisk(ctx, node, storageName, fullDiskName, diskParams)
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
func (c *Client) CreateNewDisk(ctx context.Context, vmr *VmRef, disk string, volume string) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{disk: volume})
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
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
	ctx context.Context,
	node NodeName,
	disks []string,
) error {
	for _, fullDiskName := range disks {
		storageName, volumeName := getStorageAndVolumeName(fullDiskName, ":")
		url := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storageName, volumeName)
		_, err := c.session.Post(ctx, url, nil, nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// VzDump - Create backup
func (c *Client) VzDump(ctx context.Context, vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/vzdump", vmr.node)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DeleteVolume - Delete volume
func (c *Client) DeleteVolume(ctx context.Context, vmr *VmRef, storageName string, volumeName string) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", vmr.node, storageName, volumeName)
	resp, err := c.session.Delete(ctx, url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// CreateVNCProxy - Creates a TCP VNC proxy connections
func (c *Client) CreateVNCProxy(ctx context.Context, vmr *VmRef, params map[string]interface{}) (vncProxyRes map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
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
func (c *Client) QemuAgentPing(ctx context.Context, vmr *VmRef) (pingRes map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/ping", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, nil)
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
func (c *Client) QemuAgentFileWrite(ctx context.Context, vmr *VmRef, params map[string]interface{}) (err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/file-write", vmr.node, vmr.vmId)
	_, err = c.session.Post(ctx, url, nil, nil, &reqbody)
	return
}

// QemuAgentSetUserPassword - Sets the password for the given user to the given password.
func (c *Client) QemuAgentSetUserPassword(ctx context.Context, vmr *VmRef, params map[string]interface{}) (result map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/set-user-password", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
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
func (c *Client) QemuAgentExec(ctx context.Context, vmr *VmRef, params map[string]interface{}) (result map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/agent/exec", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
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
func (c *Client) GetExecStatus(ctx context.Context, vmr *VmRef, pid string) (status map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	err = c.GetJsonRetryable(ctx, fmt.Sprintf("/nodes/%s/%s/%d/agent/exec-status?pid=%s", vmr.node, vmr.vmType, vmr.vmId, pid), &status, 3)
	if err == nil {
		status = status["data"].(map[string]interface{})
	}
	return
}

// SetQemuFirewallOptions - Set Firewall options.
func (c *Client) SetQemuFirewallOptions(ctx context.Context, vmr *VmRef, fwOptions map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(fwOptions)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// GetQemuFirewallOptions - Get VM firewall options.
func (c *Client) GetQemuFirewallOptions(ctx context.Context, vmr *VmRef) (firewallOptions map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Get(ctx, url, nil, nil)
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
func (c *Client) CreateQemuIPSet(ctx context.Context, vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// AddQemuIPSet - Add IP or Network to IPSet.
func (c *Client) AddQemuIPSet(ctx context.Context, vmr *VmRef, name string, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", vmr.node, vmr.vmId, name)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// GetQemuIPSet - List IPSets
func (c *Client) GetQemuIPSet(ctx context.Context, vmr *VmRef) (ipsets map[string]interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Get(ctx, url, nil, nil)
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
func (c *Client) DeleteQemuIPSet(ctx context.Context, vmr *VmRef, IPSetName string) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s", vmr.node, vmr.vmId, IPSetName)
	resp, err := c.session.Delete(ctx, url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

// DeleteQemuIPSetNetwork - Remove IP or Network from IPSet.
func (c *Client) DeleteQemuIPSetNetwork(ctx context.Context, vmr *VmRef, IPSetName string, network string, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(ctx, vmr)
	if err != nil {
		return nil, err
	}
	values := ParamsToValues(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset/%s/%s", vmr.node, vmr.vmId, IPSetName, network)
	resp, err := c.session.Delete(ctx, url, &values, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return "", err
		}
	}
	return
}

func (c *Client) Upload(ctx context.Context, node string, storage string, contentType string, filename string, file io.Reader) error {
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
	req, err := c.session.NewRequest(ctx, http.MethodPost, url, &headers, body)
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
	exitStatus, err := c.WaitForCompletion(ctx, taskResponse)
	if err != nil {
		return err
	}
	if exitStatus != exitStatusSuccess {
		return fmt.Errorf("moving file to destination failed: %v", exitStatus)
	}
	return nil
}

func (c *Client) UploadLargeFile(ctx context.Context, node string, storage string, contentType string, filename string, filesize int64, file io.Reader) error {
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
	req, err := c.session.NewRequest(ctx, http.MethodPost, url, &headers, body)
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
	exitStatus, err := c.WaitForCompletion(ctx, taskResponse)
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

// Still used by Terraform. Deprecated: use ConfigQemu.Update() instead
func (c *Client) UpdateVMPool(ctx context.Context, vmr *VmRef, pool string) (exitStatus interface{}, err error) {
	// Same pool
	if vmr.pool == PoolName(pool) {
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
		resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(ctx, taskResponse)

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
		resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return
}

func (c *Client) ReadVMHA(ctx context.Context, vmr *VmRef) (err error) {
	var list map[string]interface{}
	url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
	err = c.GetJsonRetryable(ctx, url, &list, 3)
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

func (c *Client) UpdateVMHA(ctx context.Context, vmr *VmRef, haState string, haGroup string) (exitStatus interface{}, err error) {
	// Same hastate & hagroup
	if vmr.haState == haState && vmr.haGroup == haGroup {
		return
	}

	// Remove HA
	if haState == "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(ctx, url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
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
		resp, err := c.session.Post(ctx, "/cluster/ha/resources", nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(ctx, taskResponse)

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
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(ctx, taskResponse)
		if err != nil {
			return nil, err
		}
	}

	return
}

// Still used by Terraform. Deprecated: use ListPoolsWithComments() instead
func (c *Client) GetPoolList(ctx context.Context) (pools map[string]interface{}, err error) {
	return c.GetItemList(ctx, "/pools")
}

// TODO: implement replacement
func (c *Client) GetPoolInfo(ctx context.Context, poolid string) (poolInfo map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface(ctx, "/pools/"+poolid, "pool", "CONFIG")
}

// Deprecated: use ConfigPool.Create() instead
func (c *Client) CreatePool(poolid string, comment string) error {
	return c.Post(context.Background(), map[string]interface{}{
		"poolid":  poolid,
		"comment": comment,
	}, "/pools")
}

// Deprecated: use ConfigPool.Update() instead
func (c *Client) UpdatePoolComment(poolid string, comment string) error {
	return c.Put(context.Background(), map[string]interface{}{
		"poolid":  poolid,
		"comment": comment,
	}, "/pools/"+poolid)
}

// Deprecated: use PoolName.Delete() instead
func (c *Client) DeletePool(poolid string) error {
	return c.Delete(context.Background(), "/pools/"+poolid)
}

// permissions check
func (c *Client) GetUserPermissions(ctx context.Context, id UserID, path string) (permissions []string, err error) {
	existence, err := CheckUserExistence(ctx, id, c)
	if err != nil {
		return nil, err
	}
	if !existence {
		return nil, fmt.Errorf("cannot get user (%s) permissions, the user does not exist", id)
	}
	permlist, err := c.GetItemList(ctx, "/access/permissions?userid="+id.String()+"&path="+path)
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
func (c *Client) GetAcmeDirectoriesUrl(ctx context.Context) (url []string, err error) {
	config, err := c.GetItemConfigInterfaceArray(ctx, "/cluster/acme/directories", "Acme directories", "CONFIG")
	url = make([]string, len(config))
	for i, element := range config {
		url[i] = element.(map[string]interface{})["url"].(string)
	}
	return
}

func (c *Client) GetAcmeTosUrl(ctx context.Context) (url string, err error) {
	return c.GetItemConfigString(ctx, "/cluster/acme/tos", "Acme T.O.S.", "CONFIG")
}

// ACME Account
func (c *Client) GetAcmeAccountList(ctx context.Context) (accounts map[string]interface{}, err error) {
	return c.GetItemList(ctx, "/cluster/acme/account")
}

func (c *Client) GetAcmeAccountConfig(ctx context.Context, id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface(ctx, "/cluster/acme/account/"+id, "acme", "CONFIG")
}

func (c *Client) CreateAcmeAccount(ctx context.Context, params map[string]interface{}) (Task, error) {
	return c.postWithTask(ctx, params, "/cluster/acme/account/")
}

func (c *Client) UpdateAcmeAccountEmails(ctx context.Context, id, emails string) (Task, error) {
	params := map[string]interface{}{
		"contact": emails,
	}
	return c.putWithTask(ctx, params, "/cluster/acme/account/"+id)
}

func (c *Client) DeleteAcmeAccount(ctx context.Context, id string) (Task, error) {
	return c.deleteWithTask(ctx, "/cluster/acme/account/"+id)
}

// ACME Plugin
func (c *Client) GetAcmePluginList(ctx context.Context) (accounts map[string]interface{}, err error) {
	return c.GetItemList(ctx, "/cluster/acme/plugins")
}

func (c *Client) GetAcmePluginConfig(ctx context.Context, id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface(ctx, "/cluster/acme/plugins/"+id, "acme plugin", "CONFIG")
}

func (c *Client) CreateAcmePlugin(ctx context.Context, params map[string]interface{}) error {
	return c.Post(ctx, params, "/cluster/acme/plugins/")
}

func (c *Client) UpdateAcmePlugin(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/cluster/acme/plugins/"+id)
}

func (c *Client) CheckAcmePluginExistence(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetAcmePluginList(ctx)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "plugin", id)
	return
}

func (c *Client) DeleteAcmePlugin(ctx context.Context, id string) (err error) {
	return c.Delete(ctx, "/cluster/acme/plugins/"+id)
}

// Metrics
func (c *Client) GetMetricServerConfig(ctx context.Context, id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface(ctx, "/cluster/metrics/server/"+id, "metrics server", "CONFIG")
}

func (c *Client) GetMetricsServerList(ctx context.Context) (metricServers map[string]interface{}, err error) {
	return c.GetItemList(ctx, "/cluster/metrics/server")
}

func (c *Client) CreateMetricServer(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Post(ctx, params, "/cluster/metrics/server/"+id)
}

func (c *Client) UpdateMetricServer(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/cluster/metrics/server/"+id)
}

func (c *Client) CheckMetricServerExistence(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetMetricsServerList(ctx)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "id", id)
	return
}

func (c *Client) DeleteMetricServer(ctx context.Context, id string) error {
	return c.Delete(ctx, "/cluster/metrics/server/"+id)
}

// storage
func (c *Client) EnableStorage(ctx context.Context, id string) error {
	return c.Put(ctx, map[string]interface{}{
		"disable": false,
	}, "/storage/"+id)
}

func (c *Client) GetStorageList(ctx context.Context) (metricServers map[string]interface{}, err error) {
	return c.GetItemList(ctx, "/storage")
}

func (c *Client) GetStorageConfig(ctx context.Context, id string) (config map[string]interface{}, err error) {
	return c.GetItemConfigMapStringInterface(ctx, "/storage/"+id, "storage", "CONFIG")
}

func (c *Client) CreateStorage(ctx context.Context, params map[string]interface{}) error {
	return c.Post(ctx, params, "/storage")
}

func (c *Client) CheckStorageExistance(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetStorageList(ctx)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "storage", id)
	return
}

func (c *Client) UpdateStorage(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/storage/"+id)
}

func (c *Client) DeleteStorage(ctx context.Context, id string) error {
	return c.Delete(ctx, "/storage/"+id)
}

// Network

// GetNetworkList gets a json encoded list of currently configured network interfaces on the
// passed in node. The typeFilter parameter can be used to filter by interface type. Pass in
// the empty string "" for typeFilter to list all network interfaces on the node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) GetNetworkList(ctx context.Context, node string, typeFilter string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	if typeFilter != "" {
		url += fmt.Sprintf("?type=%s", typeFilter)
	}
	resp, err := c.session.Get(ctx, url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// GetNetworkInterface gets a json encoded object containing the configuration of the network
// interface with the name passed in as iface from the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) GetNetworkInterface(ctx context.Context, node string, iface string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	resp, err := c.session.Get(ctx, url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// CreateNetwork creates a network with the configuration of the passed in parameters
// on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) CreateNetwork(ctx context.Context, node string, params map[string]interface{}) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network", node)
	return c.CreateItemReturnStatus(ctx, params, url)
}

// UpdateNetwork updates the network corresponding to the passed in interface name on the passed
// in node with the configuration in the passed in parameters.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) UpdateNetwork(ctx context.Context, node string, iface string, params map[string]interface{}) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	return c.UpdateItemReturnStatus(ctx, params, url)
}

// DeleteNetwork deletes the network with the passed in iface name on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) DeleteNetwork(ctx context.Context, node string, iface string) (exitStatus string, err error) {
	url := fmt.Sprintf("/nodes/%s/network/%s", node, iface)
	resp, err := c.session.Delete(ctx, url, nil, nil)
	exitStatus = c.HandleTaskError(resp)
	return
}

// ApplyNetwork applies the pending network configuration on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) ApplyNetwork(ctx context.Context, node string) (Task, error) {
	return c.putWithTask(ctx, nil, "/nodes/"+node+"/network")
}

// RevertNetwork reverts the pending network configuration on the passed in node.
// It returns the body from the API response and any HTTP error the API returns.
func (c *Client) RevertNetwork(ctx context.Context, node string) (Task, error) {
	return c.deleteWithTask(ctx, "/nodes/"+node+"/network")
}

// SDN

func (c *Client) ApplySDN(ctx context.Context) (Task, error) {
	return c.putWithTask(ctx, nil, "/cluster/sdn")
}

// GetSDNVNets returns a list of all VNet definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNVNets(ctx context.Context, pending bool) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets?pending=%d", Btoi(pending))
	err = c.GetJsonRetryable(ctx, url, &list, 3)
	return
}

// CheckSDNVNetExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNVNetExistance(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetSDNVNets(ctx, true)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "vnet", id)
	return
}

// GetSDNVNet returns details about the DNS entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNVNet struct.
func (c *Client) GetSDNVNet(ctx context.Context, name string) (dns map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s", name)
	err = c.GetJsonRetryable(ctx, url, &dns, 3)
	return
}

// CreateSDNVNet creates a new SDN DNS in the cluster
func (c *Client) CreateSDNVNet(ctx context.Context, params map[string]interface{}) error {
	return c.Post(ctx, params, "/cluster/sdn/vnets")
}

// DeleteSDNVNet deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNVNet(ctx context.Context, name string) error {
	return c.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s", name))
}

// UpdateSDNVNet updates the given DNS with the provided parameters
func (c *Client) UpdateSDNVNet(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/cluster/sdn/vnets/"+id)
}

// GetSDNSubnets returns a list of all Subnet definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNSubnets(ctx context.Context, vnet string) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", vnet)
	err = c.GetJsonRetryable(ctx, url, &list, 3)
	return
}

// CheckSDNSubnetExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNSubnetExistance(ctx context.Context, vnet, id string) (existance bool, err error) {
	list, err := c.GetSDNSubnets(ctx, vnet)
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "subnet", id)
	return
}

// GetSDNSubnet returns details about the Subnet entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned map["data"] section can be unmarshalled into a ConfigSDNSubnet struct.
func (c *Client) GetSDNSubnet(ctx context.Context, vnet, name string) (subnet map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, name)
	err = c.GetJsonRetryable(ctx, url, &subnet, 3)
	return
}

// CreateSDNSubnet creates a new SDN DNS in the cluster
func (c *Client) CreateSDNSubnet(ctx context.Context, vnet string, params map[string]interface{}) error {
	return c.Post(ctx, params, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets", vnet))
}

// DeleteSDNSubnet deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNSubnet(ctx context.Context, vnet, name string) error {
	return c.Delete(ctx, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, name))
}

// UpdateSDNSubnet updates the given DNS with the provided parameters
func (c *Client) UpdateSDNSubnet(ctx context.Context, vnet, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, fmt.Sprintf("/cluster/sdn/vnets/%s/subnets/%s", vnet, id))
}

// GetSDNDNSs returns a list of all DNS definitions in the "data" element of the returned
// map.
func (c *Client) GetSDNDNSs(ctx context.Context, typeFilter string) (list map[string]interface{}, err error) {
	url := "/cluster/sdn/dns"
	if typeFilter != "" {
		url += fmt.Sprintf("&type=%s", typeFilter)
	}
	err = c.GetJsonRetryable(ctx, url, &list, 3)
	return
}

// CheckSDNDNSExistance returns true if a DNS entry with the provided ID exists, false otherwise.
func (c *Client) CheckSDNDNSExistance(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetSDNDNSs(ctx, "")
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "dns", id)
	return
}

// GetSDNDNS returns details about the DNS entry whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNDNS struct.
func (c *Client) GetSDNDNS(ctx context.Context, name string) (dns map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/dns/%s", name)
	err = c.GetJsonRetryable(ctx, url, &dns, 3)
	return
}

// CreateSDNDNS creates a new SDN DNS in the cluster
func (c *Client) CreateSDNDNS(ctx context.Context, params map[string]interface{}) error {
	return c.Post(ctx, params, "/cluster/sdn/dns")
}

// DeleteSDNDNS deletes an existing SDN DNS in the cluster
func (c *Client) DeleteSDNDNS(ctx context.Context, name string) error {
	return c.Delete(ctx, fmt.Sprintf("/cluster/sdn/dns/%s", name))
}

// UpdateSDNDNS updates the given DNS with the provided parameters
func (c *Client) UpdateSDNDNS(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/cluster/sdn/dns/"+id)
}

// GetSDNZones returns a list of all the SDN zones defined in the cluster.
func (c *Client) GetSDNZones(ctx context.Context, pending bool, typeFilter string) (list map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/zones?pending=%d", Btoi(pending))
	if typeFilter != "" {
		url += fmt.Sprintf("&type=%s", typeFilter)
	}
	err = c.GetJsonRetryable(ctx, url, &list, 3)
	return
}

// CheckSDNZoneExistance returns true if a zone with the provided ID exists, false otherwise.
func (c *Client) CheckSDNZoneExistance(ctx context.Context, id string) (existance bool, err error) {
	list, err := c.GetSDNZones(ctx, true, "")
	existance = ItemInKeyOfArray(list["data"].([]interface{}), "zone", id)
	return
}

// GetSDNZone returns details about the zone whose name was provided.
// An error is returned if the zone doesn't exist.
// The returned zone can be unmarshalled into a ConfigSDNZone struct.
func (c *Client) GetSDNZone(ctx context.Context, zoneName string) (zone map[string]interface{}, err error) {
	url := fmt.Sprintf("/cluster/sdn/zones/%s", zoneName)
	err = c.GetJsonRetryable(ctx, url, &zone, 3)
	return
}

// CreateSDNZone creates a new SDN zone in the cluster
func (c *Client) CreateSDNZone(ctx context.Context, params map[string]interface{}) error {
	return c.Post(ctx, params, "/cluster/sdn/zones")
}

// DeleteSDNZone deletes an existing SDN zone in the cluster
func (c *Client) DeleteSDNZone(ctx context.Context, zoneName string) error {
	return c.Delete(ctx, fmt.Sprintf("/cluster/sdn/zones/%s", zoneName))
}

// UpdateSDNZone updates the given zone with the provided parameters
func (c *Client) UpdateSDNZone(ctx context.Context, id string, params map[string]interface{}) error {
	return c.Put(ctx, params, "/cluster/sdn/zones/"+id)
}

// Shared
func (c *Client) GetItemConfigMapStringInterface(ctx context.Context, url, text, message string, errorString ...string) (map[string]interface{}, error) {
	data, err := c.GetItemConfig(ctx, url, text, message, errorString...)
	if err != nil {
		return nil, err
	}
	return data["data"].(map[string]interface{}), err
}

func (c *Client) GetItemConfigString(ctx context.Context, url, text, message string) (string, error) {
	data, err := c.GetItemConfig(ctx, url, text, message)
	if err != nil {
		return "", err
	}
	return data["data"].(string), err
}

func (c *Client) GetItemConfigInterfaceArray(ctx context.Context, url, text, message string) ([]interface{}, error) {
	data, err := c.GetItemConfig(ctx, url, text, message)
	if err != nil {
		return nil, err
	}
	return data["data"].([]interface{}), err
}

func (c *Client) GetItemConfig(ctx context.Context, url, text, message string, errorString ...string) (config map[string]interface{}, err error) {
	err = c.GetJsonRetryable(ctx, url, &config, 3, errorString...)
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
func (c *Client) Post(ctx context.Context, Params map[string]interface{}, url string) (err error) {
	reqbody := ParamsToBody(Params)
	_, err = c.session.Post(ctx, url, nil, nil, &reqbody)
	return
}

// CreateItemReturnStatus creates an item on the Proxmox API.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *Client) CreateItemReturnStatus(ctx context.Context, params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(params)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	exitStatus = c.HandleTaskError(resp)
	return
}

// Makes a POST request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
// Deprecated
func (c *Client) PostWithTask(ctx context.Context, Params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(Params)
	var resp *http.Response
	resp, err = c.session.Post(ctx, url, nil, nil, &reqbody)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(ctx, resp)
}

// Makes a POST request and returns a Task to await.
func (c *Client) postWithTask(ctx context.Context, Params map[string]interface{}, url string) (Task, error) {
	reqbody := ParamsToBody(Params)
	resp, err := c.session.Post(ctx, url, nil, nil, &reqbody)
	if err != nil {
		c.HandleTaskError(resp)
		return nil, err
	}
	return c.taskResponse(ctx, resp)
}

// Makes a PUT request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *Client) Put(ctx context.Context, Params map[string]interface{}, url string) (err error) {
	reqbody := ParamsToBodyWithAllEmpty(Params)
	_, err = c.session.Put(ctx, url, nil, nil, &reqbody)
	return
}

// UpdateItemReturnStatus updates an item on the Proxmox API.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *Client) UpdateItemReturnStatus(ctx context.Context, params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBody(params)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	exitStatus = c.HandleTaskError(resp)
	return
}

// Makes a PUT request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
// Deprecated
func (c *Client) PutWithTask(ctx context.Context, Params map[string]interface{}, url string) (exitStatus string, err error) {
	reqbody := ParamsToBodyWithAllEmpty(Params)
	var resp *http.Response
	resp, err = c.session.Put(ctx, url, nil, nil, &reqbody)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(ctx, resp)
}

// Makes a PUT request and returns a Task to await.
func (c *Client) putWithTask(ctx context.Context, Params map[string]interface{}, url string) (Task, error) {
	reqbody := ParamsToBody(Params)
	resp, err := c.session.Put(ctx, url, nil, nil, &reqbody)
	if err != nil {
		c.HandleTaskError(resp)
		return nil, err
	}
	return c.taskResponse(ctx, resp)
}

// Makes a DELETE request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *Client) Delete(ctx context.Context, url string) (err error) {
	_, err = c.session.Delete(ctx, url, nil, nil)
	return
}

// Makes a DELETE request and waits on proxmox for the task to complete.
// It returns the status of the test as 'exitStatus' and the HTTP error as 'err'.
// Deprecated
func (c *Client) DeleteWithTask(ctx context.Context, url string) (exitStatus string, err error) {
	var resp *http.Response
	resp, err = c.session.Delete(ctx, url, nil, nil)
	if err != nil {
		return c.HandleTaskError(resp), err
	}
	return c.CheckTask(ctx, resp)
}

// Makes a DELETE request and returns a Task to await.
func (c *Client) deleteWithTask(ctx context.Context, url string) (Task, error) {
	resp, err := c.session.Delete(ctx, url, nil, nil)
	if err != nil {
		c.HandleTaskError(resp)
		return nil, err
	}
	return c.taskResponse(ctx, resp)
}

func (c *Client) GetItemListInterfaceArray(ctx context.Context, url string) ([]interface{}, error) {
	list, err := c.GetItemList(ctx, url)
	if err != nil {
		return nil, err
	}
	data, ok := list["data"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to cast response to list, resp: %v", list)
	}
	return data, nil
}

func (c *Client) GetItemList(ctx context.Context, url string) (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable(ctx, url, &list, 3)
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
func (c *Client) CheckTask(ctx context.Context, resp *http.Response) (exitStatus string, err error) {
	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	return c.WaitForCompletion(ctx, taskResponse)
}

// return a list of requested permissions from the cache for further processing
func (c *Client) cachedPermissions(ctx context.Context, paths []permissionPath) (map[permissionPath]privileges, error) {
	c.permissionMutex.Lock()
	defer c.permissionMutex.Unlock()
	if c.permissions == nil {
		permissionMap, err := c.getPermissions(ctx)
		if err != nil {
			return nil, err
		}
		c.permissions = permissionMap
	}
	extractedPermissions := make(map[permissionPath]privileges)
	for _, path := range paths {
		if permission, ok := c.permissions[path]; ok {
			extractedPermissions[path] = permission
		}
	}
	return extractedPermissions, nil
}

// Returns an error if the user does not have the required permissions on the given category and itme.
func (c *Client) CheckPermissions(ctx context.Context, perms []Permission) error {
	for _, perm := range perms {
		if err := perm.Validate(); err != nil {
			return err
		}
	}
	return c.checkPermissions(ctx, perms)
}

// internal function to check permissions, does not validate input.
func (c *Client) checkPermissions(ctx context.Context, perms []Permission) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if c.Username == "root@pam" { // no permissions check for root
		return nil
	}
	permissions, err := c.cachedPermissions(ctx, Permission{}.buildPathList(perms))
	if err != nil {
		return err
	}
	for _, perm := range perms {
		err = perm.check(permissions)
		if err != nil {
			return err
		}
	}
	return nil
}

// inserts a permission into the cache, this is useful for when we create an item, as refreshing the whole cache is quite expensive.
func (c *Client) insertCachedPermission(ctx context.Context, path permissionPath) error {
	rawPermissions, err := c.getPermissionsRaw(ctx)
	if err != nil {
		return err
	}
	if rawPrivileges, ok := rawPermissions[string(path)]; ok {
		privileges := privileges{}.mapToSDK(rawPrivileges.(map[string]interface{}))
		c.permissionMutex.Lock()
		c.permissions[path] = privileges
		c.permissionMutex.Unlock()
		return nil
	}
	return nil
}

// get the users permissions from the cache and decodes them for the SDK
func (c *Client) getPermissions(ctx context.Context) (map[permissionPath]privileges, error) {
	permissions, err := c.getPermissionsRaw(ctx)
	if err != nil {
		return nil, err
	}
	return permissionPath("").mapToSDK(permissions), nil
}

// returns the raw permissions from the API
func (c *Client) getPermissionsRaw(ctx context.Context) (map[string]interface{}, error) {
	return c.GetItemConfigMapStringInterface(ctx, "/access/permissions", "", "permissions")
}

// RefreshPermissions fetches the permissions from the API and updates the cache.
func (c *Client) RefreshPermissions(ctx context.Context) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	tmpPermsissions, err := c.getPermissions(ctx)
	if err != nil {
		return err
	}
	c.permissionMutex.Lock()
	c.permissions = tmpPermsissions
	c.permissionMutex.Unlock()
	return nil
}

// Returns the Client's cached version if it exists, otherwise fetches the version from the API.
func (c *Client) Version(ctx context.Context) (Version, error) {
	if c == nil {
		return Version{}, errors.New(Client_Error_Nil)
	}
	if c.version == nil {
		return c.GetVersion(ctx)
	}
	c.versionMutex.Lock()
	defer c.versionMutex.Unlock()
	return Version{
		Major: c.version.Major,
		Minor: c.version.Minor,
		Patch: c.version.Patch,
	}, nil
}

type Version struct {
	Major uint8
	Minor uint8
	Patch uint8
}

// Greater returns true if the version is greater than the other version.
func (v Version) Greater(other Version) bool {
	return uint32(v.Major)*256*256+uint32(v.Minor)*256+uint32(v.Patch) > uint32(other.Major)*256*256+uint32(other.Minor)*256+uint32(other.Patch)
}

func (Version) mapToSDK(params map[string]interface{}) (version Version) {
	if itemValue, isSet := params["version"]; isSet {
		rawVersion := strings.Split(itemValue.(string), ".")
		if len(rawVersion) > 0 {
			tmpMajor, _ := strconv.ParseUint(rawVersion[0], 10, 8)
			version.Major = uint8(tmpMajor)
		}
		if len(rawVersion) > 1 {
			tmpMinor, _ := strconv.ParseUint(rawVersion[1], 10, 8)
			version.Minor = uint8(tmpMinor)
		}
		if len(rawVersion) > 2 {
			tmpPatch, _ := strconv.ParseUint(rawVersion[2], 10, 8)
			version.Patch = uint8(tmpPatch)
		}
	}
	return
}

// return the maximum version, used during testing
func (version Version) max() Version {
	newVersion := Version{
		Major: 255,
		Minor: 255,
		Patch: 255,
	}
	if version.Major != 0 {
		newVersion.Major = version.Major
	}
	if version.Minor != 0 {
		newVersion.Minor = version.Minor
	}
	if version.Patch != 0 {
		newVersion.Patch = version.Patch
	}
	return newVersion
}

// Smaller returns true if the version is less than the other version.
func (v Version) Smaller(other Version) bool {
	return uint32(v.Major)*256*256+uint32(v.Minor)*256+uint32(v.Patch) < uint32(other.Major)*256*256+uint32(other.Minor)*256+uint32(other.Patch)
}

func (v Version) String() string {
	return strconv.FormatInt(int64(v.Major), 10) + "." + strconv.FormatInt(int64(v.Minor), 10) + "." + strconv.FormatInt(int64(v.Patch), 10)
}
