package proxmox

type QemuVirtIODisk struct {
	AsyncIO   QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup    bool              `json:"backup,omitempty"`
	Bandwidth QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache     QemuDiskCache     `json:"cache,omitempty"`
	Discard   bool              `json:"discard,omitempty"`
	Id        *uint             `json:"id,omitempty"`
	IOThread  bool              `json:"iothread,omitempty"`
	ReadOnly  bool              `json:"readonly,omitempty"`
	Replicate bool              `json:"replicate,omitempty"`
	Serial    QemuDiskSerial    `json:"serial,omitempty"`
	Size      uint              `json:"size,omitempty"`
	Storage   string            `json:"storage,omitempty"`
}

// TODO write test
func (disk QemuVirtIODisk) mapToApiValues(vmID uint, create bool) string {
	return qemuDisk{
		AsyncIO:   disk.AsyncIO,
		Backup:    disk.Backup,
		Bandwidth: disk.Bandwidth,
		Cache:     disk.Cache,
		Discard:   disk.Discard,
		IOThread:  disk.IOThread,
		ReadOnly:  disk.ReadOnly,
		Serial:    disk.Serial,
		Size:      disk.Size,
		Storage:   disk.Storage,
		Type:      virtIO,
	}.mapToApiValues(vmID, create)
}

type QemuVirtIODisks struct {
	Disk_0  *QemuVirtIOStorage `json:"0,omitempty"`
	Disk_1  *QemuVirtIOStorage `json:"1,omitempty"`
	Disk_2  *QemuVirtIOStorage `json:"2,omitempty"`
	Disk_3  *QemuVirtIOStorage `json:"3,omitempty"`
	Disk_4  *QemuVirtIOStorage `json:"4,omitempty"`
	Disk_5  *QemuVirtIOStorage `json:"5,omitempty"`
	Disk_6  *QemuVirtIOStorage `json:"6,omitempty"`
	Disk_7  *QemuVirtIOStorage `json:"7,omitempty"`
	Disk_8  *QemuVirtIOStorage `json:"8,omitempty"`
	Disk_9  *QemuVirtIOStorage `json:"9,omitempty"`
	Disk_10 *QemuVirtIOStorage `json:"10,omitempty"`
	Disk_11 *QemuVirtIOStorage `json:"11,omitempty"`
	Disk_12 *QemuVirtIOStorage `json:"12,omitempty"`
	Disk_13 *QemuVirtIOStorage `json:"13,omitempty"`
	Disk_14 *QemuVirtIOStorage `json:"14,omitempty"`
	Disk_15 *QemuVirtIOStorage `json:"15,omitempty"`
}

// TODO write test
func (disks QemuVirtIODisks) mapToApiValues(currentDisks *QemuVirtIODisks, vmID uint, params map[string]interface{}, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuVirtIODisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	disks.Disk_0.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_0.convertDataStructure(), vmID, "virtio0", params, changes)
	disks.Disk_1.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_1.convertDataStructure(), vmID, "virtio1", params, changes)
	disks.Disk_2.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_2.convertDataStructure(), vmID, "virtio2", params, changes)
	disks.Disk_3.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_3.convertDataStructure(), vmID, "virtio3", params, changes)
	disks.Disk_4.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_4.convertDataStructure(), vmID, "virtio4", params, changes)
	disks.Disk_5.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_5.convertDataStructure(), vmID, "virtio5", params, changes)
	disks.Disk_6.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_6.convertDataStructure(), vmID, "virtio6", params, changes)
	disks.Disk_7.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_7.convertDataStructure(), vmID, "virtio7", params, changes)
	disks.Disk_8.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_8.convertDataStructure(), vmID, "virtio8", params, changes)
	disks.Disk_9.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_9.convertDataStructure(), vmID, "virtio9", params, changes)
	disks.Disk_10.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_10.convertDataStructure(), vmID, "virtio10", params, changes)
	disks.Disk_11.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_11.convertDataStructure(), vmID, "virtio11", params, changes)
	disks.Disk_12.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_12.convertDataStructure(), vmID, "virtio12", params, changes)
	disks.Disk_13.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_13.convertDataStructure(), vmID, "virtio13", params, changes)
	disks.Disk_14.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_14.convertDataStructure(), vmID, "virtio14", params, changes)
	disks.Disk_15.convertDataStructure().markDiskChanges(tmpCurrentDisks.Disk_15.convertDataStructure(), vmID, "virtio15", params, changes)
}

// TODO write test
func (QemuVirtIODisks) mapToStruct(params map[string]interface{}) *QemuVirtIODisks {
	disks := QemuVirtIODisks{}
	var structPopulated bool
	if _, isSet := params["virtio0"]; isSet {
		disks.Disk_0 = QemuVirtIOStorage{}.mapToStruct(params["virtio0"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio1"]; isSet {
		disks.Disk_1 = QemuVirtIOStorage{}.mapToStruct(params["virtio1"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio2"]; isSet {
		disks.Disk_2 = QemuVirtIOStorage{}.mapToStruct(params["virtio2"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio3"]; isSet {
		disks.Disk_3 = QemuVirtIOStorage{}.mapToStruct(params["virtio3"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio4"]; isSet {
		disks.Disk_4 = QemuVirtIOStorage{}.mapToStruct(params["virtio4"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio5"]; isSet {
		disks.Disk_5 = QemuVirtIOStorage{}.mapToStruct(params["virtio5"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio6"]; isSet {
		disks.Disk_6 = QemuVirtIOStorage{}.mapToStruct(params["virtio6"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio7"]; isSet {
		disks.Disk_7 = QemuVirtIOStorage{}.mapToStruct(params["virtio7"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio8"]; isSet {
		disks.Disk_8 = QemuVirtIOStorage{}.mapToStruct(params["virtio8"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio9"]; isSet {
		disks.Disk_9 = QemuVirtIOStorage{}.mapToStruct(params["virtio9"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio10"]; isSet {
		disks.Disk_10 = QemuVirtIOStorage{}.mapToStruct(params["virtio10"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio11"]; isSet {
		disks.Disk_11 = QemuVirtIOStorage{}.mapToStruct(params["virtio11"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio12"]; isSet {
		disks.Disk_12 = QemuVirtIOStorage{}.mapToStruct(params["virtio12"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio13"]; isSet {
		disks.Disk_13 = QemuVirtIOStorage{}.mapToStruct(params["virtio13"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio14"]; isSet {
		disks.Disk_14 = QemuVirtIOStorage{}.mapToStruct(params["virtio14"].(string))
		structPopulated = true
	}
	if _, isSet := params["virtio15"]; isSet {
		disks.Disk_15 = QemuVirtIOStorage{}.mapToStruct(params["virtio15"].(string))
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

type QemuVirtIOPassthrough struct {
	AsyncIO   QemuDiskAsyncIO
	Backup    bool
	Bandwidth QemuDiskBandwidth
	Cache     QemuDiskCache
	Discard   bool
	File      string
	IOThread  bool
	ReadOnly  bool
	Serial    QemuDiskSerial `json:"serial,omitempty"`
	Size      uint           //size is only returned and setting it has no effect
}

// TODO write test
func (passthrough QemuVirtIOPassthrough) mapToApiValues() string {
	return qemuDisk{
		AsyncIO:   passthrough.AsyncIO,
		Backup:    passthrough.Backup,
		Bandwidth: passthrough.Bandwidth,
		Cache:     passthrough.Cache,
		Discard:   passthrough.Discard,
		File:      passthrough.File,
		IOThread:  passthrough.IOThread,
		ReadOnly:  passthrough.ReadOnly,
		Serial:    passthrough.Serial,
		Type:      scsi,
	}.mapToApiValues(0, false)
}

type QemuVirtIOStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuVirtIODisk
	Passthrough *QemuVirtIOPassthrough
}

// TODO write test
// converts to qemuStorage
func (storage *QemuVirtIOStorage) convertDataStructure() *qemuStorage {
	if storage == nil {
		return nil
	}
	generalizedStorage := qemuStorage{
		CdRom:     storage.CdRom,
		CloudInit: storage.CloudInit,
	}
	if storage.Disk != nil {
		generalizedStorage.Disk = &qemuDisk{
			AsyncIO:   storage.Disk.AsyncIO,
			Backup:    storage.Disk.Backup,
			Bandwidth: storage.Disk.Bandwidth,
			Cache:     storage.Disk.Cache,
			Discard:   storage.Disk.Discard,
			Id:        storage.Disk.Id,
			IOThread:  storage.Disk.IOThread,
			ReadOnly:  storage.Disk.ReadOnly,
			Replicate: storage.Disk.Replicate,
			Serial:    storage.Disk.Serial,
			Size:      storage.Disk.Size,
			Storage:   storage.Disk.Storage,
		}
	}
	if storage.Passthrough != nil {
		generalizedStorage.Passthrough = &qemuDisk{
			AsyncIO:   storage.Passthrough.AsyncIO,
			Backup:    storage.Passthrough.Backup,
			Bandwidth: storage.Passthrough.Bandwidth,
			Cache:     storage.Passthrough.Cache,
			Discard:   storage.Passthrough.Discard,
			File:      storage.Passthrough.File,
			IOThread:  storage.Passthrough.IOThread,
			ReadOnly:  storage.Passthrough.ReadOnly,
			Serial:    storage.Passthrough.Serial,
		}
	}
	return &generalizedStorage
}

// TODO write test
func (storage QemuVirtIOStorage) mapToApiValues(vmID uint, create bool) string {
	if storage.Disk != nil {
		return storage.Disk.mapToApiValues(vmID, create)
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
func (QemuVirtIOStorage) mapToStruct(param string) *QemuVirtIOStorage {
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(settings)
	if tmpCdRom != nil {
		if tmpCdRom.FileType == "" {
			return &QemuVirtIOStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuVirtIOStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(settings)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
			AsyncIO:   tmpDisk.AsyncIO,
			Backup:    tmpDisk.Backup,
			Bandwidth: tmpDisk.Bandwidth,
			Cache:     tmpDisk.Cache,
			Discard:   tmpDisk.Discard,
			Id:        tmpDisk.Id,
			IOThread:  tmpDisk.IOThread,
			ReadOnly:  tmpDisk.ReadOnly,
			Replicate: tmpDisk.Replicate,
			Serial:    tmpDisk.Serial,
			Size:      tmpDisk.Size,
			Storage:   tmpDisk.Storage,
		}}
	}
	return &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
		AsyncIO:   tmpDisk.AsyncIO,
		Backup:    tmpDisk.Backup,
		Bandwidth: tmpDisk.Bandwidth,
		Cache:     tmpDisk.Cache,
		Discard:   tmpDisk.Discard,
		File:      tmpDisk.File,
		IOThread:  tmpDisk.IOThread,
		ReadOnly:  tmpDisk.ReadOnly,
		Serial:    tmpDisk.Serial,
		Size:      tmpDisk.Size,
	}}
}
