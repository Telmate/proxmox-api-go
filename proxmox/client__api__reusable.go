package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Reusable low-level API methods

func (c *clientAPI) getResourceList(ctx context.Context, resourceType string) ([]any, error) {
	url := "/cluster/resources"
	if resourceType != "" {
		url = url + "?type=" + resourceType
	}
	return c.getList(ctx, url, "", "", nil)
}

// Primitive methods

func (c *clientAPI) getMap(ctx context.Context, url, text, message string, ignore errorIgnore) (map[string]any, error) {
	data, err := c.getRootMap(ctx, url, text, message, ignore)
	if err != nil {
		return nil, err
	}
	return data["data"].(map[string]any), err
}

func (c *clientAPI) getList(ctx context.Context, url, text, message string, ignore errorIgnore) ([]any, error) {
	list, err := c.getRootList(ctx, url, text, message, ignore)
	if err != nil {
		return nil, err
	}
	data, ok := list["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to list, resp: %v", list)
	}
	return data, nil
}

func (c *clientAPI) getRootMap(ctx context.Context, url, text, message string, ignore errorIgnore) (map[string]any, error) {
	var config map[string]any
	if err := c.getJsonRetry(ctx, url, &config, 3, ignore); err != nil {
		return nil, err
	}
	if config["data"] == nil {
		return nil, errors.New(text + " " + message + " not readable")
	}
	return config, nil
}

func (c *clientAPI) getRootList(ctx context.Context, url, text, message string, ignore errorIgnore) (map[string]any, error) {
	var data map[string]any
	if err := c.getJsonRetry(ctx, url, &data, 3, ignore); err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New(text + " " + message + " not readable")
	}
	return data, nil
}

// Makes a POST request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *clientAPI) post(ctx context.Context, url string, params map[string]any) (err error) {
	requestBody := paramsToBody(params)
	_, err = c.session.post(ctx, url, nil, nil, &requestBody)
	return
}

func (c *clientAPI) postTask(ctx context.Context, url string, params map[string]any) (exitStatus string, err error) {
	requestBody := paramsToBody(params)
	var resp *http.Response
	resp, err = c.session.post(ctx, url, nil, nil, &requestBody)
	if err != nil {
		return c.handleTaskError(resp), err
	}
	return c.checkTask(ctx, resp)
}

// handleTaskError reads the body from the passed in HTTP response and closes it.
// It returns the body of the passed in HTTP response.
func (c *clientAPI) handleTaskError(resp *http.Response) (exitStatus string) {
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

// checkTask polls the API to check if the Proxmox task has been completed.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *clientAPI) checkTask(ctx context.Context, resp *http.Response) (exitStatus string, err error) {
	taskResponse, err := responseJSON(resp)
	if err != nil {
		return "", err
	}
	return c.waitForCompletion(ctx, taskResponse)
}

// waitForCompletion - poll the API for task completion
func (c *clientAPI) waitForCompletion(ctx context.Context, taskResponse map[string]any) (string, error) {
	if taskResponse["errors"] != nil {
		errJSON, _ := json.MarshalIndent(taskResponse["errors"], "", "  ")
		return string(errJSON), fmt.Errorf("error response")
	}
	if taskResponse["data"] == nil {
		return "", nil
	}
	waited := time.Duration(0)
	taskUpid := taskResponse["data"].(string)
	for waited < c.taskTimeout {
		exitStatus, err := c.getTaskExitStatus(ctx, taskUpid)
		if err != nil {
			if err != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
				return "", err
			}
		}
		if exitStatus != nil {
			return exitStatus.(string), nil
		}
		time.Sleep(TaskStatusCheckInterval * time.Second)
		waited = waited + TaskStatusCheckInterval
	}
	return "", fmt.Errorf("Wait timeout for:" + taskUpid)
}

func (c *clientAPI) getTaskExitStatus(ctx context.Context, taskUpID string) (exitStatus any, err error) {
	node := rxTaskNode.FindStringSubmatch(taskUpID)[1]
	url := "/nodes/" + node + "/tasks/" + taskUpID + "/status"
	var data map[string]any
	_, err = c.session.getJSON(ctx, url, nil, nil, &data)
	if err == nil {
		exitStatus = data["data"].(map[string]any)["exitstatus"]
	}
	if exitStatus != nil && rxExitStatusSuccess.FindString(exitStatus.(string)) == "" {
		err = fmt.Errorf(exitStatus.(string))
	}
	return
}

func (c *clientAPI) getJsonRetry(ctx context.Context, url string, data *map[string]any, tries int, ignore errorIgnore) error {
	var err error
	for i := range time.Duration(tries) {
		_, err = c.session.getJSON(ctx, url, nil, nil, data)
		if err == nil {
			return nil
		}
		if ignore != nil && ignore(err) {
			return err
		}
		time.Sleep((i + 1) * time.Second)
	}
	return err
}
