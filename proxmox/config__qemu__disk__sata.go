package proxmox

import (
	"strconv"
	"strings"
)

type QemuSataDisk struct {
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

func (disk *QemuSataDisk) convertDataStructure() *qemuDisk {
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
		Type:            sata,
		WorldWideName:   disk.WorldWideName,
		ImportFrom:      disk.ImportFrom,
	}
}

func (disk QemuSataDisk) Validate() error {
	return disk.convertDataStructure().validate()
}

type QemuSataDisks struct {
	Disk_0 *QemuSataStorage `json:"0,omitempty"`
	Disk_1 *QemuSataStorage `json:"1,omitempty"`
	Disk_2 *QemuSataStorage `json:"2,omitempty"`
	Disk_3 *QemuSataStorage `json:"3,omitempty"`
	Disk_4 *QemuSataStorage `json:"4,omitempty"`
	Disk_5 *QemuSataStorage `json:"5,omitempty"`
}

func (q QemuSataDisks) listCloudInitDisk() string {
	diskMap := q.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].CloudInit != nil {
			return qemuPrefixApiKeyDiskSata + strconv.Itoa(int(i))
		}
	}
	return ""
}

func (disks QemuSataDisks) mapToApiValues(currentDisks *QemuSataDisks, vmID, LinkedVmId GuestID, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] == nil {
			continue
		}
		delete = diskMap[i].convertDataStructure().mapToApiValues(currentDiskMap[i].convertDataStructure(), vmID, LinkedVmId, QemuDiskId("sata"+strconv.Itoa(int(i))), params, delete)
	}
	return delete
}

func (disks QemuSataDisks) mapToIntMap() map[uint8]*QemuSataStorage {
	return map[uint8]*QemuSataStorage{
		0: disks.Disk_0,
		1: disks.Disk_1,
		2: disks.Disk_2,
		3: disks.Disk_3,
		4: disks.Disk_4,
		5: disks.Disk_5,
	}
}

func (raw RawConfigQemu) disksSata(linkedVmId *GuestID) *QemuSataDisks {
	disks := QemuSataDisks{}
	var structPopulated bool
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"0"]; isSet {
		disks.Disk_0 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"1"]; isSet {
		disks.Disk_1 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"2"]; isSet {
		disks.Disk_2 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"3"]; isSet {
		disks.Disk_3 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"4"]; isSet {
		disks.Disk_4 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if v, isSet := raw.a[qemuPrefixApiKeyDiskSata+"5"]; isSet {
		disks.Disk_5 = QemuSataStorage{}.mapToStruct(v.(string), linkedVmId)
		structPopulated = true
	}
	if structPopulated {
		return &disks
	}
	return nil
}

func (disks QemuSataDisks) markDiskChanges(currentDisks *QemuSataDisks, changes *qemuUpdateChanges) {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		diskMap[i].convertDataStructureMark().markChanges(currentDiskMap[i].convertDataStructureMark(), QemuDiskId("sata"+strconv.Itoa(int(i))), changes)
	}
}

func (disks QemuSataDisks) selectInitialResize(currentDisks *QemuSataDisks) (resize []qemuDiskResize) {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
		if diskMap[i] != nil && diskMap[i].Disk != nil && diskMap[i].Disk.SizeInKibibytes%gibibyte != 0 && (currentDiskMap[i] == nil || currentDiskMap[i].Disk == nil || diskMap[i].Disk.SizeInKibibytes < currentDiskMap[i].Disk.SizeInKibibytes) {
			resize = append(resize, qemuDiskResize{
				Id:              QemuDiskId("sata" + strconv.Itoa(int(i))),
				SizeInKibibytes: diskMap[i].Disk.SizeInKibibytes,
			})
		}
	}
	return resize
}

func (disks QemuSataDisks) Validate() (err error) {
	_, err = disks.validate()
	return
}

func (disks QemuSataDisks) validate() (numberOfCloudInitDevices uint8, err error) {
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

type QemuSataPassthrough struct {
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

func (passthrough *QemuSataPassthrough) convertDataStructure() *qemuDisk {
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
		Type:          sata,
		WorldWideName: passthrough.WorldWideName,
	}
}

func (passthrough QemuSataPassthrough) Validate() error {
	return passthrough.convertDataStructure().validate()
}

type QemuSataStorage struct {
	CdRom       *QemuCdRom           `json:"cdrom,omitempty"`
	CloudInit   *QemuCloudInitDisk   `json:"cloudinit,omitempty"`
	Disk        *QemuSataDisk        `json:"disk,omitempty"`
	Passthrough *QemuSataPassthrough `json:"passthrough,omitempty"`
	Delete      bool                 `json:"delete,omitempty"`
}

// converts to qemuStorage
func (storage *QemuSataStorage) convertDataStructure() *qemuStorage {
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
func (storage *QemuSataStorage) convertDataStructureMark() *qemuDiskMark {
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

func (QemuSataStorage) mapToStruct(param string, LinkedVmId *GuestID) *QemuSataStorage {
	diskData, _, _ := strings.Cut(param, ",")
	settings := splitStringOfSettings(param)
	tmpCdRom := qemuCdRom{}.mapToStruct(diskData, settings)
	if tmpCdRom != nil {
		if tmpCdRom.CdRom {
			return &QemuSataStorage{CdRom: QemuCdRom{}.mapToStruct(*tmpCdRom)}
		} else {
			return &QemuSataStorage{CloudInit: QemuCloudInitDisk{}.mapToStruct(*tmpCdRom)}
		}
	}

	tmpDisk := qemuDisk{}.mapToStruct(diskData, settings, LinkedVmId)
	if tmpDisk == nil {
		return nil
	}
	if tmpDisk.File == "" {
		return &QemuSataStorage{Disk: &QemuSataDisk{
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
	return &QemuSataStorage{Passthrough: &QemuSataPassthrough{
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

func (storage QemuSataStorage) Validate() (err error) {
	_, err = storage.validate()
	return
}

func (storage QemuSataStorage) validate() (CloudInit uint8, err error) {
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
