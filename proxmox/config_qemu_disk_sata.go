package proxmox

type QemuSataDisk struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool
	Replicate  bool
	Size       uint
	Storage    string
}

type QemuSataDisks struct {
	Disk_0 *QemuSataStorage
	Disk_1 *QemuSataStorage
	Disk_2 *QemuSataStorage
	Disk_3 *QemuSataStorage
	Disk_4 *QemuSataStorage
	Disk_5 *QemuSataStorage
}

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
	Size       uint
}

type QemuSataStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuSataDisk
	Passthrough *QemuSataPassthrough
}

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
		Size:       tmpDisk.Size,
	}}
}
