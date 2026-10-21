package proxmox

import (
	"context"
	"time"
)

func (c *guestClient) Status(ctx context.Context, vmr VmRef) (RawGuestStatus, error) {
	ref := vmr.toGuestRef()
	_, err := ref.check(ctx, nil, c.api)
	if err != nil {
		return nil, err
	}
	return ref.status(ctx, c.api)
}

func (c *guestClient) StatusNoCheck(ctx context.Context, vmr VmRef) (RawGuestStatus, error) {
	return vmr.toGuestRef().status(ctx, c.api)
}

// TODO add more properties to GuestStatus
type GuestStatus struct {
	HaState *string       `json:"hastate"` // TODO custom type?
	Name    GuestName     `json:"name"`
	State   PowerState    `json:"state"`
	Uptime  time.Duration `json:"uptime"`
}

type RawGuestStatus interface {
	Get() GuestStatus
	GetHaState() *string
	GetName() GuestName
	GetState() PowerState
	GetUptime() time.Duration
}

type rawGuestStatus struct {
	a map[string]any
}

var _ RawGuestStatus = (*rawGuestStatus)(nil)

func (raw *rawGuestStatus) Get() GuestStatus {
	return GuestStatus{
		HaState: raw.GetHaState(),
		Name:    raw.GetName(),
		State:   raw.GetState(),
		Uptime:  raw.GetUptime()}
}

func (raw *rawGuestStatus) GetHaState() *string {
	if v, ok := raw.a["ha"]; ok {
		if vv, ok := v.(map[string]any)["state"]; ok {
			if state, ok := vv.(string); ok {
				return &state
			}
		}
	}
	return nil
}

func (raw *rawGuestStatus) GetName() GuestName {
	if v, isSet := raw.a["name"]; isSet {
		if name, ok := v.(string); ok {
			return GuestName(name)
		}
	}
	return ""
}

func (raw *rawGuestStatus) GetState() PowerState {
	if v, isSet := raw.a["status"]; isSet {
		if state, ok := v.(string); ok {
			return PowerState(0).parse(state)
		}
	}
	return PowerStateUnknown
}

func (raw *rawGuestStatus) GetUptime() time.Duration {
	if v, isSet := raw.a["uptime"]; isSet {
		if uptime, ok := v.(float64); ok {
			return time.Duration(uptime) * time.Second
		}
	}
	return 0
}
