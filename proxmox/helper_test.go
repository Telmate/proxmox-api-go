package proxmox

import (
	"context"
	"crypto/tls"
	"net/url"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/stretchr/testify/require"
)

// Creates a test server and returns a authenticated client connected to it.
func testMockServerInit(t *testing.T) (*mockServer.Server, *Client) {
	server := mockServer.New(t)
	server.Set(mockServer.RequestsAuth(), t)
	c, err := NewClient(server.Url(), nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000)
	c.timeUnit = time.Nanosecond
	require.NoError(t, err)
	err = c.Login(context.Background(), "root@pam", "", "")
	require.NoError(t, err)
	return server, c
}

func testParamsEqual(t *testing.T, expected map[string]string, params *[]byte) {
	if params == nil {
		require.Nil(t, expected)
		return
	}

	values, err := url.ParseQuery(string(*params))
	require.NoError(t, err)

	out := make(map[string]string)
	for k, v := range values {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	require.Equal(t, expected, out)
}

// An interface for types that have a Get() method returning V
type getAble[V any] interface{ Get() V }

func testCompareRawMap[key comparable, compareObject any, get getAble[compareObject]](t *testing.T, expected map[key]compareObject, actual map[key]get) {
	if len(expected) != len(actual) {
		t.Fatalf("expected %d, got %d", len(expected), len(actual))
	}
	for k := range expected {
		v, ok := actual[k]
		if !ok {
			t.Fatalf("expected (%v) not found", k)
		}
		_ = v
		require.Equal(t, expected[k], v.Get())
	}
}
