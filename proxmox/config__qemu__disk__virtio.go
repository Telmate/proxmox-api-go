package proxmox

import (
	"strconv"
	"strings"
)

type QemuVirtIODisk struct {
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
	IOThread        bool              `json:"iothread"`
	ReadOnly        bool              `json:"readonly"`
	Replicate       bool              `json:"replicate"`
}

func (disk *QemuVirtIODisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:         disk.AsyncIO,
		Backup:          disk.Backup,
		Bandwidth:       disk.Bandwidth,
		Cache:           disk.Cache,
		Discard:         disk.Discard,
		Disk:            true,
		fileSyntax:      disk.syntax,
		Format:          disk.Format,
		Id:              disk.Id,
		IOThread:        disk.IOThread,
		LinkedDiskId:    disk.LinkedDiskId,
		ReadOnly:        disk.ReadOnly,
		Replicate:       disk.Replicate,
		Serial:          disk.Serial,
		SizeInKibibytes: disk.SizeInKibibytes,
		Storage:         disk.Storage,
		Type:            virtIO,
		WorldWideName:   disk.WorldWideName,
		ImportFrom:      disk.ImportFrom,
	}
}

func (disk QemuVirtIODisk) Validate() error {
	return disk.convertDataStructure().validate()
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

func (q QemuVirtIODisks) listCloudInitDisk() string {
	diskMap := q.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].CloudInit != nil {
			return qemuPrefixApiKeyDiskVirtIO + strconv.Itoa(int(i))
		}
	}
	return ""
}

func (disks QemuVirtIODisks) mapToApiValues(currentDisks *QemuVirtIODisks, vmID, linkedVmId GuestID, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuVirtIODisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] == nil {
			continue
		}
		delete = diskMap[i].convertDataStructure().mapToApiValues(currentDiskMap[i].convertDataStructure(), vmID, linkedVmId, QemuDiskId("virtio"+strconv.Itoa(int(i))), params, delete)
	}
	return delete
}

func (disks QemuVirtIODisks) mapToIntMap() map[uint8]*QemuVirtIOStorage {
	return map[uint8]*QemuVirtIOStorage{
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
	}
}

func (raw RawConfigQemu) disksVirtIO(linkedVmId *GuestID) *QemuVirtIODisks {
	disks := QemuVirtIODisks{}
	var structPopulated bool
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"0"]; isSet {
		disks.Disk_0 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"1"]; isSet {
		disks.Disk_1 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"2"]; isSet {
		disks.Disk_2 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"3"]; isSet {
		disks.Disk_3 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"4"]; isSet {
		disks.Disk_4 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"5"]; isSet {
		disks.Disk_5 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"6"]; isSet {
		disks.Disk_6 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"7"]; isSet {
		disks.Disk_7 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"8"]; isSet {
		disks.Disk_8 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"9"]; isSet {
		disks.Disk_9 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"10"]; isSet {
		disks.Disk_10 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"11"]; isSet {
		disks.Disk_11 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"12"]; isSet {
		disks.Disk_12 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"13"]; isSet {
		disks.Disk_13 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"14"]; isSet {
		disks.Disk_14 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskVirtIO+"15"]; isSet {
		disks.Disk_15 = QemuVirtIOStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

func (disks QemuVirtIODisks) markDiskChanges(currentDisks *QemuVirtIODisks, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuVirtIODisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		diskMap[i].convertDataStructureMark().markChanges(currentDiskMap[i].convertDataStructureMark(), QemuDiskId("virtio"+strconv.Itoa(int(i))), changes)
	}
}

func (disks QemuVirtIODisks) selectInitialResize(currentDisks *QemuVirtIODisks) (resize []qemuDiskResize) {
	tmpCurrentDisks := QemuVirtIODisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].Disk != nil && diskMap[i].Disk.SizeInKibibytes%gibibyte != 0 && (currentDiskMap[i] == nil || currentDiskMap[i].Disk == nil || diskMap[i].Disk.SizeInKibibytes < currentDiskMap[i].Disk.SizeInKibibytes) {
			resize = append(resize, qemuDiskResize{
				Id:              QemuDiskId("virtio" + strconv.Itoa(int(i))),
				SizeInKibibytes: diskMap[i].Disk.SizeInKibibytes,
			})
		}
	}
	return resize
}

func (disks QemuVirtIODisks) Validate() (err error) {
	_, err = disks.validate()
	return
}

func (disks QemuVirtIODisks) validate() (numberOfCloudInitDevices uint8, err error) {
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

type QemuVirtIOPassthrough struct {
	AsyncIO         QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Bandwidth       QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache           QemuDiskCache     `json:"cache,omitempty"`
	File            string            `json:"file"`
	Serial          QemuDiskSerial    `json:"serial,omitempty"`
	SizeInKibibytes QemuDiskSize      `json:"size"` //size is only returned and setting it has no effect
	WorldWideName   QemuWorldWideName `json:"wwn"`
	Backup          bool              `json:"backup"`
	Discard         bool              `json:"discard"`
	IOThread        bool              `json:"iothread"`
	ReadOnly        bool              `json:"readonly"`
	Replicate       bool              `json:"replicate"`
}

func (passthrough *QemuVirtIOPassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:       passthrough.AsyncIO,
		Backup:        passthrough.Backup,
		Bandwidth:     passthrough.Bandwidth,
		Cache:         passthrough.Cache,
		Discard:       passthrough.Discard,
		File:          passthrough.File,
		IOThread:      passthrough.IOThread,
		ReadOnly:      passthrough.ReadOnly,
		Replicate:     passthrough.Replicate,
		Serial:        passthrough.Serial,
		Type:          virtIO,
		WorldWideName: passthrough.WorldWideName,
	}
}

func (passthrough QemuVirtIOPassthrough) Validate() error {
	return passthrough.convertDataStructure().validate()
}

type QemuVirtIOStorage struct {
	CdRom       *QemuCdRom             `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk     `json:"cloudinit,omitempty"`
	Disk        *QemuVirtIODisk        `json:"disk,omitempty"`
	Passthrough *QemuVirtIOPassthrough `json:"passthrough,omitempty"`
	Delete      bool                   `json:"delete,omitempty"`
}

// converts to qemuStorage
func (storage *QemuVirtIOStorage) convertDataStructure() *qemuStorage {
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
func (storage *QemuVirtIOStorage) convertDataStructureMark() *qemuDiskMark {
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

func (QemuVirtIOStorage) mapToStruct(param string, LinkedVmId *GuestID) *QemuVirtIOStorage {
	diskData, _, _ := strings.Cut(param, ",")
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(diskData, settings)
	if tmpCdRom != nil {
		if tmpCdRom.CdRom {
			return &QemuVirtIOStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuVirtIOStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(diskData, settings, LinkedVmId)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
			AsyncIO:         tmpDisk.AsyncIO,
			Backup:          tmpDisk.Backup,
			Bandwidth:       tmpDisk.Bandwidth,
			Cache:           tmpDisk.Cache,
			Discard:         tmpDisk.Discard,
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
	return &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
		AsyncIO:         tmpDisk.AsyncIO,
		Backup:          tmpDisk.Backup,
		Bandwidth:       tmpDisk.Bandwidth,
		Cache:           tmpDisk.Cache,
		Discard:         tmpDisk.Discard,
		File:            tmpDisk.File,
		IOThread:        tmpDisk.IOThread,
		ReadOnly:        tmpDisk.ReadOnly,
		Replicate:       tmpDisk.Replicate,
		Serial:          tmpDisk.Serial,
		SizeInKibibytes: tmpDisk.SizeInKibibytes,
		WorldWideName:   tmpDisk.WorldWideName,
	}}
}

func (storage QemuVirtIOStorage) Validate() (err error) {
	_, err = storage.validate()
	return
}

func (storage QemuVirtIOStorage) validate() (CloudInit uint8, err error) {
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
