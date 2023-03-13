package proxmox

type QemuScsiDisk struct {
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup,omitempty"`
	Bandwidth  QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard,omitempty"`
	EmulateSSD bool              `json:"emulatessd,omitempty"`
	Format     *QemuDiskFormat   `json:"format,omitempty"`
	Id         *uint             `json:"id,omitempty"`
	IOThread   bool              `json:"iothread,omitempty"`
	ReadOnly   bool              `json:"readonly,omitempty"`
	Replicate  bool              `json:"replicate,omitempty"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size,omitempty"`
	Storage    string            `json:"storage,omitempty"`
}

func (disk *QemuScsiDisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:    disk.AsyncIO,
		Backup:     disk.Backup,
		Bandwidth:  disk.Bandwidth,
		Cache:      disk.Cache,
		Discard:    disk.Discard,
		EmulateSSD: disk.EmulateSSD,
		Format:     disk.Format,
		Id:         disk.Id,
		IOThread:   disk.IOThread,
		ReadOnly:   disk.ReadOnly,
		Replicate:  disk.Replicate,
		Serial:     disk.Serial,
		Size:       disk.Size,
		Storage:    disk.Storage,
		Type:       scsi,
	}
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
func (disks QemuScsiDisks) mapToApiValues(currentDisks *QemuScsiDisks, vmID uint, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuScsiDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_0.convertDataStructure(), vmID, "scsi0", params, changes)
	disks.Disk_1.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_1.convertDataStructure(), vmID, "scsi1", params, changes)
	disks.Disk_2.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_2.convertDataStructure(), vmID, "scsi2", params, changes)
	disks.Disk_3.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_3.convertDataStructure(), vmID, "scsi3", params, changes)
	disks.Disk_4.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_4.convertDataStructure(), vmID, "scsi4", params, changes)
	disks.Disk_5.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_5.convertDataStructure(), vmID, "scsi5", params, changes)
	disks.Disk_6.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_6.convertDataStructure(), vmID, "scsi6", params, changes)
	disks.Disk_7.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_7.convertDataStructure(), vmID, "scsi7", params, changes)
	disks.Disk_8.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_8.convertDataStructure(), vmID, "scsi8", params, changes)
	disks.Disk_9.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_9.convertDataStructure(), vmID, "scsi9", params, changes)
	disks.Disk_10.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_10.convertDataStructure(), vmID, "scsi10", params, changes)
	disks.Disk_11.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_11.convertDataStructure(), vmID, "scsi11", params, changes)
	disks.Disk_12.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_12.convertDataStructure(), vmID, "scsi12", params, changes)
	disks.Disk_13.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_13.convertDataStructure(), vmID, "scsi13", params, changes)
	disks.Disk_14.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_14.convertDataStructure(), vmID, "scsi14", params, changes)
	disks.Disk_15.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_15.convertDataStructure(), vmID, "scsi15", params, changes)
	disks.Disk_16.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_16.convertDataStructure(), vmID, "scsi16", params, changes)
	disks.Disk_17.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_17.convertDataStructure(), vmID, "scsi17", params, changes)
	disks.Disk_18.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_18.convertDataStructure(), vmID, "scsi18", params, changes)
	disks.Disk_19.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_19.convertDataStructure(), vmID, "scsi19", params, changes)
	disks.Disk_20.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_20.convertDataStructure(), vmID, "scsi20", params, changes)
	disks.Disk_21.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_21.convertDataStructure(), vmID, "scsi21", params, changes)
	disks.Disk_22.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_22.convertDataStructure(), vmID, "scsi22", params, changes)
	disks.Disk_23.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_23.convertDataStructure(), vmID, "scsi23", params, changes)
	disks.Disk_24.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_24.convertDataStructure(), vmID, "scsi24", params, changes)
	disks.Disk_25.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_25.convertDataStructure(), vmID, "scsi25", params, changes)
	disks.Disk_26.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_26.convertDataStructure(), vmID, "scsi26", params, changes)
	disks.Disk_27.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_27.convertDataStructure(), vmID, "scsi27", params, changes)
	disks.Disk_28.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_28.convertDataStructure(), vmID, "scsi28", params, changes)
	disks.Disk_29.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_29.convertDataStructure(), vmID, "scsi29", params, changes)
	disks.Disk_30.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_30.convertDataStructure(), vmID, "scsi30", params, changes)
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
	Size       uint           //size is only returned and setting it has no effect
}

func (passthrough *QemuScsiPassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:    passthrough.AsyncIO,
		Backup:     passthrough.Backup,
		Bandwidth:  passthrough.Bandwidth,
		Cache:      passthrough.Cache,
		Discard:    passthrough.Discard,
		EmulateSSD: passthrough.EmulateSSD,
		File:       passthrough.File,
		IOThread:   passthrough.IOThread,
		ReadOnly:   passthrough.ReadOnly,
		Replicate:  passthrough.Replicate,
		Serial:     passthrough.Serial,
		Type:       scsi,
	}
}

type QemuScsiStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuScsiDisk
	Passthrough *QemuScsiPassthrough
}

// TODO write test
// converts to qemuStorage
func (storage *QemuScsiStorage) convertDataStructure() *qemuStorage {
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
			Format:     tmpDisk.Format,
			Id:         tmpDisk.Id,
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
