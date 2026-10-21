package proxmox

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_VmRef_Migrate(t *testing.T) {
	t.Parallel()
	type testInput struct {
		c   *Client
		ctx context.Context
		vmr *VmRef
	}
	tests := []struct {
		name  string
		input testInput
	}{
		{name: `Client nil`,
			input: testInput{
				ctx: context.Background(),
				vmr: &VmRef{}}},
		{name: `VmRef nil`,
			input: testInput{
				c:   fakeClient(),
				ctx: context.Background()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { test.input.vmr.Migrate(test.input.ctx, test.input.c, "valid", false) })
			require.Error(t, test.input.vmr.Migrate(test.input.ctx, test.input.c, "valid", false))
		})
	}
}

func Test_VmRef_MigrateNoCheck(t *testing.T) {
	t.Parallel()
	type testInput struct {
		c   *Client
		ctx context.Context
		vmr *VmRef
	}
	tests := []struct {
		name  string
		input testInput
	}{
		{name: `Client nil`,
			input: testInput{
				ctx: context.Background(),
				vmr: &VmRef{}}},
		{name: `VmRef nil`,
			input: testInput{
				c:   fakeClient(),
				ctx: context.Background()}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.NotPanics(t, func() { test.input.vmr.MigrateNoCheck(test.input.ctx, test.input.c, "valid", false) })
			require.Error(t, test.input.vmr.MigrateNoCheck(test.input.ctx, test.input.c, "valid", false))
		})
	}
}
