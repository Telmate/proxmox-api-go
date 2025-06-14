package proxmox

import (
	"errors"
	"strings"
)

type LxcFeatures struct {
	Privileged   *PrivilegedFeatures   `json:"privileged,omitempty"`   // Mutually exclusive with Unprivileged
	Unprivileged *UnprivilegedFeatures `json:"unprivileged,omitempty"` // Mutually exclusive with Privileged
}

const (
	LxcFeatures_Error_MutuallyExclusive        = "privileged and unprivileged features are mutually exclusive"
	LxcFeatures_Error_PrivilegedInUnprivileged = "privileged features cannot be set in unprivileged containers"
	LxcFeatures_Error_UnprivilegedInPrivileged = "unprivileged features cannot be set in privileged containers"
)

func (config LxcFeatures) mapToApiCreate(params map[string]any) {
	if config.Privileged != nil {
		config.Privileged.mapToApiCreate(params)
	} else if config.Unprivileged != nil {
		config.Unprivileged.mapToApiCreate(params)
	}
}

func (config LxcFeatures) mapToApiUpdate(current LxcFeatures, params map[string]interface{}) string {
	if config.Privileged != nil && current.Privileged != nil {
		return config.Privileged.mapToApiUpdate(*current.Privileged, params)
	} else if config.Unprivileged != nil && current.Unprivileged != nil {
		return config.Unprivileged.mapToApiUpdate(*current.Unprivileged, params)
	}
	return ""
}

func (config LxcFeatures) Validate(privileged bool) error {
	if config.Privileged != nil && config.Unprivileged != nil {
		return errors.New(LxcFeatures_Error_MutuallyExclusive)
	}
	if config.Privileged != nil && !privileged {
		return errors.New(LxcFeatures_Error_PrivilegedInUnprivileged)
	}
	if config.Unprivileged != nil && privileged {
		return errors.New(LxcFeatures_Error_UnprivilegedInPrivileged)
	}
	return nil
}

type lxcFeatures [lxcFeaturesLength]bool

const (
	lxcFeaturesCreateDeviceNodes int = iota
	lxcFeaturesFUSE
	lxcFeaturesKeyCtl
	lxcFeaturesNFS
	lxcFeaturesNesting
	lxcFeaturesSMB

	lxcFeaturesLength // must be last
)

func (features lxcFeatures) String() (settings string) { // String is for fmt.Stringer.
	if features[lxcFeaturesCreateDeviceNodes] {
		settings += ",mknod=1"
	}
	if features[lxcFeaturesFUSE] {
		settings += ",fuse=1"
	}
	if features[lxcFeaturesKeyCtl] {
		settings += ",keyctl=1"
	}
	if features[lxcFeaturesNFS] {
		settings += ",mount=nfs"
		if features[lxcFeaturesSMB] {
			settings += ";cifs"
		}
	} else if features[lxcFeaturesSMB] {
		settings += ",mount=cifs"
	}
	if features[lxcFeaturesNesting] {
		settings += ",nesting=1"
	}
	return
}

type PrivilegedFeatures struct {
	CreateDeviceNodes *bool `json:"create_device_nodes,omitempty"` // Never nil when returned
	FUSE              *bool `json:"fuse,omitempty"`                // Never nil when returned
	NFS               *bool `json:"nfs,omitempty"`                 // Never nil when returned
	Nesting           *bool `json:"nesting,omitempty"`             // Never nil when returned
	SMB               *bool `json:"smb,omitempty"`                 // Never nil when returned
}

func (config PrivilegedFeatures) mapToApiCreate(params map[string]any) {
	var usedConfig lxcFeatures
	usedConfig = config.mapToApiIntermediary(usedConfig)
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
	}
}

func (config PrivilegedFeatures) mapToApiUpdate(current PrivilegedFeatures, params map[string]any) string {
	var usedConfig, currentConfig lxcFeatures
	usedConfig = current.mapToApiIntermediary(usedConfig)
	currentConfig = usedConfig
	usedConfig = config.mapToApiIntermediary(usedConfig)
	if usedConfig == currentConfig {
		return ""
	}
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
		return ""
	}
	return "," + lxcApiKeyFeatures
}

func (config PrivilegedFeatures) mapToApiIntermediary(usedConfig lxcFeatures) lxcFeatures {
	if config.CreateDeviceNodes != nil {
		usedConfig[lxcFeaturesCreateDeviceNodes] = *config.CreateDeviceNodes
	}
	if config.FUSE != nil {
		usedConfig[lxcFeaturesFUSE] = *config.FUSE
	}
	if config.NFS != nil {
		usedConfig[lxcFeaturesNFS] = *config.NFS
	}
	if config.Nesting != nil {
		usedConfig[lxcFeaturesNesting] = *config.Nesting
	}
	if config.SMB != nil {
		usedConfig[lxcFeaturesSMB] = *config.SMB
	}
	return usedConfig
}

type UnprivilegedFeatures struct {
	CreateDeviceNodes *bool `json:"create_device_nodes,omitempty"` // Never nil when returned
	FUSE              *bool `json:"fuse,omitempty"`                // Never nil when returned
	KeyCtl            *bool `json:"keyctl,omitempty"`              // Never nil when returned
	Nesting           *bool `json:"nesting,omitempty"`             // Never nil when returned
}

func (config UnprivilegedFeatures) mapToApiCreate(params map[string]any) {
	var usedConfig lxcFeatures
	usedConfig = config.mapToApiIntermediary(usedConfig)
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
	}
}

func (config UnprivilegedFeatures) mapToApiUpdate(current UnprivilegedFeatures, params map[string]any) string {
	var usedConfig, currentConfig lxcFeatures
	usedConfig = current.mapToApiIntermediary(usedConfig)
	currentConfig = usedConfig
	usedConfig = config.mapToApiIntermediary(usedConfig)
	if usedConfig == currentConfig {
		return ""
	}
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
		return ""
	}
	return "," + lxcApiKeyFeatures
}

func (config UnprivilegedFeatures) mapToApiIntermediary(usedConfig lxcFeatures) lxcFeatures {
	if config.CreateDeviceNodes != nil {
		usedConfig[lxcFeaturesCreateDeviceNodes] = *config.CreateDeviceNodes
	}
	if config.FUSE != nil {
		usedConfig[lxcFeaturesFUSE] = *config.FUSE
	}
	if config.KeyCtl != nil {
		usedConfig[lxcFeaturesKeyCtl] = *config.KeyCtl
	}
	if config.Nesting != nil {
		usedConfig[lxcFeaturesNesting] = *config.Nesting
	}
	return usedConfig
}

func (raw RawConfigLXC) Features() *LxcFeatures {
	var features lxcFeatures
	var set bool
	if v, isSet := raw[lxcApiKeyFeatures]; isSet {
		settings := splitStringOfSettings(v.(string))
		if v, isSet := settings["mknod"]; isSet {
			features[lxcFeaturesCreateDeviceNodes] = v == "1"
			set = true
		}
		if v, isSet := settings["fuse"]; isSet {
			features[lxcFeaturesFUSE] = v == "1"
			set = true
		}
		if v, isSet := settings["keyctl"]; isSet {
			features[lxcFeaturesKeyCtl] = v == "1"
			set = true
		}
		if v, isSet := settings["mount"]; isSet {
			options := strings.Split(v, ";")
			switch options[0] {
			case "cifs":
				features[lxcFeaturesSMB] = true
				if len(options) == 2 {
					features[lxcFeaturesNFS] = true
				}
			case "nfs":
				features[lxcFeaturesNFS] = true
				if len(options) == 2 {
					features[lxcFeaturesSMB] = true
				}
			}
			set = true
		}
		if v, isSet := settings["nesting"]; isSet {
			features[lxcFeaturesNesting] = v == "1"
			set = true
		}
	}
	if !set {
		return nil
	}
	if raw.isPrivileged() {
		return &LxcFeatures{
			Privileged: &PrivilegedFeatures{
				CreateDeviceNodes: &features[lxcFeaturesCreateDeviceNodes],
				FUSE:              &features[lxcFeaturesFUSE],
				NFS:               &features[lxcFeaturesNFS],
				Nesting:           &features[lxcFeaturesNesting],
				SMB:               &features[lxcFeaturesSMB]}}
	}
	return &LxcFeatures{
		Unprivileged: &UnprivilegedFeatures{
			KeyCtl:            &features[lxcFeaturesKeyCtl],
			CreateDeviceNodes: &features[lxcFeaturesCreateDeviceNodes],
			FUSE:              &features[lxcFeaturesFUSE],
			Nesting:           &features[lxcFeaturesNesting]}}
}
