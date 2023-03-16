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
		return &qemuCdRom{CdRom: true}
	}
	if settings[0][0] == "cdrom" {
		return &qemuCdRom{CdRom: true, Passthrough: true}
	}
	tmpStorage := strings.Split(settings[0][0], ":")
	if len(tmpStorage) > 1 {
		tmpFile := strings.Split(settings[0][0], "/")
		if len(tmpFile) == 2 {
			tmpFileType := strings.Split(tmpFile[1], ".")
			if len(tmpFileType) > 1 {
				fileType := QemuDiskFormat(tmpFileType[len(tmpFileType)-1])
				if fileType == "iso" {
					for _, e := range settings {
						if e[0] == "size" {
							return &qemuCdRom{
								CdRom:   true,
								Storage: tmpStorage[0],
								File:    tmpFile[1],
								Size:    e[1],
							}
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

// TODO write test
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
	File      string         // Only set for Passthrough.
	Format    QemuDiskFormat // Only set for Disk
	Id        *uint          // Only set for Disk
	IOThread  bool           // Only set for scsi,virtio
	Number    uint
	ReadOnly  bool // Only set for scsi,virtio
	Replicate bool
	Serial    QemuDiskSerial
	Size      uint
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

// TODO write test
func (disk qemuDisk) mapToApiValues(vmID uint, create bool) (settings string) {
	if disk.Storage != "" {
		if create {
			settings = disk.Storage + ":" + strconv.Itoa(int(disk.Size))
		} else {
			// test:100/vm-100-disk-0.raw
			tmpId := strconv.Itoa(int(vmID))
			settings = disk.Storage + ":" + tmpId + "/vm-" + tmpId + "-disk-" + strconv.Itoa(int(*disk.Id)) + "." + string(disk.Format)
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

	if disk.Bandwidth.Iops.ReadLimit.Concurrent != 0 {
		settings = settings + ",iops_rd=" + strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.ReadLimit.Burst != 0 {
		settings = settings + ",iops_rd_max=" + strconv.Itoa(int(disk.Bandwidth.Iops.ReadLimit.Burst))
	}
	if disk.Bandwidth.Iops.WriteLimit.Concurrent != 0 {
		settings = settings + ",iops_wr=" + strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.WriteLimit.Burst != 0 {
		settings = settings + ",iops_wr_max=" + strconv.Itoa(int(disk.Bandwidth.Iops.WriteLimit.Burst))
	}

	if (disk.Type == scsi || disk.Type == virtIO) && disk.IOThread {
		settings = settings + ",iothread=1"
	}

	if disk.Bandwidth.Data.ReadLimit.Concurrent != 0 {
		settings = settings + fmt.Sprintf(",mbps_rd=%.2f", disk.Bandwidth.Data.ReadLimit.Concurrent)
	}
	if disk.Bandwidth.Data.ReadLimit.Burst != 0 {
		settings = settings + fmt.Sprintf(",mbps_rd_max=%.2f", disk.Bandwidth.Data.ReadLimit.Burst)
	}
	if disk.Bandwidth.Data.WriteLimit.Concurrent != 0 {
		settings = settings + fmt.Sprintf(",mbps_wr=%.2f", disk.Bandwidth.Data.WriteLimit.Concurrent)
	}
	if disk.Bandwidth.Data.WriteLimit.Burst != 0 {
		settings = settings + fmt.Sprintf(",mbps_wr_max=%.2f", disk.Bandwidth.Data.WriteLimit.Burst)
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
			idAndFormat := strings.Split(diskAndNumberAndFormat[1], ".")
			if len(idAndFormat) == 2 {
				disk.Format = QemuDiskFormat(idAndFormat[1])
				tmp := strings.Split(idAndFormat[0], "-")
				if len(tmp) > 1 {
					tmpId, _ := strconv.Atoi(tmp[len(tmp)-1])
					idPointer := uint(tmpId)
					disk.Id = &idPointer
				}
			}
		}
	}

	// Replicate defaults to true
	disk.Replicate = true

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
			if e[1] == "on" {
				disk.Discard = true
			}
			continue
		}
		if e[0] == "iops_rd" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.Iops.ReadLimit.Concurrent = uint(tmp)
		}
		if e[0] == "iops_rd_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.Iops.ReadLimit.Burst = uint(tmp)
		}
		if e[0] == "iops_wr" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.Iops.WriteLimit.Concurrent = uint(tmp)
		}
		if e[0] == "iops_wr_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.Iops.WriteLimit.Burst = uint(tmp)
		}
		if e[0] == "iothread" {
			disk.IOThread, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "mbps_rd" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.Data.ReadLimit.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_rd_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.Data.ReadLimit.Burst = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.Data.WriteLimit.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.Data.WriteLimit.Burst = float32(math.Round(tmp*100) / 100)
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
	Data QemuDiskBandwidthData `json:"data,omitempty"`
	Iops QemuDiskBandwidthIops `json:"iops,omitempty"`
}

func (bandwidth QemuDiskBandwidth) Validate() error {
	err := bandwidth.Data.Validate()
	if err != nil {
		return err
	}
	return bandwidth.Iops.Validate()
}

type QemuDiskBandwidthData struct {
	ReadLimit  QemuDiskBandwidthDataLimit `json:"read,omitempty"`
	WriteLimit QemuDiskBandwidthDataLimit `json:"write,omitempty"`
}

func (data QemuDiskBandwidthData) Validate() error {
	err := data.ReadLimit.Validate()
	if err != nil {
		return err
	}
	return data.WriteLimit.Validate()
}

type QemuDiskBandwidthDataLimit struct {
	Burst      float32 `json:"burst,omitempty"`      // 0 = default
	Concurrent float32 `json:"concurrent,omitempty"` // 0 = unlimited
}

const (
	Error_QemuDiskBandwidthDataLimit_Burst      string = "burst may not be lower then 1 except for 0"
	Error_QemuDiskBandwidthDataLimit_Concurrent string = "concurrent may not be lower then 1 except for 0"
)

func (limit QemuDiskBandwidthDataLimit) Validate() error {
	if limit.Burst != 0 && limit.Burst < 1 {
		return errors.New(Error_QemuDiskBandwidthDataLimit_Burst)
	}
	if limit.Concurrent != 0 && limit.Concurrent < 1 {
		return errors.New(Error_QemuDiskBandwidthDataLimit_Concurrent)
	}
	return nil
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
	Burst      uint `json:"burst,omitempty"`      // 0 = default
	Concurrent uint `json:"concurrent,omitempty"` // 0 = unlimited
}

const (
	Error_QemuDiskBandwidthIopsLimit_Burst      string = "burst may not be lower then 10 except for 0"
	Error_QemuDiskBandwidthIopsLimit_Concurrent string = "concurrent may not be lower then 10 except for 0"
)

func (limit QemuDiskBandwidthIopsLimit) Validate() error {
	if limit.Burst != 0 && limit.Burst < 10 {
		return errors.New(Error_QemuDiskBandwidthIopsLimit_Burst)
	}
	if limit.Concurrent != 0 && limit.Concurrent < 10 {
		return errors.New(Error_QemuDiskBandwidthIopsLimit_Concurrent)
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
	Id              string
	SizeInGigaBytes uint
}

// Increase the disk size to the specified amount in gigabytes
// Decrease of disk size is not permitted.
func (disk qemuDiskResize) resize(vmr *VmRef, client *Client) (exitStatus string, err error) {
	return client.PutWithTask(map[string]interface{}{"disk": disk.Id, "size": strconv.Itoa(int(disk.SizeInGigaBytes)) + "G"}, fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId))
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

type qemuStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *qemuDisk
	Passthrough *qemuDisk
}

func (storage *qemuStorage) markDiskChanges(currentStorage *qemuStorage, vmID uint, id string, params map[string]interface{}, changes *qemuUpdateChanges) {
	if storage == nil {
		return
	}
	// CDROM
	if storage.CdRom != nil {
		// Create or Update
		params[id] = storage.CdRom.mapToApiValues()
		return
	} else if currentStorage != nil && currentStorage.CdRom != nil && storage.CloudInit == nil && storage.Disk == nil && storage.Passthrough == nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// CloudInit
	if storage.CloudInit != nil {
		// Create or Update
		params[id] = storage.CloudInit.mapToApiValues()
		return
	} else if currentStorage != nil && currentStorage.CloudInit != nil && storage.Disk == nil && storage.Passthrough == nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// Disk
	if storage.Disk != nil {
		if currentStorage == nil || currentStorage.Disk == nil {
			// Create
			params[id] = storage.Disk.mapToApiValues(vmID, true)
			return
		} else {
			if storage.Disk.Size >= currentStorage.Disk.Size {
				// Update
				if storage.Disk.Size > currentStorage.Disk.Size {
					changes.Resize = append(changes.Resize, qemuDiskResize{
						Id:              id,
						SizeInGigaBytes: storage.Disk.Size,
					})
				}
				if storage.Disk.Id == nil {
					storage.Disk.Id = currentStorage.Disk.Id
				}
				if storage.Disk.Storage != currentStorage.Disk.Storage {
					changes.Move = append(changes.Move, qemuDiskShort{
						Id:      id,
						Storage: storage.Disk.Storage,
					})
				}
				params[id] = storage.Disk.mapToApiValues(vmID, false)
			} else {
				// Delete and Create
				// creating a disk on top of an existing disk is the same as detaching the disk and creating a new one.
				params[id] = storage.Disk.mapToApiValues(vmID, true)
			}
			return
		}
	} else if currentStorage != nil && currentStorage.Disk != nil && storage.Passthrough == nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// Passthrough
	if storage.Passthrough != nil {
		// Create or Update
		params[id] = storage.Passthrough.mapToApiValues(0, false)
		return
	} else if currentStorage != nil && currentStorage.Passthrough != nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// Delete if no subtype was specified
	if currentStorage != nil {
		changes.Delete = AddToList(changes.Delete, id)
	}
}

type QemuStorages struct {
	Ide    *QemuIdeDisks    `json:"ide,omitempty"`
	Sata   *QemuSataDisks   `json:"sata,omitempty"`
	Scsi   *QemuScsiDisks   `json:"scsi,omitempty"`
	VirtIO *QemuVirtIODisks `json:"virtio,omitempty"`
}

func (storages QemuStorages) mapToApiValues(vmID uint, params map[string]interface{}) {
	if storages.Ide != nil {
		storages.Ide.mapToApiValues(nil, vmID, params, nil)
	}
	if storages.Sata != nil {
		storages.Sata.mapToApiValues(nil, vmID, params, nil)
	}
	if storages.Scsi != nil {
		storages.Scsi.mapToApiValues(nil, vmID, params, nil)
	}
	if storages.VirtIO != nil {
		storages.VirtIO.mapToApiValues(nil, vmID, params, nil)
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

func (storages QemuStorages) markDiskChanges(currentStorages QemuStorages, vmID uint, params map[string]interface{}) *qemuUpdateChanges {
	changes := &qemuUpdateChanges{}
	if storages.Ide != nil {
		storages.Ide.mapToApiValues(currentStorages.Ide, vmID, params, changes)
	}
	if storages.Sata != nil {
		storages.Sata.mapToApiValues(currentStorages.Sata, vmID, params, changes)
	}
	if storages.Scsi != nil {
		storages.Scsi.mapToApiValues(currentStorages.Scsi, vmID, params, changes)
	}
	if storages.VirtIO != nil {
		storages.VirtIO.mapToApiValues(currentStorages.VirtIO, vmID, params, changes)
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
	Delete string
	Move   []qemuDiskShort
	Resize []qemuDiskResize
}

func diskSubtypeSet(set bool) error {
	if set {
		return errors.New(Error_QemuDisk_MutuallyExclusive)
	}
	return nil
}
