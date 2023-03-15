package proxmox

import "strconv"

type QemuSataDisk struct {
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Bandwidth  QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard,omitempty"`
	EmulateSSD bool              `json:"emulatessd,omitempty"`
	Format     QemuDiskFormat    `json:"format,omitempty"`
	Id         *uint             `json:"id,omitempty"`
	Replicate  bool              `json:"replicate,omitempty"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size,omitempty"`
	Storage    string            `json:"storage,omitempty"`
}

func (disk *QemuSataDisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:    disk.AsyncIO,
		Backup:     disk.Backup,
		Bandwidth:  disk.Bandwidth,
		Cache:      disk.Cache,
		Discard:    disk.Discard,
		Disk:       true,
		EmulateSSD: disk.EmulateSSD,
		Format:     disk.Format,
		Id:         disk.Id,
		Replicate:  disk.Replicate,
		Serial:     disk.Serial,
		Size:       disk.Size,
		Storage:    disk.Storage,
		Type:       sata,
	}
}

type QemuSataDisks struct {
	Disk_0 *QemuSataStorage `json:"0,omitempty"`
	Disk_1 *QemuSataStorage `json:"1,omitempty"`
	Disk_2 *QemuSataStorage `json:"2,omitempty"`
	Disk_3 *QemuSataStorage `json:"3,omitempty"`
	Disk_4 *QemuSataStorage `json:"4,omitempty"`
	Disk_5 *QemuSataStorage `json:"5,omitempty"`
}

// TODO write test
func (disks QemuSataDisks) mapToApiValues(currentDisks *QemuSataDisks, vmID uint, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		diskMap[i].convertDataStructure().markDiskChanges(currentDiskMap[i].convertDataStructure(), vmID, "sata"+strconv.Itoa(int(i)), params, changes)
	}
}

func (disks QemuSataDisks) mapToIntMap() map[uint8]*QemuSataStorage {
	return map[uint8]*QemuSataStorage{
		0: disks.Disk_0,
		1: disks.Disk_1,
		2: disks.Disk_2,
		3: disks.Disk_3,
		4: disks.Disk_4,
		5: disks.Disk_5,
	}
}

// TODO write test
func (QemuSataDisks) mapToStruct(params map[string]interface{}) *QemuSataDisks {
	disks := QemuSataDisks{}
	var structPopulated bool
	if _, isSet := params["sata0"]; isSet {
		disks.Disk_0 = QemuSataStorage{}.mapToStruct(params["sata0"].(string))
		structPopulated = true
	}
	if _, isSet := params["sata1"]; isSet {
		disks.Disk_1 = QemuSataStorage{}.mapToStruct(params["sata1"].(string))
		structPopulated = true
	}
	if _, isSet := params["sata2"]; isSet {
		disks.Disk_2 = QemuSataStorage{}.mapToStruct(params["sata2"].(string))
		structPopulated = true
	}
	if _, isSet := params["sata3"]; isSet {
		disks.Disk_3 = QemuSataStorage{}.mapToStruct(params["sata3"].(string))
		structPopulated = true
	}
	if _, isSet := params["sata4"]; isSet {
		disks.Disk_4 = QemuSataStorage{}.mapToStruct(params["sata4"].(string))
		structPopulated = true
	}
	if _, isSet := params["sata5"]; isSet {
		disks.Disk_5 = QemuSataStorage{}.mapToStruct(params["sata5"].(string))
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

type QemuSataPassthrough struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool
	File       string
	Replicate  bool
	Serial     QemuDiskSerial `json:"serial,omitempty"`
	Size       uint           //size is only returned and setting it has no effect
}

func (passthrough *QemuSataPassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:    passthrough.AsyncIO,
		Backup:     passthrough.Backup,
		Bandwidth:  passthrough.Bandwidth,
		Cache:      passthrough.Cache,
		Discard:    passthrough.Discard,
		EmulateSSD: passthrough.EmulateSSD,
		File:       passthrough.File,
		Replicate:  passthrough.Replicate,
		Serial:     passthrough.Serial,
		Type:       sata,
	}
}

type QemuSataStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuSataDisk
	Passthrough *QemuSataPassthrough
}

// TODO write test
// converts to qemuStorage
func (storage *QemuSataStorage) convertDataStructure() *qemuStorage {
	if storage == nil {
		return nil
	}
	generalizedStorage := qemuStorage{
		CdRom:     storage.CdRom,
		CloudInit: storage.CloudInit,
	}
	if storage.Disk != nil {
		generalizedStorage.Disk = storage.Disk.convertDataStructure()
	}
	if storage.Passthrough != nil {
		generalizedStorage.Passthrough = storage.Passthrough.convertDataStructure()
	}
	return &generalizedStorage
}

// TODO write test
func (QemuSataStorage) mapToStruct(param string) *QemuSataStorage {
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(settings)
	if tmpCdRom != nil {
		if tmpCdRom.CdRom {
			return &QemuSataStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuSataStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(settings)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuSataStorage{Disk: &QemuSataDisk{
			AsyncIO:    tmpDisk.AsyncIO,
			Backup:     tmpDisk.Backup,
			Bandwidth:  tmpDisk.Bandwidth,
			Cache:      tmpDisk.Cache,
			Discard:    tmpDisk.Discard,
			EmulateSSD: tmpDisk.EmulateSSD,
			Format:     tmpDisk.Format,
			Id:         tmpDisk.Id,
			Replicate:  tmpDisk.Replicate,
			Serial:     tmpDisk.Serial,
			Size:       tmpDisk.Size,
			Storage:    tmpDisk.Storage,
		}}
	}
	return &QemuSataStorage{Passthrough: &QemuSataPassthrough{
		AsyncIO:    tmpDisk.AsyncIO,
		Backup:     tmpDisk.Backup,
		Bandwidth:  tmpDisk.Bandwidth,
		Cache:      tmpDisk.Cache,
		Discard:    tmpDisk.Discard,
		EmulateSSD: tmpDisk.EmulateSSD,
		File:       tmpDisk.File,
		Replicate:  tmpDisk.Replicate,
		Serial:     tmpDisk.Serial,
		Size:       tmpDisk.Size,
	}}
}
