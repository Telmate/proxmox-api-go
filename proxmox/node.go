package proxmox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Only the following characters are allowed: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-".
// May not start with a hyphen.
// May not end with a hyphen.
// Must contain at least one alphabetical character.
// Max length 63 characters.
type NodeName string

const (
	NodeName_Error_Alphabetical string = "Node name must contain at least one alphabetical character"
	NodeName_Error_Empty        string = "Node name cannot be empty"
	NodeName_Error_HyphenEnd    string = "Node name cannot end with a hyphen"
	NodeName_Error_HyphenStart  string = "Node name cannot start with a hyphen"
	NodeName_Error_Illegal      string = "Node name may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	NodeName_Error_Length       string = "Node name must be less than 64 characters"
)

func (name NodeName) Validate() error {
	if name == "" {
		return errors.New(NodeName_Error_Empty)
	}
	if len(name) > 63 {
		return errors.New(NodeName_Error_Length)
	}
	if name[0] == '-' {
		return errors.New(NodeName_Error_HyphenStart)
	}
	if name[len(name)-1] == '-' {
		return errors.New(NodeName_Error_HyphenEnd)
	}
	var hasAlpha bool
	for i := range name {
		if (name[i] >= 'a' && name[i] <= 'z') || (name[i] >= 'A' && name[i] <= 'Z') {
			hasAlpha = true
			break
		}
	}
	if !hasAlpha {
		return errors.New(NodeName_Error_Alphabetical)
	}
	for i := range name {
		if !((name[i] >= 'a' && name[i] <= 'z') || (name[i] >= 'A' && name[i] <= 'Z') || (name[i] >= '0' && name[i] <= '9') || name[i] == '-') {
			return errors.New(NodeName_Error_Illegal)
		}
	}
	return nil
}

func (name NodeName) String() string {
	return string(name)
}

func (c *Client) nodeStatusCommand(ctx context.Context, node, command string) (exitStatus string, err error) {
	nodes, err := c.GetNodeList(ctx)
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
	resp, err = c.session.Post(ctx, url, nil, nil, &reqbody)
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

func (c *Client) ShutdownNode(ctx context.Context, node string) (exitStatus string, err error) {
	return c.nodeStatusCommand(ctx, node, "shutdown")
}

func (c *Client) RebootNode(ctx context.Context, node string) (exitStatus string, err error) {
	return c.nodeStatusCommand(ctx, node, "reboot")
}
