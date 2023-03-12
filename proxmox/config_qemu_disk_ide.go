package proxmox

type QemuIdeDisk struct {
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Bandwidth  QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard,omitempty"`
	EmulateSSD bool              `json:"emulatessd,omitempty"`
	Format     *QemuDiskFormat   `json:"format,omitempty"`
	Id         *uint             `json:"id,omitempty"`
	Replicate  bool              `json:"replicate,omitempty"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size,omitempty"`
	Storage    string            `json:"storage,omitempty"`
}

type QemuIdeDisks struct {
	Disk_0 *QemuIdeStorage `json:"0,omitempty"`
	Disk_1 *QemuIdeStorage `json:"1,omitempty"`
	Disk_2 *QemuIdeStorage `json:"2,omitempty"`
	Disk_3 *QemuIdeStorage `json:"3,omitempty"`
}

// TODO write test
func (disks QemuIdeDisks) mapToApiValues(currentDisks *QemuIdeDisks, vmID uint, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_0.convertDataStructure(), vmID, "ide0", params, changes)
	disks.Disk_1.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_1.convertDataStructure(), vmID, "ide1", params, changes)
	disks.Disk_2.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_2.convertDataStructure(), vmID, "ide2", params, changes)
	disks.Disk_3.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_3.convertDataStructure(), vmID, "ide3", params, changes)
}

// TODO write test
func (QemuIdeDisks) mapToStruct(params map[string]interface{}) *QemuIdeDisks {
	disks := QemuIdeDisks{}
	var structPopulated bool
	if _, isSet := params["ide0"]; isSet {
		disks.Disk_0 = QemuIdeStorage{}.mapToStruct(params["ide0"].(string))
		structPopulated = true
	}
	if _, isSet := params["ide1"]; isSet {
		disks.Disk_1 = QemuIdeStorage{}.mapToStruct(params["ide1"].(string))
		structPopulated = true
	}
	if _, isSet := params["ide2"]; isSet {
		disks.Disk_2 = QemuIdeStorage{}.mapToStruct(params["ide2"].(string))
		structPopulated = true
	}
	if _, isSet := params["ide3"]; isSet {
		disks.Disk_3 = QemuIdeStorage{}.mapToStruct(params["ide3"].(string))
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

type QemuIdePassthrough struct {
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

type QemuIdeStorage struct {
	CdRom       *QemuCdRom          `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk  `json:"cloudinit,omitempty"`
	Disk        *QemuIdeDisk        `json:"disk,omitempty"`
	Passthrough *QemuIdePassthrough `json:"passthrough,omitempty"`
}

// TODO write test
// converts to qemuStorage
func (storage *QemuIdeStorage) convertDataStructure() *qemuStorage {
	if storage == nil {
		return nil
	}
	generalizedStorage := qemuStorage{
		CdRom:     storage.CdRom,
		CloudInit: storage.CloudInit,
	}
	if storage.Disk != nil {
		generalizedStorage.Disk = &qemuDisk{
			AsyncIO:    storage.Disk.AsyncIO,
			Backup:     storage.Disk.Backup,
			Bandwidth:  storage.Disk.Bandwidth,
			Cache:      storage.Disk.Cache,
			Discard:    storage.Disk.Discard,
			EmulateSSD: storage.Disk.EmulateSSD,
			Format:     storage.Disk.Format,
			Id:         storage.Disk.Id,
			Replicate:  storage.Disk.Replicate,
			Serial:     storage.Disk.Serial,
			Size:       storage.Disk.Size,
			Storage:    storage.Disk.Storage,
			Type:       ide,
		}
	}
	if storage.Passthrough != nil {
		generalizedStorage.Passthrough = &qemuDisk{
			AsyncIO:    storage.Passthrough.AsyncIO,
			Backup:     storage.Passthrough.Backup,
			Bandwidth:  storage.Passthrough.Bandwidth,
			Cache:      storage.Passthrough.Cache,
			Discard:    storage.Passthrough.Discard,
			EmulateSSD: storage.Passthrough.EmulateSSD,
			File:       storage.Passthrough.File,
			Replicate:  storage.Passthrough.Replicate,
			Serial:     storage.Passthrough.Serial,
			Type:       ide,
		}
	}
	return &generalizedStorage
}

// TODO write test
func (QemuIdeStorage) mapToStruct(param string) *QemuIdeStorage {
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(settings)
	if tmpCdRom != nil {
		if tmpCdRom.FileType == "" {
			return &QemuIdeStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuIdeStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(settings)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuIdeStorage{Disk: &QemuIdeDisk{
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
	return &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
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
