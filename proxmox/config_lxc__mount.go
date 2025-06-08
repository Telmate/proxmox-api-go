package proxmox

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type LxcBootMount struct {
	ACL             *TriBool
	Options         *LxcBootMountOptions
	Replication     *bool
	SizeInKibibytes *LxcMountSize // Required during creation, never nil when returned
	Storage         *string       // Required during creation, never nil when returned
	rawDisk         string
}

const (
	LxcBootMount_Error_NoSizeDuringCreation    = "size must be set during creation"
	LxcBootMount_Error_NoStorageDuringCreation = "storage must be set during creation"
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
	if mount.Replication != nil {
		usedConfig.Replication = mount.Replication
	}
	if mount.ACL != nil {
		usedConfig.ACL = mount.ACL
	}
	return usedConfig
}

func (config LxcBootMount) mapToApiCreate() string {
	rootFs := config.string()
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

func (config LxcBootMount) mapToApiUpdate_Unsafe(current *LxcBootMount, params map[string]any) {
	var usedConfig LxcBootMount
	usedConfig = config.combine(current.combine(usedConfig))
	rootFs := usedConfig.string()
	if usedConfig.Storage != nil {
		// we can ignore adding the size, the call will work without it
		rootFs = *usedConfig.Storage + ":" + current.rawDisk + rootFs
		if current.Storage != nil && rootFs == *current.Storage+":"+current.rawDisk+current.string() {
			return
		}
	}
	params[lxcApiKeyRootFS] = rootFs
}

func (config LxcBootMount) string() (rootFs string) {
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
	if config.Replication != nil && !*config.Replication {
		rootFs += ",replicate=0"
	}
	return
}

func (config LxcBootMount) Validate(current *LxcBootMount) error {
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
	return err
}

type LxcBootMountOptions struct {
	Discard  *bool // Never nil when returned
	LazyTime *bool // Never nil when returned
	NoATime  *bool // Never nil when returned
	NoSuid   *bool // Never nil when returned
}

type LxcMountSize uint

const (
	LxcMountSize_Error_Minimum = "mount point size must be greater than 131071"
	lxcMountSize_Minimum       = LxcMountSize(gibiByteOneEighth)
	gibiByteLxc                = mebiByte * 1024
)

func (size LxcMountSize) String() string { return strconv.Itoa(int(size)) } // String is for fmt.Stringer.

func (size LxcMountSize) Validate() error {
	if size < lxcMountSize_Minimum {
		return errors.New(LxcMountSize_Error_Minimum)
	}
	return nil
}

func (raw RawConfigLXC) BootMount() *LxcBootMount {
	var size LxcMountSize
	var storage string
	config := LxcBootMount{
		SizeInKibibytes: &size,
		Storage:         &storage}
	var settings map[string]string
	if v, isSet := raw[lxcApiKeyRootFS]; isSet {
		if tmpString := strings.SplitN(v.(string), ",", 2); len(tmpString) == 2 {
			if index := strings.IndexRune(tmpString[0], ':'); index != -1 {
				storage = tmpString[0][:index]
				config.rawDisk = tmpString[0][index+1:]
				settings = splitStringOfSettings(tmpString[1])
			}
		}
	} else {
		return nil
	}
	if v, isSet := settings["size"]; isSet {
		size = LxcMountSize(parseDiskSize(v))
	}
	if v, isSet := settings["acl"]; isSet {
		if v == "1" {
			config.ACL = util.Pointer(TriBoolTrue)
		} else {
			config.ACL = util.Pointer(TriBoolFalse)
		}
	} else {
		config.ACL = util.Pointer(TriBoolNone)
	}
	if v, isSet := settings["mountoptions"]; isSet {
		tmpOptions := strings.Split(v, ";")
		options := make(map[string]struct{}, len(tmpOptions))
		for i := range tmpOptions {
			options[tmpOptions[i]] = struct{}{}
		}
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
	if v, isSet := settings["replicate"]; isSet {
		config.Replication = util.Pointer(v == "1")
	} else {
		config.Replication = util.Pointer(true)
	}
	return &config
}
