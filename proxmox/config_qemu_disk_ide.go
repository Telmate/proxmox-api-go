package proxmox

type QemuIdeDisk struct {
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

type QemuIdeDisks struct {
	Disk_0 *QemuIdeStorage
	Disk_1 *QemuIdeStorage
	Disk_2 *QemuIdeStorage
	Disk_3 *QemuIdeStorage
}

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
	Size       uint
}

type QemuIdeStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuIdeDisk
	Passthrough *QemuIdePassthrough
}

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
		Size:       tmpDisk.Size,
	}}
}
