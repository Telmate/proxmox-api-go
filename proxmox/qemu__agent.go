package proxmox

import (
	"context"
	"strings"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/body"
)

type (
	QemuAgentInterface interface {
		Exec(ctx context.Context, vmr VmRef, commands []string) (ExecStatus, error)
		ExecNoCheck(ctx context.Context, vmr VmRef, commands []string) (ExecStatus, error)

		NetworkInterfaces(context.Context, VmRef) (RawAgentNetworkInterfaces, GuestAgentState, error)
		NetworkInterfacesNoCheck(context.Context, VmRef) (RawAgentNetworkInterfaces, GuestAgentState, error)

		FsTrim(context.Context, VmRef) error
		FsTrimNoCheck(context.Context, VmRef) error

		Time(context.Context, VmRef) (time.Time, error)
		TimeNoCheck(context.Context, VmRef) (time.Time, error)

		Ping(context.Context, VmRef) error
		PingNoCheck(context.Context, VmRef) error

		SetUserPassword(ctx context.Context, vmr VmRef, username string, password PamPassword) error
		SetUserPasswordNoCheck(ctx context.Context, vmr VmRef, username string, password PamPassword) error

		// path + data is limited to 61440 bytes after encoding.
		WriteFile(ctx context.Context, vmr VmRef, path string, data []byte) error
		WriteFileNoCheck(ctx context.Context, vmr VmRef, path string, data []byte) error
	}

	qemuAgentClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ QemuAgentInterface = (*qemuAgentClient)(nil)

func (c *qemuAgentClient) FsTrim(ctx context.Context, vmr VmRef) error {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return err
	}
	return c.FsTrimNoCheck(ctx, ref.legacy())
}

func (c *qemuAgentClient) FsTrimNoCheck(ctx context.Context, vmr VmRef) error {
	// TODO Figure out how to return the data
	_, err := c.api.postAgent(ctx, vmr.vmId, vmr.node, "fstrim", nil, "qemu agent", "fstrim")
	return err
}

func (c *qemuAgentClient) Time(ctx context.Context, vmr VmRef) (time.Time, error) {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return time.Time{}, err
	}
	return c.TimeNoCheck(ctx, ref.legacy())
}

func (c *qemuAgentClient) TimeNoCheck(ctx context.Context, vmr VmRef) (time.Time, error) {
	a, err := c.api.getAgent(ctx, vmr.vmId, vmr.node, "get-time", "qemu agent", "get-time")
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(a.(float64))), nil
}

func (c *qemuAgentClient) Ping(ctx context.Context, vmr VmRef) error {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return err
	}
	return c.PingNoCheck(ctx, ref.legacy())
}

func (c *qemuAgentClient) PingNoCheck(ctx context.Context, vmr VmRef) error {
	_, err := c.api.postAgent(ctx, vmr.vmId, vmr.node, "ping", nil, "qemu agent", "ping")
	return err
}

func (c *qemuAgentClient) SetUserPassword(ctx context.Context, vmr VmRef, username string, password PamPassword) error {
	if err := password.Validate(); err != nil {
		return err
	}
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return err
	}
	return c.SetUserPasswordNoCheck(ctx, ref.legacy(), username, password)
}

func (c *qemuAgentClient) SetUserPasswordNoCheck(ctx context.Context, vmr VmRef, username string, password PamPassword) error {
	var b strings.Builder
	b.WriteString("username=")
	b.WriteString(body.Escape(username))
	b.WriteString("&password=")
	b.WriteString(body.Escape(password.String()))
	_, err := c.api.postAgent(ctx, vmr.vmId, vmr.node, "set-user-password", new([]byte(b.String())), "qemu agent", "set-user-password")
	return err
}

func (c *qemuAgentClient) WriteFile(ctx context.Context, vmr VmRef, path string, data []byte) error {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	_, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return err
	}
	return c.WriteFileNoCheck(ctx, ref.legacy(), path, data)
}

func (c *qemuAgentClient) WriteFileNoCheck(ctx context.Context, vmr VmRef, path string, data []byte) error {
	b := append([]byte("file="+body.Escape(path)+"&content="), []byte(body.Escape(string(data)))...)
	return c.api.post(ctx, "/nodes/"+vmr.node.String()+"/qemu/"+vmr.vmId.String()+"/agent/file-write", &b)
}

// minimum length of 5
type PamPassword string

const PamPassword_Error_Invalid = "password has a minimum length of 5"

func (p PamPassword) String() string { return string(p) }

func (p PamPassword) Validate() error {
	if len(p) < 5 {
		return newValidateError(PamPassword_Error_Invalid)
	}
	return nil
}
