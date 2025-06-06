package proxmox

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LxcMountSize_String(t *testing.T) {
	require.Equal(t, "547434", LxcMountSize(547434).String())
}
