package proxmox

// inspired by https://github.com/Telmate/vagrant-proxmox/blob/master/lib/vagrant-proxmox/proxmox/connection.rb

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// TaskTimeout - default async task call timeout in seconds
const TaskTimeout = 90

// TaskStatusCheckInterval - time between async checks in seconds
const TaskStatusCheckInterval = 2

// Client - URL, user and password to specifc Proxmox node
type Client struct {
	session  *Session
	ApiUrl   string
	Username string
	Password string
}

// VmRef - virtual machine ref parts
// map[type:qemu node:proxmox1-xx id:qemu/132 diskread:5.57424738e+08 disk:0 netin:5.9297450593e+10 mem:3.3235968e+09 uptime:1.4567097e+07 vmid:132 template:0 maxcpu:2 netout:6.053310416e+09 maxdisk:3.4359738368e+10 maxmem:8.592031744e+09 diskwrite:1.49663619584e+12 status:running cpu:0.00386980694947209 name:appt-app1-dev.xxx.xx]
type VmRef struct {
	vmId   int
	node   string
	vmType string
}

func (vmr *VmRef) SetNode(node string) {
	vmr.node = node
	return
}

func (vmr *VmRef) SetVmType(vmType string) {
	vmr.vmType = vmType
	return
}

func (vmr *VmRef) VmId() int {
	return vmr.vmId
}

func (vmr *VmRef) Node() string {
	return vmr.node
}

func NewVmRef(vmId int) (vmr *VmRef) {
	vmr = &VmRef{vmId: vmId, node: "", vmType: ""}
	return
}

func NewClient(apiUrl string, hclient *http.Client, tls *tls.Config) (client *Client, err error) {
	var sess *Session
	sess, err = NewSession(apiUrl, hclient, tls)
	if err == nil {
		client = &Client{session: sess, ApiUrl: apiUrl}
	}
	return client, err
}

func (c *Client) Login(username string, password string) (err error) {
	c.Username = username
	c.Password = password
	return c.session.Login(username, password)
}

func (c *Client) GetJsonRetryable(url string, data *map[string]interface{}, tries int) error {
	var statErr error
	for ii := 0; ii < tries; ii++ {
		_, statErr = c.session.GetJSON(url, nil, nil, data)
		if statErr == nil {
			return nil
		}
		// if statErr != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
		//   return statErr
		// }
		time.Sleep(5 * time.Second)
	}
	return statErr
}

func (c *Client) GetNodeList() (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable("/nodes", &list, 3)
	return
}

func (c *Client) GetVmList() (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable("/cluster/resources?type=vm", &list, 3)
	return
}

func (c *Client) CheckVmRef(vmr *VmRef) (err error) {
	if vmr.node == "" || vmr.vmType == "" {
		_, err = c.GetVmInfo(vmr)
	}
	return
}

func (c *Client) GetVmInfo(vmr *VmRef) (vmInfo map[string]interface{}, err error) {
	resp, err := c.GetVmList()
	vms := resp["data"].([]interface{})
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if int(vm["vmid"].(float64)) == vmr.vmId {
			vmInfo = vm
			vmr.node = vmInfo["node"].(string)
			vmr.vmType = vmInfo["type"].(string)
			return
		}
	}
	return nil, errors.New(fmt.Sprintf("Vm '%d' not found", vmr.vmId))
}

func (c *Client) GetVmRefByName(vmName string) (vmr *VmRef, err error) {
	resp, err := c.GetVmList()
	vms := resp["data"].([]interface{})
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if vm["name"] != nil && vm["name"].(string) == vmName {
			vmr = NewVmRef(int(vm["vmid"].(float64)))
			vmr.node = vm["node"].(string)
			vmr.vmType = vm["type"].(string)
			return
		}
	}
	return nil, errors.New(fmt.Sprintf("Vm '%s' not found", vmName))
}

func (c *Client) GetVmState(vmr *VmRef) (vmState map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/status/current", vmr.node, vmr.vmType, vmr.vmId)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Vm STATE not readable")
	}
	vmState = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetVmConfig(vmr *VmRef) (vmConfig map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Vm CONFIG not readable")
	}
	vmConfig = data["data"].(map[string]interface{})
	return
}

func (c *Client) MonitorCmd(vmr *VmRef, command string) (monitorRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(map[string]interface{}{"command": command})
	url := fmt.Sprintf("/nodes/%s/%s/%d/monitor", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	monitorRes = ResponseJSON(resp)
	return
}

// WaitForCompletion - poll the API for task completion
func (c *Client) WaitForCompletion(taskResponse map[string]interface{}) (waitExitStatus string, err error) {
	if taskResponse["errors"] != nil {
		errJSON, _ := json.MarshalIndent(taskResponse["errors"], "", "  ")
		return string(errJSON), errors.New("Error reponse")
	}
	if taskResponse["data"] == nil {
		return "", nil
	}
	waited := 0
	taskUpid := taskResponse["data"].(string)
	for waited < TaskTimeout {
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
	return "", errors.New("Wait timeout for:" + taskUpid)
}

var rxTaskNode = regexp.MustCompile("UPID:(.*?):")

func (c *Client) GetTaskExitstatus(taskUpid string) (exitStatus interface{}, err error) {
	node := rxTaskNode.FindStringSubmatch(taskUpid)[1]
	url := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskUpid)
	var data map[string]interface{}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		exitStatus = data["data"].(map[string]interface{})["exitstatus"]
	}
	return
}

func (c *Client) StatusChangeVm(vmr *VmRef, setStatus string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/status/%s", vmr.node, vmr.vmType, vmr.vmId, setStatus)
	var taskResponse map[string]interface{}
	for i := 0; i < 3; i++ {
		_, err = c.session.PostJSON(url, nil, nil, nil, &taskResponse)
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if exitStatus == "" {
			time.Sleep(TaskStatusCheckInterval * time.Second)
		} else {
			return
		}
	}
	return
}

func (c *Client) StartVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "start")
}

func (c *Client) StopVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "stop")
}

func (c *Client) ShutdownVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "shutdown")
}

func (c *Client) DeleteVm(vmr *VmRef) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d", vmr.node, vmr.vmType, vmr.vmId)
	var taskResponse map[string]interface{}
	_, err = c.session.RequestJSON("DELETE", url, nil, nil, nil, &taskResponse)
	exitStatus, err = c.WaitForCompletion(taskResponse)
	return
}

func (c *Client) CreateQemuVm(node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu", node)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse := ResponseJSON(resp)
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) CloneQemuVm(vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/clone", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse := ResponseJSON(resp)
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) RollbackQemuVm(vmr *VmRef, snapshot string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/%s/rollback", vmr.node, vmr.vmType, vmr.vmId, snapshot)
	var taskResponse map[string]interface{}
	_, err = c.session.PostJSON(url, nil, nil, nil, &taskResponse)
	exitStatus, err = c.WaitForCompletion(taskResponse)
	return
}

// SetVmConfig - send config options
func (c *Client) SetVmConfig(vmr *VmRef, vmParams map[string]interface{}) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse := ResponseJSON(resp)
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) ResizeQemuDisk(vmr *VmRef, disk string, moreSizeGB int) (exitStatus interface{}, err error) {
	// PUT
	//disk:virtio0
	//size:+2G
	if disk == "" {
		disk = "virtio0"
	}
	size := fmt.Sprintf("+%dG", moreSizeGB)
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "size": size})
	url := fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse := ResponseJSON(resp)
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// GetNextID - Get next free VMID
func (c *Client) GetNextID(currentID int) (nextID int, err error) {
	var data map[string]interface{}
	var url string
	if currentID > 0 {
		url = fmt.Sprintf("/cluster/nextid?vmid=%d", currentID)
	} else {
		url = "/cluster/nextid"
	}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		if data["errors"] != nil {
			if currentID != 0 {
				return c.GetNextID(0)
			} else {
				return -1, errors.New("error using /cluster/nextid")
			}
		}
		nextID, err = strconv.Atoi(data["data"].(string))
	}
	return
}
