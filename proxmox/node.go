package proxmox

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) nodeStatusCommand(node, command string) (exitStatus string, err error) {
	nodes, err := c.GetNodeList()
	if err != nil {
		return
	}

	nodeFound := false
	// nodes contains a key named "data" which is a slice of nodes
	// the list of nodes is a list of map[string]interface{}
	for _, n := range nodes["data"].([]interface{}) {
		if n.(map[string]interface{})["node"].(string) == node {
			nodeFound = true
			break
		}
	}

	if !nodeFound {
		err = fmt.Errorf("Node %s not found", node)
		return
	}

	reqbody := ParamsToBody(map[string]interface{}{"command": command})
	url := fmt.Sprintf("/nodes/%s/status", node)

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

	return
}

func (c *Client) ShutdownNode(node string) (exitStatus string, err error) {
	return c.nodeStatusCommand(node, "shutdown")
}

func (c *Client) RebootNode(node string) (exitStatus string, err error) {
	return c.nodeStatusCommand(node, "reboot")
}
