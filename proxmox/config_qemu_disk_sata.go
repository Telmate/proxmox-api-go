package proxmox

type QemuSataDisk struct {
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Bandwidth  QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard,omitempty"`
	EmulateSSD bool              `json:"emulatessd,omitempty"`
	Replicate  bool              `json:"replicate,omitempty"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size,omitempty"`
	Storage    string            `json:"storage,omitempty"`
}

// TODO write test
func (disk QemuSataDisk) mapToApiValues(create bool) string {
	return qemuDisk{
		AsyncIO:    disk.AsyncIO,
		Backup:     disk.Backup,
		Bandwidth:  disk.Bandwidth,
		Cache:      disk.Cache,
		Discard:    disk.Discard,
		EmulateSSD: disk.EmulateSSD,
		Replicate:  disk.Replicate,
		Serial:     disk.Serial,
		Size:       disk.Size,
		Storage:    disk.Storage,
		Type:       sata,
	}.mapToApiValues(create)
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
func (disks QemuSataDisks) mapToApiValues(currentDisks *QemuSataDisks, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.markDiskChanges(tmpCurrentDisks.Disk_0, "sata0", params, changes)
	disks.Disk_1.markDiskChanges(tmpCurrentDisks.Disk_1, "sata1", params, changes)
	disks.Disk_2.markDiskChanges(tmpCurrentDisks.Disk_2, "sata2", params, changes)
	disks.Disk_3.markDiskChanges(tmpCurrentDisks.Disk_3, "sata3", params, changes)
	disks.Disk_4.markDiskChanges(tmpCurrentDisks.Disk_4, "sata4", params, changes)
	disks.Disk_5.markDiskChanges(tmpCurrentDisks.Disk_5, "sata5", params, changes)
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
	Size       uint
}

// TODO write test
func (passthrough QemuSataPassthrough) mapToApiValues() string {
	return qemuDisk{
		AsyncIO:    passthrough.AsyncIO,
		Backup:     passthrough.Backup,
		Bandwidth:  passthrough.Bandwidth,
		Cache:      passthrough.Cache,
		Discard:    passthrough.Discard,
		EmulateSSD: passthrough.EmulateSSD,
		File:       passthrough.File,
		Replicate:  passthrough.Replicate,
		Serial:     passthrough.Serial,
		Size:       passthrough.Size,
		Type:       sata,
	}.mapToApiValues(false)
}

type QemuSataStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuSataDisk
	Passthrough *QemuSataPassthrough
}

// TODO write test
func (storage QemuSataStorage) mapToApiValues(create bool) string {
	if storage.Disk != nil {
		return storage.Disk.mapToApiValues(create)
	}
	if storage.CdRom != nil {
		return storage.CdRom.mapToApiValues()
	}
	if storage.CloudInit != nil {
		return storage.CloudInit.mapToApiValues()
	}
	if storage.Passthrough != nil {
		return storage.Passthrough.mapToApiValues()
	}
	return ""
}

// TODO write test
func (storage *QemuSataStorage) markDiskChanges(currentStorage *QemuSataStorage, id string, params map[string]interface{}, changes *qemuUpdateChanges) {
	if storage == nil {
		if currentStorage != nil {
			changes.Delete = AddToList(changes.Delete, id)
		}
		return
	}
	// CDROM
	if storage.CdRom != nil {
		// Create or Update
		params[id] = storage.CdRom.mapToApiValues()
		return
	} else if currentStorage != nil && currentStorage.CdRom != nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// CloudInit
	if storage.CloudInit != nil {
		// Create or Update
		params[id] = storage.CloudInit.mapToApiValues()
		return
	} else if currentStorage != nil && currentStorage.CloudInit != nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// Disk
	if storage.Disk != nil {
		if currentStorage == nil || currentStorage.Disk == nil {
			// Create
			params[id] = storage.Disk.mapToApiValues(true)
		} else {
			if storage.Disk.Size >= currentStorage.Disk.Size {
				// Update
				if storage.Disk.Storage != currentStorage.Disk.Storage {
					changes.Move = append(changes.Move, qemuDiskShort{
						Id:      id,
						Storage: storage.Disk.Storage,
					})
				}
				params[id] = storage.Disk.mapToApiValues(false)
			} else {
				// Delete and Create
				changes.Delete = AddToList(changes.Delete, id)
				params[id] = storage.Disk.mapToApiValues(true)
			}
		}
		return
	} else if currentStorage != nil && currentStorage.Disk != nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
	// Passthrough
	if storage.Passthrough != nil {
		// Create or Update
		changes.MigrationImpossible = true
		params[id] = storage.Passthrough.mapToApiValues()
		return
	} else if currentStorage != nil && currentStorage.Passthrough != nil {
		// Delete
		changes.Delete = AddToList(changes.Delete, id)
		return
	}
}

// TODO write test
func (QemuSataStorage) mapToStruct(param string) *QemuSataStorage {
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(settings)
	if tmpCdRom != nil {
		if tmpCdRom.FileType == "" {
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
