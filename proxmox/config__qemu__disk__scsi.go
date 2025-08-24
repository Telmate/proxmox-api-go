package proxmox

import (
	"strconv"
	"strings"
)

type QemuScsiDisk struct {
	AsyncIO         QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Bandwidth       QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache           QemuDiskCache     `json:"cache,omitempty"`
	Format          QemuDiskFormat    `json:"format"`
	Id              uint              `json:"id"`     //Id is only returned and setting it has no effect
	LinkedDiskId    *GuestID          `json:"linked"` //LinkedCloneId is only returned and setting it has no effect
	Serial          QemuDiskSerial    `json:"serial,omitempty"`
	SizeInKibibytes QemuDiskSize      `json:"size"`
	Storage         string            `json:"storage"`
	syntax          diskSyntaxEnum
	WorldWideName   QemuWorldWideName `json:"wwn"`
	ImportFrom      string            `json:"import_from,omitempty"`
	Backup          bool              `json:"backup"`
	Discard         bool              `json:"discard"`
	EmulateSSD      bool              `json:"emulatessd"`
	IOThread        bool              `json:"iothread"`
	ReadOnly        bool              `json:"readonly"`
	Replicate       bool              `json:"replicate"`
}

func (disk *QemuScsiDisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:         disk.AsyncIO,
		Backup:          disk.Backup,
		Bandwidth:       disk.Bandwidth,
		Cache:           disk.Cache,
		Discard:         disk.Discard,
		Disk:            true,
		EmulateSSD:      disk.EmulateSSD,
		Format:          disk.Format,
		Id:              disk.Id,
		IOThread:        disk.IOThread,
		LinkedDiskId:    disk.LinkedDiskId,
		ReadOnly:        disk.ReadOnly,
		Replicate:       disk.Replicate,
		Serial:          disk.Serial,
		SizeInKibibytes: disk.SizeInKibibytes,
		Storage:         disk.Storage,
		fileSyntax:      disk.syntax,
		Type:            scsi,
		WorldWideName:   disk.WorldWideName,
		ImportFrom:      disk.ImportFrom,
	}
}

func (disk QemuScsiDisk) Validate() error {
	return disk.convertDataStructure().validate()
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

func (q QemuScsiDisks) listCloudInitDisk() string {
	diskMap := q.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].CloudInit != nil {
			return qemuPrefixApiKeyDiskSCSI + strconv.Itoa(int(i))
		}
	}
	return ""
}

func (disks QemuScsiDisks) mapToApiValues(currentDisks *QemuScsiDisks, vmID, linkedVmId GuestID, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuScsiDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] == nil {
			continue
		}
		delete = diskMap[i].convertDataStructure().mapToApiValues(currentDiskMap[i].convertDataStructure(), vmID, linkedVmId, QemuDiskId("scsi"+strconv.Itoa(int(i))), params, delete)
	}
	return delete
}

func (disks QemuScsiDisks) mapToIntMap() map[uint8]*QemuScsiStorage {
	return map[uint8]*QemuScsiStorage{
		0:  disks.Disk_0,
		1:  disks.Disk_1,
		2:  disks.Disk_2,
		3:  disks.Disk_3,
		4:  disks.Disk_4,
		5:  disks.Disk_5,
		6:  disks.Disk_6,
		7:  disks.Disk_7,
		8:  disks.Disk_8,
		9:  disks.Disk_9,
		10: disks.Disk_10,
		11: disks.Disk_11,
		12: disks.Disk_12,
		13: disks.Disk_13,
		14: disks.Disk_14,
		15: disks.Disk_15,
		16: disks.Disk_16,
		17: disks.Disk_17,
		18: disks.Disk_18,
		19: disks.Disk_19,
		20: disks.Disk_20,
		21: disks.Disk_21,
		22: disks.Disk_22,
		23: disks.Disk_23,
		24: disks.Disk_24,
		25: disks.Disk_25,
		26: disks.Disk_26,
		27: disks.Disk_27,
		28: disks.Disk_28,
		29: disks.Disk_29,
		30: disks.Disk_30,
	}
}

func (raw RawConfigQemu) disksSCSI(linkedVmId *GuestID) *QemuScsiDisks {
	disks := QemuScsiDisks{}
	var structPopulated bool
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"0"]; isSet {
		disks.Disk_0 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"1"]; isSet {
		disks.Disk_1 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"2"]; isSet {
		disks.Disk_2 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"3"]; isSet {
		disks.Disk_3 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"4"]; isSet {
		disks.Disk_4 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"5"]; isSet {
		disks.Disk_5 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"6"]; isSet {
		disks.Disk_6 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"7"]; isSet {
		disks.Disk_7 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"8"]; isSet {
		disks.Disk_8 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"9"]; isSet {
		disks.Disk_9 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"10"]; isSet {
		disks.Disk_10 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"11"]; isSet {
		disks.Disk_11 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"12"]; isSet {
		disks.Disk_12 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"13"]; isSet {
		disks.Disk_13 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"14"]; isSet {
		disks.Disk_14 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"15"]; isSet {
		disks.Disk_15 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"16"]; isSet {
		disks.Disk_16 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"17"]; isSet {
		disks.Disk_17 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"18"]; isSet {
		disks.Disk_18 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"19"]; isSet {
		disks.Disk_19 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"20"]; isSet {
		disks.Disk_20 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"21"]; isSet {
		disks.Disk_21 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"22"]; isSet {
		disks.Disk_22 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"23"]; isSet {
		disks.Disk_23 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"24"]; isSet {
		disks.Disk_24 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"25"]; isSet {
		disks.Disk_25 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"26"]; isSet {
		disks.Disk_26 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"27"]; isSet {
		disks.Disk_27 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"28"]; isSet {
		disks.Disk_28 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"29"]; isSet {
		disks.Disk_29 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSCSI+"30"]; isSet {
		disks.Disk_30 = QemuScsiStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

func (disks QemuScsiDisks) markDiskChanges(currentDisks *QemuScsiDisks, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuScsiDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		diskMap[i].convertDataStructureMark().markChanges(currentDiskMap[i].convertDataStructureMark(), QemuDiskId("scsi"+strconv.Itoa(int(i))), changes)
	}
}

func (disks QemuScsiDisks) selectInitialResize(currentDisks *QemuScsiDisks) (resize []qemuDiskResize) {
	tmpCurrentDisks := QemuScsiDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].Disk != nil && diskMap[i].Disk.SizeInKibibytes%gibibyte != 0 && (currentDiskMap[i] == nil || currentDiskMap[i].Disk == nil || diskMap[i].Disk.SizeInKibibytes < currentDiskMap[i].Disk.SizeInKibibytes) {
			resize = append(resize, qemuDiskResize{
				Id:              QemuDiskId("scsi" + strconv.Itoa(int(i))),
				SizeInKibibytes: diskMap[i].Disk.SizeInKibibytes,
			})
		}
	}
	return resize
}

func (disks QemuScsiDisks) Validate() (err error) {
	_, err = disks.validate()
	return
}

func (disks QemuScsiDisks) validate() (numberOfCloudInitDevices uint8, err error) {
	diskMap := disks.mapToIntMap()
	var cloudInit uint8
	for _, e := range diskMap {
		if e != nil {
			cloudInit, err = e.validate()
			if err != nil {
				return
			}
			numberOfCloudInitDevices += cloudInit
			if err = (QemuCloudInitDisk{}.checkDuplicates(numberOfCloudInitDevices)); err != nil {
				return
			}
		}
	}
	return
}

type QemuScsiPassthrough struct {
	AsyncIO         QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Bandwidth       QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache           QemuDiskCache     `json:"cache,omitempty"`
	File            string            `json:"file"`
	Serial          QemuDiskSerial    `json:"serial,omitempty"`
	SizeInKibibytes QemuDiskSize      `json:"size"` //size is only returned and setting it has no effect
	WorldWideName   QemuWorldWideName `json:"wwn"`
	Backup          bool              `json:"backup"`
	Discard         bool              `json:"discard"`
	EmulateSSD      bool              `json:"emulatessd"`
	IOThread        bool              `json:"iothread"`
	ReadOnly        bool              `json:"readonly"`
	Replicate       bool              `json:"replicate"`
}

func (passthrough *QemuScsiPassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:       passthrough.AsyncIO,
		Backup:        passthrough.Backup,
		Bandwidth:     passthrough.Bandwidth,
		Cache:         passthrough.Cache,
		Discard:       passthrough.Discard,
		EmulateSSD:    passthrough.EmulateSSD,
		File:          passthrough.File,
		IOThread:      passthrough.IOThread,
		ReadOnly:      passthrough.ReadOnly,
		Replicate:     passthrough.Replicate,
		Serial:        passthrough.Serial,
		Type:          scsi,
		WorldWideName: passthrough.WorldWideName,
	}
}

func (passthrough QemuScsiPassthrough) Validate() error {
	return passthrough.convertDataStructure().validate()
}

type QemuScsiStorage struct {
	CdRom       *QemuCdRom           `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk   `json:"cloudinit,omitempty"`
	Disk        *QemuScsiDisk        `json:"disk,omitempty"`
	Passthrough *QemuScsiPassthrough `json:"passthrough,omitempty"`
	Delete      bool                 `json:"delete,omitempty"`
}

// converts to qemuStorage
func (storage *QemuScsiStorage) convertDataStructure() *qemuStorage {
	if storage == nil {
		return nil
	}
	generalizedStorage := qemuStorage{
		CdRom:     storage.CdRom,
		CloudInit: storage.CloudInit,
		delete:    storage.Delete,
	}
	if storage.Disk != nil {
		generalizedStorage.Disk = storage.Disk.convertDataStructure()
	}
	if storage.Passthrough != nil {
		generalizedStorage.Passthrough = storage.Passthrough.convertDataStructure()
	}
	return &generalizedStorage
}

// converts to qemuDiskMark
func (storage *QemuScsiStorage) convertDataStructureMark() *qemuDiskMark {
	if storage == nil {
		return nil
	}
	if storage.Disk != nil {
		return &qemuDiskMark{
			Format:  storage.Disk.Format,
			Size:    storage.Disk.SizeInKibibytes,
			Storage: storage.Disk.Storage,
			Type:    ide,
		}
	}
	return nil
}

func (QemuScsiStorage) mapToStruct(param string, LinkedVmId *GuestID) *QemuScsiStorage {
	diskData, _, _ := strings.Cut(param, ",")
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(diskData, settings)
	if tmpCdRom != nil {
		if tmpCdRom.CdRom {
			return &QemuScsiStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuScsiStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(diskData, settings, LinkedVmId)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuScsiStorage{Disk: &QemuScsiDisk{
			AsyncIO:         tmpDisk.AsyncIO,
			Backup:          tmpDisk.Backup,
			Bandwidth:       tmpDisk.Bandwidth,
			Cache:           tmpDisk.Cache,
			Discard:         tmpDisk.Discard,
			EmulateSSD:      tmpDisk.EmulateSSD,
			Format:          tmpDisk.Format,
			Id:              tmpDisk.Id,
			IOThread:        tmpDisk.IOThread,
			LinkedDiskId:    tmpDisk.LinkedDiskId,
			ReadOnly:        tmpDisk.ReadOnly,
			Replicate:       tmpDisk.Replicate,
			Serial:          tmpDisk.Serial,
			SizeInKibibytes: tmpDisk.SizeInKibibytes,
			Storage:         tmpDisk.Storage,
			syntax:          tmpDisk.fileSyntax,
			WorldWideName:   tmpDisk.WorldWideName,
		}}
	}
	return &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
		AsyncIO:         tmpDisk.AsyncIO,
		Backup:          tmpDisk.Backup,
		Bandwidth:       tmpDisk.Bandwidth,
		Cache:           tmpDisk.Cache,
		Discard:         tmpDisk.Discard,
		EmulateSSD:      tmpDisk.EmulateSSD,
		File:            tmpDisk.File,
		IOThread:        tmpDisk.IOThread,
		ReadOnly:        tmpDisk.ReadOnly,
		Replicate:       tmpDisk.Replicate,
		Serial:          tmpDisk.Serial,
		SizeInKibibytes: tmpDisk.SizeInKibibytes,
		WorldWideName:   tmpDisk.WorldWideName,
	}}
}

func (storage QemuScsiStorage) Validate() (err error) {
	_, err = storage.validate()
	return
}

func (storage QemuScsiStorage) validate() (CloudInit uint8, err error) {
	// First check if more than one item is nil
	var subTypeSet bool
	if storage.CdRom != nil {
		subTypeSet = true
	}
	if storage.CloudInit != nil {
		if err = diskSubtypeSet(subTypeSet); err != nil {
			return
		}
		subTypeSet = true
		CloudInit = 1
	}
	if storage.Disk != nil {
		if err = diskSubtypeSet(subTypeSet); err != nil {
			return
		}
		subTypeSet = true
	}
	if storage.Passthrough != nil {
		if err = diskSubtypeSet(subTypeSet); err != nil {
			return
		}
	}
	// Validate sub items
	if storage.CdRom != nil {
		if err = storage.CdRom.Validate(); err != nil {
			return
		}
	}
	if storage.CloudInit != nil {
		if err = storage.CloudInit.Validate(); err != nil {
			return
		}
	}
	if storage.Disk != nil {
		if err = storage.Disk.Validate(); err != nil {
			return
		}
	}
	if storage.Passthrough != nil {
		err = storage.Passthrough.Validate()
	}
	return
}
