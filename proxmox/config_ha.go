package proxmox

import (
	"errors"
	"strconv"
	"strings"
)

type GuestHA struct {
	Comment     *string      `json:"comment"`          // Description.
	Delete      bool         `json:"remove,omitempty"` // When true, remove HA settings for the Guest.
	Group       *HaGroupName `json:"group"`            // May be empty, in which case the guest is not part of a group.
	Reallocates *HaRelocate  `json:"reallocates"`
	Restarts    *HaRestart   `json:"restarts"`
	State       *HaState     `json:"state"`
}

// TODO change type when we have a custom type for guestID
func (g GuestHA) mapToApi(guestID int) map[string]interface{} {
	params := map[string]interface{}{}
	if g.State != nil {
		params["state"] = string(*g.State)
	}
	if g.Restarts != nil {
		params["max_restart"] = int(*g.Restarts)
	}
	if g.Reallocates != nil {
		params["max_relocate"] = int(*g.Reallocates)
	}
	if guestID > 0 { // Update
		if g.Comment != nil {
			params["comment"] = *g.Comment
		}
		params["sid"] = guestID
		if g.Group != nil {
			if *g.Group != "" {
				params["group"] = string(*g.Group)
			} else {
				params["delete"] = "group"
			}
		}
	} else { // Create
		if g.Comment != nil && *g.Comment != "" {
			params["comment"] = *g.Comment
		}
		if g.Group != nil && *g.Group != "" {
			params["group"] = string(*g.Group)
		}
	}
	return params
}

func (GuestHA) mapToSDK(params map[string]interface{}) (config GuestHA) {
	if itemValue, isSet := params["comment"]; isSet {
		comment := itemValue.(string)
		config.Comment = &comment
	}
	if itemValue, isSet := params["group"]; isSet {
		group := HaGroupName(itemValue.(string))
		config.Group = &group
	}
	if itemValue, isSet := params["max_relocate"]; isSet {
		relocate := HaRelocate(itemValue.(float64))
		config.Reallocates = &relocate
	}
	if itemValue, isSet := params["max_restart"]; isSet {
		restarts := HaRestart(itemValue.(float64))
		config.Restarts = &restarts
	}
	if itemValue, isSet := params["state"]; isSet {
		state := HaState(itemValue.(string))
		config.State = &state
	}
	return
}

func (g GuestHA) Set(vmr *VmRef, client *Client) (err error) {
	if err = g.Validate(); err != nil {
		return
	}
	ha, err := NewGuestHAFromApi(vmr, client)
	if err != nil {
		return
	}
	return g.Set_Unsafe(ha, vmr, client)
}

func (g GuestHA) Set_Unsafe(current *GuestHA, vmr *VmRef, client *Client) (err error) {
	if current == nil { // create
		if err = client.Post(g.mapToApi(vmr.vmId), "/cluster/ha/resources"); err != nil {
			return
		}
		if g.State != nil {
			vmr.haState = *g.State
		}
		if g.Group != nil {
			vmr.haGroup = *g.Group
		}
	}
	if g.Delete { // delete
		if err = client.Delete("/cluster/ha/resources/" + strconv.FormatInt(int64(vmr.vmId), 10)); err != nil {
			return
		}
		vmr.haState = ""
		vmr.haGroup = ""
		return
	}
	// update
	if err = client.Put(g.mapToApi(0), "/cluster/ha/resources/"+strconv.FormatInt(int64(vmr.vmId), 10)); err != nil {
		return
	}
	if g.State != nil {
		vmr.haState = *g.State
	}
	if g.Group != nil {
		vmr.haGroup = *g.Group
	}
	return
}

func (g GuestHA) Validate() (err error) {
	if g.Group != nil && *g.Group != "" {
		if err = g.Group.Validate(); err != nil {
			return
		}
	}
	if g.State != nil {
		if err = g.State.Validate(); err != nil {
			return
		}
	}
	if g.Reallocates != nil {
		if err = g.Reallocates.Validate(); err != nil {
			return
		}
	}
	if g.Restarts != nil {
		if err = g.Restarts.Validate(); err != nil {
			return
		}
	}
	return
}

type HaGroupName string

const (
	HaGroupName_Error_Length              string = "HaGroupName should be at least 2 characters long"
	HaGroupName_Error_Illegal_Start       string = "HaGroupName may only start with the following characters:" + haGroupName_Characters_Legal_Starting
	HaGroupName_Error_Illegal_End         string = "HaGroupName may only end with the following characters:" + haGroupName_Characters_Legal_Ending
	HaGroupName_Error_Illegal             string = "HaGroupName may only contain the following characters:" + haGroupName_Characters_Legal
	haGroupName_Characters_Legal_Starting string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	haGroupName_Characters_Legal_Ending   string = haGroupName_Characters_Legal_Starting + "12345678890"
	haGroupName_Characters_Legal          string = haGroupName_Characters_Legal_Starting + ".-_" + haGroupName_Characters_Legal_Ending
)

func (g HaGroupName) Validate() error {
	if len(g) < 2 {
		return errors.New(HaGroupName_Error_Length)
	}
	if !strings.Contains(haGroupName_Characters_Legal_Starting, string(g[0])) {
		return errors.New(HaGroupName_Error_Illegal_Start)
	}
	if !strings.Contains(haGroupName_Characters_Legal_Ending, string(g[len(g)-1])) {
		return errors.New(HaGroupName_Error_Illegal_End)
	}
	for _, c := range g {
		if !strings.Contains(haGroupName_Characters_Legal, string(c)) {
			return errors.New(HaGroupName_Error_Illegal)
		}
	}
	return nil
}

type HaRelocate uint8

const HaRelocate_Error_UpperBound string = "HaRelocate should be less or equal to 10"

func (r HaRelocate) Validate() error {
	if r > 10 {
		return errors.New(HaRelocate_Error_UpperBound)
	}
	return nil
}

type HaRestart uint8

const HaRestart_Error_UpperBound string = "HaRestart should be less or equal to 10"

func (r HaRestart) Validate() error {
	if r > 10 {
		return errors.New(HaRestart_Error_UpperBound)
	}
	return nil
}

type HaState string // enum

const (
	HaState_Disabled      HaState = "disabled"
	HaState_Ignored       HaState = "ignored"
	HaState_Started       HaState = "started"
	HaState_Stopped       HaState = "stopped"
	HaState_Error_Invalid string  = string("HaState should be one of: " + HaState_Disabled + "," + HaState_Ignored + "," + HaState_Started + "," + HaState_Stopped)
)

func (s HaState) Validate() error {
	switch s {
	case HaState_Disabled, HaState_Ignored, HaState_Started, HaState_Stopped:
		return nil
	}
	return errors.New(HaState_Error_Invalid)
}

func NewGuestHAFromApi(vmr *VmRef, client *Client) (*GuestHA, error) {
	if err := vmr.nilCheck(); err != nil {
		return nil, err
	}
	if client == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	guestID := strconv.FormatInt(int64(vmr.vmId), 10)
	params, err := client.GetItemConfigMapStringInterface("/cluster/ha/resources/"+guestID, "", "")
	if err != nil {
		if err.Error() == noSuchResource+" 'vm:"+guestID+"'" || err.Error() == noSuchResource+" 'ct:"+guestID+"'" {
			return nil, nil
		}
		return nil, err
	}
	ha := GuestHA{}.mapToSDK(params)
	vmr.haState = *ha.State
	if ha.Group != nil {
		vmr.haGroup = *ha.Group
	} else {
		vmr.haGroup = ""
	}
	return &ha, nil
}
