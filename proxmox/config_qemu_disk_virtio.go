package proxmox

type QemuVirtIODisk struct {
	AsyncIO   QemuDiskAsyncIO
	Backup    bool
	Bandwidth QemuDiskBandwidth
	Cache     QemuDiskCache
	Discard   bool
	IOThread  bool
	ReadOnly  bool
	Size      uint
	Storage   string
}

type QemuVirtIODisks struct {
	Disk_0  *QemuVirtIOStorage
	Disk_1  *QemuVirtIOStorage
	Disk_2  *QemuVirtIOStorage
	Disk_3  *QemuVirtIOStorage
	Disk_4  *QemuVirtIOStorage
	Disk_5  *QemuVirtIOStorage
	Disk_6  *QemuVirtIOStorage
	Disk_7  *QemuVirtIOStorage
	Disk_8  *QemuVirtIOStorage
	Disk_9  *QemuVirtIOStorage
	Disk_10 *QemuVirtIOStorage
	Disk_11 *QemuVirtIOStorage
	Disk_12 *QemuVirtIOStorage
	Disk_13 *QemuVirtIOStorage
	Disk_14 *QemuVirtIOStorage
	Disk_15 *QemuVirtIOStorage
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
	Size      uint
}

type QemuVirtIOStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuVirtIODisk
	Passthrough *QemuVirtIOPassthrough
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
		Size:      tmpDisk.Size,
	}}
}
