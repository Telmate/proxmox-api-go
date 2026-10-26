package proxmox

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	atomicError "github.com/Telmate/proxmox-api-go/internal/atomicerror"
	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/notify"
)

func (c *qemuAgentClient) Exec(ctx context.Context, vmr VmRef, commands []string) (ExecStatus, error) {
	if len(commands) == 0 {
		return new(execFake), nil
	}
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return nil, err
	}
	return c.ExecNoCheck(ctx, ref.legacy(), commands)
}

func (c *qemuAgentClient) ExecNoCheck(ctx context.Context, vmr VmRef, commands []string) (ExecStatus, error) {
	var b strings.Builder
	for i := range commands {
		b.WriteString("&command=")
		b.WriteString(body.Escape(commands[i]))
	}
	var reqBody []byte
	if b.Len() > 0 {
		reqBody = []byte(b.String()[1:])
	}
	rawID, err := c.api.postMap(ctx, "/nodes/"+vmr.node.String()+"/qemu/"+vmr.vmId.String()+"/agent/exec", &reqBody, "", "")
	if err != nil {
		return nil, err
	}
	return newExecPoller(ctx, c.api, vmr.vmId, vmr.node, int(rawID["pid"].(float64)), 200*time.Millisecond), nil
}

type ExecStatus interface {
	Await() error
	Error() error
	Get() (stdOut CommandStatus, stdErr CommandStatus, exitCode uint)
}

var _ ExecStatus = (*exec)(nil)

type CommandStatus struct {
	Output      string
	IsTruncated bool
}

func newExecPoller(ctx context.Context, c *clientAPI, id GuestID, node NodeName, pID int, pollingInterval time.Duration) *exec {
	x := &exec{
		closingCh:       notify.New(),
		id:              id.String(),
		node:            node.String(),
		pID:             strconv.Itoa(pID),
		pollingInterval: pollingInterval,
		statusCh:        notify.New(),
	}
	go func() {
		var err error
		var gotFirstStatus bool
		for {
			var data map[string]any
			data, err = c.getMap(ctx, "/nodes/"+x.node+"/qemu/"+x.id+"/agent/exec-status?pid="+x.pID, "", "")
			if err != nil {
				x.setError(err)
				return
			}
			var tmpExitCode uint
			var tmpOutTruncated, tmpErrTruncated, tmpExisted bool
			var tmpOutData, tmpErrData string
			var updateCache bool
			if v, ok := data["exitcode"]; ok {
				tmpExitCode = uint(v.(float64))
				if x.lastExitCode == tmpExitCode {
					x.lastExitCode = tmpExitCode
					updateCache = true
				}
			}
			if v, ok := data["out-truncated"]; ok {
				tmpOutTruncated = int(v.(float64)) == 1
				if x.lastOutTruncated == tmpOutTruncated {
					x.lastOutTruncated = tmpOutTruncated
					updateCache = true
				}
			}
			if v, ok := data["err-truncated"]; ok {
				tmpErrTruncated = int(v.(float64)) == 1
				if x.lastErrTruncated == tmpErrTruncated {
					x.lastErrTruncated = tmpErrTruncated
					updateCache = true
				}
			}
			if v, ok := data["out-data"]; ok {
				tmpOutData = v.(string)
				if x.lastOutputData == tmpOutData {
					x.lastOutputData = tmpOutData
					updateCache = true
				}
			}
			if v, ok := data["err-data"]; ok {
				tmpErrData = v.(string)
				if x.lastErrorData == tmpErrData {
					x.lastErrorData = tmpErrData
					updateCache = true
				}
			}
			if v, ok := data["exited"]; ok {
				tmpExisted = int(v.(float64)) == 1
				if x.lastExited == tmpExisted {
					x.lastExited = tmpExisted
					updateCache = true
				}
			}
			if updateCache {
				x.cacheMutex.Lock()
				x.cacheErrData = tmpErrData
				x.cacheErrTruncated = tmpErrTruncated
				x.cacheExitCode = tmpExitCode
				x.cacheExited = tmpExisted
				x.cacheOutData = tmpOutData
				x.cacheOutTruncated = tmpOutTruncated
				x.cacheMutex.Unlock()
			}
			if !gotFirstStatus { // Close the channel to indicate that the status has been fetched for the first time.
				x.statusCh.Close()
				gotFirstStatus = true
			}
			if tmpExisted {
				x.closingCh.Close()
				return
			}
			<-time.After(x.pollingInterval)
		}
	}()
	return x
}

type exec struct {
	closingCh         *notify.Channel // This channel is closed when the task ends.
	statusCh          *notify.Channel // This closes once we got api data
	cacheErrData      string
	cacheOutData      string
	id                string
	lastErrorData     string
	lastOutputData    string
	node              string
	pID               string
	err               atomicError.AtomicError
	cacheExitCode     uint
	lastExitCode      uint
	pollingInterval   time.Duration
	cacheMutex        sync.RWMutex
	cacheErrTruncated bool
	cacheExited       bool
	cacheOutTruncated bool
	lastErrTruncated  bool
	lastExited        bool
	lastOutTruncated  bool
}

func (x *exec) Await() error {
	<-x.closingCh.Done() // Block until the task has ended.
	return x.err.Load()
}

func (x *exec) Error() error {
	<-x.statusCh.Done() // Block until the task has started.
	return x.err.Load()
}

func (x *exec) Get() (CommandStatus, CommandStatus, uint) {
	<-x.statusCh.Done() // Block until the task has started.
	x.cacheMutex.RLock()
	defer x.cacheMutex.RUnlock()
	return CommandStatus{
		Output:      x.cacheOutData,
		IsTruncated: x.cacheOutTruncated,
	}, CommandStatus{
		Output: x.cacheErrData,
	}, x.cacheExitCode
}

// Set the error and close the control channel.
// Also closes the status and log channels to prevent blocking when users try to get data from the task.
func (x *exec) setError(err error) {
	x.err.Store(err)
	// Close the channel to stop the goroutines.
	x.closingCh.Close()
	x.statusCh.Close()
}

type execFake struct{}

var _ ExecStatus = (*execFake)(nil)

func (x *execFake) Await() error { return nil }

func (x *execFake) Error() error { return nil }

func (x *execFake) Get() (CommandStatus, CommandStatus, uint) {
	return CommandStatus{
		Output:      "",
		IsTruncated: false,
	}, CommandStatus{
		Output:      "",
		IsTruncated: false,
	}, 0
}
