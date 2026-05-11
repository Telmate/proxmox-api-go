package proxmox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type IsoFile struct {
	File    string `json:"file"`
	Storage string `json:"storage"`
	// SizeInKibibytes can only be retrieved, setting it has no effect
	SizeInKibibytes string `json:"size"`
}

const (
	Error_IsoFile_File    string = "file may not be empty"
	Error_IsoFile_Storage string = "storage may not be empty"
)

func (iso IsoFile) Validate() error {
	if iso.File == "" {
		return errors.New(Error_IsoFile_File)
	}
	if iso.Storage == "" {
		return errors.New(Error_IsoFile_Storage)
	}
	return nil
}

type QemuCdRom struct {
	Iso *IsoFile `json:"iso,omitempty"`
	// Passthrough and File are mutually exclusive
	Passthrough bool `json:"passthrough,omitempty"`
}

const (
	Error_QemuCdRom_MutuallyExclusive string = "iso and passthrough are mutually exclusive"
)

func (cdRom QemuCdRom) mapToApiValues() string {
	if cdRom.Passthrough {
		return "cdrom,media=cdrom"
	}
	if cdRom.Iso != nil {
		return cdRom.Iso.Storage + ":iso/" + cdRom.Iso.File + ",media=cdrom"
	}
	return "none,media=cdrom"
}

func (QemuCdRom) mapToStruct(settings qemuCdRom) *QemuCdRom {
	if settings.File != "" {
		return &QemuCdRom{
			Iso: &IsoFile{
				Storage:         settings.Storage,
				File:            settings.File,
				SizeInKibibytes: settings.SizeInKibibytes,
			},
		}
	}
	return &QemuCdRom{Passthrough: settings.Passthrough}
}

func (cdRom QemuCdRom) Validate() error {
	if cdRom.Iso != nil {
		err := cdRom.Iso.Validate()
		if err != nil {
			return err
		}
		if cdRom.Passthrough {
			return errors.New(Error_QemuCdRom_MutuallyExclusive)
		}
	}
	return nil
}

type qemuCdRom struct {
	CdRom bool
	// "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"
	Passthrough     bool
	Storage         string
	Format          QemuDiskFormat // Only set for Cloud-init drives
	File            string
	SizeInKibibytes string
}

func (qemuCdRom) mapToStruct(diskData string, settings map[string]string) *qemuCdRom {
	if setting, isSet := settings["media"]; isSet {
		if setting != "cdrom" {
			return nil
		}
	} else {
		return nil
	}
	if _, isSet := settings["none"]; isSet {
		return &qemuCdRom{CdRom: true}
	}
	if _, isSet := settings["cdrom"]; isSet {
		return &qemuCdRom{CdRom: true, Passthrough: true}
	}
	tmpStorage := strings.Split(diskData, ":")
	if len(tmpStorage) > 1 {
		tmpFile := strings.Split(diskData, "/")
		switch len(tmpFile) {
		case 1:
			return &qemuCdRom{
				Storage: tmpStorage[0],
				File:    tmpStorage[1],
				Format:  QemuDiskFormat_Raw,
			}
		case 2:
			tmpFileType := strings.Split(tmpFile[1], ".")
			if len(tmpFileType) > 1 {
				fileType := QemuDiskFormat(tmpFileType[len(tmpFileType)-1])
				if fileType == "iso" {
					if setting, isSet := settings["size"]; isSet {
						return &qemuCdRom{
							CdRom:           true,
							Storage:         tmpStorage[0],
							File:            tmpFile[1],
							SizeInKibibytes: setting,
						}
					}
				} else {
					return &qemuCdRom{
						Storage: tmpStorage[0],
						File:    tmpFile[1],
						Format:  fileType,
					}
				}
			}
		}
	}
	return nil
}

type QemuCloudInitDisk struct {
	Format  QemuDiskFormat `json:"format,omitempty"`
	Storage string         `json:"storage,omitempty"`
}

const (
	Error_QemuCloudInitDisk_Storage string = "storage should not be empty"
	Error_QemuCloudInitDisk_OnlyOne string = "only one cloud init disk may exist"
)

func (QemuCloudInitDisk) checkDuplicates(numberOFCloudInitDrives uint8) error {
	if numberOFCloudInitDrives > 1 {
		return errors.New(Error_QemuCloudInitDisk_OnlyOne)
	}
	return nil
}

func (cloudInit QemuCloudInitDisk) mapToApiValues() string {
	return cloudInit.Storage + ":cloudinit,format=" + string(cloudInit.Format)
}

func (QemuCloudInitDisk) mapToStruct(settings qemuCdRom) *QemuCloudInitDisk {
	return &QemuCloudInitDisk{
		Storage: settings.Storage,
		Format:  settings.Format,
	}
}

func (cloudInit QemuCloudInitDisk) Validate() error {
	if err := cloudInit.Format.Validate(); err != nil {
		return err
	}
	if cloudInit.Storage == "" {
		return errors.New(Error_QemuCloudInitDisk_Storage)
	}
	return nil
}

type qemuDisk struct {
	LinkedDiskId    *GuestID // Only set for Disk
	AsyncIO         QemuDiskAsyncIO
	Cache           QemuDiskCache
	Format          QemuDiskFormat // Only set for Disk
	Serial          QemuDiskSerial
	SizeInKibibytes QemuDiskSize
	Type            qemuDiskType
	WorldWideName   QemuWorldWideName
	// TODO custom type
	File       string // Only set for Passthrough.
	ImportFrom string
	// TODO custom type
	Storage    string // Only set for Disk
	VolumePath string
	Bandwidth  QemuDiskBandwidth
	Id         uint // Only set for Disk
	Backup     bool
	Discard    bool
	Disk       bool // true = disk, false = passthrough
	EmulateSSD bool // Only set for ide,sata,scsi
	IOThread   bool // Only set for scsi,virtio
	ReadOnly   bool // Only set for scsi,virtio
	Replicate  bool
}

const (
	Error_QemuDisk_File              string = "file may not be empty"
	Error_QemuDisk_MutuallyExclusive string = "settings cdrom,cloudinit,disk,passthrough are mutually exclusive"
	Error_QemuDisk_Storage           string = "storage may not be empty"
)

// Maps all the disk related settings to api values proxmox understands.
func (disk qemuDisk) mapToApiValues(create bool) string {
	builder := strings.Builder{}
	if disk.Storage != "" {
		builder.WriteString(disk.Storage)
		if create {
			if disk.ImportFrom != "" {
				builder.WriteString(":0,import-from=")
				builder.WriteString(disk.ImportFrom)
			} else if disk.SizeInKibibytes%gibibyte == 0 {
				builder.WriteRune(':')
				builder.WriteString(strconv.FormatInt(int64(disk.SizeInKibibytes/gibibyte), 10))
			} else {
				builder.WriteString(":0.001")
			}
		} else {
			builder.WriteRune(':')
			builder.WriteString(disk.VolumePath)
		}
	}

	if disk.File != "" {
		builder.WriteString(disk.File)
	}
	// Set File

	if disk.AsyncIO != "" {
		builder.WriteString(",aio=")
		builder.WriteString(disk.AsyncIO.String())
	}
	if !disk.Backup {
		builder.WriteString(",backup=0")
	}
	if disk.Cache != "" {
		builder.WriteString(",cache=")
		builder.WriteString(disk.Cache.String())
	}
	if disk.Discard {
		builder.WriteString(",discard=on")
	}
	if disk.Format != "" && create {
		builder.WriteString(",format=")
		builder.WriteString(disk.Format.String())
	}
	// media

	if disk.Bandwidth.Iops.ReadLimit.Concurrent != 0 {
		builder.WriteString(",iops_rd=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Concurrent)))
	}
	if disk.Bandwidth.Iops.ReadLimit.Burst != 0 {
		builder.WriteString(",iops_rd_max=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Burst)))
	}
	if disk.Bandwidth.Iops.ReadLimit.BurstDuration != 0 {
		builder.WriteString(",iops_rd_max_length=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.BurstDuration)))
	}

	if disk.Bandwidth.Iops.WriteLimit.Concurrent != 0 {
		builder.WriteString(",iops_wr=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Concurrent)))
	}
	if disk.Bandwidth.Iops.WriteLimit.Burst != 0 {
		builder.WriteString(",iops_wr_max=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Burst)))
	}
	if disk.Bandwidth.Iops.WriteLimit.BurstDuration != 0 {
		builder.WriteString(",iops_wr_max_length=")
		builder.WriteString(strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.BurstDuration)))
	}

	if (disk.Type == scsi || disk.Type == virtIO) && disk.IOThread {
		builder.WriteString(",iothread=1")
	}

	if disk.Bandwidth.MBps.ReadLimit.Concurrent != 0 {
		builder.WriteString(",mbps_rd=")
		builder.WriteString(floatToTrimmedString(float64(disk.Bandwidth.MBps.ReadLimit.Concurrent), 2))
	}
	if disk.Bandwidth.MBps.ReadLimit.Burst != 0 {
		builder.WriteString(",mbps_rd_max=")
		builder.WriteString(floatToTrimmedString(float64(disk.Bandwidth.MBps.ReadLimit.Burst), 2))
	}

	if disk.Bandwidth.MBps.WriteLimit.Concurrent != 0 {
		builder.WriteString(",mbps_wr=")
		builder.WriteString(floatToTrimmedString(float64(disk.Bandwidth.MBps.WriteLimit.Concurrent), 2))
	}
	if disk.Bandwidth.MBps.WriteLimit.Burst != 0 {
		builder.WriteString(",mbps_wr_max=")
		builder.WriteString(floatToTrimmedString(float64(disk.Bandwidth.MBps.WriteLimit.Burst), 2))
	}

	if !disk.Replicate {
		builder.WriteString(",replicate=0")
	}
	if (disk.Type == scsi || disk.Type == virtIO) && disk.ReadOnly {
		builder.WriteString(",ro=1")
	}
	if disk.Serial != "" {
		builder.WriteString(",serial=")
		builder.WriteString(disk.Serial.String())
	}
	if disk.Type != virtIO && disk.EmulateSSD {
		builder.WriteString(",ssd=1")
	}
	if disk.WorldWideName != "" {
		builder.WriteString(",wwn=")
		builder.WriteString(disk.WorldWideName.String())
	}
	return builder.String()
}

// Maps all the disk related settings to our own data structure.
func (qemuDisk) mapToStruct(diskData string, settings map[string]string, linkedVmId *GuestID) *qemuDisk {
	disk := qemuDisk{
		Backup:    true,
		Replicate: true}
	diskPTR := &disk
	if diskData[0:1] == "/" { // Passthrough
		disk.File = diskData
	} else { // Disk
		diskPTR.parseDisk(diskData, linkedVmId)
	}

	if len(settings) == 0 {
		return nil
	}

	if value, isSet := settings["aio"]; isSet {
		disk.AsyncIO = QemuDiskAsyncIO(value)
	}
	if value, isSet := settings["backup"]; isSet {
		disk.Backup, _ = strconv.ParseBool(value)
	}
	if value, isSet := settings["cache"]; isSet {
		disk.Cache = QemuDiskCache(value)
	}
	if value, isSet := settings["discard"]; isSet {
		if value == "on" {
			disk.Discard = true
		}
	}
	if value, isSet := settings["iops_rd"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.ReadLimit.Concurrent = QemuDiskBandwidthIopsLimitConcurrent(tmp)
	}
	if value, isSet := settings["iops_rd_max"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.ReadLimit.Burst = QemuDiskBandwidthIopsLimitBurst(tmp)
	}
	if value, isSet := settings["iops_rd_max_length"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.ReadLimit.BurstDuration = uint(tmp)
	}
	if value, isSet := settings["iops_wr"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.WriteLimit.Concurrent = QemuDiskBandwidthIopsLimitConcurrent(tmp)
	}
	if value, isSet := settings["iops_wr_max"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.WriteLimit.Burst = QemuDiskBandwidthIopsLimitBurst(tmp)
	}
	if value, isSet := settings["iops_wr_max_length"]; isSet {
		tmp, _ := strconv.Atoi(value)
		disk.Bandwidth.Iops.WriteLimit.BurstDuration = uint(tmp)
	}
	if value, isSet := settings["iothread"]; isSet {
		disk.IOThread, _ = strconv.ParseBool(value)
	}
	if value, isSet := settings["mbps_rd"]; isSet {
		tmp, _ := strconv.ParseFloat(value, 32)
		disk.Bandwidth.MBps.ReadLimit.Concurrent = QemuDiskBandwidthMBpsLimitConcurrent(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_rd_max"]; isSet {
		tmp, _ := strconv.ParseFloat(value, 32)
		disk.Bandwidth.MBps.ReadLimit.Burst = QemuDiskBandwidthMBpsLimitBurst(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_wr"]; isSet {
		tmp, _ := strconv.ParseFloat(value, 32)
		disk.Bandwidth.MBps.WriteLimit.Concurrent = QemuDiskBandwidthMBpsLimitConcurrent(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_wr_max"]; isSet {
		tmp, _ := strconv.ParseFloat(value, 32)
		disk.Bandwidth.MBps.WriteLimit.Burst = QemuDiskBandwidthMBpsLimitBurst(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["replicate"]; isSet {
		disk.Replicate, _ = strconv.ParseBool(value)
	}
	if value, isSet := settings["ro"]; isSet {
		disk.ReadOnly, _ = strconv.ParseBool(value)
	}
	if value, isSet := settings["serial"]; isSet {
		disk.Serial = QemuDiskSerial(value)
	}
	if value, isSet := settings["size"]; isSet {
		disk.SizeInKibibytes = QemuDiskSize(0).parse(value)
	}
	if value, isSet := settings["ssd"]; isSet {
		disk.EmulateSSD, _ = strconv.ParseBool(value)
	}
	if value, isSet := settings["wwn"]; isSet {
		disk.WorldWideName = QemuWorldWideName(value)
	}
	return diskPTR
}

// parse and extract the values from the disk data
// storage:110/base-110-disk-1.qcow2/100/vm-100-disk-0.qcow2
// storage:100/vm-100-disk-0.qcow2
// storage:vm-100-disk-0
// storage:vm-100-disk-0.qcow2 (for volume syntax with non-raw format)
// storage:base-110-disk-1.qcow2/vm-100-disk-0
// storage:base-110-disk-1/vm-100-disk-0.qcow2 (for linked clone volume syntax with non-raw format)
func (disk *qemuDisk) parseDisk(diskData string, linkedVmId *GuestID) {
	index := strings.IndexByte(diskData, ':')
	if index <= 0 {
		return
	}
	disk.Storage = diskData[0:index]
	disk.VolumePath = diskData[index+1:]
	pathParts := strings.Split(disk.VolumePath, "/")
	switch len(pathParts) {
	case 2:
		if pathParts[0][0:1] == "b" { // Linked Volumes are prefixed with "base"
			index = strings.IndexByte(pathParts[0], '.')
			if index <= 0 {
				index = len(pathParts[0])
			}
			tmp := strings.Split(pathParts[0][0:index], "-")
			if len(tmp) > 1 {
				if tmpVmId, err := strconv.Atoi(tmp[1]); err == nil {
					*linkedVmId = GuestID(tmpVmId)
				}
				if tmpDiskId, err := strconv.Atoi(tmp[len(tmp)-1]); err == nil {
					disk.LinkedDiskId = new(GuestID(tmpDiskId))
				}
			}
		}
	case 4: // Linked File
		if tmpVmId, err := strconv.Atoi(pathParts[0]); err == nil {
			*linkedVmId = GuestID(tmpVmId)
		}
		index = strings.IndexByte(pathParts[1], '.')
		if index <= 0 {
			index = len(pathParts[1])
		}
		tmp := strings.Split(pathParts[1][0:index], "-")
		if len(tmp) > 1 {
			if tmpDiskId, err := strconv.Atoi(tmp[len(tmp)-1]); err == nil {
				disk.LinkedDiskId = new(GuestID(tmpDiskId))
			}
		}
	}

	diskNameAndFormat := strings.Split(pathParts[len(pathParts)-1], ".")

	index = strings.LastIndexByte(diskNameAndFormat[0], '-')
	if index > 0 {
		if tmpDiskId, err := strconv.Atoi(diskNameAndFormat[0][index+1:]); err == nil {
			disk.Id = uint(tmpDiskId)
		}
	}

	if len(diskNameAndFormat) == 2 { // set disk format, default to raw
		disk.Format = QemuDiskFormat(diskNameAndFormat[1])
	} else {
		disk.Format = QemuDiskFormat_Raw
	}
}

func (disk *qemuDisk) validate() (err error) {
	if disk == nil {
		return
	}
	if err = disk.AsyncIO.Validate(); err != nil {
		return
	}
	if err = disk.Bandwidth.Validate(); err != nil {
		return
	}
	if err = disk.Cache.Validate(); err != nil {
		return
	}
	if err = disk.Serial.Validate(); err != nil {
		return
	}
	if err = disk.WorldWideName.Validate(); err != nil {
		return
	}
	if disk.Disk {
		// disk
		if disk.ImportFrom == "" {
			// size/format is mandatory except if import-from is set.
			if err = disk.Format.Validate(); err != nil {
				return
			}
			if err = disk.SizeInKibibytes.Validate(); err != nil {
				return
			}
		}
		if disk.Storage == "" {
			return errors.New(Error_QemuDisk_Storage)
		}
	} else {
		// passthrough
		if disk.File == "" {
			return errors.New(Error_QemuDisk_File)
		}
	}
	return
}

type QemuDiskAsyncIO string

const (
	QemuDiskAsyncIO_Native  QemuDiskAsyncIO = "native"
	QemuDiskAsyncIO_Threads QemuDiskAsyncIO = "threads"
	QemuDiskAsyncIO_IOuring QemuDiskAsyncIO = "io_uring"
)

func (QemuDiskAsyncIO) Error() error {
	return fmt.Errorf("asyncio can only be one of the following values: %s,%s,%s", QemuDiskAsyncIO_Native, QemuDiskAsyncIO_Threads, QemuDiskAsyncIO_IOuring)
}

func (asyncIO QemuDiskAsyncIO) String() string { return string(asyncIO) } // For fmt.Stringer

func (asyncIO QemuDiskAsyncIO) Validate() error {
	switch asyncIO {
	case "", QemuDiskAsyncIO_Native, QemuDiskAsyncIO_Threads, QemuDiskAsyncIO_IOuring:
		return nil
	}
	return QemuDiskAsyncIO("").Error()
}

type QemuDiskBandwidth struct {
	MBps QemuDiskBandwidthMBps `json:"mbps,omitempty"`
	Iops QemuDiskBandwidthIops `json:"iops,omitempty"`
}

func (bandwidth QemuDiskBandwidth) Validate() error {
	err := bandwidth.MBps.Validate()
	if err != nil {
		return err
	}
	return bandwidth.Iops.Validate()
}

type QemuDiskBandwidthIops struct {
	ReadLimit  QemuDiskBandwidthIopsLimit `json:"read,omitempty"`
	WriteLimit QemuDiskBandwidthIopsLimit `json:"write,omitempty"`
}

func (iops QemuDiskBandwidthIops) Validate() error {
	err := iops.ReadLimit.Validate()
	if err != nil {
		return err
	}
	return iops.WriteLimit.Validate()
}

type QemuDiskBandwidthIopsLimit struct {
	Burst         QemuDiskBandwidthIopsLimitBurst      `json:"burst,omitempty"`          // 0 = unlimited
	BurstDuration uint                                 `json:"burst_duration,omitempty"` // burst duration in seconds
	Concurrent    QemuDiskBandwidthIopsLimitConcurrent `json:"concurrent,omitempty"`     // 0 = unlimited
}

func (limit QemuDiskBandwidthIopsLimit) Validate() (err error) {
	if err = limit.Burst.Validate(); err != nil {
		return
	}
	err = limit.Concurrent.Validate()
	return
}

type QemuDiskBandwidthIopsLimitBurst uint

const (
	Error_QemuDiskBandwidthIopsLimitBurst string = "burst may not be lower than 10 except for 0"
)

func (limit QemuDiskBandwidthIopsLimitBurst) Validate() error {
	if limit != 0 && limit < 10 {
		return errors.New(Error_QemuDiskBandwidthIopsLimitBurst)
	}
	return nil
}

type QemuDiskBandwidthIopsLimitConcurrent uint

const (
	Error_QemuDiskBandwidthIopsLimitConcurrent string = "concurrent may not be lower than 10 except for 0"
)

func (limit QemuDiskBandwidthIopsLimitConcurrent) Validate() error {
	if limit != 0 && limit < 10 {
		return errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)
	}
	return nil
}

type QemuDiskBandwidthMBps struct {
	ReadLimit  QemuDiskBandwidthMBpsLimit `json:"read,omitempty"`
	WriteLimit QemuDiskBandwidthMBpsLimit `json:"write,omitempty"`
}

func (data QemuDiskBandwidthMBps) Validate() error {
	err := data.ReadLimit.Validate()
	if err != nil {
		return err
	}
	return data.WriteLimit.Validate()
}

type QemuDiskBandwidthMBpsLimit struct {
	Burst      QemuDiskBandwidthMBpsLimitBurst      `json:"burst,omitempty"`      // 0 = unlimited
	Concurrent QemuDiskBandwidthMBpsLimitConcurrent `json:"concurrent,omitempty"` // 0 = unlimited
}

func (limit QemuDiskBandwidthMBpsLimit) Validate() (err error) {
	if err = limit.Burst.Validate(); err != nil {
		return
	}
	err = limit.Concurrent.Validate()
	return
}

const (
	Error_QemuDiskBandwidthMBpsLimitBurst string = "burst may not be lower than 1 except for 0"
)

type QemuDiskBandwidthMBpsLimitBurst float32

func (limit QemuDiskBandwidthMBpsLimitBurst) Validate() error {
	if limit != 0 && limit < 1 {
		return errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)
	}
	return nil
}

const (
	Error_QemuDiskBandwidthMBpsLimitConcurrent string = "concurrent may not be lower than 1 except for 0"
)

type QemuDiskBandwidthMBpsLimitConcurrent float32

func (limit QemuDiskBandwidthMBpsLimitConcurrent) Validate() error {
	if limit != 0 && limit < 1 {
		return errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)
	}
	return nil
}

type QemuDiskCache string

const (
	QemuDiskCache_None         QemuDiskCache = "none"
	QemuDiskCache_WriteThrough QemuDiskCache = "writethrough"
	QemuDiskCache_WriteBack    QemuDiskCache = "writeback"
	QemuDiskCache_Unsafe       QemuDiskCache = "unsafe"
	QemuDiskCache_DirectSync   QemuDiskCache = "directsync"
)

func (QemuDiskCache) Error() error {
	return fmt.Errorf("cache can only be one of the following values: %s,%s,%s,%s,%s", QemuDiskCache_None, QemuDiskCache_WriteThrough, QemuDiskCache_WriteBack, QemuDiskCache_Unsafe, QemuDiskCache_DirectSync)
}

func (cache QemuDiskCache) String() string { return string(cache) } // For fmt.Stringer

func (cache QemuDiskCache) Validate() error {
	switch cache {
	case "", QemuDiskCache_None, QemuDiskCache_WriteThrough, QemuDiskCache_WriteBack, QemuDiskCache_Unsafe, QemuDiskCache_DirectSync:
		return nil
	}
	return QemuDiskCache("").Error()
}

type QemuDiskFormat string

const (
	QemuDiskFormat_Cow   QemuDiskFormat = "cow"
	QemuDiskFormat_Cloop QemuDiskFormat = "cloop"
	QemuDiskFormat_Qcow  QemuDiskFormat = "qcow"
	QemuDiskFormat_Qcow2 QemuDiskFormat = "qcow2"
	QemuDiskFormat_Qed   QemuDiskFormat = "qed"
	QemuDiskFormat_Vmdk  QemuDiskFormat = "vmdk"
	QemuDiskFormat_Raw   QemuDiskFormat = "raw"
)

func (QemuDiskFormat) Error() error {
	return fmt.Errorf("format can only be one of the following values: %s,%s,%s,%s,%s,%s,%s", QemuDiskFormat_Cow, QemuDiskFormat_Cloop, QemuDiskFormat_Qcow, QemuDiskFormat_Qcow2, QemuDiskFormat_Qed, QemuDiskFormat_Vmdk, QemuDiskFormat_Raw)
}

func (format *QemuDiskFormat) parse(path string) {
	index := strings.LastIndexByte(path, '.')
	if index == -1 {
		*format = QemuDiskFormat_Raw
		return
	}
	switch path[index+1:] {
	case "cow", "cloop", "qcow", "qcow2", "qed", "vmdk":
		*format = QemuDiskFormat(path[index+1:])
		return
	}
	*format = QemuDiskFormat_Raw
}

func (format QemuDiskFormat) String() string { return string(format) } // String is for fmt.Stringer.

func (format QemuDiskFormat) Validate() error {
	switch format {
	case QemuDiskFormat_Cow, QemuDiskFormat_Cloop, QemuDiskFormat_Qcow, QemuDiskFormat_Qcow2, QemuDiskFormat_Qed, QemuDiskFormat_Vmdk, QemuDiskFormat_Raw:
		return nil
	}
	return QemuDiskFormat("").Error()
}

type QemuDiskId string

const (
	ERROR_QemuDiskId_Invalid string = "invalid Disk ID"
)

func (id QemuDiskId) String() string { return string(id) } // String is for fmt.Stringer.

func (id QemuDiskId) Validate() error {
	if len(id) >= 7 {
		if id[0:6] == "virtio" {
			if id[6:] != "0" && strings.HasPrefix(string(id[6:]), "0") {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			number, err := strconv.Atoi(string(id[6:]))
			if err != nil {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			if number >= 0 && number <= 15 {
				return nil
			}
		}
	}
	if len(id) >= 5 {
		if id[0:4] == "sata" {
			if id[4:] != "0" && strings.HasPrefix(string(id[4:]), "0") {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			number, err := strconv.Atoi(string(id[4:]))
			if err != nil {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			if number >= 0 && number <= 5 {
				return nil
			}
		}
		if id[0:4] == "scsi" {
			if id[4:] != "0" && strings.HasPrefix(string(id[4:]), "0") {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			number, err := strconv.Atoi(string(id[4:]))
			if err != nil {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			if number >= 0 && number <= 30 {
				return nil
			}
		}
	}
	if len(id) == 4 {
		if id[0:3] == "ide" {
			number, err := strconv.Atoi(string(id[3]))
			if err != nil {
				return errors.New(ERROR_QemuDiskId_Invalid)
			}
			if number >= 0 && number <= 3 {
				return nil
			}
		}
	}
	return errors.New(ERROR_QemuDiskId_Invalid)
}

type qemuDiskMark struct {
	Format  QemuDiskFormat
	Id      QemuDiskId
	Size    QemuDiskSize
	Storage string
	Type    qemuDiskType
}

// Generate lists of disks that need to be moved and or resized
func (disk *qemuDiskMark) markChanges(currentDisk *qemuDiskMark, id QemuDiskId, changes *qemuUpdateChanges) {
	if disk == nil || currentDisk == nil {
		return
	}
	// Disk
	if disk.Size >= currentDisk.Size {
		// Update
		if disk.Size > currentDisk.Size {
			changes.Resize = append(changes.Resize, qemuDiskResize{
				Id:              id,
				SizeInKibibytes: disk.Size,
			})
		}
		if disk.Storage != currentDisk.Storage || disk.Format != currentDisk.Format {
			var format *QemuDiskFormat
			if disk.Format != currentDisk.Format {
				format = &disk.Format
			}
			changes.Move = append(changes.Move, qemuDiskMove{
				Format:  format,
				Id:      id,
				Storage: disk.Storage,
			})
		}
	}
}

type QemuDiskSerial string

const (
	Error_QemuDiskSerial_IllegalCharacter string = "serial may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_"
	Error_QemuDiskSerial_IllegalLength    string = "serial may only be 60 characters long"
)

func (serial QemuDiskSerial) String() string { return string(serial) } // For fmt.Stringer

// QemuDiskSerial may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_
// And has a max length of 60 characters
func (serial QemuDiskSerial) Validate() error {
	regex, _ := regexp.Compile(`^([a-z]|[A-Z]|[0-9]|_|-)*$`)
	if !regex.Match([]byte(serial)) {
		return errors.New(Error_QemuDiskSerial_IllegalCharacter)
	}
	if len(serial) > 60 {
		return errors.New(Error_QemuDiskSerial_IllegalLength)
	}
	return nil
}

// Amount of Kibibytes the disk should be.
// Disk size must be greater than 4096.
type QemuDiskSize uint

const (
	QemuDiskSize_Error_Minimum string       = "disk size must be greater than 4096"
	qemuDiskSize_Minimum       QemuDiskSize = 4097
	mebibyte                   QemuDiskSize = 1024
	gibibyte                   QemuDiskSize = 1048576
	tebibyte                   QemuDiskSize = 1073741824
)

func (QemuDiskSize) parse(rawSize string) (size QemuDiskSize) {
	tmpSize, _ := strconv.ParseInt(rawSize[:len(rawSize)-1], 10, 0)
	switch rawSize[len(rawSize)-1:] {
	case "T":
		size = QemuDiskSize(tmpSize) * tebibyte
	case "G":
		size = QemuDiskSize(tmpSize) * gibibyte
	case "M":
		size = QemuDiskSize(tmpSize) * mebibyte
	case "K":
		size = QemuDiskSize(tmpSize)
	}
	return
}

func (size QemuDiskSize) Validate() error {
	if size < qemuDiskSize_Minimum {
		return errors.New(QemuDiskSize_Error_Minimum)
	}
	return nil
}

type qemuDiskResize struct {
	Id              QemuDiskId
	SizeInKibibytes QemuDiskSize
}

// Increase the disk size to the specified amount in gigabytes
// Decrease of disk size is not permitted.
func (disk qemuDiskResize) resize(ctx context.Context, vmr *VmRef, client *Client) (exitStatus string, err error) {
	return client.PutWithTask(ctx, map[string]interface{}{"disk": disk.Id, "size": strconv.FormatInt(int64(disk.SizeInKibibytes), 10) + "K"}, fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId))
}

type qemuDiskMove struct {
	Format  *QemuDiskFormat
	Id      QemuDiskId
	Storage string
}

func (disk qemuDiskMove) mapToApiValues(delete bool) (params map[string]interface{}) {
	params = map[string]interface{}{"disk": string(disk.Id), "storage": string(disk.Storage)}
	if delete {
		params["delete"] = "1"
	}
	if disk.Format != nil {
		params["format"] = string(*disk.Format)
	}
	return
}

func (disk qemuDiskMove) move(ctx context.Context, delete bool, vmr *VmRef, client *Client) (exitStatus interface{}, err error) {
	return client.PostWithTask(ctx, disk.mapToApiValues(delete), fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmr.node, vmr.vmType, vmr.vmId))
}

func (disk qemuDiskMove) Validate() (err error) {
	if disk.Format != nil {
		err = disk.Format.Validate()
		if err != nil {
			return
		}
	}
	err = disk.Id.Validate()
	// TODO validate storage when it has custom type
	return
}

type qemuDiskType int

const (
	ide    qemuDiskType = 0
	sata   qemuDiskType = 1
	scsi   qemuDiskType = 2
	virtIO qemuDiskType = 3
)

type qemuStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *qemuDisk
	Passthrough *qemuDisk
	delete      bool
}

func (storage qemuStorage) mapToApiValues(currentStorage *qemuStorage, id QemuDiskId, params map[string]any, delete *strings.Builder) {
	if storage.delete {
		if currentStorage == nil {
			return
		}
		delete.WriteRune(',')
		delete.WriteString(id.String())
		return
	}
	// CDROM
	if storage.CdRom != nil {
		if currentStorage == nil || currentStorage.CdRom == nil { // Create
			params[id.String()] = storage.CdRom.mapToApiValues()
		} else { // Update
			cdRom := storage.CdRom.mapToApiValues()
			if cdRom != currentStorage.CdRom.mapToApiValues() {
				params[id.String()] = cdRom
			}
		}
		return
	}
	// CloudInit
	if storage.CloudInit != nil {
		if currentStorage == nil || currentStorage.CloudInit == nil { // Create
			params[id.String()] = storage.CloudInit.mapToApiValues()
		} else { // Update
			cloudInit := storage.CloudInit.mapToApiValues()
			if cloudInit != currentStorage.CloudInit.mapToApiValues() {
				params[id.String()] = cloudInit
			}
		}
		return
	}
	// Disk
	if storage.Disk != nil {
		if currentStorage == nil || currentStorage.Disk == nil { // Create
			params[id.String()] = storage.Disk.mapToApiValues(true)
		} else {
			if storage.Disk.SizeInKibibytes >= currentStorage.Disk.SizeInKibibytes { // Update
				storage.Disk.Id = currentStorage.Disk.Id
				storage.Disk.LinkedDiskId = currentStorage.Disk.LinkedDiskId
				storage.Disk.VolumePath = currentStorage.Disk.VolumePath
				disk := storage.Disk.mapToApiValues(false)
				if disk != currentStorage.Disk.mapToApiValues(false) {
					params[id.String()] = disk
				}
			} else { // Delete and Create
				// creating a disk on top of an existing disk is the same as detaching the disk and creating a new one.
				params[id.String()] = storage.Disk.mapToApiValues(true)
			}
		}
		return
	}
	// Passthrough
	if storage.Passthrough != nil {
		if currentStorage == nil || currentStorage.Passthrough == nil { // Create
			params[id.String()] = storage.Passthrough.mapToApiValues(false)
		} else { // Update
			passthrough := storage.Passthrough.mapToApiValues(false)
			if passthrough != currentStorage.Passthrough.mapToApiValues(false) {
				params[id.String()] = passthrough
			}
		}
	}
}

type QemuStorages struct {
	Ide    *QemuIdeDisks    `json:"ide,omitempty"`
	Sata   *QemuSataDisks   `json:"sata,omitempty"`
	Scsi   *QemuScsiDisks   `json:"scsi,omitempty"`
	VirtIO *QemuVirtIODisks `json:"virtio,omitempty"`
}

// Return the cloud init disk that should be removed.
func (newStorages QemuStorages) cloudInitRemove(currentStorages QemuStorages, delete *strings.Builder) {
	newCloudInit := newStorages.listCloudInitDisk()
	currentCloudInit := currentStorages.listCloudInitDisk()
	if newCloudInit != "" && currentCloudInit != "" && newCloudInit != currentCloudInit {
		delete.WriteString(comma + currentCloudInit)
	}
}

func (q QemuStorages) listCloudInitDisk() string {
	if q.Ide != nil {
		if disk := q.Ide.listCloudInitDisk(); disk != "" {
			return disk
		}
	}
	if q.Sata != nil {
		if disk := q.Sata.listCloudInitDisk(); disk != "" {
			return disk
		}
	}
	if q.Scsi != nil {
		if disk := q.Scsi.listCloudInitDisk(); disk != "" {
			return disk
		}
	}
	if q.VirtIO != nil {
		if disk := q.VirtIO.listCloudInitDisk(); disk != "" {
			return disk
		}
	}
	return ""
}

func (storages QemuStorages) mapToApiCreate(params map[string]any) {
	storages.mapToApiValues(QemuStorages{}, params, nil)
}

func (storages QemuStorages) mapToApiUpdate(current QemuStorages, params map[string]any, delete *strings.Builder) {
	storages.mapToApiValues(current, params, delete)
}

func (storages QemuStorages) mapToApiValues(currentStorages QemuStorages, params map[string]any, delete *strings.Builder) {
	if storages.Ide != nil {
		storages.Ide.mapToApiValues(currentStorages.Ide, params, delete)
	}
	if storages.Sata != nil {
		storages.Sata.mapToApiValues(currentStorages.Sata, params, delete)
	}
	if storages.Scsi != nil {
		storages.Scsi.mapToApiValues(currentStorages.Scsi, params, delete)
	}
	if storages.VirtIO != nil {
		storages.VirtIO.mapToApiValues(currentStorages.VirtIO, params, delete)
	}
}

func (raw *rawConfigQemu) GetDisks() (disks *QemuStorages, linkedID *GuestID) {
	tmpLinkedID := util.Pointer(GuestID(0))
	storage := QemuStorages{
		Ide:    raw.disksIde(tmpLinkedID),
		Sata:   raw.disksSata(tmpLinkedID),
		Scsi:   raw.disksSCSI(tmpLinkedID),
		VirtIO: raw.disksVirtIO(tmpLinkedID),
	}
	if *tmpLinkedID != 0 {
		linkedID = tmpLinkedID
	}
	if storage.Ide != nil || storage.Sata != nil || storage.Scsi != nil || storage.VirtIO != nil {
		return &storage, linkedID
	}
	return nil, nil
}

// mark disk that need to be moved or resized
func (storages QemuStorages) markDiskChanges(currentStorages QemuStorages) *qemuUpdateChanges {
	changes := &qemuUpdateChanges{}
	if storages.Ide != nil {
		storages.Ide.markDiskChanges(currentStorages.Ide, changes)
	}
	if storages.Sata != nil {
		storages.Sata.markDiskChanges(currentStorages.Sata, changes)
	}
	if storages.Scsi != nil {
		storages.Scsi.markDiskChanges(currentStorages.Scsi, changes)
	}
	if storages.VirtIO != nil {
		storages.VirtIO.markDiskChanges(currentStorages.VirtIO, changes)
	}
	return changes
}

// Select all new disks that do not have their size in gibigytes for resizing to the specified size
func (newStorages QemuStorages) selectInitialResize(currentStorages *QemuStorages) (resize []qemuDiskResize) {
	if currentStorages == nil {
		if newStorages.Ide != nil {
			resize = newStorages.Ide.selectInitialResize(nil)
		}
		if newStorages.Sata != nil {
			resize = append(resize, newStorages.Sata.selectInitialResize(nil)...)
		}
		if newStorages.Scsi != nil {
			resize = append(resize, newStorages.Scsi.selectInitialResize(nil)...)
		}
		if newStorages.VirtIO != nil {
			resize = append(resize, newStorages.VirtIO.selectInitialResize(nil)...)
		}
		return
	}
	if newStorages.Ide != nil {
		resize = newStorages.Ide.selectInitialResize(currentStorages.Ide)
	}
	if newStorages.Sata != nil {
		resize = append(resize, newStorages.Sata.selectInitialResize(currentStorages.Sata)...)
	}
	if newStorages.Scsi != nil {
		resize = append(resize, newStorages.Scsi.selectInitialResize(currentStorages.Scsi)...)
	}
	if newStorages.VirtIO != nil {
		resize = append(resize, newStorages.VirtIO.selectInitialResize(currentStorages.VirtIO)...)
	}
	return
}

func (storages QemuStorages) Validate() (err error) {
	var numberOfCloudInitDevices uint8
	var CloudInit uint8
	if storages.Ide != nil {
		CloudInit, err = storages.Ide.validate()
		if err != nil {
			return
		}
		numberOfCloudInitDevices += CloudInit
		if err = (QemuCloudInitDisk{}.checkDuplicates(numberOfCloudInitDevices)); err != nil {
			return
		}
	}
	if storages.Sata != nil {
		CloudInit, err = storages.Sata.validate()
		if err != nil {
			return
		}
		numberOfCloudInitDevices += CloudInit
		if err = (QemuCloudInitDisk{}.checkDuplicates(numberOfCloudInitDevices)); err != nil {
			return
		}
	}
	if storages.Scsi != nil {
		CloudInit, err = storages.Scsi.validate()
		if err != nil {
			return
		}
		numberOfCloudInitDevices += CloudInit
		if err = (QemuCloudInitDisk{}.checkDuplicates(numberOfCloudInitDevices)); err != nil {
			return
		}
	}
	if storages.VirtIO != nil {
		CloudInit, err = storages.VirtIO.validate()
		if err != nil {
			return
		}
		numberOfCloudInitDevices += CloudInit
		err = QemuCloudInitDisk{}.checkDuplicates(numberOfCloudInitDevices)
	}
	return
}

type qemuUpdateChanges struct {
	Move   []qemuDiskMove
	Resize []qemuDiskResize
}

type QemuWorldWideName string

const Error_QemuWorldWideName_Invalid string = "world wide name should be prefixed with 0x followed by 8 hexadecimal values"

var regexp_QemuWorldWideName = regexp.MustCompile(`^0x[0-9A-Fa-f]{16}$`)

func (wwn QemuWorldWideName) String() string { return string(wwn) } // For fmt.Stringer

func (wwn QemuWorldWideName) Validate() error {
	if wwn == "" || regexp_QemuWorldWideName.MatchString(string(wwn)) {
		return nil
	}
	return errors.New(Error_QemuWorldWideName_Invalid)
}

func diskSubtypeSet(set bool) error {
	if set {
		return errors.New(Error_QemuDisk_MutuallyExclusive)
	}
	return nil
}

func MoveQemuDisk(ctx context.Context, format *QemuDiskFormat, diskId QemuDiskId, storage string, deleteAfterMove bool, vmr *VmRef, client *Client) (err error) {
	disk := qemuDiskMove{
		Format:  format,
		Id:      diskId,
		Storage: storage,
	}
	err = disk.Validate()
	if err != nil {
		return
	}
	_, err = disk.move(ctx, deleteAfterMove, vmr, client)
	return
}

// increase Disks in size
func resizeDisks(ctx context.Context, vmr *VmRef, client *Client, disks []qemuDiskResize) (err error) {
	for _, e := range disks {
		_, err = e.resize(ctx, vmr, client)
		if err != nil {
			return
		}
	}
	return
}

// Resize newly created disks
func resizeNewDisks(ctx context.Context, vmr *VmRef, client *Client, newDisks, currentDisks *QemuStorages) (err error) {
	if newDisks == nil {
		return
	}
	resize := newDisks.selectInitialResize(currentDisks)
	if len(resize) > 0 {
		if err = resizeDisks(ctx, vmr, client, resize); err != nil {
			return
		}
	}
	return
}
