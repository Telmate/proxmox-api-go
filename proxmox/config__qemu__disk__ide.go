package proxmox

import (
	"strconv"
	"strings"
)

type QemuIdeDisk struct {
	AsyncIO         QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Bandwidth       QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache           QemuDiskCache     `json:"cache,omitempty"`
	Format          QemuDiskFormat    `json:"format"`
	Id              uint              `json:"id"`     //Id is only returned and setting it has no effect
	LinkedDiskId    *GuestID          `json:"linked"` //LinkedClone is only returned and setting it has no effect
	Serial          QemuDiskSerial    `json:"serial,omitempty"`
	SizeInKibibytes QemuDiskSize      `json:"size"`
	Storage         string            `json:"storage"`
	syntax          diskSyntaxEnum
	WorldWideName   QemuWorldWideName `json:"wwn"`
	ImportFrom      string            `json:"import_from,omitempty"`
	Backup          bool              `json:"backup"`
	Discard         bool              `json:"discard"`
	EmulateSSD      bool              `json:"emulatessd"`
	Replicate       bool              `json:"replicate"`
}

func (disk *QemuIdeDisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:         disk.AsyncIO,
		Backup:          disk.Backup,
		Bandwidth:       disk.Bandwidth,
		Cache:           disk.Cache,
		Discard:         disk.Discard,
		Disk:            true,
		EmulateSSD:      disk.EmulateSSD,
		fileSyntax:      disk.syntax,
		Format:          disk.Format,
		Id:              disk.Id,
		LinkedDiskId:    disk.LinkedDiskId,
		Replicate:       disk.Replicate,
		Serial:          disk.Serial,
		SizeInKibibytes: disk.SizeInKibibytes,
		Storage:         disk.Storage,
		Type:            ide,
		WorldWideName:   disk.WorldWideName,
		ImportFrom:      disk.ImportFrom,
	}
}

func (disk QemuIdeDisk) Validate() error {
	return disk.convertDataStructure().validate()
}

type QemuIdeDisks struct {
	Disk_0 *QemuIdeStorage `json:"0,omitempty"`
	Disk_1 *QemuIdeStorage `json:"1,omitempty"`
	Disk_2 *QemuIdeStorage `json:"2,omitempty"`
	Disk_3 *QemuIdeStorage `json:"3,omitempty"`
}

func (q QemuIdeDisks) listCloudInitDisk() string {
	diskMap := q.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].CloudInit != nil {
			return "ide" + strconv.Itoa(int(i))
		}
	}
	return ""
}

func (disks QemuIdeDisks) mapToApiValues(currentDisks *QemuIdeDisks, vmID GuestID, LinkedVmId GuestID, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] == nil {
			continue
		}
		delete = diskMap[i].convertDataStructure().mapToApiValues(currentDiskMap[i].convertDataStructure(), vmID, LinkedVmId, QemuDiskId("ide"+strconv.Itoa(int(i))), params, delete)
	}
	return delete
}

func (disks QemuIdeDisks) mapToIntMap() map[uint8]*QemuIdeStorage {
	return map[uint8]*QemuIdeStorage{
		0: disks.Disk_0,
		1: disks.Disk_1,
		2: disks.Disk_2,
		3: disks.Disk_3,
	}
}

func (raw RawConfigQemu) disksIde(linkedVmId *GuestID) *QemuIdeDisks {
	disks := QemuIdeDisks{}
	var structPopulated bool
	if v, isSet := raw.a[qemuPrefixApiKeyDiskIde+"0"]; isSet {
		disks.Disk_0 = QemuIdeStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskIde+"1"]; isSet {
		disks.Disk_1 = QemuIdeStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskIde+"2"]; isSet {
		disks.Disk_2 = QemuIdeStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskIde+"3"]; isSet {
		disks.Disk_3 = QemuIdeStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

func (disks QemuIdeDisks) markDiskChanges(currentDisks *QemuIdeDisks, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		diskMap[i].convertDataStructureMark().markChanges(currentDiskMap[i].convertDataStructureMark(), QemuDiskId("ide"+strconv.Itoa(int(i))), changes)
	}
}

func (disks QemuIdeDisks) selectInitialResize(currentDisks *QemuIdeDisks) (resize []qemuDiskResize) {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].Disk != nil && diskMap[i].Disk.SizeInKibibytes%gibibyte != 0 && (currentDiskMap[i] == nil || currentDiskMap[i].Disk == nil || diskMap[i].Disk.SizeInKibibytes < currentDiskMap[i].Disk.SizeInKibibytes) {
			aaa := diskMap[i].Disk.SizeInKibibytes % gibibyte
			_ = aaa
			resize = append(resize, qemuDiskResize{
				Id:              QemuDiskId(qemuPrefixApiKeyDiskIde + strconv.Itoa(int(i))),
				SizeInKibibytes: diskMap[i].Disk.SizeInKibibytes,
			})
		}
	}
	return resize
}

func (disks QemuIdeDisks) Validate() (err error) {
	_, err = disks.validate()
	return
}

func (disks QemuIdeDisks) validate() (numberOfCloudInitDevices uint8, err error) {
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

type QemuIdePassthrough struct {
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
	Replicate       bool              `json:"replicate"`
}

func (passthrough *QemuIdePassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:       passthrough.AsyncIO,
		Backup:        passthrough.Backup,
		Bandwidth:     passthrough.Bandwidth,
		Cache:         passthrough.Cache,
		Discard:       passthrough.Discard,
		EmulateSSD:    passthrough.EmulateSSD,
		File:          passthrough.File,
		Replicate:     passthrough.Replicate,
		Serial:        passthrough.Serial,
		Type:          ide,
		WorldWideName: passthrough.WorldWideName,
	}
}

func (passthrough QemuIdePassthrough) Validate() error {
	return passthrough.convertDataStructure().validate()
}

type QemuIdeStorage struct {
	CdRom       *QemuCdRom          `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk  `json:"cloudinit,omitempty"`
	Disk        *QemuIdeDisk        `json:"disk,omitempty"`
	Passthrough *QemuIdePassthrough `json:"passthrough,omitempty"`
	Delete      bool                `json:"delete,omitempty"`
}

// converts to qemuStorage
func (storage *QemuIdeStorage) convertDataStructure() *qemuStorage {
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
func (storage *QemuIdeStorage) convertDataStructureMark() *qemuDiskMark {
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

func (QemuIdeStorage) mapToStruct(param string, LinkedVmId *GuestID) *QemuIdeStorage {
	diskData, _, _ := strings.Cut(param, ",")
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(diskData, settings)
	if tmpCdRom != nil {
		if tmpCdRom.CdRom {
			return &QemuIdeStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuIdeStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(diskData, settings, LinkedVmId)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuIdeStorage{Disk: &QemuIdeDisk{
			AsyncIO:         tmpDisk.AsyncIO,
			Backup:          tmpDisk.Backup,
			Bandwidth:       tmpDisk.Bandwidth,
			Cache:           tmpDisk.Cache,
			Discard:         tmpDisk.Discard,
			EmulateSSD:      tmpDisk.EmulateSSD,
			Format:          tmpDisk.Format,
			Id:              tmpDisk.Id,
			LinkedDiskId:    tmpDisk.LinkedDiskId,
			Replicate:       tmpDisk.Replicate,
			Serial:          tmpDisk.Serial,
			SizeInKibibytes: tmpDisk.SizeInKibibytes,
			Storage:         tmpDisk.Storage,
			syntax:          tmpDisk.fileSyntax,
			WorldWideName:   tmpDisk.WorldWideName,
		}}
	}
	return &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
		AsyncIO:         tmpDisk.AsyncIO,
		Backup:          tmpDisk.Backup,
		Bandwidth:       tmpDisk.Bandwidth,
		Cache:           tmpDisk.Cache,
		Discard:         tmpDisk.Discard,
		EmulateSSD:      tmpDisk.EmulateSSD,
		File:            tmpDisk.File,
		Replicate:       tmpDisk.Replicate,
		Serial:          tmpDisk.Serial,
		SizeInKibibytes: tmpDisk.SizeInKibibytes,
		WorldWideName:   tmpDisk.WorldWideName,
	}}
}

func (storage QemuIdeStorage) Validate() (err error) {
	_, err = storage.validate()
	return
}

func (storage QemuIdeStorage) validate() (CloudInit uint8, err error) {
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
