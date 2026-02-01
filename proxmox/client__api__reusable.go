package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Reusable low-level API methods

const RequestRetryCount = 3

func (c *clientAPI) getResourceList(ctx context.Context, resourceType string) ([]any, error) {
	url := "/cluster/resources"
	if resourceType != "" {
		url = url + "?type=" + resourceType
	}
	return c.getList(ctx, url, "", "")
}

// Primitive methods

// Makes a DELETE request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *clientAPI) delete(ctx context.Context, url string) (retry bool, err error) {
	_, retry, err = c.session.delete(ctx, url, nil, nil)
	return
}

func (c *clientAPI) deleteRetry(ctx context.Context, url string, tries int) (err error) {
	var retry bool
	for i := range time.Duration(tries) {
		_, retry, err = c.session.delete(ctx, url, nil, nil)
		if err == nil {
			return nil
		}
		if !retry {
			return
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	return
}

func (c *clientAPI) deleteTask(ctx context.Context, url string) error {
	var response *http.Response
	var retry bool
	var err error
	for i := range time.Duration(RequestRetryCount) {
		response, retry, err = c.session.delete(ctx, url, nil, nil)
		if err == nil || !retry {
			break
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	if err != nil {
		return err
	}
	return c.checkTask(ctx, response)
}

func (c *clientAPI) getMap(ctx context.Context, url, text, message string) (map[string]any, error) {
	data, err := c.getRootMap(ctx, url, text, message)
	if err != nil {
		return nil, err
	}
	return data["data"].(map[string]any), err
}

func (c *clientAPI) getList(ctx context.Context, url, text, message string) ([]any, error) {
	list, err := c.getRootList(ctx, url, text, message)
	if err != nil {
		return nil, err
	}
	data, ok := list["data"].([]any)
	if !ok {
		return nil, fmt.Errorf("failed to cast response to list, resp: %v", list)
	}
	return data, nil
}

func (c *clientAPI) getRootMap(ctx context.Context, url, text, message string) (map[string]any, error) {
	var config map[string]any
	if err := c.getJsonRetry(ctx, url, &config, 3); err != nil {
		return nil, err
	}
	if config["data"] == nil {
		return nil, errors.New(text + " " + message + " not readable")
	}
	return config, nil
}

func (c *clientAPI) getRootList(ctx context.Context, url, text, message string) (map[string]any, error) {
	var data map[string]any
	if err := c.getJsonRetry(ctx, url, &data, 3); err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New(text + " " + message + " not readable")
	}
	return data, nil
}

// Makes a POST request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *clientAPI) post(ctx context.Context, url string, params map[string]any) (retry bool, err error) {
	requestBody := paramsToBody(params)
	_, retry, err = c.session.post(ctx, url, nil, nil, &requestBody)
	return
}

func (c *clientAPI) postMap(ctx context.Context, url string, body *[]byte, text, message string) (map[string]any, error) {
	data, err := c.postRootMap(ctx, url, body, text, message)
	if err != nil {
		return nil, err
	}
	return data["data"].(map[string]any), err
}

func (c *clientAPI) postRootMap(ctx context.Context, url string, body *[]byte, text, message string) (map[string]any, error) {
	config, err := c.postJsonRetry(ctx, url, body, 3)
	if err != nil {
		return nil, err
	}
	if config["data"] == nil {
		return nil, errors.New(text + " " + message + " not readable")
	}
	return config, nil
}

func (c *clientAPI) postRawRetry(ctx context.Context, url string, body *[]byte, tries int) (err error) {
	var retry bool
	for i := range time.Duration(tries) {
		_, retry, err = c.session.post(ctx, url, nil, nil, body)
		if err == nil {
			return
		}
		if !retry {
			return
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	return
}

func (c *clientAPI) postJsonRetry(ctx context.Context, url string, body *[]byte, tries int) (response map[string]any, err error) {
	var retry bool
	for i := range time.Duration(tries) {
		_, retry, err = c.session.postJSON(ctx, url, nil, nil, body, &response)
		if err == nil {
			return
		}
		if !retry {
			return
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	return
}

func (c *clientAPI) postTask(ctx context.Context, url string, params map[string]any) error {
	requestBody := paramsToBody(params)
	resp, _, err := c.session.post(ctx, url, nil, nil, &requestBody)
	if err != nil {
		return err
	}
	return c.checkTask(ctx, resp)
}

func (c *clientAPI) postRawTask(ctx context.Context, url string, body *[]byte) error {
	var response *http.Response
	var retry bool
	var err error
	for i := range time.Duration(RequestRetryCount) {
		response, retry, err = c.session.post(ctx, url, nil, nil, body)
		if err == nil || !retry {
			break
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	if err != nil {
		return err
	}
	return c.checkTask(ctx, response)
}

// Makes a PUT request without waiting on proxmox for the task to complete.
// It returns the HTTP error as 'err'.
func (c *clientAPI) put(ctx context.Context, url string, params map[string]any) (retry bool, err error) {
	reqbody := paramsToBodyWithAllEmpty(params)
	_, retry, err = c.session.put(ctx, url, nil, nil, &reqbody)
	return
}

func (c *clientAPI) putRawRetry(ctx context.Context, url string, body *[]byte, tries int) (err error) {
	var retry bool
	for i := range time.Duration(tries) {
		_, retry, err = c.session.put(ctx, url, nil, nil, body)
		if err == nil {
			return
		}
		if !retry {
			return
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	return
}

// checkTask polls the API to check if the Proxmox task has been completed.
// It returns the body of the HTTP response and any HTTP error occurred during the request.
func (c *clientAPI) checkTask(ctx context.Context, resp *http.Response) error {
	taskResponse, err := responseJSON(resp)
	if err != nil {
		return err
	}
	return c.waitForCompletion(ctx, taskResponse)
}

// waitForCompletion - poll the API for task completion
func (c *clientAPI) waitForCompletion(ctx context.Context, taskResponse map[string]any) error {
	if taskResponse["errors"] != nil {
		err, _ := json.MarshalIndent(taskResponse["errors"], "", "  ")
		return errors.New(string(err))
	}
	if taskResponse["data"] == nil {
		return nil
	}
	waited := time.Duration(0)
	taskUpid := taskResponse["data"].(string)
	for waited < c.taskTimeout {
		retry, err := c.getTaskExitStatus(ctx, taskUpid)
		if err != nil {
			return err
		}
		if !retry {
			return nil
		}
		time.Sleep(TaskStatusCheckInterval * time.Second)
		waited = waited + TaskStatusCheckInterval
	}
	return fmt.Errorf("Wait timeout for:" + taskUpid)
}

func (c *clientAPI) getTaskExitStatus(ctx context.Context, taskUpID string) (bool, error) {

	// "UPID:pve-01:00068180:17BE8318:697285F2:qmdelsnapshot:801:root@pam:"
	const prefixLen = len("UPID:")
	secondColon := strings.IndexByte(taskUpID[prefixLen:], ':') // find position of second colon
	node := taskUpID[prefixLen : prefixLen+secondColon]

	var err error
	var data map[string]any
	dataPtr := &data
	url := "/nodes/" + node + "/tasks/" + taskUpID + "/status"
	for i := range time.Duration(RequestRetryCount) {
		var retry bool
		_, retry, err = c.session.getJSON(ctx, url, nil, nil, dataPtr)
		if err == nil || !retry {
			break
		}
		if err == io.ErrUnexpectedEOF { // Early EOF can happen, don't retry
			return true, nil
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	if err != nil {
		return false, err
	}
	var exitStatus string
	if v, isSet := data["data"].(map[string]any)["exitstatus"]; isSet {
		exitStatus = v.(string)
	} else {
		return true, nil // still running
	}

	const taskSuccess = "OK"
	const taskWarning = "WARNING"

	if exitStatus == taskSuccess {
		return false, nil
	}
	if strings.HasPrefix(exitStatus, taskWarning) {
		return false, nil
	}
	return false, TaskError{
		Message: exitStatus,
		TaskID:  taskUpID}
}

func (c *clientAPI) getJsonRetry(ctx context.Context, url string, data *map[string]any, tries int) error {
	var err error
	var retry bool
	for i := range time.Duration(tries) {
		_, retry, err = c.session.getJSON(ctx, url, nil, nil, data)
		if err == nil {
			return nil
		}
		if !retry {
			return err
		}
		time.Sleep((i + 1) * c.timeUnit)
	}
	return err
}
