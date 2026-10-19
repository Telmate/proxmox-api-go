package proxmox

import (
	"errors"
	"strings"
)

type TpmState struct {
	Delete  bool        `json:"remove,omitempty"`  // If true, the tpmstate will be deleted.
	Storage string      `json:"storage"`           // TODO change to proper type once the type is added.
	Version *TpmVersion `json:"version,omitempty"` // Changing version will delete the current tpmstate and create a new one. Optional during update, required during create.
}

const TmpState_Error_VersionRequired string = "version is required"

func (t TpmState) mapToApi(params map[string]any, currentTpm *TpmState) string {
	if t.Delete {
		return qemuApiKeyTPM
	}
	if currentTpm == nil { // create
		params[qemuApiKeyTPM] = t.Storage + ":1,version=" + t.Version.mapToApi()
	}
	return ""
}

func (t TpmState) markChanges(currentTpm TpmState, delete *strings.Builder) (disk *qemuDiskMove) {
	if t.Delete { // Explicit delete happens elsewhere
		return nil
	}
	if t.Version != nil && t.Version.mapToApi() != string(*currentTpm.Version) {
		delete.WriteString(comma + qemuApiKeyTPM)
		return nil
	}
	if t.Storage != currentTpm.Storage {
		return &qemuDiskMove{Storage: t.Storage, Id: qemuApiKeyTPM}
	}
	return nil
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

func (raw *rawConfigQemu) GetTPM() *TpmState {
	if v, ok := raw.a[qemuApiKeyTPM]; ok {
		setting := splitStringOfSettings(v.(string))
		splitString := strings.Split(v.(string), ":")
		tmp := TpmState{}
		if len(splitString) > 1 {
			tmp.Storage = splitString[0]
		}
		if v, ok := setting["version"]; ok {
			tmp.Version = new(TpmVersion(v))
		}
		return &tmp
	}
	return nil
}
