package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type QemuGuestAgent struct {
	Enable *bool               `json:"enable,omitempty"` // Optional
	Type   *QemuGuestAgentType `json:"type,omitempty"`   // Optional
	Freeze *bool               `json:"freeze,omitempty"` // Optional
	FsTrim *bool               `json:"trim,omitempty"`   // Optional
}

func (newSetting QemuGuestAgent) mapToAPI(currentSettings *QemuGuestAgent) string {
	var params string
	tmpEnable := "0"
	if newSetting.Enable != nil {
		if *newSetting.Enable {
			tmpEnable = "1"
		}
	} else if currentSettings != nil && currentSettings.Enable != nil {
		if *currentSettings.Enable {
			tmpEnable = "1"
		}
	}
	if newSetting.Freeze != nil {
		params += ",freeze-fs-on-backup=" + boolToIntString(*newSetting.Freeze)
	} else if currentSettings != nil && currentSettings.Freeze != nil {
		params += ",freeze-fs-on-backup=" + boolToIntString(*currentSettings.Freeze)
	}
	if newSetting.FsTrim != nil {
		params += ",fstrim_cloned_disks=" + boolToIntString(*newSetting.FsTrim)
	} else if currentSettings != nil && currentSettings.FsTrim != nil {
		params += ",fstrim_cloned_disks=" + boolToIntString(*currentSettings.FsTrim)
	}
	if newSetting.Type != nil {
		if *newSetting.Type != QemuGuestAgentType_None {
			params += ",type=" + strings.ToLower(string(*newSetting.Type))
		}
	} else if currentSettings != nil && currentSettings.Type != nil {
		params += ",type=" + strings.ToLower(string(*currentSettings.Type))
	}
	return tmpEnable + params
}

func (QemuGuestAgent) mapToSDK(params string) *QemuGuestAgent {
	config := QemuGuestAgent{}
	tmpEnable, _ := strconv.ParseBool(params[0:1])
	config.Enable = &tmpEnable
	tmpParams := splitStringOfSettings(params)
	if v, isSet := tmpParams["freeze-fs-on-backup"]; isSet {
		tmpBool, _ := strconv.ParseBool(v)
		config.Freeze = &tmpBool
	}
	if v, isSet := tmpParams["fstrim_cloned_disks"]; isSet {
		tmpBool, _ := strconv.ParseBool(v)
		config.FsTrim = &tmpBool
	}
	if v, isSet := tmpParams["type"]; isSet {
		config.Type = util.Pointer(QemuGuestAgentType(v))
	}
	return &config
}

func (setting QemuGuestAgent) Validate() error {
	if setting.Type != nil {
		return setting.Type.Validate()
	}
	return nil
}

type QemuGuestAgentType string // enum

const (
	QemuGuestAgentType_Isa           QemuGuestAgentType = "isa"
	QemuGuestAgentType_VirtIO        QemuGuestAgentType = "virtio"
	QemuGuestAgentType_None          QemuGuestAgentType = "" // Used to unset the value. Proxmox enforces the default.
	QemuGuestAgentType_Error_Invalid string             = `invalid qemu guest agent type, should one of [` + string(QemuGuestAgentType_Isa) + `, ` + string(QemuGuestAgentType_VirtIO) + `, ""]`
)

func (q QemuGuestAgentType) Validate() error {
	if q == QemuGuestAgentType_None {
		return nil
	}
	switch QemuGuestAgentType(strings.ToLower(string(q))) {
	case QemuGuestAgentType_Isa, QemuGuestAgentType_VirtIO:
		return nil
	}
	return errors.New(QemuGuestAgentType_Error_Invalid)
}
