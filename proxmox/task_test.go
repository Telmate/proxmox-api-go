package proxmox

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Test if the status is returned correctly.
func Test_task_Status(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()
	statusChannel := make(chan func() (map[string]interface{}, error))
	ctx, cancel := context.WithCancel(context.Background())
	c := &mocClient{
		getItemConfigFunc: func(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error) {
			// Block until data is sent into the channel
			select {
			case f := <-statusChannel:
				return f()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}}
	task := newTask(ctx, c, "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:", 10*time.Millisecond)

	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"status": "running"}}, nil
		}
	}()
	require.Equal(t, "running", task.Status())

	time.Sleep(50 * time.Millisecond)
	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"status":             "done",
					taskApiKeyExitStatus: "OK"}}, nil
		}
	}()

	time.Sleep(50 * time.Millisecond)

	require.Equal(t, "done", task.Status())
	ended, _ := task.Ended()
	require.True(t, ended)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines {
		t.Errorf("Potential goroutine leak: initial = %d, final = %d", initialGoroutines, finalGoroutines)
	} else {
		t.Log("No goroutine leaks detected")
	}
	cancel()
}

// Test if the log is returned correctly.
func Test_task_Logs(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()
	statusChannel := make(chan func() (map[string]interface{}, error))
	logChannel := make(chan func() (map[string]interface{}, error))
	ctx, cancel := context.WithCancel(context.Background())
	c := &mocClient{
		getItemConfigFunc: func(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error) {
			// Block until data is sent into the channel
			select {
			case f := <-statusChannel:
				return f()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		getItemListFunc: func(ctx context.Context, url string) (map[string]interface{}, error) {
			// Block until data is sent into the channel
			select {
			case f := <-logChannel:
				return f()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
	task := newTask(ctx, c, "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:", 10*time.Millisecond)

	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"status": "running"}}, nil
		}
		logChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"total": float64(4),
				"data": []interface{}{
					map[string]interface{}{"t": "1"},
					map[string]interface{}{"t": "2"},
					map[string]interface{}{"t": "3"},
					map[string]interface{}{"t": "4"},
				}}, nil
		}
	}()
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "running", task.Status())
	require.Equal(t, []string{
		"1",
		"2",
		"3",
		"4",
	}, task.Log())

	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"status":             "done",
					taskApiKeyExitStatus: "OK"}}, nil
		}
		logChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"total": float64(7),
				"data": []interface{}{
					map[string]interface{}{"t": "5"},
					map[string]interface{}{"t": "6"},
					map[string]interface{}{"t": "TASK OK"},
				}}, nil
		}
	}()
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "done", task.Status())
	require.Equal(t, []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"TASK OK",
	}, task.Log())
	ended, err := task.Ended()
	require.True(t, ended)
	require.NoError(t, err)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines {
		t.Errorf("Potential goroutine leak: initial = %d, final = %d", initialGoroutines, finalGoroutines)
	} else {
		t.Log("No goroutine leaks detected")
	}
	cancel()
}

// Test if all go routines are cleaned up after an error occurs.
func Test_task_Log_error(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()
	statusChannel := make(chan func() (map[string]interface{}, error))
	logChannel := make(chan func() (map[string]interface{}, error))
	ctx, cancel := context.WithCancel(context.Background())
	c := &mocClient{
		getItemConfigFunc: func(ctx context.Context, url, text, message string, errorStrings []string) (map[string]interface{}, error) {
			// Block until data is sent into the channel
			select {
			case f := <-statusChannel:
				return f()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
		getItemListFunc: func(ctx context.Context, url string) (map[string]interface{}, error) {
			// Block until data is sent into the channel
			select {
			case f := <-logChannel:
				return f()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		},
	}
	task := newTask(ctx, c, "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:", 10*time.Millisecond)

	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"data": map[string]interface{}{
					"status": "running"}}, nil
		}
		logChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"total": float64(4),
				"data": []interface{}{
					map[string]interface{}{"t": "1"},
					map[string]interface{}{"t": "2"},
					map[string]interface{}{"t": "3"},
					map[string]interface{}{"t": "4"},
				}}, nil
		}
	}()
	time.Sleep(50 * time.Millisecond)
	require.Equal(t, "running", task.Status())
	require.Equal(t, []string{
		"1",
		"2",
		"3",
		"4",
	}, task.Log())

	go func() {
		statusChannel <- func() (map[string]interface{}, error) {
			return map[string]interface{}{
					"data": map[string]interface{}{
						"status":             "stopped",
						taskApiKeyExitStatus: "ERROR"}},
				errors.New("error status")
		}
	}()
	time.Sleep(50 * time.Millisecond)
	ended, err := task.Ended()
	require.True(t, ended)
	require.Error(t, err)

	finalGoroutines := runtime.NumGoroutine()
	if finalGoroutines > initialGoroutines {
		t.Errorf("Potential goroutine leak: initial = %d, final = %d", initialGoroutines, finalGoroutines)
	} else {
		t.Log("No goroutine leaks detected")
	}
	cancel()
}

func Test_nodeFromUpID_Unsafe(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output task
	}{
		{name: "1",
			input: "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:",
			output: task{
				id:            "UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:",
				node:          "pve-test",
				operationType: "qmmove",
				user: UserID{
					Name:  "root",
					Realm: "pam"}}},
		{name: "2",
			input: "UPID:pve:002860A9:051E01C1:67536165:qmshutdown:102:test-user@realm:",
			output: task{
				id:            "UPID:pve:002860A9:051E01C1:67536165:qmshutdown:102:test-user@realm:",
				node:          "pve",
				operationType: "qmshutdown",
				user: UserID{
					Name:  "test-user",
					Realm: "realm"}}},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			tmpTask := &task{}
			tmpTask.mapToSDK_Unsafe(tests[i].input)
			require.Equal(t, tests[i].output.id, tmpTask.id)
			require.Equal(t, tests[i].output.node, tmpTask.node)
			require.Equal(t, tests[i].output.operationType, tmpTask.operationType)
			require.Equal(t, tests[i].output.user, tmpTask.user)
		})
	}
}
