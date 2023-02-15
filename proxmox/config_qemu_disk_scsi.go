package proxmox

type QemuScsiDisk struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool
	IOThread   bool
	ReadOnly   bool
	Replicate  bool
	Size       uint
	Storage    string
}

type QemuScsiDisks struct {
	Disk_0  *QemuScsiStorage
	Disk_1  *QemuScsiStorage
	Disk_2  *QemuScsiStorage
	Disk_3  *QemuScsiStorage
	Disk_4  *QemuScsiStorage
	Disk_5  *QemuScsiStorage
	Disk_6  *QemuScsiStorage
	Disk_7  *QemuScsiStorage
	Disk_8  *QemuScsiStorage
	Disk_9  *QemuScsiStorage
	Disk_10 *QemuScsiStorage
	Disk_11 *QemuScsiStorage
	Disk_12 *QemuScsiStorage
	Disk_13 *QemuScsiStorage
	Disk_14 *QemuScsiStorage
	Disk_15 *QemuScsiStorage
	Disk_16 *QemuScsiStorage
	Disk_17 *QemuScsiStorage
	Disk_18 *QemuScsiStorage
	Disk_19 *QemuScsiStorage
	Disk_20 *QemuScsiStorage
	Disk_21 *QemuScsiStorage
	Disk_22 *QemuScsiStorage
	Disk_23 *QemuScsiStorage
	Disk_24 *QemuScsiStorage
	Disk_25 *QemuScsiStorage
	Disk_26 *QemuScsiStorage
	Disk_27 *QemuScsiStorage
	Disk_28 *QemuScsiStorage
	Disk_29 *QemuScsiStorage
	Disk_30 *QemuScsiStorage
}

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
	Size       uint
}

type QemuScsiStorage struct {
	CdRom       *QemuCdRom
	CloudInit   *QemuCloudInitDisk
	Disk        *QemuScsiDisk
	Passthrough *QemuScsiPassthrough
}

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
		Size:       tmpDisk.Size,
	}}

}
