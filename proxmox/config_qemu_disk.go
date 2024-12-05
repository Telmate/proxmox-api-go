package proxmox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type diskSyntaxEnum bool

const (
	diskSyntaxFile   diskSyntaxEnum = false
	diskSyntaxVolume diskSyntaxEnum = true
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
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	Disk       bool // true = disk, false = passthrough
	EmulateSSD bool // Only set for ide,sata,scsi
	// TODO custom type
	File            string         // Only set for Passthrough.
	fileSyntax      diskSyntaxEnum // private enum to determine the syntax of the file path, as this changes depending on the type of backing storage. ie nfs, lvm, local, etc.
	Format          QemuDiskFormat // Only set for Disk
	Id              uint           // Only set for Disk
	IOThread        bool           // Only set for scsi,virtio
	LinkedDiskId    *uint          // Only set for Disk
	ReadOnly        bool           // Only set for scsi,virtio
	Replicate       bool
	Serial          QemuDiskSerial
	SizeInKibibytes QemuDiskSize
	// TODO custom type
	Storage       string // Only set for Disk
	Type          qemuDiskType
	WorldWideName QemuWorldWideName
}

const (
	Error_QemuDisk_File              string = "file may not be empty"
	Error_QemuDisk_MutuallyExclusive string = "settings cdrom,cloudinit,disk,passthrough are mutually exclusive"
	Error_QemuDisk_Storage           string = "storage may not be empty"
)

// create the disk string for the api
func (disk qemuDisk) formatDisk(vmID, LinkedVmId uint, currentStorage string, currentFormat QemuDiskFormat, syntax diskSyntaxEnum) (settings string) {
	tmpId := strconv.Itoa(int(vmID))
	// vm-100-disk-0
	settings = "vm-" + tmpId + "-disk-" + strconv.Itoa(int(disk.Id))
	switch syntax {
	case diskSyntaxFile:
		// format is ignored when syntax is diskSyntaxVolume
		// normal disk syntax
		// 100/vm-100-disk-0.raw
		settings = tmpId + "/" + settings + "." + string(disk.Format)
		if disk.LinkedDiskId != nil && disk.Storage == currentStorage && disk.Format == currentFormat {
			// linked clone disk syntax
			// 110/base-110-disk-1.raw/100/vm-100-disk-0.raw
			tmpId = strconv.Itoa(int(LinkedVmId))
			settings = tmpId + "/base-" + tmpId + "-disk-" + strconv.Itoa(int(*disk.LinkedDiskId)) + "." + string(disk.Format) + "/" + settings
		}
	case diskSyntaxVolume:
		// normal disk syntax
		// vm-100-disk-0
		if disk.LinkedDiskId != nil && disk.Storage == currentStorage {
			// linked clone disk syntax
			// base-110-disk-1/vm-100-disk-0
			tmpId = strconv.Itoa(int(LinkedVmId))
			settings = "base-" + tmpId + "-disk-" + strconv.Itoa(int(*disk.LinkedDiskId)) + "/" + settings
		}
	}
	// storage:100/vm-100-disk-0.raw
	// storage:110/base-110-disk-1.raw/100/vm-100-disk-0.raw
	// storage:vm-100-disk-0
	// storage:base-110-disk-1/vm-100-disk-0
	settings = disk.Storage + ":" + settings
	return
}

// Maps all the disk related settings to api values proxmox understands.
func (disk qemuDisk) mapToApiValues(vmID, LinkedVmId uint, currentStorage string, currentFormat QemuDiskFormat, syntax diskSyntaxEnum, create bool) (settings string) {
	if disk.Storage != "" {
		if create {
			if disk.SizeInKibibytes%gibibyte == 0 {
				settings = disk.Storage + ":" + strconv.FormatInt(int64(disk.SizeInKibibytes/gibibyte), 10)
			} else {
				settings = disk.Storage + ":0.001"
			}
		} else {
			settings = disk.formatDisk(vmID, LinkedVmId, currentStorage, currentFormat, syntax)
		}
	}

	if disk.File != "" {
		settings = disk.File
	}
	// Set File

	if disk.AsyncIO != "" {
		settings = settings + ",aio=" + string(disk.AsyncIO)
	}
	if !disk.Backup {
		settings = settings + ",backup=0"
	}
	if disk.Cache != "" {
		settings = settings + ",cache=" + string(disk.Cache)
	}
	if disk.Discard {
		settings = settings + ",discard=on"
	}
	if disk.Format != "" && create {
		settings = settings + ",format=" + string(disk.Format)
	}
	// media

	if disk.Bandwidth.Iops.ReadLimit.Concurrent != 0 {
		settings = settings + ",iops_rd=" + strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.ReadLimit.Burst != 0 {
		settings = settings + ",iops_rd_max=" + strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Burst))
	}
	if disk.Bandwidth.Iops.ReadLimit.BurstDuration != 0 {
		settings = settings + ",iops_rd_max_length=" + strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.BurstDuration))
	}

	if disk.Bandwidth.Iops.WriteLimit.Concurrent != 0 {
		settings = settings + ",iops_wr=" + strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.WriteLimit.Burst != 0 {
		settings = settings + ",iops_wr_max=" + strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Burst))
	}
	if disk.Bandwidth.Iops.WriteLimit.BurstDuration != 0 {
		settings = settings + ",iops_wr_max_length=" + strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.BurstDuration))
	}

	if (disk.Type == scsi || disk.Type == virtIO) && disk.IOThread {
		settings = settings + ",iothread=1"
	}

	if disk.Bandwidth.MBps.ReadLimit.Concurrent != 0 {
		settings = settings + ",mbps_rd=" + floatToTrimmedString(float64(disk.Bandwidth.MBps.ReadLimit.Concurrent), 2)
	}
	if disk.Bandwidth.MBps.ReadLimit.Burst != 0 {
		settings = settings + ",mbps_rd_max=" + floatToTrimmedString(float64(disk.Bandwidth.MBps.ReadLimit.Burst), 2)
	}

	if disk.Bandwidth.MBps.WriteLimit.Concurrent != 0 {
		settings = settings + ",mbps_wr=" + floatToTrimmedString(float64(disk.Bandwidth.MBps.WriteLimit.Concurrent), 2)
	}
	if disk.Bandwidth.MBps.WriteLimit.Burst != 0 {
		settings = settings + ",mbps_wr_max=" + floatToTrimmedString(float64(disk.Bandwidth.MBps.WriteLimit.Burst), 2)
	}

	if !disk.Replicate {
		settings = settings + ",replicate=0"
	}
	if (disk.Type == scsi || disk.Type == virtIO) && disk.ReadOnly {
		settings = settings + ",ro=1"
	}
	if disk.Serial != "" {
		settings = settings + ",serial=" + string(disk.Serial)
	}
	if disk.Type != virtIO && disk.EmulateSSD {
		settings = settings + ",ssd=1"
	}
	if disk.WorldWideName != "" {
		settings = settings + ",wwn=" + string(disk.WorldWideName)
	}
	return
}

// Maps all the disk related settings to our own data structure.
func (qemuDisk) mapToStruct(diskData string, settings map[string]string, linkedVmId *uint) *qemuDisk {
	disk := qemuDisk{Backup: true}

	if diskData[0:1] == "/" {
		disk.File = diskData
	} else {
		disk.Id, disk.Storage, disk.Format, disk.LinkedDiskId, disk.fileSyntax = qemuDisk{}.parseDisk(diskData, linkedVmId)
	}

	if len(settings) == 0 {
		return nil
	}

	// Replicate defaults to true
	disk.Replicate = true

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
	return &disk
}

// parse and extract the values from the disk data
// storage:110/base-110-disk-1.qcow2/100/vm-100-disk-0.qcow2
// storage:100/vm-100-disk-0.qcow2
// storage:base-110-disk-1/vm-100-disk-0
// storage:vm-100-disk-0
func (qemuDisk) parseDisk(diskData string, linkedVmId *uint) (diskId uint, storage string, format QemuDiskFormat, linkedDiskId *uint, syntax diskSyntaxEnum) {
	parts := strings.Split(diskData, ":")
	storage = parts[0]

	if len(parts) != 2 {
		return
	}

	pathParts := strings.Split(parts[1], "/")
	switch len(pathParts) {
	case 1:
		syntax = diskSyntaxVolume
	case 2:
		if _, err := strconv.Atoi(pathParts[0]); err != nil { // linked clone
			tmp := strings.Split(strings.Split(pathParts[0], ".")[0], "-")
			if len(tmp) > 1 {
				if tmpVmId, err := strconv.Atoi(tmp[1]); err == nil {
					*linkedVmId = uint(tmpVmId)
				}
				if tmpDiskId, err := strconv.Atoi(tmp[len(tmp)-1]); err == nil {
					tmpDiskIdPointer := uint(tmpDiskId)
					linkedDiskId = &tmpDiskIdPointer
				}
				syntax = diskSyntaxVolume
			}
		} else {
			syntax = diskSyntaxFile
		}
	case 4: // Linked Clone
		if tmpVmId, err := strconv.Atoi(pathParts[0]); err == nil {
			*linkedVmId = uint(tmpVmId)
		}
		tmp := strings.Split(strings.Split(pathParts[1], ".")[0], "-")
		if len(tmp) > 1 {
			if tmpDiskId, err := strconv.Atoi(tmp[len(tmp)-1]); err == nil {
				tmpDiskIdPointer := uint(tmpDiskId)
				linkedDiskId = &tmpDiskIdPointer
			}
		}
		syntax = diskSyntaxFile
	}

	diskNameAndFormat := strings.Split(pathParts[len(pathParts)-1], ".")
	if len(diskNameAndFormat) > 0 {
		tmp := strings.Split(diskNameAndFormat[0], "-")
		if len(tmp) > 1 {
			if tmpDiskId, err := strconv.Atoi(tmp[len(tmp)-1]); err == nil {
				diskId = uint(tmpDiskId)
			}
		}

		// set disk format, default to raw
		if len(diskNameAndFormat) == 2 {
			format = QemuDiskFormat(diskNameAndFormat[1])
		} else {
			format = QemuDiskFormat_Raw
		}
	}

	return
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
		if err = disk.Format.Validate(); err != nil {
			return
		}
		if err = disk.SizeInKibibytes.Validate(); err != nil {
			return
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
	Error_QemuDiskBandwidthIopsLimitBurst string = "burst may not be lower then 10 except for 0"
)

func (limit QemuDiskBandwidthIopsLimitBurst) Validate() error {
	if limit != 0 && limit < 10 {
		return errors.New(Error_QemuDiskBandwidthIopsLimitBurst)
	}
	return nil
}

type QemuDiskBandwidthIopsLimitConcurrent uint

const (
	Error_QemuDiskBandwidthIopsLimitConcurrent string = "concurrent may not be lower then 10 except for 0"
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
	Error_QemuDiskBandwidthMBpsLimitBurst string = "burst may not be lower then 1 except for 0"
)

type QemuDiskBandwidthMBpsLimitBurst float32

func (limit QemuDiskBandwidthMBpsLimitBurst) Validate() error {
	if limit != 0 && limit < 1 {
		return errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)
	}
	return nil
}

const (
	Error_QemuDiskBandwidthMBpsLimitConcurrent string = "concurrent may not be lower then 1 except for 0"
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
// Disk size must be greater then 4096.
type QemuDiskSize uint

const (
	QemuDiskSize_Error_Minimum string       = "disk size must be greater then 4096"
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
}

func (storage *qemuStorage) mapToApiValues(currentStorage *qemuStorage, vmID, linkedVmId uint, id QemuDiskId, params map[string]interface{}, delete string) string {
	if storage == nil {
		return delete
	}
	// CDROM
	if storage.CdRom != nil {
		if currentStorage == nil || currentStorage.CdRom == nil {
			// Create
			params[string(id)] = storage.CdRom.mapToApiValues()
		} else {
			// Update
			cdRom := storage.CdRom.mapToApiValues()
			if cdRom != currentStorage.CdRom.mapToApiValues() {
				params[string(id)] = cdRom
			}
		}
		return delete
	} else if currentStorage != nil && currentStorage.CdRom != nil && storage.CloudInit == nil && storage.Disk == nil && storage.Passthrough == nil {
		// Delete
		return AddToList(delete, string(id))
	}
	// CloudInit
	if storage.CloudInit != nil {
		if currentStorage == nil || currentStorage.CloudInit == nil {
			// Create
			params[string(id)] = storage.CloudInit.mapToApiValues()
		} else {
			// Update
			cloudInit := storage.CloudInit.mapToApiValues()
			if cloudInit != currentStorage.CloudInit.mapToApiValues() {
				params[string(id)] = cloudInit
			}
		}
		return delete
	} else if currentStorage != nil && currentStorage.CloudInit != nil && storage.Disk == nil && storage.Passthrough == nil {
		// Delete
		return AddToList(delete, string(id))
	}
	// Disk
	if storage.Disk != nil {
		if currentStorage == nil || currentStorage.Disk == nil {
			// Create
			params[string(id)] = storage.Disk.mapToApiValues(vmID, 0, "", "", false, true)
		} else {
			if storage.Disk.SizeInKibibytes >= currentStorage.Disk.SizeInKibibytes {
				// Update
				storage.Disk.Id = currentStorage.Disk.Id
				storage.Disk.LinkedDiskId = currentStorage.Disk.LinkedDiskId
				disk := storage.Disk.mapToApiValues(vmID, linkedVmId, currentStorage.Disk.Storage, currentStorage.Disk.Format, currentStorage.Disk.fileSyntax, false)
				if disk != currentStorage.Disk.mapToApiValues(vmID, linkedVmId, currentStorage.Disk.Storage, currentStorage.Disk.Format, currentStorage.Disk.fileSyntax, false) {
					params[string(id)] = disk
				}
			} else {
				// Delete and Create
				// creating a disk on top of an existing disk is the same as detaching the disk and creating a new one.
				params[string(id)] = storage.Disk.mapToApiValues(vmID, 0, "", "", false, true)
			}
		}
		return delete
	} else if currentStorage != nil && currentStorage.Disk != nil && storage.Passthrough == nil {
		// Delete
		return AddToList(delete, string(id))
	}
	// Passthrough
	if storage.Passthrough != nil {
		if currentStorage == nil || currentStorage.Passthrough == nil {
			// Create
			params[string(id)] = storage.Passthrough.mapToApiValues(0, 0, "", "", false, false)
		} else {
			// Update
			passthrough := storage.Passthrough.mapToApiValues(0, 0, "", "", false, false)
			if passthrough != currentStorage.Passthrough.mapToApiValues(0, 0, "", "", false, false) {
				params[string(id)] = passthrough
			}
		}
		return delete
	} else if currentStorage != nil && currentStorage.Passthrough != nil {
		// Delete
		return AddToList(delete, string(id))
	}
	// Delete if no subtype was specified
	if currentStorage != nil {
		return AddToList(delete, string(id))
	}
	return delete
}

type QemuStorages struct {
	Ide    *QemuIdeDisks    `json:"ide,omitempty"`
	Sata   *QemuSataDisks   `json:"sata,omitempty"`
	Scsi   *QemuScsiDisks   `json:"scsi,omitempty"`
	VirtIO *QemuVirtIODisks `json:"virtio,omitempty"`
}

// Return the cloud init disk that should be removed.
func (newStorages QemuStorages) cloudInitRemove(currentStorages QemuStorages) string {
	newCloudInit := newStorages.listCloudInitDisk()
	currentCloudInit := currentStorages.listCloudInitDisk()
	if newCloudInit != "" && currentCloudInit != "" && newCloudInit != currentCloudInit {
		return currentCloudInit
	}
	return ""
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

func (storages QemuStorages) mapToApiValues(currentStorages QemuStorages, vmID, linkedVmId uint, params map[string]interface{}) (delete string) {
	if storages.Ide != nil {
		delete = storages.Ide.mapToApiValues(currentStorages.Ide, vmID, linkedVmId, params, delete)
	}
	if storages.Sata != nil {
		delete = storages.Sata.mapToApiValues(currentStorages.Sata, vmID, linkedVmId, params, delete)
	}
	if storages.Scsi != nil {
		delete = storages.Scsi.mapToApiValues(currentStorages.Scsi, vmID, linkedVmId, params, delete)
	}
	if storages.VirtIO != nil {
		delete = storages.VirtIO.mapToApiValues(currentStorages.VirtIO, vmID, linkedVmId, params, delete)
	}
	return delete
}

func (QemuStorages) mapToStruct(params map[string]interface{}, linkedVmId *uint) *QemuStorages {
	storage := QemuStorages{
		Ide:    QemuIdeDisks{}.mapToStruct(params, linkedVmId),
		Sata:   QemuSataDisks{}.mapToStruct(params, linkedVmId),
		Scsi:   QemuScsiDisks{}.mapToStruct(params, linkedVmId),
		VirtIO: QemuVirtIODisks{}.mapToStruct(params, linkedVmId),
	}
	if storage.Ide != nil || storage.Sata != nil || storage.Scsi != nil || storage.VirtIO != nil {
		return &storage
	}
	return nil
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
