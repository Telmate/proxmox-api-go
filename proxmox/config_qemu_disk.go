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

func (iso IsoFile) Validate() error {
	if iso.File == "" {
		return errors.New("file may not be empty")
	}
	if iso.Storage == "" {
		return errors.New("storage may not be empty")
	}
	return nil
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

func (cdRom QemuCdRom) Validate() error {
	if cdRom.Iso != nil {
		err := cdRom.Iso.Validate()
		if err != nil {
			return err
		}
		if cdRom.Passthrough {
			return errors.New("iso and passthrough are mutually exclusive")
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
		return errors.New("storage should not be empty")
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

	if disk.Bandwidth.Iops.ReadLimit.Concurrent != nil {
		settings = settings + ",iops_rd=" + strconv.Itoa(int(*disk.Bandwidth.Iops.ReadLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.ReadLimit.Burst != nil {
		settings = settings + ",iops_rd_max=" + strconv.Itoa(int(*disk.Bandwidth.Iops.ReadLimit.Burst))
	}
	if disk.Bandwidth.Iops.WriteLimit.Concurrent != nil {
		settings = settings + ",iops_wr=" + strconv.Itoa(int(*disk.Bandwidth.Iops.WriteLimit.Concurrent))
	}
	if disk.Bandwidth.Iops.WriteLimit.Burst != nil {
		settings = settings + ",iops_wr_max=" + strconv.Itoa(int(*disk.Bandwidth.Iops.WriteLimit.Burst))
	}

	if (disk.Type == scsi || disk.Type == virtIO) && disk.IOThread {
		settings = settings + ",iothread=1"
	}

	if disk.Bandwidth.Data.ReadLimit.Concurrent != nil {
		settings = settings + fmt.Sprintf(",mbps_rd=%.2f", *disk.Bandwidth.Data.ReadLimit.Concurrent)
	}
	if disk.Bandwidth.Data.ReadLimit.Burst != nil {
		settings = settings + fmt.Sprintf(",mbps_rd_max=%.2f", *disk.Bandwidth.Data.ReadLimit.Burst)
	}
	if disk.Bandwidth.Data.WriteLimit.Concurrent != nil {
		settings = settings + fmt.Sprintf(",mbps_wr=%.2f", *disk.Bandwidth.Data.WriteLimit.Concurrent)
	}
	if disk.Bandwidth.Data.WriteLimit.Burst != nil {
		settings = settings + fmt.Sprintf(",mbps_wr_max=%.2f", *disk.Bandwidth.Data.WriteLimit.Burst)
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
			pointer := uint(tmp)
			disk.Bandwidth.Iops.ReadLimit.Concurrent = &pointer
		}
		if e[0] == "iops_rd_max" {
			tmp, _ := strconv.Atoi(e[1])
			pointer := uint(tmp)
			disk.Bandwidth.Iops.ReadLimit.Burst = &pointer
		}
		if e[0] == "iops_wr" {
			tmp, _ := strconv.Atoi(e[1])
			pointer := uint(tmp)
			disk.Bandwidth.Iops.WriteLimit.Concurrent = &pointer
		}
		if e[0] == "iops_wr_max" {
			tmp, _ := strconv.Atoi(e[1])
			pointer := uint(tmp)
			disk.Bandwidth.Iops.WriteLimit.Burst = &pointer
		}
		if e[0] == "iothread" {
			disk.IOThread, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "mbps_rd" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			pointer := float32(math.Round(tmp*100) / 100)
			disk.Bandwidth.Data.ReadLimit.Concurrent = &pointer
		}
		if e[0] == "mbps_rd_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			pointer := float32(math.Round(tmp*100) / 100)
			disk.Bandwidth.Data.ReadLimit.Burst = &pointer
		}
		if e[0] == "mbps_wr" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			pointer := float32(math.Round(tmp*100) / 100)
			disk.Bandwidth.Data.WriteLimit.Concurrent = &pointer
		}
		if e[0] == "mbps_wr_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			pointer := float32(math.Round(tmp*100) / 100)
			disk.Bandwidth.Data.WriteLimit.Burst = &pointer
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
	Data QemuDiskBandwidthData
	Iops QemuDiskBandwidthIops
}

func (bandwidth QemuDiskBandwidth) Validate() error {
	err := bandwidth.Data.Validate()
	if err != nil {
		return err
	}
	return bandwidth.Iops.Validate()
}

type QemuDiskBandwidthData struct {
	ReadLimit  QemuDiskBandwidthDataLimit
	WriteLimit QemuDiskBandwidthDataLimit
}

func (data QemuDiskBandwidthData) Validate() error {
	err := data.ReadLimit.Validate()
	if err != nil {
		return err
	}
	return data.WriteLimit.Validate()
}

type QemuDiskBandwidthDataLimit struct {
	Burst      *float32 // nil = default
	Concurrent *float32 // nil = unlimited
}

func (limit QemuDiskBandwidthDataLimit) Validate() error {
	if limit.Burst != nil && *limit.Burst < 1 {
		return errors.New("burst may not be lower then 1")
	}
	if limit.Concurrent != nil && *limit.Concurrent < 1 {
		return errors.New("concurrent may not be lower then 1")
	}
	return nil
}

type QemuDiskBandwidthIops struct {
	ReadLimit  QemuDiskBandwidthIopsLimit
	WriteLimit QemuDiskBandwidthIopsLimit
}

func (iops QemuDiskBandwidthIops) Validate() error {
	err := iops.ReadLimit.Validate()
	if err != nil {
		return err
	}
	return iops.WriteLimit.Validate()
}

type QemuDiskBandwidthIopsLimit struct {
	Burst      *uint // nil = default
	Concurrent *uint // nil = unlimited
}

func (limit QemuDiskBandwidthIopsLimit) Validate() error {
	if limit.Burst != nil && *limit.Burst < 10 {
		return errors.New("burst may not be lower then 10")
	}
	if limit.Concurrent != nil && *limit.Concurrent < 10 {
		return errors.New("concurrent may not be lower then 1")
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
	CdRom       *QemuCdRom         `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk `json:"cloudinit,omitempty"`
	Disk        *qemuDisk          `json:"disk,omitempty"`
	Passthrough *qemuDisk          `json:"passthrough,omitempty"`
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

type qemuUpdateChanges struct {
	Delete string
	Move   []qemuDiskShort
	Resize []qemuDiskResize
}
