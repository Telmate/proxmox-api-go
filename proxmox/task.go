package proxmox

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	atomicError "github.com/Telmate/proxmox-api-go/internal/atomicerror"
	"github.com/Telmate/proxmox-api-go/internal/notify"
)

const (
	taskApiKeyEndTime    = "endtime"
	taskApiKeyExitStatus = "exitstatus"
	taskApiKeyProcessID  = "pid"
	taskApiKeyStartTime  = "starttime"
	taskApiKeyStatus     = "status"
)

type demoTask interface {

	// Cancels the task.
	Cancel() error

	// Returns true if the task has ended.
	Ended() bool

	// Return the task error.
	Error() error

	// Returns the time the task ended. If the task has not ended, returns false.
	EndTime() (time.Time, bool)

	// Returns the exit status of the task. If the task has not ended, an empty string is returned.
	ExitStatus() string

	// Returns the ID of the task.
	ID() string

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

var _ demoTask = (*task)(nil)

type task struct {
	id            string
	node          string
	operationType string

	// status points to the latest immutable status snapshot.
	// The writer replaces the pointer, it never modifies an existing map.
	status      *map[string]any
	statusMutex sync.Mutex
	user        UserID
	client      *clientAPI

	pollingInterval time.Duration

	// ctx is the context that is inherited from the original api call.
	ctx context.Context

	// logCache logCache

	// This channel is open by default and closed when the task ends.
	closingCh *notify.Channel

	// While statusCh is open, the status has not been fetched yet.
	// We use this to block reads until we have some data.
	statusCh *notify.Channel

	// err is the error that is stored when an error occurs in the status loop.
	err atomicError.AtomicError

	// fetchLog   sync.Once
	cancelTask sync.Once
}

func (t *task) Cancel() error {
	if t.id == "" {
		return nil
	}
	return t.cancel()
}

func (t *task) Ended() bool {
	if t.id == "" {
		return true
	}
	select {
	case <-t.closingCh.Done():
		return true
	default:
		return false
	}
}

func (t *task) Error() error { return t.err.Load() }

func (t *task) EndTime() (time.Time, bool) {
	if t.id == "" {
		return time.Time{}, false
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	status := *t.status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyEndTime]; isSet {
		return time.Unix(int64(v.(float64)), 0), true
	}
	return time.Time{}, false
}

func (t *task) ExitStatus() string {
	if t.id == "" {
		return ""
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	status := *t.status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyExitStatus]; isSet {
		return v.(string)
	}
	return ""
}

func (t *task) ID() string { return t.id }

func (t *task) Node() string { return t.node }

func (t *task) OperationType() string { return t.operationType }

func (t *task) ProcessID() uint {
	if t.id == "" {
		return 0
	}
	<-t.statusCh.Done()
	t.statusMutex.Lock()
	status := *t.status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyProcessID]; isSet {
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
	status := *t.status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyStartTime]; isSet {
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
	status := *t.status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyStatus]; isSet {
		return v.(string)
	}
	return ""
}

func (t *task) User() UserID { return t.user }

func (t *task) WaitForCompletion() (err error) {
	if t.id == "" { // if we didn't get a task ID, the function that instantiated the task should be changed to not return a task.
		return nil
	}
	<-t.closingCh.Done() // Block until the task has ended.
	return t.err.Load()
}

// Cancels the task and closes the control channel.
func (t *task) cancel() (err error) {
	t.cancelTask.Do(func() {
		ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
		defer cancel()
		err = t.client.deleteRetry(ctx, "/nodes/"+t.node+"/tasks/"+t.id, 3)
		if err != nil {
			t.setError(err)
			return
		}
		t.closingCh.Close()
		t.statusCh.Close()
	})
	return
}

func newTask(ctx context.Context, c *clientAPI, upID string, pollingInterval time.Duration) *task {
	const taskSuccess = "OK"
	const taskWarning = "WARNING"
	t := &task{
		client:          c,
		closingCh:       notify.New(),
		ctx:             ctx,
		pollingInterval: pollingInterval,
		statusCh:        notify.New(),
	}
	t.mapToSDK_Unsafe(upID)
	go func() { // Start the status fetcher, which periodically fetches the status from the API and stores it in the status field.
		var err error
		var gotFirstStatus bool
		for {
			var data map[string]any
			data, err = t.client.getMap(t.ctx, "/nodes/"+t.node+"/tasks/"+t.id+"/status", "", "")
			if err != nil {
				t.setError(err)
				return
			}
			t.statusMutex.Lock()
			t.status = &data
			t.statusMutex.Unlock()
			if !gotFirstStatus { // Close the channel to indicate that the status has been fetched for the first time.
				t.statusCh.Close()
				gotFirstStatus = true
			}
			if v, isSet := data[taskApiKeyExitStatus]; isSet { // we don't need to lock as we are the only writer, and map is safe from concurrent reads
				exitStatus := v.(string)
				if exitStatus != taskSuccess && !strings.HasPrefix(exitStatus, taskWarning) {
					t.setError(&TaskError{
						TaskID:  t.id,
						Message: exitStatus})
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

// Set the error and close the control channel.
// Also closes the status and log channels to prevent blocking when users try to get data from the task.
func (t *task) setError(err error) {
	t.err.Store(err)
	// Close the channel to stop the goroutines.
	t.closingCh.Close()
	t.statusCh.Close()
}

// UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:
// parse the UPID string and map it to the task struct.
func (t *task) mapToSDK_Unsafe(upID string) {
	t.id = upID
	indexA := strings.Index(upID[5:], ":") + 5
	t.node = upID[5:indexA]
	indexB := strings.Index(upID[indexA+28:], ":") + indexA + 28
	t.operationType = upID[indexA+28 : indexB]
	indexA = strings.Index(upID[indexB+1:], ":") + indexB + 1 + 1 // +1 because we are skipping a field
	var user UserID
	// This can never fail
	_ = user.Parse(upID[indexA : strings.Index(upID[indexA:], ":")+indexA])
	t.user = user
}

type logCache struct {
	// Essentially a linked list
	Chunks *logChunk

	// is only for the writer
	Current *logChunk

	numberOfElements atomic.Uint64
}

const chunkSize = 256

type logChunk struct {
	next  *logChunk
	chunk [chunkSize]string
}

func (log *logCache) Load() []string {
	numberOfElements := log.numberOfElements.Load()
	innerIndex := numberOfElements % chunkSize
	outerIndex := numberOfElements / chunkSize

	output := make([]string, int(numberOfElements))

	current := log.Chunks

	var outputIndex int
	for range outerIndex {
		for i := range chunkSize { // Somehow loop is faster than copy
			output[outputIndex] = current.chunk[i]
			outputIndex++
		}

		current = current.next
	}

	for i := range innerIndex {
		output[outputIndex] = current.chunk[i]
		outputIndex++
	}
	return output
}

func (log *logCache) Store(raw []any) {
	numberOfElements := log.numberOfElements.Load()
	innerIndex := numberOfElements % chunkSize
	for i := range len(raw) {
		if innerIndex == chunkSize {
			innerIndex = 1
			newChunk := &logChunk{}
			newChunk.chunk[0] = raw[i].(map[string]any)["t"].(string)
			log.Current.next = newChunk
			log.Current = newChunk
			continue
		}
		log.Current.chunk[innerIndex] = raw[i].(map[string]any)["t"].(string)
		innerIndex++
	}
	if innerIndex == chunkSize {
		newChunk := &logChunk{}
		log.Current.next = newChunk
		log.Current = newChunk
	}
	log.numberOfElements.Store(numberOfElements + uint64(len(raw)))
}
