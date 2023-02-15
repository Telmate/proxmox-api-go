package proxmox

import (
	"math"
	"strconv"
	"strings"
)

type IsoFile struct {
	Storage string
	File    string
	// Size can only be retrieved, setting it has no effect
	Size string
}

type QemuCdRom struct {
	Iso *IsoFile
	// Passthrough and File are mutually exclusive
	Passthrough bool
}

func (QemuCdRom) mapToStruct(settings qemuCdRom) *QemuCdRom {
	if !settings.Passthrough {
		return &QemuCdRom{
			Iso: &IsoFile{
				Storage: settings.Storage,
				File:    settings.File,
				Size:    settings.Size,
			},
		}
	}
	return &QemuCdRom{Passthrough: false}
}

type qemuCdRom struct {
	// "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"
	Passthrough bool
	Storage     string
	// FileType is only set for Cloud init drives, this value will be used to check if it is a normal cdrom or cloud init drive.
	FileType string
	File     string
	Size     string
}

func (qemuCdRom) mapToStruct(settings [][]string) *qemuCdRom {
	var isCdRom bool
	for _, e := range settings {
		if e[0] == "media" {
			if e[1] == "cdrom" {
				isCdRom = true
				break
			}
		}
	}
	if !isCdRom {
		return nil
	}
	if settings[0][0] == "none" {
		return &qemuCdRom{}
	}
	if settings[0][0] == "cdrom" {
		return &qemuCdRom{Passthrough: true}
	}
	tmpStorage := strings.Split(settings[0][0], ":")
	if len(tmpStorage) > 1 {
		tmpFile := strings.Split(settings[0][0], "/")
		if len(tmpFile) == 2 {
			tmpFileType := strings.Split(tmpFile[1], ".")
			if len(tmpFileType) > 1 {
				fileType := tmpFileType[len(tmpFileType)-1]
				if fileType == "iso" {
					for _, e := range settings {
						if e[0] == "size" {
							return &qemuCdRom{
								Storage: tmpStorage[0],
								File:    tmpFile[1],
								Size:    e[1],
							}
						}
					}
				} else {
					return &qemuCdRom{
						Storage:  tmpStorage[0],
						File:     tmpFile[1],
						FileType: fileType,
					}
				}
			}
		}
	}
	return nil
}

type QemuCloudInitDisk struct {
	Storage  string
	FileType string
}

func (QemuCloudInitDisk) mapToStruct(settings qemuCdRom) *QemuCloudInitDisk {
	return &QemuCloudInitDisk{
		Storage:  settings.Storage,
		FileType: settings.FileType,
	}
}

// TODO add enum
type QemuDiskAsyncIO string

type QemuDiskBandwidth struct {
	ReadLimit_Data  QemuDisk_Bandwidth_Data
	WriteLimit_Data QemuDisk_Bandwidth_Data
	ReadLimit_Iops  QemuDisk_Bandwidth_Iops
	WriteLimit_Iops QemuDisk_Bandwidth_Iops
}

type QemuDisk_Bandwidth_Data struct {
	Concurrent float32
	Burst      float32
}

type QemuDisk_Bandwidth_Iops struct {
	Concurrent uint
	Burst      uint
}

// TODO add enum
type QemuDiskCache string

// TODO add enum
type QemuDiskFormat string

func (QemuDiskFormat) mapToStruct(setting string) QemuDiskFormat {
	settings := strings.Split(setting, ".")
	if len(settings) < 2 {
		return ""
	}
	return QemuDiskFormat(settings[len(settings)-1])
}

type qemuDisk struct {
	AsyncIO    QemuDiskAsyncIO
	Backup     bool
	Bandwidth  QemuDiskBandwidth
	Cache      QemuDiskCache
	Discard    bool
	EmulateSSD bool
	// TODO custom type
	// File is only set for Passthrough.
	File      string
	Format    QemuDiskFormat
	IOThread  bool
	ReadOnly  bool
	Replicate bool
	Size      uint
	// TODO custom type
	// Storage is only set for Disk
	Storage string
}

// Maps all the disk related settings
func (qemuDisk) mapToStruct(settings [][]string) *qemuDisk {
	if len(settings) == 0 {
		return nil
	}
	disk := qemuDisk{Backup: true}

	if settings[0][0][0:1] == "/" {
		disk.File = settings[0][0]
	} else {
		// "test2:105/vm-105-disk-53.qcow2,
		disk.Storage = strings.Split(settings[0][0], ":")[0]
		disk.Format = QemuDiskFormat("").mapToStruct(settings[0][0])
	}

	for _, e := range settings {
		if e[0] == "aio" {
			disk.AsyncIO = QemuDiskAsyncIO(e[1])
			continue
		}
		if e[0] == "backup" {
			disk.Backup, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "cache" {
			disk.Cache = QemuDiskCache(e[1])
			continue
		}
		if e[0] == "discard" {
			disk.Discard, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "iops_rd" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.ReadLimit_Iops.Concurrent = uint(tmp)
		}
		if e[0] == "iops_rd_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.ReadLimit_Iops.Burst = uint(tmp)
		}
		if e[0] == "iops_wr" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.WriteLimit_Iops.Concurrent = uint(tmp)
		}
		if e[0] == "iops_wr_max" {
			tmp, _ := strconv.Atoi(e[1])
			disk.Bandwidth.WriteLimit_Iops.Burst = uint(tmp)
		}
		if e[0] == "iothread" {
			disk.IOThread, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "mbps_rd" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.ReadLimit_Data.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_rd_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.ReadLimit_Data.Burst = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.WriteLimit_Data.Concurrent = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "mbps_wr_max" {
			tmp, _ := strconv.ParseFloat(e[1], 32)
			disk.Bandwidth.WriteLimit_Data.Burst = float32(math.Round(tmp*100) / 100)
		}
		if e[0] == "replicate" {
			disk.Replicate, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "ro" {
			disk.ReadOnly, _ = strconv.ParseBool(e[1])
			continue
		}
		if e[0] == "size" {
			diskSize, _ := strconv.Atoi(strings.TrimSuffix(e[1], "G"))
			disk.Size = uint(diskSize)
			continue
		}
		if e[0] == "ssd" {
			disk.EmulateSSD, _ = strconv.ParseBool(e[1])
		}
	}
	return &disk
}

type QemuStorages struct {
	Ide    *QemuIdeDisks
	Sata   *QemuSataDisks
	Scsi   *QemuScsiDisks
	VirtIO *QemuVirtIODisks
}

func (QemuStorages) mapToStruct(params map[string]interface{}) *QemuStorages {
	storage := QemuStorages{
		Ide:    QemuIdeDisks{}.mapToStruct(params),
		Sata:   QemuSataDisks{}.mapToStruct(params),
		Scsi:   QemuScsiDisks{}.mapToStruct(params),
		VirtIO: QemuVirtIODisks{}.mapToStruct(params),
	}
	if storage.Ide != nil || storage.Sata != nil || storage.Scsi != nil || storage.VirtIO != nil {
		return &storage
	}
	return nil
}
