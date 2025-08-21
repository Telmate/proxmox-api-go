package proxmox

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type LxcBootMount struct {
	ACL             *TriBool // Never nil when returned
	Options         *LxcBootMountOptions
	Quota           *bool         // Only for privileged guests.
	Replicate       *bool         // Never nil when returned
	SizeInKibibytes *LxcMountSize // Required during creation, never nil when returned
	Storage         *string       // Required during creation, never nil when returned
	rawDisk         string
}

const (
	LxcBootMount_Error_NoSizeDuringCreation    = "size must be set during creation"
	LxcBootMount_Error_NoStorageDuringCreation = "storage must be set during creation"
	LxcBootMount_Error_QuotaNotPrivileged      = "quota can only be set for privileged guest"
)

func (mount LxcBootMount) combine(usedConfig LxcBootMount) LxcBootMount {
	if mount.Storage != nil {
		usedConfig.Storage = mount.Storage
	}
	if mount.SizeInKibibytes != nil {
		usedConfig.SizeInKibibytes = mount.SizeInKibibytes
	}
	if mount.Options != nil {
		if usedConfig.Options == nil {
			usedConfig.Options = &LxcBootMountOptions{}
		}
		if mount.Options.Discard != nil {
			usedConfig.Options.Discard = mount.Options.Discard
		}
		if mount.Options.LazyTime != nil {
			usedConfig.Options.LazyTime = mount.Options.LazyTime
		}
		if mount.Options.NoATime != nil {
			usedConfig.Options.NoATime = mount.Options.NoATime
		}
		if mount.Options.NoSuid != nil {
			usedConfig.Options.NoSuid = mount.Options.NoSuid
		}
	}
	if mount.Replicate != nil {
		usedConfig.Replicate = mount.Replicate
	}
	if mount.Quota != nil {
		usedConfig.Quota = mount.Quota
	}
	if mount.ACL != nil {
		usedConfig.ACL = mount.ACL
	}
	return usedConfig
}

func (config LxcBootMount) mapToApiCreate(privileged bool) string {
	rootFs := config.string(privileged)
	if config.Storage != nil && config.SizeInKibibytes != nil {
		var size float64
		if *config.SizeInKibibytes < gibiByteLxc { // only approximate if the size is less than 1 GiB
			size = approximateDiskSize(int64(*config.SizeInKibibytes))
		} else {
			size = float64(*config.SizeInKibibytes / gibiByteLxc)
		}
		rootFs = *config.Storage + ":" + strconv.FormatFloat(size, 'f', -1, 64) + rootFs
	}
	return rootFs
}

func (config LxcBootMount) mapToApiUpdate(current LxcBootMount, privileged bool, params map[string]any) {
	currentRootFs := current.string(privileged)
	var usedConfig LxcBootMount
	usedConfig = config.combine(current.combine(usedConfig))
	rootFs := usedConfig.string(privileged)
	if currentRootFs == rootFs { // No changes
		return
	}
	if usedConfig.SizeInKibibytes != nil {
		rootFs += ",size=" + usedConfig.SizeInKibibytes.String()
	}
	if usedConfig.Storage != nil {
		rootFs = *usedConfig.Storage + ":" + current.rawDisk + rootFs
		if current.Storage != nil && rootFs == *current.Storage+":"+current.rawDisk+current.string(privileged) {
			return
		}
	}
	params[lxcApiKeyRootFS] = rootFs
}

func (config LxcBootMount) markMountChanges_Unsafe(current *LxcBootMount) lxcUpdateChanges {
	changes := lxcUpdateChanges{}
	if config.SizeInKibibytes != nil && *config.SizeInKibibytes > *current.SizeInKibibytes { // Resize
		changes.resize = []lxcMountResize{{
			sizeInKibibytes: *config.SizeInKibibytes,
			id:              "rootfs"}}
	}
	if config.Storage != nil && *config.Storage != *current.Storage { // Move
		changes.move = []lxcMountMove{{
			storage: *config.Storage,
			id:      "rootfs"}}
	}
	return changes
}

func (config LxcBootMount) string(privileged bool) (rootFs string) {
	// zfs  // local-zfs:subvol-101-disk-0
	// ext4 // local-ext4:101/vm-101-disk-0.raw
	// lvm  // local-lvm:vm-101-disk-0
	if config.ACL != nil {
		switch *config.ACL {
		case TriBoolTrue:
			rootFs += ",acl=1"
		case TriBoolFalse:
			rootFs += ",acl=0"
		}
	}
	if config.Options != nil {
		var options string
		if config.Options.Discard != nil && *config.Options.Discard {
			options += ";discard"
		}
		if config.Options.LazyTime != nil && *config.Options.LazyTime {
			options += ";lazytime"
		}
		if config.Options.NoATime != nil && *config.Options.NoATime {
			options += ";noatime"
		}
		if config.Options.NoSuid != nil && *config.Options.NoSuid {
			options += ";nosuid"
		}
		if options != "" {
			rootFs += ",mountoptions=" + options[1:]
		}
	}
	if privileged && config.Quota != nil && *config.Quota {
		rootFs += ",quota=1"
	}
	if config.Replicate != nil && !*config.Replicate {
		rootFs += ",replicate=0"
	}
	return
}

func (config LxcBootMount) Validate(current *LxcBootMount, privileged bool) error {
	var err error
	if config.ACL != nil {
		if err = config.ACL.Validate(); err != nil {
			return err
		}
	}
	if current == nil { // Create
		if config.Storage == nil {
			return errors.New(LxcBootMount_Error_NoStorageDuringCreation)
		}
		if config.SizeInKibibytes == nil {
			return errors.New(LxcBootMount_Error_NoSizeDuringCreation)
		}
	}
	if config.SizeInKibibytes != nil {
		err = config.SizeInKibibytes.Validate()
	}
	if config.Quota != nil && !privileged {
		return errors.New(LxcBootMount_Error_QuotaNotPrivileged)
	}
	return err
}

type LxcBootMountOptions struct {
	Discard  *bool `json:"discard,omitempty"`   // Never nil when returned
	LazyTime *bool `json:"lazy_time,omitempty"` // Never nil when returned
	NoATime  *bool `json:"no_atime,omitempty"`  // Never nil when returned
	NoSuid   *bool `json:"no_suid,omitempty"`   // Never nil when returned
}

type LxcMounts map[LxcMountID]LxcMount

const LxcMountsAmount = 256

func (config LxcMounts) markMountChanges(current LxcMounts) lxcUpdateChanges {
	changes := lxcUpdateChanges{}
	for id, v := range config {
		vv, isSet := current[id]
		if !isSet {
			continue
		}
		if v.Detach {
			changes.offState = true
			continue
		}
		if v.BindMount != nil {
			if vv.BindMount != nil {
				if v.BindMount.HostPath != nil && vv.BindMount.HostPath != nil && *v.BindMount.HostPath != *vv.BindMount.HostPath {
					changes.offState = true // HostPath changed
					continue
				}
				if v.BindMount.GuestPath != nil && vv.BindMount.GuestPath != nil && *v.BindMount.GuestPath != *vv.BindMount.GuestPath {
					changes.offState = true // GuestPath changed
					continue
				}
			}
			if vv.DataMount != nil {
				changes.offState = true // BindMount will replace DataMount
				continue
			}
			continue // bindmount will be newly created
		}
		if v.DataMount != nil {
			if vv.DataMount != nil {
				if v.DataMount.SizeInKibibytes != nil {
					if *v.DataMount.SizeInKibibytes > *vv.DataMount.SizeInKibibytes { // Resize
						changes.resize = append(changes.resize, lxcMountResize{
							id:              lxcPrefixApiKeyMount + id.String(),
							sizeInKibibytes: *v.DataMount.SizeInKibibytes})
					} else if *v.DataMount.SizeInKibibytes < *vv.DataMount.SizeInKibibytes { // Recreate is handled elsewhere
						changes.offState = true
						continue
					}
				}
				if v.DataMount.Storage != nil {
					if *v.DataMount.Storage != *vv.DataMount.Storage {
						changes.offState = true
						changes.move = append(changes.move, lxcMountMove{
							id:      lxcPrefixApiKeyMount + id.String(),
							storage: *v.DataMount.Storage})
					}
				}
				continue
			}
			if vv.BindMount != nil {
				changes.offState = true // DataMount will replace BindMount
				continue
			}
			continue // DataMount will be newly created
		}
	}
	return changes
}

func (config LxcMounts) mapToAPICreate(privileged bool, params map[string]any) {
	for id, v := range config {
		if v.Detach {
			continue
		}
		if v.DataMount != nil {
			v.DataMount.mapToAPICreate(id, privileged, params)
			continue
		}
		if v.BindMount != nil {
			params[lxcPrefixApiKeyMount+id.String()] = v.BindMount.string()
		}
	}
}

func (config LxcMounts) mapToAPIUpdate(current LxcMounts, privileged bool, params map[string]any) (delete string) {
	for id, v := range config {
		if vv, isSet := current[id]; isSet { // Update
			if v.Detach {
				delete += "," + lxcPrefixApiKeyMount + id.String()
				continue
			}
			if v.DataMount != nil {
				if vv.DataMount != nil {
					v.DataMount.mapToAPIUpdate(*vv.DataMount, id, privileged, params)
				} else {
					v.DataMount.mapToAPICreate(id, privileged, params)
				}
			} else if v.BindMount != nil {
				if vv.BindMount != nil {
					v.BindMount.mapToAPIUpdate(*vv.BindMount, id, params)
				} else {
					params[lxcPrefixApiKeyMount+id.String()] = v.BindMount.string()
				}
			}
		} else {
			if v.Detach {
				continue
			}
			if v.DataMount != nil {
				v.DataMount.mapToAPICreate(id, privileged, params)
			} else if v.BindMount != nil {
				params[lxcPrefixApiKeyMount+id.String()] = v.BindMount.string()
			}
		}
	}
	return
}

func (config LxcMounts) mapToSdkBindMount(id LxcMountID, params string) {
	var hostPath LxcHostPath
	var path LxcMountPath
	readOnly := false
	replicate := true
	mount := LxcBindMount{
		HostPath:  &hostPath,
		ReadOnly:  &readOnly,
		Replicate: &replicate,
		GuestPath: &path,
	}
	if index := strings.IndexRune(params, ','); index != -1 {
		hostPath = LxcHostPath(params[:index])
		settings := splitStringOfSettings(params[index:])
		if v, isSet := settings["mountoptions"]; isSet {
			mountOptionsPointer := &LxcMountOptions{}
			mountOptionsPointer.mapToSDK(v)
			mount.Options = mountOptionsPointer
		}
		if v, isSet := settings["mp"]; isSet {
			path = LxcMountPath(v)
		}
		if v, isSet := settings["ro"]; isSet {
			readOnly = v == "1"
		}
		if v, isSet := settings["replicate"]; isSet {
			replicate = v == "1"
		}
		config[id] = LxcMount{BindMount: &mount}
	}
}

func (config LxcMounts) mapToSdkDataMount(id LxcMountID, params string, privileged bool) {
	acl := TriBoolNone
	backup, quota, readOnly := false, false, false
	replicate := true
	var path LxcMountPath
	var size LxcMountSize
	mount := LxcDataMount{
		ACL:             &acl,
		Backup:          &backup,
		Path:            &path,
		ReadOnly:        &readOnly,
		Replicate:       &replicate,
		SizeInKibibytes: &size,
		Storage:         util.Pointer(params[:strings.IndexRune(params, ':')])}
	if privileged {
		mount.Quota = &quota
	}
	if index := strings.IndexRune(params, ','); index != -1 {
		mount.rawDisk = params[:index]
		settings := splitStringOfSettings(params[index:])
		if v, isSet := settings["acl"]; isSet {
			switch v {
			case "1":
				acl = TriBoolTrue
			case "0":
				acl = TriBoolFalse
			}
		}
		if v, isSet := settings["backup"]; isSet {
			backup = v == "1"
		}
		if v, isSet := settings["mountoptions"]; isSet {
			mountOptionsPointer := &LxcMountOptions{}
			mountOptionsPointer.mapToSDK(v)
			mount.Options = mountOptionsPointer
		}
		if v, isSet := settings["mp"]; isSet {
			path = LxcMountPath(v)
		}
		if v, isSet := settings["quota"]; isSet {
			quota = v == "1"
		}
		if v, isSet := settings["ro"]; isSet {
			readOnly = v == "1"
		}
		if v, isSet := settings["replicate"]; isSet {
			replicate = v == "1"
		}
		if v, isSet := settings["size"]; isSet {
			size = LxcMountSize(parseDiskSize(v))
		}
	} else {
		mount.rawDisk = params
	}
	config[id] = LxcMount{DataMount: &mount}
}

func (config LxcMounts) Validate(current LxcMounts, privileged bool) error {
	if current != nil {
		return config.validateUpdate(current, privileged)
	}
	return config.validateCreate(privileged)
}

func (config LxcMounts) validateCreate(privileged bool) error {
	for _, v := range config {
		if v.Detach {
			return nil
		}
		if err := v.validateCreate(privileged); err != nil {
			return err
		}
	}
	return nil
}

func (config LxcMounts) validateUpdate(current LxcMounts, privileged bool) error {
	for k, v := range config {
		if v.Detach {
			continue
		}
		if _, isSet := current[k]; isSet {
			if err := v.validateUpdate(privileged); err != nil {
				return err
			}
		} else {
			if err := v.validateCreate(privileged); err != nil {
				return err
			}
		}
	}
	return nil
}

type LxcMountID uint8

const LxcMountIDMaximum = 255

const (
	LxcMountID0 LxcMountID = iota
	LxcMountID1
	LxcMountID2
	LxcMountID3
	LxcMountID4
	LxcMountID5
	LxcMountID6
	LxcMountID7
	LxcMountID8
	LxcMountID9
	LxcMountID10
	LxcMountID11
	LxcMountID12
	LxcMountID13
	LxcMountID14
	LxcMountID15
	LxcMountID16
	LxcMountID17
	LxcMountID18
	LxcMountID19
	LxcMountID20
	LxcMountID21
	LxcMountID22
	LxcMountID23
	LxcMountID24
	LxcMountID25
	LxcMountID26
	LxcMountID27
	LxcMountID28
	LxcMountID29
	LxcMountID30
	LxcMountID31
	LxcMountID32
	LxcMountID33
	LxcMountID34
	LxcMountID35
	LxcMountID36
	LxcMountID37
	LxcMountID38
	LxcMountID39
	LxcMountID40
	LxcMountID41
	LxcMountID42
	LxcMountID43
	LxcMountID44
	LxcMountID45
	LxcMountID46
	LxcMountID47
	LxcMountID48
	LxcMountID49
	LxcMountID50
	LxcMountID51
	LxcMountID52
	LxcMountID53
	LxcMountID54
	LxcMountID55
	LxcMountID56
	LxcMountID57
	LxcMountID58
	LxcMountID59
	LxcMountID60
	LxcMountID61
	LxcMountID62
	LxcMountID63
	LxcMountID64
	LxcMountID65
	LxcMountID66
	LxcMountID67
	LxcMountID68
	LxcMountID69
	LxcMountID70
	LxcMountID71
	LxcMountID72
	LxcMountID73
	LxcMountID74
	LxcMountID75
	LxcMountID76
	LxcMountID77
	LxcMountID78
	LxcMountID79
	LxcMountID80
	LxcMountID81
	LxcMountID82
	LxcMountID83
	LxcMountID84
	LxcMountID85
	LxcMountID86
	LxcMountID87
	LxcMountID88
	LxcMountID89
	LxcMountID90
	LxcMountID91
	LxcMountID92
	LxcMountID93
	LxcMountID94
	LxcMountID95
	LxcMountID96
	LxcMountID97
	LxcMountID98
	LxcMountID99
	LxcMountID100
	LxcMountID101
	LxcMountID102
	LxcMountID103
	LxcMountID104
	LxcMountID105
	LxcMountID106
	LxcMountID107
	LxcMountID108
	LxcMountID109
	LxcMountID110
	LxcMountID111
	LxcMountID112
	LxcMountID113
	LxcMountID114
	LxcMountID115
	LxcMountID116
	LxcMountID117
	LxcMountID118
	LxcMountID119
	LxcMountID120
	LxcMountID121
	LxcMountID122
	LxcMountID123
	LxcMountID124
	LxcMountID125
	LxcMountID126
	LxcMountID127
	LxcMountID128
	LxcMountID129
	LxcMountID130
	LxcMountID131
	LxcMountID132
	LxcMountID133
	LxcMountID134
	LxcMountID135
	LxcMountID136
	LxcMountID137
	LxcMountID138
	LxcMountID139
	LxcMountID140
	LxcMountID141
	LxcMountID142
	LxcMountID143
	LxcMountID144
	LxcMountID145
	LxcMountID146
	LxcMountID147
	LxcMountID148
	LxcMountID149
	LxcMountID150
	LxcMountID151
	LxcMountID152
	LxcMountID153
	LxcMountID154
	LxcMountID155
	LxcMountID156
	LxcMountID157
	LxcMountID158
	LxcMountID159
	LxcMountID160
	LxcMountID161
	LxcMountID162
	LxcMountID163
	LxcMountID164
	LxcMountID165
	LxcMountID166
	LxcMountID167
	LxcMountID168
	LxcMountID169
	LxcMountID170
	LxcMountID171
	LxcMountID172
	LxcMountID173
	LxcMountID174
	LxcMountID175
	LxcMountID176
	LxcMountID177
	LxcMountID178
	LxcMountID179
	LxcMountID180
	LxcMountID181
	LxcMountID182
	LxcMountID183
	LxcMountID184
	LxcMountID185
	LxcMountID186
	LxcMountID187
	LxcMountID188
	LxcMountID189
	LxcMountID190
	LxcMountID191
	LxcMountID192
	LxcMountID193
	LxcMountID194
	LxcMountID195
	LxcMountID196
	LxcMountID197
	LxcMountID198
	LxcMountID199
	LxcMountID200
	LxcMountID201
	LxcMountID202
	LxcMountID203
	LxcMountID204
	LxcMountID205
	LxcMountID206
	LxcMountID207
	LxcMountID208
	LxcMountID209
	LxcMountID210
	LxcMountID211
	LxcMountID212
	LxcMountID213
	LxcMountID214
	LxcMountID215
	LxcMountID216
	LxcMountID217
	LxcMountID218
	LxcMountID219
	LxcMountID220
	LxcMountID221
	LxcMountID222
	LxcMountID223
	LxcMountID224
	LxcMountID225
	LxcMountID226
	LxcMountID227
	LxcMountID228
	LxcMountID229
	LxcMountID230
	LxcMountID231
	LxcMountID232
	LxcMountID233
	LxcMountID234
	LxcMountID235
	LxcMountID236
	LxcMountID237
	LxcMountID238
	LxcMountID239
	LxcMountID240
	LxcMountID241
	LxcMountID242
	LxcMountID243
	LxcMountID244
	LxcMountID245
	LxcMountID246
	LxcMountID247
	LxcMountID248
	LxcMountID249
	LxcMountID250
	LxcMountID251
	LxcMountID252
	LxcMountID253
	LxcMountID254
	LxcMountID255
)

func (id LxcMountID) String() string { return strconv.FormatUint(uint64(id), 10) } // String is for fmt.Stringer.

type LxcMount struct {
	BindMount *LxcBindMount `json:"bind,omitempty"`
	DataMount *LxcDataMount `json:"data,omitempty"`
	Detach    bool          `json:"detach,omitempty"`
}

const LxcMountErrorMutuallyExclusive = "bindMount and dataMount are mutually exclusive"

func (config LxcMount) Validate(current *LxcMount, privileged bool) error {
	if config.Detach {
		return nil
	}
	if current != nil {
		return config.validateUpdate(privileged)
	}
	return config.validateCreate(privileged)
}

func (config LxcMount) validateCreate(privileged bool) error {
	if config.DataMount != nil {
		if config.BindMount != nil {
			return errors.New(LxcMountErrorMutuallyExclusive)
		}
		return config.DataMount.validateCreate(privileged)
	}
	if config.BindMount != nil {
		return config.BindMount.validateCreate()
	}
	return nil
}

func (config LxcMount) validateUpdate(privileged bool) error {
	if config.DataMount != nil {
		if config.BindMount != nil {
			return errors.New(LxcMountErrorMutuallyExclusive)
		}
		return config.DataMount.validateUpdate(privileged)
	}
	if config.BindMount != nil {
		return config.BindMount.validateUpdate()
	}
	return nil
}

type LxcBindMount struct {
	GuestPath *LxcMountPath    `json:"guest_path,omitempty"` // Required during creation, never nil when returned
	HostPath  *LxcHostPath     `json:"host_path,omitempty"`  // Required during creation, never nil when returned
	Options   *LxcMountOptions `json:"mount_options,omitempty"`
	ReadOnly  *bool            `json:"read_only,omitempty"` // Never nil when returned
	Replicate *bool            `json:"replicate,omitempty"` // Never nil when returned
}

const (
	LxcBindMountErrorHostPathRequired  = "host path is required for creation"
	LxcBindMountErrorGuestPathRequired = "guest path is required for creation"
)

func (config LxcBindMount) combine(current LxcBindMount) LxcBindMount {
	if config.HostPath != nil {
		current.HostPath = config.HostPath
	}
	if config.GuestPath != nil {
		current.GuestPath = config.GuestPath
	}
	if config.Options != nil {
		if current.Options == nil {
			current.Options = config.Options
		} else {
			if config.Options.Discard != nil {
				current.Options.Discard = config.Options.Discard
			}
			if config.Options.LazyTime != nil {
				current.Options.LazyTime = config.Options.LazyTime
			}
			if config.Options.NoATime != nil {
				current.Options.NoATime = config.Options.NoATime
			}
			if config.Options.NoSuid != nil {
				current.Options.NoSuid = config.Options.NoSuid
			}
			if config.Options.NoDevice != nil {
				current.Options.NoDevice = config.Options.NoDevice
			}
			if config.Options.NoExec != nil {
				current.Options.NoExec = config.Options.NoExec
			}
		}
	}
	if config.ReadOnly != nil {
		current.ReadOnly = config.ReadOnly
	}
	if config.Replicate != nil {
		current.Replicate = config.Replicate
	}
	return current
}

func (config LxcBindMount) mapToAPIUpdate(current LxcBindMount, id LxcMountID, params map[string]any) {
	var usedConfig LxcBindMount
	currentMount := current.string()
	usedConfig = config.combine(current)
	mount := usedConfig.string()
	if mount != currentMount {
		params[lxcPrefixApiKeyMount+id.String()] = mount
	}
}

func (config LxcBindMount) string() (settings string) {
	if config.HostPath != nil {
		settings = string(*config.HostPath)
	}
	if config.GuestPath != nil {
		settings += ",mp=" + config.GuestPath.String()
	}
	if config.Options != nil {
		if v := config.Options.string(); v != "" {
			settings += ",mountoptions=" + v[1:] // removes the first simicolon
		}
	}
	if config.ReadOnly != nil && *config.ReadOnly {
		settings += ",ro=1"
	}
	if config.Replicate != nil && !*config.Replicate {
		settings += ",replicate=0"
	}
	return settings
}

func (config LxcBindMount) Validate(current *LxcBindMount) error {
	if current != nil {
		return config.validateUpdate()
	}
	return config.validateCreate()
}

func (config LxcBindMount) validateCreate() error {
	if config.HostPath == nil {
		return errors.New(LxcBindMountErrorHostPathRequired)
	}
	if config.GuestPath == nil {
		return errors.New(LxcBindMountErrorGuestPathRequired)
	}
	return config.validateUpdate()
}

func (config LxcBindMount) validateUpdate() error {
	if config.HostPath != nil {
		if err := config.HostPath.Validate(); err != nil {
			return err
		}
	}
	if config.GuestPath != nil {
		return config.GuestPath.Validate()
	}
	return nil
}

type LxcDataMount struct {
	ACL             *TriBool         `json:"acl,omitempty"`    // Never nil when returned
	Backup          *bool            `json:"backup,omitempty"` // Never nil when returned
	Options         *LxcMountOptions `json:"mount_options,omitempty"`
	Path            *LxcMountPath    `json:"path,omitempty"` // Required during creation, never nil when returned
	Quota           *bool            `json:"quota,omitempty"`
	ReadOnly        *bool            `json:"read_only,omitempty"`         // Never nil when returned
	Replicate       *bool            `json:"replicate,omitempty"`         // Never nil when returned
	SizeInKibibytes *LxcMountSize    `json:"size_in_kibibytes,omitempty"` // Required during creation, never nil when returned
	Storage         *string          `json:"storage,omitempty"`           // Required during creation, never nil when returned
	rawDisk         string
}

const (
	LxcDataMountErrorPathRequired      = "path is required for creation"
	LxcDataMountErrorQuotaUnprivileged = "quota can only be set for privileged containers"
	LxcDataMountErrorSizeRequired      = "size is required for creation"
	LxcDataMountErrorStorageRequired   = "storage is required for creation"
)

func (config LxcDataMount) combine(current LxcDataMount) LxcDataMount {
	if config.ACL != nil {
		current.ACL = config.ACL
	}
	if config.Backup != nil {
		current.Backup = config.Backup
	}
	if config.Options != nil {
		if current.Options == nil {
			current.Options = config.Options
		} else {
			if config.Options.Discard != nil {
				current.Options.Discard = config.Options.Discard
			}
			if config.Options.LazyTime != nil {
				current.Options.LazyTime = config.Options.LazyTime
			}
			if config.Options.NoATime != nil {
				current.Options.NoATime = config.Options.NoATime
			}
			if config.Options.NoSuid != nil {
				current.Options.NoSuid = config.Options.NoSuid
			}
			if config.Options.NoDevice != nil {
				current.Options.NoDevice = config.Options.NoDevice
			}
			if config.Options.NoExec != nil {
				current.Options.NoExec = config.Options.NoExec
			}
		}
	}
	if config.Path != nil {
		current.Path = config.Path
	}
	if config.Quota != nil {
		current.Quota = config.Quota
	}
	if config.ReadOnly != nil {
		current.ReadOnly = config.ReadOnly
	}
	if config.Replicate != nil {
		current.Replicate = config.Replicate
	}
	return current
}

func (config LxcDataMount) mapToAPICreate(id LxcMountID, privileged bool, params map[string]any) {
	rootFs := config.string(privileged)
	if config.Storage != nil && config.SizeInKibibytes != nil {
		var size string
		if *config.SizeInKibibytes < gibiByteLxc {
			size = "0.001"
		} else {
			size = strconv.FormatFloat(float64(*config.SizeInKibibytes/gibiByteLxc), 'f', -1, 64)
		}
		rootFs = *config.Storage + ":" + size + rootFs
	}
	params[lxcPrefixApiKeyMount+id.String()] = rootFs
}

func (config LxcDataMount) mapToAPIUpdate(current LxcDataMount, id LxcMountID, privileged bool, params map[string]any) {
	var usedConfig LxcDataMount
	currentMount := current.string(privileged)
	usedConfig = config.combine(current)
	// When the size is decreased we recreate the mountpoint. PVE will detach the old mount and create a new one.
	if config.SizeInKibibytes != nil && *config.SizeInKibibytes < *current.SizeInKibibytes {
		usedConfig.SizeInKibibytes = config.SizeInKibibytes
		if config.Storage != nil {
			usedConfig.Storage = config.Storage
		} else { // Use the current storage
			usedConfig.Storage = current.Storage
		}
		usedConfig.mapToAPICreate(id, privileged, params)
		return
	}
	mount := usedConfig.string(privileged)
	if mount != currentMount {
		if usedConfig.SizeInKibibytes != nil {
			mount = ",size=" + usedConfig.SizeInKibibytes.String() + mount
		}
		// local-ext4:100/vm-100-disk-0.raw,mp=./mnt/,acl=1,mountoptions=lazytime;noatime;nodev;noexec;nosuid;discard,replicate=0,size=8G,backup=1,ro=1
		params[lxcPrefixApiKeyMount+id.String()] = usedConfig.rawDisk + mount
	}
}

func (config LxcDataMount) string(privileged bool) (settings string) {
	if config.ACL != nil {
		switch *config.ACL {
		case TriBoolTrue:
			settings += ",acl=1"
		case TriBoolFalse:
			settings += ",acl=0"
		}
	}
	if config.Backup != nil && *config.Backup {
		settings += ",backup=1"
	}
	if config.Options != nil {
		if v := config.Options.string(); v != "" {
			settings += ",mountoptions=" + v[1:] // removes the first simicolon
		}
	}
	if config.Path != nil {
		settings += ",mp=" + config.Path.String()
	}
	if config.Quota != nil && *config.Quota && privileged {
		settings += ",quota=1"
	}
	if config.ReadOnly != nil && *config.ReadOnly {
		settings += ",ro=1"
	}
	if config.Replicate != nil && !*config.Replicate {
		settings += ",replicate=0"
	}
	return
}

func (config LxcDataMount) Validate(current *LxcDataMount, privileged bool) error {
	if current != nil {
		return config.validateUpdate(privileged)
	}
	return config.validateCreate(privileged)
}

func (config LxcDataMount) validateCreate(privileged bool) error {
	if config.Path == nil {
		return errors.New(LxcDataMountErrorPathRequired)
	}
	if config.SizeInKibibytes == nil {
		return errors.New(LxcDataMountErrorSizeRequired)
	}
	if config.Storage == nil {
		return errors.New(LxcDataMountErrorStorageRequired)
	}
	return config.validateUpdate(privileged)
}

func (config LxcDataMount) validateUpdate(privileged bool) error {
	var err error
	if config.ACL != nil {
		if err = config.ACL.Validate(); err != nil {
			return err
		}
	}
	if config.Quota != nil && !privileged {
		return errors.New(LxcDataMountErrorQuotaUnprivileged)
	}
	if config.Path != nil {
		if err = config.Path.Validate(); err != nil {
			return err
		}
	}
	if config.SizeInKibibytes != nil {
		return config.SizeInKibibytes.Validate()
	}
	return nil
}

type LxcMountOptions struct {
	Discard  *bool `json:"discard,omitempty"`   // Never nil when returned
	LazyTime *bool `json:"lazy_time,omitempty"` // Never nil when returned
	NoATime  *bool `json:"no_atime,omitempty"`  // Never nil when returned
	NoDevice *bool `json:"no_device,omitempty"` // Never nil when returned
	NoExec   *bool `json:"no_exec,omitempty"`   // Never nil when returned
	NoSuid   *bool `json:"no_suid,omitempty"`   // Never nil when returned
}

func (options *LxcMountOptions) mapToSDK(params string) {
	var discard, lazyTime, noATime, noDevice, noExec, noSuid bool
	*options = LxcMountOptions{
		Discard:  &discard,
		LazyTime: &lazyTime,
		NoATime:  &noATime,
		NoDevice: &noDevice,
		NoExec:   &noExec,
		NoSuid:   &noSuid}
	tmpOptions := splitStringOfOptions(params)
	if _, isSet := tmpOptions["discard"]; isSet {
		discard = true
	}
	if _, isSet := tmpOptions["lazytime"]; isSet {
		lazyTime = true
	}
	if _, isSet := tmpOptions["noatime"]; isSet {
		noATime = true
	}
	if _, isSet := tmpOptions["nodev"]; isSet {
		noDevice = true
	}
	if _, isSet := tmpOptions["noexec"]; isSet {
		noExec = true
	}
	if _, isSet := tmpOptions["nosuid"]; isSet {
		noSuid = true
	}
}

func (options LxcMountOptions) string() (settings string) {
	if options.Discard != nil && *options.Discard {
		settings += ";discard"
	}
	if options.LazyTime != nil && *options.LazyTime {
		settings += ";lazytime"
	}
	if options.NoATime != nil && *options.NoATime {
		settings += ";noatime"
	}
	if options.NoDevice != nil && *options.NoDevice {
		settings += ";nodev"
	}
	if options.NoExec != nil && *options.NoExec {
		settings += ";noexec"
	}
	if options.NoSuid != nil && *options.NoSuid {
		settings += ";nosuid"
	}
	return
}

// LxcHostPath is the absolute path to the mount point on the PVE host.
// Example: `/mnt/data`
// The path must start with a slash and must not contain `,` character.
type LxcHostPath string

const (
	LxcHostPathErrorInvalid          = "host path must not be empty"
	LxcHostPathErrorRelative         = "host path must be absolute"
	LxcHostPathErrorInvalidCharacter = "host path must not contain ',' character"
)

func (path LxcHostPath) String() string { return string(path) } // String is for fmt.Stringer.

func (path LxcHostPath) Validate() error {
	if path == "" {
		return errors.New(LxcHostPathErrorInvalid)
	}
	if path[0] != '/' {
		return errors.New(LxcHostPathErrorRelative)
	}
	if strings.ContainsRune(string(path), ',') {
		return errors.New(LxcHostPathErrorInvalidCharacter)
	}
	return nil
}

// LxcMountPath is the absolute path to the mount point inside the container.
// Example: `/mnt/data`
// The path must start with a slash and must not contain `,` character.
type LxcMountPath string

const (
	LxcMountPathErrorInvalid          = "mount point path must not be empty"
	LxcMountPathErrorRelative         = "mount point path must be absolute"
	LxcMountPathErrorInvalidCharacter = "mount point path must not contain ',' character"
)

func (path LxcMountPath) String() string { return string(path) } // String is for fmt.Stringer.

func (path LxcMountPath) Validate() error {
	if path == "" {
		return errors.New(LxcMountPathErrorInvalid)
	}
	if path[0] != '/' {
		return errors.New(LxcMountPathErrorRelative)
	}
	if strings.ContainsRune(string(path), ',') {
		return errors.New(LxcMountPathErrorInvalidCharacter)
	}
	return nil
}

type LxcMountSize uint

const (
	LxcMountSizeErrorMinimum = "mount point size must be greater than 131071"
	lxcMountSizeMinimum      = LxcMountSize(gibiByteOneEighth)
	gibiByteLxc              = mebiByte * 1024
)

func (size LxcMountSize) String() string { // String is for fmt.Stringer.
	if size%tebiByte == 0 {
		return strconv.Itoa(int(size/tebiByte)) + "T"
	}
	if size%gibiByte == 0 {
		return strconv.Itoa(int(size/gibiByte)) + "G"
	}
	if size%mebiByte == 0 {
		return strconv.Itoa(int(size/mebiByte)) + "M"
	}
	return strconv.Itoa(int(size)) + "K"
}

func (size LxcMountSize) Validate() error {
	if size < lxcMountSizeMinimum {
		return errors.New(LxcMountSizeErrorMinimum)
	}
	return nil
}

type lxcUpdateChanges struct {
	move     []lxcMountMove
	resize   []lxcMountResize
	offState bool // if true there is a change that can only be applied when the VM is off
}

type lxcMountMove struct {
	id      string
	storage string
}

func (disk lxcMountMove) move(ctx context.Context, delete bool, vmr *VmRef, client *Client) (exitStatus any, err error) {
	return client.PostWithTask(ctx, disk.mapToAPI(delete),
		"/nodes/"+vmr.node.String()+"/lxc/"+vmr.vmId.String()+"/move_volume")
}

func (disk lxcMountMove) mapToAPI(delete bool) map[string]any {
	params := map[string]any{
		"volume":  disk.id,
		"storage": disk.storage}
	if delete {
		params["delete"] = "1"
	}
	return params
}

type lxcMountResize struct {
	id              string
	sizeInKibibytes LxcMountSize
}

// Increase the disk size to the specified amount.
// Decrease of disk size is not permitted.
func (disk lxcMountResize) resize(ctx context.Context, vmr *VmRef, client *Client) (exitStatus string, err error) {
	return client.PutWithTask(ctx, map[string]any{"disk": disk.id, "size": disk.sizeInKibibytes.String()},
		"/nodes/"+vmr.node.String()+"/lxc/"+vmr.vmId.String()+"/resize")
}

func (raw RawConfigLXC) BootMount() *LxcBootMount {
	return raw.GetBootMount(raw.isPrivileged())
}

func (raw RawConfigLXC) GetBootMount(privileged bool) *LxcBootMount {
	var acl TriBool
	var quota bool
	var size LxcMountSize
	var storage string
	replicate := true
	config := LxcBootMount{
		ACL:             &acl,
		Replicate:       &replicate,
		SizeInKibibytes: &size,
		Storage:         &storage}
	if privileged {
		config.Quota = &quota
	}
	var settings map[string]string
	if v, isSet := raw[lxcApiKeyRootFS].(string); isSet {
		storage = v[:strings.IndexRune(v, ':')]
		if index := strings.IndexRune(v, ','); index != -1 {
			config.rawDisk = v[:index]
			settings = splitStringOfSettings(v[index:])
		} else {
			config.rawDisk = v
		}
	} else {
		return nil
	}
	if v, isSet := settings["size"]; isSet {
		size = LxcMountSize(parseDiskSize(v))
	}
	if v, isSet := settings["acl"]; isSet {
		if v == "1" {
			acl = TriBoolTrue
		} else {
			acl = TriBoolFalse
		}
	} else {
		config.ACL = util.Pointer(TriBoolNone)
	}
	if v, isSet := settings["mountoptions"]; isSet {
		options := splitStringOfOptions(v)
		var discard, lazyTime, noATime, noSuid bool
		mountOptions := LxcBootMountOptions{
			Discard:  &discard,
			LazyTime: &lazyTime,
			NoATime:  &noATime,
			NoSuid:   &noSuid}
		if _, isSet := options["discard"]; isSet {
			discard = true
		}
		if _, isSet := options["lazytime"]; isSet {
			lazyTime = true
		}
		if _, isSet := options["noatime"]; isSet {
			noATime = true
		}
		if _, isSet := options["nosuid"]; isSet {
			noSuid = true
		}
		config.Options = &mountOptions
	}
	if v, isSet := settings["quota"]; isSet {
		quota = v == "1"
	}
	if v, isSet := settings["replicate"]; isSet {
		replicate = v == "1"
	}
	return &config
}

func (raw RawConfigLXC) Mounts() LxcMounts {
	return raw.getMounts(raw.isPrivileged())
}

func (raw RawConfigLXC) getMounts(privileged bool) LxcMounts {
	mounts := LxcMounts{}
	for i := range LxcMountsAmount {
		if v, isSet := raw[lxcPrefixApiKeyMount+strconv.Itoa(i)].(string); isSet {
			if v[0] == '/' { // Bind mount
				mounts.mapToSdkBindMount(LxcMountID(i), v)
			} else { // Data mount
				mounts.mapToSdkDataMount(LxcMountID(i), v, privileged)
			}
		}
	}
	if len(mounts) == 0 {
		return nil
	}
	return mounts
}
