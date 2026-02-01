package proxmox

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_guestDoesNotExist(b *testing.B) {
	id := GuestID(123)
	for i := 0; i < b.N; i++ {
		err := errorMsg{}.guestDoesNotExist(id)
		_ = errors.Is(err, Error.GuestDoesNotExist())
	}
}

func Test_guestDoesNotExist(t *testing.T) {
	t.Parallel()
	id := GuestID(123)
	err := errorMsg{}.guestDoesNotExist(id)
	require.True(t, errors.Is(err, Error.GuestDoesNotExist()))
}
