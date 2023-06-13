package proxmox

import (
	"strconv"
	"strings"
)

type QemuIdeDisk struct {
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

func (disk *QemuIdeDisk) convertDataStructure() *qemuDisk {
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
		Type:         ide,
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

func (disks QemuIdeDisks) mapToApiValues(currentDisks *QemuIdeDisks, vmID, LinkedVmId uint, params map[string]interface{}, delete string) string {
	tmpCurrentDisks := QemuIdeDisks{}
	if currentDisks != nil {
		tmpCurrentDisks = *currentDisks
	}
	diskMap := disks.mapToIntMap()
	currentDiskMap := tmpCurrentDisks.mapToIntMap()
	for i := range diskMap {
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

func (QemuIdeDisks) mapToStruct(params map[string]interface{}, linkedVmId *uint) *QemuIdeDisks {
	disks := QemuIdeDisks{}
	var structPopulated bool
	if _, isSet := params["ide0"]; isSet {
		disks.Disk_0 = QemuIdeStorage{}.mapToStruct(params["ide0"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["ide1"]; isSet {
		disks.Disk_1 = QemuIdeStorage{}.mapToStruct(params["ide1"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["ide2"]; isSet {
		disks.Disk_2 = QemuIdeStorage{}.mapToStruct(params["ide2"].(string), linkedVmId)
		structPopulated = true
	}
	if _, isSet := params["ide3"]; isSet {
		disks.Disk_3 = QemuIdeStorage{}.mapToStruct(params["ide3"].(string), linkedVmId)
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

func (passthrough *QemuIdePassthrough) convertDataStructure() *qemuDisk {
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
		Type:       ide,
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
}

// converts to qemuStorage
func (storage *QemuIdeStorage) convertDataStructure() *qemuStorage {
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
func (storage *QemuIdeStorage) convertDataStructureMark() *qemuDiskMark {
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

func (QemuIdeStorage) mapToStruct(param string, LinkedVmId *uint) *QemuIdeStorage {
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
	return &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
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
