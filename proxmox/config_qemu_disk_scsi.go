package proxmox

type QemuScsiDisk struct {
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Bandwidth  QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard,omitempty"`
	EmulateSSD bool              `json:"emulatessd,omitempty"`
	IOThread   bool              `json:"iothread,omitempty"`
	ReadOnly   bool              `json:"readonly,omitempty"`
	Replicate  bool              `json:"replicate,omitempty"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size,omitempty"`
	Storage    string            `json:"storage,omitempty"`
}

// TODO write test
func (disk QemuScsiDisk) mapToApiValues(create bool) string {
	return qemuDisk{
		AsyncIO:    disk.AsyncIO,
		Backup:     disk.Backup,
		Bandwidth:  disk.Bandwidth,
		Cache:      disk.Cache,
		Discard:    disk.Discard,
		EmulateSSD: disk.EmulateSSD,
		IOThread:   disk.IOThread,
		ReadOnly:   disk.ReadOnly,
		Replicate:  disk.Replicate,
		Serial:     disk.Serial,
		Size:       disk.Size,
		Storage:    disk.Storage,
		Type:       scsi,
	}.mapToApiValues(create)
}

type QemuScsiDisks struct {
	Disk_0  *QemuScsiStorage `json:"0,omitempty"`
	Disk_1  *QemuScsiStorage `json:"1,omitempty"`
	Disk_2  *QemuScsiStorage `json:"2,omitempty"`
	Disk_3  *QemuScsiStorage `json:"3,omitempty"`
	Disk_4  *QemuScsiStorage `json:"4,omitempty"`
	Disk_5  *QemuScsiStorage `json:"5,omitempty"`
	Disk_6  *QemuScsiStorage `json:"6,omitempty"`
	Disk_7  *QemuScsiStorage `json:"7,omitempty"`
	Disk_8  *QemuScsiStorage `json:"8,omitempty"`
	Disk_9  *QemuScsiStorage `json:"9,omitempty"`
	Disk_10 *QemuScsiStorage `json:"10,omitempty"`
	Disk_11 *QemuScsiStorage `json:"11,omitempty"`
	Disk_12 *QemuScsiStorage `json:"12,omitempty"`
	Disk_13 *QemuScsiStorage `json:"13,omitempty"`
	Disk_14 *QemuScsiStorage `json:"14,omitempty"`
	Disk_15 *QemuScsiStorage `json:"15,omitempty"`
	Disk_16 *QemuScsiStorage `json:"16,omitempty"`
	Disk_17 *QemuScsiStorage `json:"17,omitempty"`
	Disk_18 *QemuScsiStorage `json:"18,omitempty"`
	Disk_19 *QemuScsiStorage `json:"19,omitempty"`
	Disk_20 *QemuScsiStorage `json:"20,omitempty"`
	Disk_21 *QemuScsiStorage `json:"21,omitempty"`
	Disk_22 *QemuScsiStorage `json:"22,omitempty"`
	Disk_23 *QemuScsiStorage `json:"23,omitempty"`
	Disk_24 *QemuScsiStorage `json:"24,omitempty"`
	Disk_25 *QemuScsiStorage `json:"25,omitempty"`
	Disk_26 *QemuScsiStorage `json:"26,omitempty"`
	Disk_27 *QemuScsiStorage `json:"27,omitempty"`
	Disk_28 *QemuScsiStorage `json:"28,omitempty"`
	Disk_29 *QemuScsiStorage `json:"29,omitempty"`
	Disk_30 *QemuScsiStorage `json:"30,omitempty"`
}

// TODO write test
func (disks QemuScsiDisks) mapToApiValues(currentDisks *QemuScsiDisks, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuScsiDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.markDiskChanges(tmpCurrentDisks.Disk_0, "scsi0", params, changes)
	disks.Disk_1.markDiskChanges(tmpCurrentDisks.Disk_1, "scsi1", params, changes)
	disks.Disk_2.markDiskChanges(tmpCurrentDisks.Disk_2, "scsi2", params, changes)
	disks.Disk_3.markDiskChanges(tmpCurrentDisks.Disk_3, "scsi3", params, changes)
	disks.Disk_4.markDiskChanges(tmpCurrentDisks.Disk_4, "scsi4", params, changes)
	disks.Disk_5.markDiskChanges(tmpCurrentDisks.Disk_5, "scsi5", params, changes)
	disks.Disk_6.markDiskChanges(tmpCurrentDisks.Disk_6, "scsi6", params, changes)
	disks.Disk_7.markDiskChanges(tmpCurrentDisks.Disk_7, "scsi7", params, changes)
	disks.Disk_8.markDiskChanges(tmpCurrentDisks.Disk_8, "scsi8", params, changes)
	disks.Disk_9.markDiskChanges(tmpCurrentDisks.Disk_9, "scsi9", params, changes)
	disks.Disk_10.markDiskChanges(tmpCurrentDisks.Disk_10, "scsi10", params, changes)
	disks.Disk_11.markDiskChanges(tmpCurrentDisks.Disk_11, "scsi11", params, changes)
	disks.Disk_12.markDiskChanges(tmpCurrentDisks.Disk_12, "scsi12", params, changes)
	disks.Disk_13.markDiskChanges(tmpCurrentDisks.Disk_13, "scsi13", params, changes)
	disks.Disk_14.markDiskChanges(tmpCurrentDisks.Disk_14, "scsi14", params, changes)
	disks.Disk_15.markDiskChanges(tmpCurrentDisks.Disk_15, "scsi15", params, changes)
	disks.Disk_16.markDiskChanges(tmpCurrentDisks.Disk_16, "scsi16", params, changes)
	disks.Disk_17.markDiskChanges(tmpCurrentDisks.Disk_17, "scsi17", params, changes)
	disks.Disk_18.markDiskChanges(tmpCurrentDisks.Disk_18, "scsi18", params, changes)
	disks.Disk_19.markDiskChanges(tmpCurrentDisks.Disk_19, "scsi19", params, changes)
	disks.Disk_20.markDiskChanges(tmpCurrentDisks.Disk_20, "scsi20", params, changes)
	disks.Disk_21.markDiskChanges(tmpCurrentDisks.Disk_21, "scsi21", params, changes)
	disks.Disk_22.markDiskChanges(tmpCurrentDisks.Disk_22, "scsi22", params, changes)
	disks.Disk_23.markDiskChanges(tmpCurrentDisks.Disk_23, "scsi23", params, changes)
	disks.Disk_24.markDiskChanges(tmpCurrentDisks.Disk_24, "scsi24", params, changes)
	disks.Disk_25.markDiskChanges(tmpCurrentDisks.Disk_25, "scsi25", params, changes)
	disks.Disk_26.markDiskChanges(tmpCurrentDisks.Disk_26, "scsi26", params, changes)
	disks.Disk_27.markDiskChanges(tmpCurrentDisks.Disk_27, "scsi27", params, changes)
	disks.Disk_28.markDiskChanges(tmpCurrentDisks.Disk_28, "scsi28", params, changes)
	disks.Disk_29.markDiskChanges(tmpCurrentDisks.Disk_29, "scsi29", params, changes)
	disks.Disk_30.markDiskChanges(tmpCurrentDisks.Disk_30, "scsi30", params, changes)
}

// TODO write test
func (QemuScsiDisks) mapToStruct(params map[string]interface{}) *QemuScsiDisks {
	disks := QemuScsiDisks{}
	var structPopulated bool
	if _, isSet := params["scsi0"]; isSet {
		disks.Disk_0 = QemuScsiStorage{}.mapToStruct(params["scsi0"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi1"]; isSet {
		disks.Disk_1 = QemuScsiStorage{}.mapToStruct(params["scsi1"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi2"]; isSet {
		disks.Disk_2 = QemuScsiStorage{}.mapToStruct(params["scsi2"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi3"]; isSet {
		disks.Disk_3 = QemuScsiStorage{}.mapToStruct(params["scsi3"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi4"]; isSet {
		disks.Disk_4 = QemuScsiStorage{}.mapToStruct(params["scsi4"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi5"]; isSet {
		disks.Disk_5 = QemuScsiStorage{}.mapToStruct(params["scsi5"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi6"]; isSet {
		disks.Disk_6 = QemuScsiStorage{}.mapToStruct(params["scsi6"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi7"]; isSet {
		disks.Disk_7 = QemuScsiStorage{}.mapToStruct(params["scsi7"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi8"]; isSet {
		disks.Disk_8 = QemuScsiStorage{}.mapToStruct(params["scsi8"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi9"]; isSet {
		disks.Disk_9 = QemuScsiStorage{}.mapToStruct(params["scsi9"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi10"]; isSet {
		disks.Disk_10 = QemuScsiStorage{}.mapToStruct(params["scsi10"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi11"]; isSet {
		disks.Disk_11 = QemuScsiStorage{}.mapToStruct(params["scsi11"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi12"]; isSet {
		disks.Disk_12 = QemuScsiStorage{}.mapToStruct(params["scsi12"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi13"]; isSet {
		disks.Disk_13 = QemuScsiStorage{}.mapToStruct(params["scsi13"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi14"]; isSet {
		disks.Disk_14 = QemuScsiStorage{}.mapToStruct(params["scsi14"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi15"]; isSet {
		disks.Disk_15 = QemuScsiStorage{}.mapToStruct(params["scsi15"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi16"]; isSet {
		disks.Disk_16 = QemuScsiStorage{}.mapToStruct(params["scsi16"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi17"]; isSet {
		disks.Disk_17 = QemuScsiStorage{}.mapToStruct(params["scsi17"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi18"]; isSet {
		disks.Disk_18 = QemuScsiStorage{}.mapToStruct(params["scsi18"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi19"]; isSet {
		disks.Disk_19 = QemuScsiStorage{}.mapToStruct(params["scsi19"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi20"]; isSet {
		disks.Disk_20 = QemuScsiStorage{}.mapToStruct(params["scsi20"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi21"]; isSet {
		disks.Disk_21 = QemuScsiStorage{}.mapToStruct(params["scsi21"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi22"]; isSet {
		disks.Disk_22 = QemuScsiStorage{}.mapToStruct(params["scsi22"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi23"]; isSet {
		disks.Disk_23 = QemuScsiStorage{}.mapToStruct(params["scsi23"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi24"]; isSet {
		disks.Disk_24 = QemuScsiStorage{}.mapToStruct(params["scsi24"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi25"]; isSet {
		disks.Disk_25 = QemuScsiStorage{}.mapToStruct(params["scsi25"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi26"]; isSet {
		disks.Disk_26 = QemuScsiStorage{}.mapToStruct(params["scsi26"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi27"]; isSet {
		disks.Disk_27 = QemuScsiStorage{}.mapToStruct(params["scsi27"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi28"]; isSet {
		disks.Disk_28 = QemuScsiStorage{}.mapToStruct(params["scsi28"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi29"]; isSet {
		disks.Disk_29 = QemuScsiStorage{}.mapToStruct(params["scsi29"].(string))
		structPopulated = true
	}
	if _, isSet := params["scsi30"]; isSet {
		disks.Disk_30 = QemuScsiStorage{}.mapToStruct(params["scsi30"].(string))
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

type QemuScsiPassthrough struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool
	File       string
	IOThread   bool
	ReadOnly   bool
	Replicate  bool
	Serial     QemuDiskSerial `json:"serial,omitempty"`
	Size       uint
}

// TODO write function
// TODO write test
func (passthrough QemuScsiPassthrough) mapToApiValues() string {
	return ""
}

type QemuScsiStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuScsiDisk
	Passthrough *QemuScsiPassthrough
}

// TODO write test
func (storage QemuScsiStorage) mapToApiValues(create bool) string {
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
func (storage *QemuScsiStorage) markDiskChanges(currentStorage *QemuScsiStorage, id string, params map[string]interface{}, changes *qemuUpdateChanges) {
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
func (QemuScsiStorage) mapToStruct(param string) *QemuScsiStorage {
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(settings)
	if tmpCdRom != nil {
		if tmpCdRom.FileType == "" {
			return &QemuScsiStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuScsiStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(settings)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuScsiStorage{Disk: &QemuScsiDisk{
			AsyncIO:    tmpDisk.AsyncIO,
			Backup:     tmpDisk.Backup,
			Bandwidth:  tmpDisk.Bandwidth,
			Cache:      tmpDisk.Cache,
			Discard:    tmpDisk.Discard,
			EmulateSSD: tmpDisk.EmulateSSD,
			IOThread:   tmpDisk.IOThread,
			ReadOnly:   tmpDisk.ReadOnly,
			Replicate:  tmpDisk.Replicate,
			Serial:     tmpDisk.Serial,
			Size:       tmpDisk.Size,
			Storage:    tmpDisk.Storage,
		}}
	}
	return &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
		AsyncIO:    tmpDisk.AsyncIO,
		Backup:     tmpDisk.Backup,
		Bandwidth:  tmpDisk.Bandwidth,
		Cache:      tmpDisk.Cache,
		Discard:    tmpDisk.Discard,
		EmulateSSD: tmpDisk.EmulateSSD,
		File:       tmpDisk.File,
		IOThread:   tmpDisk.IOThread,
		ReadOnly:   tmpDisk.ReadOnly,
		Replicate:  tmpDisk.Replicate,
		Serial:     tmpDisk.Serial,
		Size:       tmpDisk.Size,
	}}

}