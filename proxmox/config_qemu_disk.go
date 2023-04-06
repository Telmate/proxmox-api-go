package proxmox

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type IsoFile struct {
	File    string `json:"file"`
	Storage string `json:"storage"`
	// Size can only be retrieved, setting it has no effect
	Size string `json:"size"`
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
				Storage: settings.Storage,
				File:    settings.File,
				Size:    settings.Size,
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
	Passthrough bool
	Storage     string
	Format      QemuDiskFormat // Only set for Cloud-init drives
	File        string
	Size        string
}

func (qemuCdRom) mapToStruct(diskData string, settings map[string]interface{}) *qemuCdRom {
	var isCdRom bool
	if setting, isSet := settings["media"]; isSet {
		isCdRom = setting.(string) == "cdrom"
	}
	if !isCdRom {
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
		if len(tmpFile) == 2 {
			tmpFileType := strings.Split(tmpFile[1], ".")
			if len(tmpFileType) > 1 {
				fileType := QemuDiskFormat(tmpFileType[len(tmpFileType)-1])
				if fileType == "iso" {
					if setting, isSet := settings["size"]; isSet {
						return &qemuCdRom{
							CdRom:   true,
							Storage: tmpStorage[0],
							File:    tmpFile[1],
							Size:    setting.(string),
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
	File         string         // Only set for Passthrough.
	Format       QemuDiskFormat // Only set for Disk
	Id           uint           // Only set for Disk
	IOThread     bool           // Only set for scsi,virtio
	LinkedDiskId *uint          // Only set for Disk
	ReadOnly     bool           // Only set for scsi,virtio
	Replicate    bool
	Serial       QemuDiskSerial
	Size         uint
	// TODO custom type
	Storage string // Only set for Disk
	Type    qemuDiskType
}

const (
	Error_QemuDisk_File              string = "file may not be empty"
	Error_QemuDisk_MutuallyExclusive string = "settings cdrom,cloudinit,disk,passthrough are mutually exclusive"
	Error_QemuDisk_Size              string = "size must be greater then 0"
	Error_QemuDisk_Storage           string = "storage may not be empty"
)

// Maps all the disk related settings to api values proxmox understands.
func (disk qemuDisk) mapToApiValues(vmID, LinkedVmId uint, currentStorage string, currentFormat QemuDiskFormat, create bool) (settings string) {
	if disk.Storage != "" {
		if create {
			settings = disk.Storage + ":" + strconv.Itoa(int(disk.Size))
		} else {
			tmpId := strconv.Itoa(int(vmID))
			settings = tmpId + "/vm-" + tmpId + "-disk-" + strconv.Itoa(int(disk.Id)) + "." + string(disk.Format)
			// storage:100/vm-100-disk-0.raw
			if disk.LinkedDiskId != nil && disk.Storage == currentStorage && disk.Format == currentFormat {
				// storage:110/base-110-disk-1.raw/100/vm-100-disk-0.raw
				tmpId = strconv.Itoa(int(LinkedVmId))
				settings = tmpId + "/base-" + tmpId + "-disk-" + strconv.Itoa(int(*disk.LinkedDiskId)) + "." + string(disk.Format) + "/" + settings
			}
			settings = disk.Storage + ":" + settings
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
		settings = settings + fmt.Sprintf(",mbps_rd=%.2f", disk.Bandwidth.MBps.ReadLimit.Concurrent)
	}
	if disk.Bandwidth.MBps.ReadLimit.Burst != 0 {
		settings = settings + fmt.Sprintf(",mbps_rd_max=%.2f", disk.Bandwidth.MBps.ReadLimit.Burst)
	}

	if disk.Bandwidth.MBps.WriteLimit.Concurrent != 0 {
		settings = settings + fmt.Sprintf(",mbps_wr=%.2f", disk.Bandwidth.MBps.WriteLimit.Concurrent)
	}
	if disk.Bandwidth.MBps.WriteLimit.Burst != 0 {
		settings = settings + fmt.Sprintf(",mbps_wr_max=%.2f", disk.Bandwidth.MBps.WriteLimit.Burst)
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

	return
}

// Maps all the disk related settings to our own data structure.
func (qemuDisk) mapToStruct(diskData string, settings map[string]interface{}, linkedVmId *uint) *qemuDisk {
	disk := qemuDisk{Backup: true}

	if diskData[0:1] == "/" {
		disk.File = diskData
	} else {
		// storage:110/base-110-disk-1.qcow2/100/vm-100-disk-0.qcow2
		// storage:100/vm-100-disk-0.qcow2
		diskAndPath := strings.Split(diskData, ":")
		disk.Storage = diskAndPath[0]
		if len(diskAndPath) == 2 {
			pathParts := strings.Split(diskAndPath[1], "/")
			if len(pathParts) == 4 {
				var tmpDiskId int
				tmpVmId, _ := strconv.Atoi(pathParts[0])
				*linkedVmId = uint(tmpVmId)
				tmp := strings.Split(strings.Split(pathParts[1], ".")[0], "-")
				if len(tmp) > 1 {
					tmpDiskId, _ = strconv.Atoi(tmp[len(tmp)-1])
					tmpDiskIdPointer := uint(tmpDiskId)
					disk.LinkedDiskId = &tmpDiskIdPointer
				}
			}
			if len(pathParts) > 1 {
				diskNameAndFormat := strings.Split(pathParts[len(pathParts)-1], ".")
				if len(diskNameAndFormat) == 2 {
					disk.Format = QemuDiskFormat(diskNameAndFormat[1])
					tmp := strings.Split(diskNameAndFormat[0], "-")
					if len(tmp) > 1 {
						tmpDiskId, _ := strconv.Atoi(tmp[len(tmp)-1])
						disk.Id = uint(tmpDiskId)
					}
				}
			}
		}
	}

	if len(settings) == 0 {
		return nil
	}

	// Replicate defaults to true
	disk.Replicate = true

	if value, isSet := settings["aio"]; isSet {
		disk.AsyncIO = QemuDiskAsyncIO(value.(string))
	}
	if value, isSet := settings["backup"]; isSet {
		disk.Backup, _ = strconv.ParseBool(value.(string))
	}
	if value, isSet := settings["cache"]; isSet {
		disk.Cache = QemuDiskCache(value.(string))
	}
	if value, isSet := settings["discard"]; isSet {
		if value.(string) == "on" {
			disk.Discard = true
		}
	}
	if value, isSet := settings["iops_rd"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.ReadLimit.Concurrent = QemuDiskBandwidthIopsLimitConcurrent(tmp)
	}
	if value, isSet := settings["iops_rd_max"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.ReadLimit.Burst = QemuDiskBandwidthIopsLimitBurst(tmp)
	}
	if value, isSet := settings["iops_rd_max_length"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.ReadLimit.BurstDuration = uint(tmp)
	}
	if value, isSet := settings["iops_wr"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.WriteLimit.Concurrent = QemuDiskBandwidthIopsLimitConcurrent(tmp)
	}
	if value, isSet := settings["iops_wr_max"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.WriteLimit.Burst = QemuDiskBandwidthIopsLimitBurst(tmp)
	}
	if value, isSet := settings["iops_wr_max_length"]; isSet {
		tmp, _ := strconv.Atoi(value.(string))
		disk.Bandwidth.Iops.WriteLimit.BurstDuration = uint(tmp)
	}
	if value, isSet := settings["iothread"]; isSet {
		disk.IOThread, _ = strconv.ParseBool(value.(string))
	}
	if value, isSet := settings["mbps_rd"]; isSet {
		tmp, _ := strconv.ParseFloat(value.(string), 32)
		disk.Bandwidth.MBps.ReadLimit.Concurrent = QemuDiskBandwidthMBpsLimitConcurrent(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_rd_max"]; isSet {
		tmp, _ := strconv.ParseFloat(value.(string), 32)
		disk.Bandwidth.MBps.ReadLimit.Burst = QemuDiskBandwidthMBpsLimitBurst(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_wr"]; isSet {
		tmp, _ := strconv.ParseFloat(value.(string), 32)
		disk.Bandwidth.MBps.WriteLimit.Concurrent = QemuDiskBandwidthMBpsLimitConcurrent(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["mbps_wr_max"]; isSet {
		tmp, _ := strconv.ParseFloat(value.(string), 32)
		disk.Bandwidth.MBps.WriteLimit.Burst = QemuDiskBandwidthMBpsLimitBurst(math.Round(tmp*100) / 100)
	}
	if value, isSet := settings["replicate"]; isSet {
		disk.Replicate, _ = strconv.ParseBool(value.(string))
	}
	if value, isSet := settings["ro"]; isSet {
		disk.ReadOnly, _ = strconv.ParseBool(value.(string))
	}
	if value, isSet := settings["serial"]; isSet {
		disk.Serial = QemuDiskSerial(value.(string))
	}
	if value, isSet := settings["size"]; isSet {
		sizeString := value.(string)
		if len(sizeString) > 1 {
			diskSize, _ := strconv.Atoi(sizeString[0 : len(sizeString)-1])
			switch sizeString[len(sizeString)-1:] {
			case "G":
				disk.Size = uint(diskSize)
			case "T":
				disk.Size = uint(diskSize * 1024)
			}
		}
	}
	if value, isSet := settings["ssd"]; isSet {
		disk.EmulateSSD, _ = strconv.ParseBool(value.(string))
	}
	return &disk
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
	if disk.Disk {
		// disk
		if err = disk.Format.Validate(); err != nil {
			return
		}
		if disk.Size == 0 {
			return errors.New(Error_QemuDisk_Size)
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
	Size    uint
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
				SizeInGigaBytes: disk.Size,
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

type qemuDiskResize struct {
	Id              QemuDiskId
	SizeInGigaBytes uint
}

// Increase the disk size to the specified amount in gigabytes
// Decrease of disk size is not permitted.
func (disk qemuDiskResize) resize(vmr *VmRef, client *Client) (exitStatus string, err error) {
	return client.PutWithTask(map[string]interface{}{"disk": disk.Id, "size": strconv.Itoa(int(disk.SizeInGigaBytes)) + "G"}, fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId))
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

func (disk qemuDiskMove) move(delete bool, vmr *VmRef, client *Client) (exitStatus interface{}, err error) {
	return client.PostWithTask(disk.mapToApiValues(delete), fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmr.node, vmr.vmType, vmr.vmId))
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
			params[string(id)] = storage.Disk.mapToApiValues(vmID, 0, "", "", true)
		} else {
			if storage.Disk.Size >= currentStorage.Disk.Size {
				// Update
				storage.Disk.Id = currentStorage.Disk.Id
				storage.Disk.LinkedDiskId = currentStorage.Disk.LinkedDiskId
				disk := storage.Disk.mapToApiValues(vmID, linkedVmId, currentStorage.Disk.Storage, currentStorage.Disk.Format, false)
				if disk != currentStorage.Disk.mapToApiValues(vmID, linkedVmId, currentStorage.Disk.Storage, currentStorage.Disk.Format, false) {
					params[string(id)] = disk
				}
			} else {
				// Delete and Create
				// creating a disk on top of an existing disk is the same as detaching the disk and creating a new one.
				params[string(id)] = storage.Disk.mapToApiValues(vmID, 0, "", "", true)
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
			params[string(id)] = storage.Passthrough.mapToApiValues(0, 0, "", "", false)
		} else {
			// Update
			passthrough := storage.Passthrough.mapToApiValues(0, 0, "", "", false)
			if passthrough != currentStorage.Passthrough.mapToApiValues(0, 0, "", "", false) {
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

func diskSubtypeSet(set bool) error {
	if set {
		return errors.New(Error_QemuDisk_MutuallyExclusive)
	}
	return nil
}

func MoveQemuDisk(format *QemuDiskFormat, diskId QemuDiskId, storage string, deleteAfterMove bool, vmr *VmRef, client *Client) (err error) {
	disk := qemuDiskMove{
		Format:  format,
		Id:      diskId,
		Storage: storage,
	}
	err = disk.Validate()
	if err != nil {
		return
	}
	_, err = disk.move(deleteAfterMove, vmr, client)
	return
}
