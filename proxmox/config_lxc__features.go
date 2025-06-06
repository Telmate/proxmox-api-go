package proxmox

import "strings"

type LxcFeatures struct {
	CreateDeviceNodes *bool `json:"create_device_nodes,omitempty"`
	FUSE              *bool `json:"fuse,omitempty"`
	KeyCtl            *bool `json:"keyctl,omitempty"`
	NFS               *bool `json:"nfs,omitempty"`
	Nesting           *bool `json:"nesting,omitempty"`
	SMB               *bool `json:"smb,omitempty"`
}

func (config LxcFeatures) mapToApiCreate(params map[string]any) {
	var usedConfig lxcFeatures
	usedConfig = config.mapToApiIntermediary_Unsafe(usedConfig)
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
	}
}

func (config LxcFeatures) mapToApiUpdate(current LxcFeatures, params map[string]interface{}) string {
	var usedConfig, currentConfig lxcFeatures
	usedConfig = current.mapToApiIntermediary_Unsafe(usedConfig)
	currentConfig = usedConfig
	usedConfig = config.mapToApiIntermediary_Unsafe(usedConfig)
	if usedConfig == currentConfig {
		return ""
	}
	if v := usedConfig.String(); v != "" {
		params[lxcApiKeyFeatures] = v[1:]
		return ""
	}
	return "," + lxcApiKeyFeatures
}

func (config LxcFeatures) mapToApiIntermediary_Unsafe(usedConfig lxcFeatures) lxcFeatures {
	if config.CreateDeviceNodes != nil {
		usedConfig[lxcFeaturesCreateDeviceNodes] = *config.CreateDeviceNodes
	}
	if config.FUSE != nil {
		usedConfig[lxcFeaturesFUSE] = *config.FUSE
	}
	if config.KeyCtl != nil {
		usedConfig[lxcFeaturesKeyCtl] = *config.KeyCtl
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
	return &LxcFeatures{
		CreateDeviceNodes: &features[lxcFeaturesCreateDeviceNodes],
		FUSE:              &features[lxcFeaturesFUSE],
		KeyCtl:            &features[lxcFeaturesKeyCtl],
		NFS:               &features[lxcFeaturesNFS],
		Nesting:           &features[lxcFeaturesNesting],
		SMB:               &features[lxcFeaturesSMB]}
}
