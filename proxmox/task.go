package proxmox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	atomicError "github.com/Telmate/proxmox-api-go/internal/atomicerror"
	"github.com/Telmate/proxmox-api-go/internal/notify"
)

func (c *Client) taskResponse(ctx context.Context, resp *http.Response) (Task, error) {
	if c.Features == nil || !c.Features.AsyncTask {
		exit, err := c.CheckTask(ctx, resp)
		return newDummyTask(exit, err), nil
	}
	var jbody map[string]interface{}
	var err error
	if err = decodeResponse(resp, &jbody); err != nil {
		return &task{}, err
	}
	if v, isSet := jbody["errors"]; isSet {
		errJSON, _ := json.MarshalIndent(v, "", "  ")
		return &task{}, fmt.Errorf("error: %s", errJSON)
	}
	if v, isSet := jbody["data"]; isSet && v != nil {
		return newTask(ctx, &client{c: c}, v.(string), taskPollingInterval), nil
	}
	return &task{}, nil
}

const (
	logChunkSize        uint   = 510
	logChunkSizeString  string = "510"
	taskPollingInterval        = 1 * time.Second
)

// Only `WaitForCompletion()` is required to be used.
// For now all other functionality is opt-in with feature flag with exception of `WaitForCompletion()`.
// The context that was inherited from the original API call.
// To cancel the task, cancel the context.
// You can also cancel the task by calling the Cancel() method.
type Task interface {

	// Cancels the task.
	Cancel() error

	// Returns true if the task has ended, and an error if an error occurred.
	// That the task has ended does not mean the log is fully fetched.
	Ended() (bool, error)

	// Returns the time the task ended. If the task has not ended, the zero time is returned.
	EndTime() time.Time

	// Returns the exit status of the task. If the task has not ended, an empty string is returned.
	ExitStatus() string

	// Returns the ID of the task.
	ID() string

	// Returns the current log of the task.
	Log() []string

	// Inputs the log into the provided channel.
	// TODO add more tests for this functionality.
	_LogStream(log chan<- string) error

	// Returns the node the task was executed on.
	Node() string

	// Returns the operation type of the task.
	OperationType() string

	// Returns the process ID of the task.
	ProcessID() uint

	// Returns the time the task started.
	// If the task has not started, the zero time is returned.
	StartTime() time.Time

	// Returns the status of the task.
	Status() string

	// Returns the user that started the task.
	User() UserID

	// Blocks until the task is completed, or returned an error.
	WaitForCompletion() error
}

type task struct {
	id            string
	node          string
	operationType string
	status        map[string]interface{}
	statusMutex   sync.Mutex
	user          UserID
	client        clientInterface

	pollingInterval time.Duration

	// ctx is the context that is inherited from the original api call.
	ctx context.Context

	logCache sync.Map

	// This channel is open by default and closed when the task ends.
	closingCh *notify.Channel

	// While statusCh is open, the status ha not been fetched yet.
	statusCh *notify.Channel

	// While logClosedCh is open, the end of the log has not been reached yet.
	logClosedCh *notify.Channel

	// While logStartedCh is open, the log has not been fetched yet.
	logStartedCh *notify.Channel

	// err is the error that is stored when an error occurs in the status loop.
	err atomicError.AtomicError

	fetchLog   sync.Once
	cancelTask sync.Mutex
}

const (
	taskApiKeyEndTime    = "endtime"
	taskApiKeyExitStatus = "exitstatus"
	taskApiKeyProcessID  = "pid"
	taskApiKeyStartTime  = "starttime"
	taskApiKeyStatus     = "status"
)

func (t *task) Cancel() error {
	if t.id == "" {
		return nil
	}
	return t.cancel(nil)
}

func (t *task) Ended() (bool, error) {
	if t.id == "" {
		return true, nil
	}
	<-t.statusCh.Done()
	return t.ended()
}

func (t *task) EndTime() time.Time {
	if t.id == "" {
		return time.Time{}
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyEndTime]; isSet {
		return time.Unix(int64(v.(float64)), 0)
	}
	return time.Time{}
}

func (t *task) ExitStatus() string {
	if t.id == "" {
		return ""
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyExitStatus]; isSet {
		return v.(string)
	}
	return ""
}

func (t *task) ID() string {
	return t.id
}

func (t *task) Log() []string {
	if t.id == "" {
		return nil
	}
	t.startLogFetcher()
	localLog := make(map[int]string)
	<-t.logStartedCh.Done() // Wait for the log to be fetched for the first time.
	t.logCache.Range(func(key, value interface{}) bool {
		localLog[key.(int)] = value.(string)
		return true // continue iteration
	})
	logArray := make([]string, len(localLog))
	for i := int(0); i < int(len(localLog)); i++ {
		logArray[i] = localLog[i]
	}
	return logArray
}

func (t *task) _LogStream(out chan<- string) error {
	if t.id == "" {
		return nil
	}
	t.startLogFetcher()
	var index uint
	var v interface{}
	var ok bool
	for {
		select {
		case <-t.closingCh.Done():
			return t.err.Get()
		default:
			v, ok = t.logCache.Load(index)
			if !ok {
				time.Sleep(t.pollingInterval)
				continue
			}
			out <- v.(string)
			index++
		}
	}
}

func (t *task) Node() string {
	return t.node
}

func (t *task) OperationType() string {
	return t.operationType
}

func (t *task) ProcessID() uint {
	if t.id == "" {
		return 0
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyProcessID]; isSet {
		return uint(v.(float64))
	}
	return 0
}

func (t *task) StartTime() time.Time {
	if t.id == "" {
		return time.Time{}
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyStartTime]; isSet {
		return time.Unix(int64(v.(float64)), 0)
	}
	return time.Time{}
}

func (t *task) Status() string {
	if t.id == "" {
		return ""
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyStatus]; isSet {
		return v.(string)
	}
	return ""
}

func (t *task) User() UserID {
	return t.user
}

func (t *task) WaitForCompletion() (err error) {
	if t.id == "" { // if we didn't get a task ID, the function that instantiated the task should be changed to not return a task.
		return nil
	}
	<-t.statusCh.Done()  // Block until the task status has been fetched for the first time.
	<-t.closingCh.Done() // Block until the task has ended.
	select {
	case <-t.logStartedCh.Done(): // We started fetching the log, so we should wait for it to finish.
		<-t.logClosedCh.Done() // Block until the log has been fully fetched.
		return t.err.Get()
	default: // We haven't started fetching the log yet, so we can just return.
		return t.err.Get()
	}
}

func newTask(ctx context.Context, c clientInterface, upID string, pollingInterval time.Duration) *task {
	t := &task{
		client:          c,
		closingCh:       notify.New(),
		ctx:             ctx,
		logClosedCh:     notify.New(),
		logStartedCh:    notify.New(),
		pollingInterval: pollingInterval,
		statusCh:        notify.New()}
	t.mapToSDK_Unsafe(upID)
	go func() { // Start the status fetcher, which periodically fetches the status from the API and stores it in the status field.
		var err error
		var gotFirstStatus bool
		for {
			var params map[string]interface{}
			params, err = t.client.getItemConfig(t.ctx, "/nodes/"+t.node+"/tasks/"+t.id+"/status", "", "", nil)
			if err != nil {
				t.setError(err)
				return
			}
			status := params["data"].(map[string]interface{})
			t.statusMutex.Lock()
			t.status = status
			t.statusMutex.Unlock()
			if !gotFirstStatus { // Close the channel to indicate that the status has been fetched for the first time.
				t.statusCh.Close()
				gotFirstStatus = true
			}
			if v, isSet := status[taskApiKeyExitStatus]; isSet {
				exitStatus := v.(string)
				if !strings.HasPrefix(exitStatus, "OK") && !strings.HasPrefix(exitStatus, "WARNINGS") {
					t.setError(errors.New(exitStatus))
					return
				}
				t.closingCh.Close()
				return
			}
			select {
			case <-t.closingCh.Done():
				return
			case <-time.After(t.pollingInterval):
			}
		}
	}()
	return t
}

// Cancels the task and closes the control channel.
func (t *task) cancel(previousErr error) (err error) {
	t.cancelTask.Lock()
	defer t.cancelTask.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = t.client.delete(ctx, "/nodes/"+t.node+"/tasks/"+t.id)
	if err != nil {
		if previousErr != nil {
			err = fmt.Errorf("%s: %w", previousErr, err)
		}
		t.setError(err)
		return
	}
	t.closingCh.Close()
	t.logStartedCh.Close()
	t.statusCh.Close()
	return
}

// Returns true if the task has ended, and an error if an error occurred.
func (t *task) ended() (bool, error) {
	select {
	case <-t.closingCh.Done():
		return true, t.err.Get()
	default:
		return false, nil
	}
}

// Set the error and close the control channel.
// Also closes the status and log channels to prevent blocking when users try to get data from the task.
func (t *task) setError(err error) {
	t.err.Set(err)
	// Close the channel to stop the goroutines.
	t.closingCh.Close()
	// Close the channels to prevent blocking when users try to get data from task.
	t.logClosedCh.Close()
	t.logStartedCh.Close()
	t.statusCh.Close()
}

// retrieves a chunk of the log from the API, parses and stores it in the log cache.
func (t *task) addLogChunkToCache(start uint) (total, cacheLen uint, err error) {
	// The GUI uses a limit of 510. Not sure if going higher causes issues.
	// Therefor we will also request the log in chunks of 510.
	var params map[string]interface{}
	params, err = t.client.getItemList(t.ctx, "/nodes/"+t.node+"/tasks/"+t.id+"/log?start="+strconv.FormatUint(uint64(start), 10)+"&limit="+logChunkSizeString)
	if err != nil {
		return
	}
	total = uint(params["total"].(float64))
	logEntries := params["data"].([]interface{})
	if len(logEntries) == 0 {
		return total, start, nil
	}
	startInt := int(start) // Convert to int for the loop, so we don't have to convert it every iteration.
	for i := range logEntries {
		logMessage := logEntries[i].(map[string]interface{})["t"].(string)
		t.logCache.Store(i+startInt, logMessage)
		if len(logMessage) >= 4 && logMessage[:4] == "TASK" { // The last log message will always be `TASK` followed by the status.
			t.logClosedCh.Close()
		}
	}
	cacheLen = uint(len(logEntries)) + start
	return
}

// Starts the log fetcher if it hasn't been started yet.
// The log fetcher is the goroutine that fetches the log from the API and stores it in the log cache.
func (t *task) startLogFetcher() {
	t.fetchLog.Do(func() {
		go func() {
			var total, cacheLen uint
			var err error
			var gotFirstLog bool
			for {
				select {
				case <-t.closingCh.Done():
					if t.err.Get() != nil {
						return
					}
					for { // Fetch the log one last time before returning to ensure we have all the logs.
						cacheLen, total, err = t.addLogChunkToCache(cacheLen)
						if err != nil {
							t.setError(err)
							return
						}
						if !gotFirstLog {
							t.logStartedCh.Close()
							gotFirstLog = true
						}
						if total == cacheLen {
							return
						}
					}
				default:
					// Fetch new log lines from the API
					cacheLen, total, err = t.addLogChunkToCache(cacheLen)
					if err != nil {
						t.setError(err)
					}
					if !gotFirstLog {
						t.logStartedCh.Close()
						gotFirstLog = true
					}
					if total == cacheLen {
						time.Sleep(t.pollingInterval)
					}
				}
			}
		}()
	})
}

// UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:
// Requires the caller to ensure (t *task) is not nil, or it will panic.
// parse the UPID string and map it to the task struct.
func (t *task) mapToSDK_Unsafe(upID string) {
	t.id = upID
	indexA := strings.Index(upID[5:], ":") + 5
	t.node = upID[5:indexA]
	indexB := strings.Index(upID[indexA+28:], ":") + indexA + 28
	t.operationType = upID[indexA+28 : indexB]
	indexA = strings.Index(upID[indexB+1:], ":") + indexB + 1 + 1 // +1 because we are skipping a field
	t.user = UserID{}.mapToStruct(upID[indexA : strings.Index(upID[indexA:], ":")+indexA])
}

type dummyTask struct {
	err    error
	status string
}

func (d *dummyTask) Cancel() error {
	return nil
}

func (d *dummyTask) Ended() (bool, error) {
	return true, d.err
}

func (d *dummyTask) EndTime() time.Time {
	return time.Time{}
}

func (d *dummyTask) ExitStatus() string {
	return d.status
}
func (d *dummyTask) ID() string {
	return ""
}
func (d *dummyTask) Log() []string {
	return []string{}
}
func (d *dummyTask) _LogStream(log chan<- string) error {
	return nil
}
func (d *dummyTask) Node() string {
	return ""
}
func (d *dummyTask) OperationType() string {
	return ""
}
func (d *dummyTask) ProcessID() uint {
	return 0
}
func (d *dummyTask) StartTime() time.Time {
	return time.Time{}
}
func (d *dummyTask) Status() string {
	return ""
}
func (d *dummyTask) User() UserID {
	return UserID{}
}
func (d *dummyTask) WaitForCompletion() error {
	return d.err
}

func newDummyTask(exitStatus string, err error) *dummyTask {
	return &dummyTask{
		err:    err,
		status: exitStatus}
}
