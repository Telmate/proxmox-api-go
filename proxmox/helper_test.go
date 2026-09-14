package proxmox

import (
	"context"
	"crypto/tls"
	"iter"
	"net"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/mockServer"
	"github.com/stretchr/testify/require"
)

// Creates a test server and returns a authenticated client connected to it.
func testMockServerInit(t *testing.T) (*mockServer.Server, *Client) {
	t.Helper()
	server := mockServer.New(t)
	server.Set(mockServer.RequestsAuth(), t)
	c, err := NewClient(server.Url(), nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, false)
	c.timeUnit = time.Nanosecond
	require.NoError(t, err)
	err = c.Login(context.Background(), "root@pam", "", "")
	require.NoError(t, err)
	return server, c
}

func testParamsEqual(t *testing.T, expected map[string]string, params *[]byte, msgAndArgs ...any) {
	t.Helper()
	if params == nil {
		require.Nil(t, expected, msgAndArgs...)
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
	require.Equal(t, expected, out, msgAndArgs...)
}

func testParamsEqualRaw(t *testing.T, expected map[string]string, params *[]byte, msgAndArgs ...any) {
	t.Helper()
	if params == nil {
		require.Nil(t, expected, msgAndArgs...)
		return
	}

	values := strings.Split(string(*params), "&")

	out := make(map[string]string)
	for i := range values {
		if len(values[i]) > 0 {
			index := strings.IndexByte(values[i], '=')
			if index > 0 {
				out[values[i][0:index]] = values[i][index+1:]
			}
		}
	}
	require.Equal(t, expected, out, msgAndArgs...)
}

// An interface for types that have a Get() method returning V
type getAble[V any] interface{ Get() V }

func testCompareRawMap[key comparable, compareObject any, get getAble[compareObject]](t *testing.T, expected map[key]compareObject, actual map[key]get) {
	t.Helper()
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

func generateUPID(node NodeName, task string, guest GuestID, user UserID) string {
	return "UPID:" + node.String() + ":0006E4CB:17C8E729:6972A08C:" + task + ":" + guest.String() + ":" + user.String() + ":"
}

type mapToApiTest[V any] struct {
	output map[string]string
	input  V
	name   string
}

func parseMAC(mac string) net.HardwareAddr {
	parsedMac, err := net.ParseMAC(mac)
	if err != nil {
		panic(err)
	}
	return parsedMac
}

func parseCIDR(cidr string) (net.IP, *net.IPNet) {
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}
	return ip, net
}

type testIterInterface[T any] interface {
	Iter() iter.Seq[T]
}

func testIter[T any](t *testing.T, a testIterInterface[T], b []T) {
	t.Helper()
	var result []T
	// Test iterating over all items
	for pool := range a.Iter() {
		result = append(result, pool)
	}
	require.Equal(t, len(b), len(result))
	for i := range result {
		require.Equal(t, b[i], result[i])
	}
	if len(b) > 0 { // Test early termination (break after first item)
		count := 0
		for range a.Iter() {
			count++
			break
		}
		require.Equal(t, 1, count)
	}
}
