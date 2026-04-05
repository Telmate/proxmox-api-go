package proxmox

import (
	"errors"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type EfiDisk struct {
	Format *QemuDiskFormat `json:"format,omitempty"` // Never nil when returned

	// Due to a bug in PVE MsCertType can not be set, only returned.
	MsCertType      *EfiMsCertType `json:"ms-cert-type,omitempty"`
	PreEnrolledKeys *bool          `json:"pre-enrolled-keys,omitempty"` // Never nil when returned. Change causes recreate.
	Storage         *StorageName   `json:"storage,omitempty"`           // Never nil when returned
	Size            EfiDiskSize    `json:"size,omitempty"`              // Only returned. Size seems to be determined by the backing storage type.
	Type            *EfiDiskType   `json:"type,omitempty"`              // Change causes recreate.
	Delete          bool           `json:"delete,omitempty"`
}

const (
	EFiDisk_Error_StorageRequired = "efi disk storage is required"
)

func (efi EfiDisk) mapToApiCreate(b *strings.Builder) {
	if efi.Delete {
		return
	}

	b.WriteString("&" + qemuApiKeyEfiDisk + "=")
	if efi.Storage != nil {
		b.WriteString(efi.Storage.String())
	}
	b.WriteString(colon + "1")
	if efi.Format != nil {
		b.WriteString(comma + qemuEfiDiskSettingsKeyFormat + equal)
		b.WriteString(efi.Format.String())
	}
	// if efi.MsCertType != nil {
	// 	b.WriteString(comma + qemuEfiDiskSettingsKeyMsCert + equal)
	// 	b.WriteString(efi.MsCertType.String())
	// }
	if efi.PreEnrolledKeys != nil && *efi.PreEnrolledKeys {
		b.WriteString(comma + qemuEfiDiskSettingsKeyPreEnrolledKeys + equal + "1")
	}
	if efi.Type != nil && *efi.Type != EfiDiskTypeUnset {
		b.WriteString(comma + qemuEfiDiskSettingsKeyType + equal)
		b.WriteString(efi.Type.String())
	}
}

func (efi EfiDisk) mapToApiUpdate(current *EfiDisk, builder, delete *strings.Builder) {
	if efi.Delete {
		delete.WriteString("," + qemuApiKeyEfiDisk)
		return
	}
	if efi.replace(current) { // We can recreate it over the existing disk
		if efi.Storage != nil {
			if *efi.Storage != *current.Storage { // Format is only inherited when storage does not change
				current.Format = nil
				current.Storage = efi.Storage
			}
		}
		if efi.Format != nil {
			current.Format = efi.Format
		}
		if efi.PreEnrolledKeys != nil {
			current.PreEnrolledKeys = efi.PreEnrolledKeys
		}
		if efi.Type != nil {
			current.Type = efi.Type
		}
		current.mapToApiCreate(builder)
	}
}

func (efi *EfiDisk) markChangesUnsafe(current *EfiDisk) (disk *qemuDiskMove) {
	if efi.Delete { // explicit delete is handled elsewhere
		return nil
	}
	if efi.replace(current) {
		return nil
	}

	var move bool
	var storage StorageName
	var format *QemuDiskFormat

	if efi.Storage != nil {
		if *efi.Storage != *current.Storage {
			move = true
		}
		storage = *efi.Storage
	} else {
		storage = *current.Storage
	}
	if efi.Format != nil {
		if *efi.Format != *current.Format {
			move = true
		}
		format = efi.Format
	}
	if move {
		return &qemuDiskMove{
			Storage: string(storage),
			Format:  format,
			Id:      QemuDiskId(qemuApiKeyEfiDisk)}
	}
	return nil
}

func (efi EfiDisk) replace(current *EfiDisk) bool {
	if efi.PreEnrolledKeys != nil && *efi.PreEnrolledKeys != *current.PreEnrolledKeys {
		return true
	}
	if efi.Type != nil && current.Type != nil && *efi.Type != *current.Type {
		return true
	}
	return false
}

// TODO test
func (efi EfiDisk) Validate(current *EfiDisk) error {
	if current == nil {
		return efi.validateCreate()
	}
	return efi.validateUpdate()
}

func (efi *EfiDisk) validateCreate() error {
	if efi.Delete {
		return nil
	}
	if efi.Storage == nil {
		return errors.New(EFiDisk_Error_StorageRequired)
	}
	if err := efi.Storage.Validate(); err != nil {
		return err
	}
	return efi.validateShared()
}

func (efi *EfiDisk) validateUpdate() error {
	if efi.Delete {
		return nil
	}
	if efi.Storage != nil {
		if err := efi.Storage.Validate(); err != nil {
			return err
		}
	}
	return efi.validateShared()
}

func (efi *EfiDisk) validateShared() error {
	if efi.Format != nil {
		if err := efi.Format.Validate(); err != nil {
			return err
		}
	}
	if efi.Type != nil {
		return efi.Type.Validate()
	}
	// if efi.MsCertType != nil {
	// 	if err := efi.MsCertType.Validate(); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

// EfiDiskSize is the size of the efi disk in kibibytes.
type EfiDiskSize uint

// Enum
type EfiDiskType string

const (
	EfiDiskTypeUnset EfiDiskType = ""
	EfiDiskType2M    EfiDiskType = "2m"
	EfiDiskType4M    EfiDiskType = "4m"
)

const EfiDiskType_Error = "invalid efi disk type"

func (e EfiDiskType) String() string { return string(e) }

// TODO test
func (e EfiDiskType) Validate() error {
	switch e {
	case EfiDiskTypeUnset, EfiDiskType2M, EfiDiskType4M:
		return nil
	default:
		return errors.New(EfiDiskType_Error)
	}
}

// Enum
type EfiMsCertType int16

const (
	EfiMsCertTypeUnset EfiMsCertType = 0
	EfiMsCertType2011  EfiMsCertType = 2011
	EfiMsCertType2023  EfiMsCertType = 2023
)

const EfiMsCertType_Error = "invalid efi ms-cert type"

func (e EfiMsCertType) String() string {
	switch e {
	case EfiMsCertType2011:
		return "2011"
	case EfiMsCertType2023:
		return "2023"
	default:
		return ""
	}
}

// TODO test
func (e EfiMsCertType) Validate() error {
	switch e {
	case EfiMsCertTypeUnset, EfiMsCertType2011, EfiMsCertType2023:
		return nil
	default:
		return errors.New(EfiMsCertType_Error)
	}
}

func (raw *rawConfigQemu) GetEfiDisk() *EfiDisk {
	if v, isSet := raw.a[qemuApiKeyEfiDisk]; isSet {
		// "local-zfs:vm-1020-disk-0,size=1M"
		raw := v.(string)
		comma := strings.IndexByte(raw, ',')
		var preEnrolledKeys bool
		var format QemuDiskFormat
		disk := EfiDisk{
			Format:          &format,
			PreEnrolledKeys: &preEnrolledKeys,
			Storage:         util.Pointer(StorageName(raw[:strings.IndexByte(raw, ':')])),
		}

		format.parse(raw[:comma])

		settings := splitStringOfSettings(raw[comma:])
		if v, ok := settings[qemuEfiDiskSettingsKeyMsCert]; ok {
			switch v {
			case "2011":
				disk.MsCertType = util.Pointer(EfiMsCertType2011)
			case "2023":
				disk.MsCertType = util.Pointer(EfiMsCertType2023)
			}
		}
		if v, ok := settings[qemuEfiDiskSettingsKeyPreEnrolledKeys]; ok {
			preEnrolledKeys = v == "1"
		}
		if v, ok := settings[qemuEfiDiskSettingsKeySize]; ok {
			disk.Size = EfiDiskSize(parseDiskSize(v))
		}
		if v, ok := settings[qemuEfiDiskSettingsKeyType]; ok {
			switch v {
			case "2m":
				disk.Type = util.Pointer(EfiDiskType2M)
			case "4m":
				disk.Type = util.Pointer(EfiDiskType4M)
			}
		}
		return &disk
	}
	return nil
}

const (
	qemuEfiDiskSettingsKeyPreEnrolledKeys = "pre-enrolled-keys"
	qemuEfiDiskSettingsKeySize            = "size"
	qemuEfiDiskSettingsKeyType            = "efitype"
	qemuEfiDiskSettingsKeyMsCert          = "ms-cert"
	qemuEfiDiskSettingsKeyFormat          = "format"
)
