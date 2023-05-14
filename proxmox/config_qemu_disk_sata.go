package proxmox

import (
	"strconv"
	"strings"
)

type QemuSataDisk struct {
	AsyncIO      QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup       bool              `json:"backup"`
	Bandwidth    QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache        QemuDiskCache     `json:"cache,omitempty"`
	Discard      bool              `json:"discard"`
	EmulateSSD   bool              `json:"emulatessd"`
	Format       QemuDiskFormat    `json:"format"`
	Id           uint              `json:"id"`     //Id is only returned and setting it has no effect
	LinkedDiskId *uint             `json:"linked"` //LinkedClone is only returned and setting it has no effect
	Replicate    bool              `json:"replicate"`
	Serial       QemuDiskSerial    `json:"serial,omitempty"`
	Size         uint              `json:"size"`
	Storage      string            `json:"storage"`
}

func (disk *QemuSataDisk) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:      disk.AsyncIO,
		Backup:       disk.Backup,
		Bandwidth:    disk.Bandwidth,
		Cache:        disk.Cache,
		Discard:      disk.Discard,
		Disk:         true,
		EmulateSSD:   disk.EmulateSSD,
		Format:       disk.Format,
		Id:           disk.Id,
		LinkedDiskId: disk.LinkedDiskId,
		Replicate:    disk.Replicate,
		Serial:       disk.Serial,
		Size:         disk.Size,
		Storage:      disk.Storage,
		Type:         sata,
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

func (disks QemuSataDisks) mapToApiValues(currentDisks *QemuSataDisks, vmID, LinkedVmId uint, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuSataDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
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

func (QemuSataDisks) mapToStruct(params map[string]interface{}, linkedVmId *uint) *QemuSataDisks {
	disks := QemuSataDisks{}
	var structPopulated bool
	if _, isSet := params["sata0"]; isSet {
		disks.Disk_0 = QemuSataStorage{}.mapToStruct(params["sata0"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["sata1"]; isSet {
		disks.Disk_1 = QemuSataStorage{}.mapToStruct(params["sata1"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["sata2"]; isSet {
		disks.Disk_2 = QemuSataStorage{}.mapToStruct(params["sata2"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["sata3"]; isSet {
		disks.Disk_3 = QemuSataStorage{}.mapToStruct(params["sata3"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["sata4"]; isSet {
		disks.Disk_4 = QemuSataStorage{}.mapToStruct(params["sata4"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["sata5"]; isSet {
		disks.Disk_5 = QemuSataStorage{}.mapToStruct(params["sata5"].(string), linkedVmId)
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
	AsyncIO    QemuDiskAsyncIO   `json:"asyncio,omitempty"`
	Backup     bool              `json:"backup"`
	Bandwidth  QemuDiskBandwidth `json:"bandwidth,omitempty"`
	Cache      QemuDiskCache     `json:"cache,omitempty"`
	Discard    bool              `json:"discard"`
	EmulateSSD bool              `json:"emulatessd"`
	File       string            `json:"file"`
	Replicate  bool              `json:"replicate"`
	Serial     QemuDiskSerial    `json:"serial,omitempty"`
	Size       uint              `json:"size"` //size is only returned and setting it has no effect
}

func (passthrough *QemuSataPassthrough) convertDataStructure() *qemuDisk {
	return &qemuDisk{
		AsyncIO:    passthrough.AsyncIO,
		Backup:     passthrough.Backup,
		Bandwidth:  passthrough.Bandwidth,
		Cache:      passthrough.Cache,
		Discard:    passthrough.Discard,
		EmulateSSD: passthrough.EmulateSSD,
		File:       passthrough.File,
		Replicate:  passthrough.Replicate,
		Serial:     passthrough.Serial,
		Type:       sata,
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
}

// converts to qemuStorage
func (storage *QemuSataStorage) convertDataStructure() *qemuStorage {
	if storage == nil {
		return nil
	}
	generalizedStorage := qemuStorage{
		CdRom:     storage.CdRom,
		CloudInit: storage.CloudInit,
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
			Size:    storage.Disk.Size,
			Storage: storage.Disk.Storage,
			Type:    ide,
		}
	}
	return nil
}

func (QemuSataStorage) mapToStruct(param string, LinkedVmId *uint) *QemuSataStorage {
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
			AsyncIO:      tmpDisk.AsyncIO,
			Backup:       tmpDisk.Backup,
			Bandwidth:    tmpDisk.Bandwidth,
			Cache:        tmpDisk.Cache,
			Discard:      tmpDisk.Discard,
			EmulateSSD:   tmpDisk.EmulateSSD,
			Format:       tmpDisk.Format,
			Id:           tmpDisk.Id,
			LinkedDiskId: tmpDisk.LinkedDiskId,
			Replicate:    tmpDisk.Replicate,
			Serial:       tmpDisk.Serial,
			Size:         tmpDisk.Size,
			Storage:      tmpDisk.Storage,
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
		Serial:     tmpDisk.Serial,
		Size:       tmpDisk.Size,
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
