package proxmox

import (
	"errors"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type TpmState struct {
	Delete  bool        `json:"remove,omitempty"`  // If true, the tpmstate will be deleted.
	Storage string      `json:"storage"`           // TODO change to proper type once the type is added.
	Version *TpmVersion `json:"version,omitempty"` // Changing version will delete the current tpmstate and create a new one. Optional during update, required during create.
}

const TmpState_Error_VersionRequired string = "version is required"

func (t TpmState) mapToApi(params map[string]interface{}, currentTpm *TpmState) string {
	if t.Delete {
		return "tpmstate0"
	}
	if currentTpm == nil { // create
		params["tpmstate0"] = t.Storage + ":1,version=" + t.Version.mapToApi()
	}
	return ""
}

func (TpmState) mapToSDK(param string) *TpmState {
	setting := splitStringOfSettings(param)
	splitString := strings.Split(param, ":")
	tmp := TpmState{}
	if len(splitString) > 1 {
		tmp.Storage = splitString[0]
	}
	if itemValue, isSet := setting["version"]; isSet {
		tmp.Version = util.Pointer(TpmVersion(itemValue))
	}
	return &tmp

}

func (t TpmState) markChanges(currentTpm TpmState) (delete string, disk *qemuDiskMove) {
	if t.Delete {
		return "", nil
	}
	if t.Version != nil && t.Version.mapToApi() != string(*currentTpm.Version) {
		return "tpmstate0", nil
	}
	if t.Storage != currentTpm.Storage {
		return "", &qemuDiskMove{Storage: t.Storage, Id: "tpmstate0"}
	}
	return "", nil
}

func (t TpmState) Validate(current *TpmState) error {
	if t.Storage == "" {
		return errors.New("storage is required")
	}
	if t.Version == nil {
		if current == nil { // create
			return errors.New(TmpState_Error_VersionRequired)
		}
	} else {
		if err := t.Version.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type TpmVersion string // enum

const (
	TpmVersion_1_2           TpmVersion = "v1.2"
	TpmVersion_2_0           TpmVersion = "v2.0"
	TpmVersion_Error_Invalid string     = "enum TmpVersion should be one of: " + string(TpmVersion_1_2) + ", " + string(TpmVersion_2_0)
)

func (t TpmVersion) mapToApi() string {
	switch t {
	case TpmVersion_1_2, "1.2":
		return string(t)
	case TpmVersion_2_0, "v2", "2.0", "2":
		return string(TpmVersion_2_0)
	}
	return ""
}

func (t TpmVersion) Validate() error {
	if t.mapToApi() == "" {
		return errors.New(TpmVersion_Error_Invalid)
	}
	return nil
}
