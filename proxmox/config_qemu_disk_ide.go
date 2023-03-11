package proxmox

type QemuIdeDisk struct {
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
func (disk QemuIdeDisk) mapToApiValues(create bool) string {
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
		Type:       ide,
	}.mapToApiValues(create)
}

type QemuIdeDisks struct {
	Disk_0 *QemuIdeStorage `json:"0,omitempty"`
	Disk_1 *QemuIdeStorage `json:"1,omitempty"`
	Disk_2 *QemuIdeStorage `json:"2,omitempty"`
	Disk_3 *QemuIdeStorage `json:"3,omitempty"`
}

// TODO write test
func (disks QemuIdeDisks) mapToApiValues(currentDisks *QemuIdeDisks, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.markDiskChanges(tmpCurrentDisks.Disk_0, "ide0", params, changes)
	disks.Disk_1.markDiskChanges(tmpCurrentDisks.Disk_1, "ide1", params, changes)
	disks.Disk_2.markDiskChanges(tmpCurrentDisks.Disk_2, "ide2", params, changes)
	disks.Disk_3.markDiskChanges(tmpCurrentDisks.Disk_3, "ide3", params, changes)
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

// TODO write test
func (passthrough QemuIdePassthrough) mapToApiValues() string {
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
		Type:       ide,
	}.mapToApiValues(false)
}

type QemuIdeStorage struct {
	CdRom       *QemuCdRom          `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk  `json:"cloudinit,omitempty"`
	Disk        *QemuIdeDisk        `json:"disk,omitempty"`
	Passthrough *QemuIdePassthrough `json:"passthrough,omitempty"`
}

// TODO write test
func (storage QemuIdeStorage) mapToApiValues(create bool) string {
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
func (storage *QemuIdeStorage) markDiskChanges(currentStorage *QemuIdeStorage, id string, params map[string]interface{}, changes *qemuUpdateChanges) {
	if storage == nil {
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
		params[id] = storage.Passthrough.mapToApiValues()
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
