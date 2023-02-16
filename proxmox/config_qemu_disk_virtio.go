package proxmox

type QemuVirtIODisk struct {
	AsyncIO   QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup    bool              `json:"backup,omitempty"`
	Bandwidth QemuDiskBandwidth `json:"bandwith,omitempty"`
	Cache     QemuDiskCache     `json:"cache,omitempty"`
	Discard   bool              `json:"discard,omitempty"`
	IOThread  bool              `json:"iothread,omitempty"`
	ReadOnly  bool              `json:"readonly,omitempty"`
	Replicate bool              `json:"replicate,omitempty"`
	Serial    QemuDiskSerial    `json:"serial,omitempty"`
	Size      uint              `json:"size,omitempty"`
	Storage   string            `json:"storage,omitempty"`
}

func (disk QemuVirtIODisk) mapToApiValues(create bool) string {
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
	}.mapToApiValues(create)
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

func (disks QemuVirtIODisks) mapToApiValues(create bool, params map[string]interface{}) {
	if disks.Disk_0 != nil {
		params["virtio0"] = disks.Disk_0.mapToApiValues(create)
	}
	if disks.Disk_1 != nil {
		params["virtio1"] = disks.Disk_1.mapToApiValues(create)
	}
	if disks.Disk_2 != nil {
		params["virtio2"] = disks.Disk_2.mapToApiValues(create)
	}
	if disks.Disk_3 != nil {
		params["virtio3"] = disks.Disk_3.mapToApiValues(create)
	}
	if disks.Disk_4 != nil {
		params["virtio4"] = disks.Disk_4.mapToApiValues(create)
	}
	if disks.Disk_5 != nil {
		params["virtio5"] = disks.Disk_5.mapToApiValues(create)
	}
	if disks.Disk_6 != nil {
		params["virtio6"] = disks.Disk_6.mapToApiValues(create)
	}
	if disks.Disk_7 != nil {
		params["virtio7"] = disks.Disk_7.mapToApiValues(create)
	}
	if disks.Disk_8 != nil {
		params["virtio8"] = disks.Disk_8.mapToApiValues(create)
	}
	if disks.Disk_9 != nil {
		params["virtio9"] = disks.Disk_9.mapToApiValues(create)
	}
	if disks.Disk_10 != nil {
		params["virtio10"] = disks.Disk_10.mapToApiValues(create)
	}
	if disks.Disk_11 != nil {
		params["virtio11"] = disks.Disk_11.mapToApiValues(create)
	}
	if disks.Disk_12 != nil {
		params["virtio12"] = disks.Disk_12.mapToApiValues(create)
	}
	if disks.Disk_13 != nil {
		params["virtio13"] = disks.Disk_13.mapToApiValues(create)
	}
	if disks.Disk_14 != nil {
		params["virtio14"] = disks.Disk_14.mapToApiValues(create)
	}
	if disks.Disk_15 != nil {
		params["virtio15"] = disks.Disk_15.mapToApiValues(create)
	}
}

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
	Size      uint
}

// TODO write function
func (passthrough QemuVirtIOPassthrough) mapToApiValues() string {
	return ""
}

type QemuVirtIOStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuVirtIODisk
	Passthrough *QemuVirtIOPassthrough
}

func (storage QemuVirtIOStorage) mapToApiValues(create bool) string {
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
