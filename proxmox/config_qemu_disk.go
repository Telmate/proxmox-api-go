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
	Storage string `json:"storage"`
	File    string `json:"file"`
	// Size can only be retrieved, setting it has no effect
	Size string `json:"size"`
}

type QemuCdRom struct {
	Iso *IsoFile `json:"iso,omitempty"`
	// Passthrough and File are mutually exclusive
	Passthrough bool `json:"passthrough,omitempty"`
}

// TODO write test
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
	if !settings.Passthrough {
		return &QemuCdRom{
			Iso: &IsoFile{
				Storage: settings.Storage,
				File:    settings.File,
				Size:    settings.Size,
			},
		}
	}
	return &QemuCdRom{Passthrough: false}
}

type qemuCdRom struct {
	// "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"
	Passthrough bool
	Storage     string
	// FileType is only set for Cloud init drives, this value will be used to check if it is a normal cdrom or cloud init drive.
	FileType string
	File     string
	Size     string
}

func (qemuCdRom) mapToStruct(settings [][]string) *qemuCdRom {
	var isCdRom bool
	for _, e := range settings {
		if e[0] == "media" {
			if e[1] == "cdrom" {
				isCdRom = true
				break
			}
		}
	}
	if !isCdRom {
		return nil
	}
	if settings[0][0] == "none" {
		return &qemuCdRom{}
	}
	if settings[0][0] == "cdrom" {
		return &qemuCdRom{Passthrough: true}
	}
	tmpStorage := strings.Split(settings[0][0], ":")
	if len(tmpStorage) > 1 {
		tmpFile := strings.Split(settings[0][0], "/")
		if len(tmpFile) == 2 {
			tmpFileType := strings.Split(tmpFile[1], ".")
			if len(tmpFileType) > 1 {
				fileType := tmpFileType[len(tmpFileType)-1]
				if fileType == "iso" {
					for _, e := range settings {
						if e[0] == "size" {
							return &qemuCdRom{
								Storage: tmpStorage[0],
								File:    tmpFile[1],
								Size:    e[1],
							}
						}
					}
				} else {
					return &qemuCdRom{
						Storage:  tmpStorage[0],
						File:     tmpFile[1],
						FileType: fileType,
					}
				}
			}
		}
	}
	return nil
}

type QemuCloudInitDisk struct {
	Storage  string
	FileType string
}

// TODO write function
func (cloudInit QemuCloudInitDisk) mapToApiValues() string {
	return ""
}

func (QemuCloudInitDisk) mapToStruct(settings qemuCdRom) *QemuCloudInitDisk {
	return &QemuCloudInitDisk{
		Storage:  settings.Storage,
		FileType: settings.FileType,
	}
}

type qemuDisk struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool // Only set for ide,sata,scsi
	// TODO custom type
	File      string // Only set for Passthrough.
	Format    QemuDiskFormat
	ID        uint // Only set for Disk
	IOThread  bool // Only set for scsi,virtio
	Number    uint
	ReadOnly  bool // Only set for scsi,virtio
	Replicate bool
	Serial    QemuDiskSerial
	Size      uint
	// TODO custom type
	Storage string // Only set for Disk
	Type    qemuDiskType
}

// TODO write test
func (disk qemuDisk) mapToApiValues(create bool) (settings string) {
	if create {
		if disk.Storage != "" {
			settings = disk.Storage + ":" + strconv.Itoa(int(disk.Size))
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
	// format
	// media

	if disk.Bandwidth.ReadLimit_Iops.Concurrent >= 10 {
		settings = settings + ",iops_rd=" + strconv.Itoa(int(disk.Bandwidth.ReadLimit_Iops.Concurrent))
	}
	if disk.Bandwidth.ReadLimit_Iops.Burst >= 10 {
		settings = settings + ",iops_rd_max=" + strconv.Itoa(int(disk.Bandwidth.ReadLimit_Iops.Burst))
	}
	if disk.Bandwidth.WriteLimit_Iops.Concurrent >= 10 {
		settings = settings + ",iops_wr=" + strconv.Itoa(int(disk.Bandwidth.WriteLimit_Iops.Concurrent))
	}
	if disk.Bandwidth.WriteLimit_Iops.Burst >= 10 {
		settings = settings + ",iops_wr_max=" + strconv.Itoa(int(disk.Bandwidth.WriteLimit_Iops.Burst))
	}

	if (disk.Type == scsi || disk.Type == virtIO) && disk.IOThread {
		settings = settings + ",iothread=1"
	}

	if disk.Bandwidth.ReadLimit_Data.Concurrent >= float32(1) {
		settings = settings + fmt.Sprintf(",mbps_rd=%.2f", disk.Bandwidth.ReadLimit_Data.Concurrent)
	}
	if disk.Bandwidth.ReadLimit_Data.Burst >= float32(1) {
		settings = settings + fmt.Sprintf(",mbps_rd_max=%.2f", disk.Bandwidth.ReadLimit_Data.Burst)
	}
	if disk.Bandwidth.WriteLimit_Data.Concurrent >= float32(1) {
		settings = settings + fmt.Sprintf(",mbps_wr=%.2f", disk.Bandwidth.WriteLimit_Data.Concurrent)
	}
	if disk.Bandwidth.WriteLimit_Data.Burst >= float32(1) {
		settings = settings + fmt.Sprintf(",mbps_wr_max=%.2f", disk.Bandwidth.WriteLimit_Data.Burst)
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

// Maps all the disk related settings
// TODO write test
func (qemuDisk) mapToStruct(settings [][]string) *qemuDisk {
	if len(settings) == 0 {
		return nil
	}
	disk := qemuDisk{Backup: true}

	if settings[0][0][0:1] == "/" {
		disk.File = settings[0][0]
	} else {
		// "test2:105/vm-105-disk-53.qcow2,
		diskAndNumberAndFormat := strings.Split(settings[0][0], ":")
		disk.Storage = diskAndNumberAndFormat[0]
		if len(diskAndNumberAndFormat) == 2 {
			idAndFormat := strings.Split(diskAndNumberAndFormat[1], "-")
			if len(idAndFormat) == 2 {
				tmp := strings.Split(idAndFormat[1], ".")
				tmpId, _ := strconv.Atoi(tmp[0])
				disk.ID = uint(tmpId)
				if len(tmp) == 2 {
					disk.Format = QemuDiskFormat(tmp[1])
				}
			}
		}
	}

	for _, e := range settings {
		if e[0] == "aio" {
			disk.AsyncIO = QemuDiskAsyncIO(e[1])
			continue
		}
		if e[0] == "backup" {
			disk.Backup, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "cache" {
			disk.Cache = QemuDiskCache(e[1])
			continue
		}
		if e[0] == "discard" {
			disk.Discard, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "iops_rd" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.ReadLimit_Iops.Concurrent = uint(tmp)
		}
		if e[0] == "iops_rd_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.ReadLimit_Iops.Burst = uint(tmp)
		}
		if e[0] == "iops_wr" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.WriteLimit_Iops.Concurrent = uint(tmp)
		}
		if e[0] == "iops_wr_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.WriteLimit_Iops.Burst = uint(tmp)
		}
		if e[0] == "iothread" {
			disk.IOThread, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "mbps_rd" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.ReadLimit_Data.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_rd_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.ReadLimit_Data.Burst = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.WriteLimit_Data.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.WriteLimit_Data.Burst = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "replicate" {
			disk.Replicate, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "ro" {
			disk.ReadOnly, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "serial" {
			disk.Serial = QemuDiskSerial(e[1])
			continue
		}
		if e[0] == "size" {
			diskSize, _ := strconv.Atoi(strings.TrimSuffix(e[1], "G"))
			disk.Size = uint(diskSize)
			continue
		}
		if e[0] == "ssd" {
			disk.EmulateSSD, _ = strconv.ParseBool(e[1])
		}
	}
	return &disk
}

type QemuDiskAsyncIO string

const (
	QemuDiskAsyncIO_Native  QemuDiskAsyncIO = "native"
	QemuDiskAsyncIO_Threads QemuDiskAsyncIO = "threads"
	QemuDiskAsyncIO_IOuring QemuDiskAsyncIO = "io_uring"
)

func (asyncIO QemuDiskAsyncIO) Validate() error {
	switch asyncIO {
	case QemuDiskAsyncIO_Native, QemuDiskAsyncIO_Threads, QemuDiskAsyncIO_IOuring:
		return nil
	}
	return fmt.Errorf("asyncio can only be one of the following values: %s,%s,%s", QemuDiskAsyncIO_Native, QemuDiskAsyncIO_Threads, QemuDiskAsyncIO_IOuring)
}

type QemuDiskBandwidth struct {
	ReadLimit_Data  QemuDisk_Bandwidth_Data
	WriteLimit_Data QemuDisk_Bandwidth_Data
	ReadLimit_Iops  QemuDisk_Bandwidth_Iops
	WriteLimit_Iops QemuDisk_Bandwidth_Iops
}

type QemuDisk_Bandwidth_Data struct {
	Concurrent float32
	Burst      float32
}

type QemuDisk_Bandwidth_Iops struct {
	Concurrent uint
	Burst      uint
}

type QemuDiskCache string

const (
	QemuDiskCache_None         QemuDiskCache = "none"
	QemuDiskCache_WriteThrough QemuDiskCache = "writethrough"
	QemuDiskCache_WriteBack    QemuDiskCache = "writeback"
	QemuDiskCache_Unsafe       QemuDiskCache = "unsafe"
	QemuDiskCache_DirectSync   QemuDiskCache = "directsync"
)

func (cache QemuDiskCache) Validate() error {
	switch cache {
	case QemuDiskCache_None, QemuDiskCache_WriteThrough, QemuDiskCache_WriteBack, QemuDiskCache_Unsafe, QemuDiskCache_DirectSync:
		return nil
	}
	return fmt.Errorf("cache can only be one of the following values: %s,%s,%s,%s,%s", QemuDiskCache_None, QemuDiskCache_WriteThrough, QemuDiskCache_WriteBack, QemuDiskCache_Unsafe, QemuDiskCache_DirectSync)
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

func (format QemuDiskFormat) Validate() error {
	switch format {
	case QemuDiskFormat_Cow, QemuDiskFormat_Cloop, QemuDiskFormat_Qcow, QemuDiskFormat_Qcow2, QemuDiskFormat_Qed, QemuDiskFormat_Vmdk, QemuDiskFormat_Raw:
		return nil
	}
	return fmt.Errorf("format can only be one of the following values: %s,%s,%s,%s,%s,%s,%s", QemuDiskFormat_Cow, QemuDiskFormat_Cloop, QemuDiskFormat_Qcow, QemuDiskFormat_Qcow2, QemuDiskFormat_Qed, QemuDiskFormat_Vmdk, QemuDiskFormat_Raw)
}

type QemuDiskSerial string

// QemuDiskSerial may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_
// And has a max length of 60 characters
func (serial QemuDiskSerial) Validate() error {
	regex, _ := regexp.Compile(`^([a-z]|[A-Z]|[0-9]|_|-)*$`)
	if !regex.Match([]byte(serial)) {
		return errors.New("serial may only contain the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-_")
	}
	if len(serial) > 60 {
		return errors.New("serial may only be 60 characters long")
	}
	return nil
}

type qemuDiskShort struct {
	Storage string
	Id      string
}

type qemuDiskType int

const (
	ide    qemuDiskType = 0
	sata   qemuDiskType = 1
	scsi   qemuDiskType = 2
	virtIO qemuDiskType = 3
)

type QemuStorages struct {
	Ide    *QemuIdeDisks    `json:"ide,omitempty"`
	Sata   *QemuSataDisks   `json:"sata,omitempty"`
	Scsi   *QemuScsiDisks   `json:"scsi,omitempty"`
	VirtIO *QemuVirtIODisks `json:"virtio,omitempty"`
}

func (storages QemuStorages) markDiskChanges(currentStorages QemuStorages, params map[string]interface{}) *qemuUpdateChanges {
	changes := &qemuUpdateChanges{}
	if currentStorages.Ide != nil {
		storages.Ide.mapToApiValues(currentStorages.Ide, params, changes)
	}
	if currentStorages.Sata != nil {
		storages.Sata.mapToApiValues(currentStorages.Sata, params, changes)
	}
	if currentStorages.Scsi != nil {
		storages.Scsi.mapToApiValues(currentStorages.Scsi, params, changes)
	}
	if currentStorages.VirtIO != nil {
		storages.VirtIO.mapToApiValues(currentStorages.VirtIO, params, changes)
	}
	return changes
}

func (storages QemuStorages) mapToApiValues(params map[string]interface{}) {
	if storages.Ide != nil {
		storages.Ide.mapToApiValues(nil, params, nil)
	}
	if storages.Sata != nil {
		storages.Sata.mapToApiValues(nil, params, nil)
	}
	if storages.Scsi != nil {
		storages.Scsi.mapToApiValues(nil, params, nil)
	}
	if storages.VirtIO != nil {
		storages.VirtIO.mapToApiValues(nil, params, nil)
	}
}

func (QemuStorages) mapToStruct(params map[string]interface{}) *QemuStorages {
	storage := QemuStorages{
		Ide:    QemuIdeDisks{}.mapToStruct(params),
		Sata:   QemuSataDisks{}.mapToStruct(params),
		Scsi:   QemuScsiDisks{}.mapToStruct(params),
		VirtIO: QemuVirtIODisks{}.mapToStruct(params),
	}
	if storage.Ide != nil || storage.Sata != nil || storage.Scsi != nil || storage.VirtIO != nil {
		return &storage
	}
	return nil
}

type qemuUpdateChanges struct {
	Move                []qemuDiskShort
	Delete              string
	MigrationImpossible bool
}
