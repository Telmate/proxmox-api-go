package proxmox

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

// matrix of storage types and which content types they support.
var storageContentTypesAPI = []string{"backup", "rootdir", "images", "iso", "snippets", "vztmpl"}
var storageContentTypesStruct = []string{"backup", "container", "diskimage", "iso", "snippets", "template"}
var storageContentTypes = map[string]interface{}{
	"directory":      []bool{true, true, true, true, true, true},
	"lvm":            []bool{false, true, true, false, false, false},
	"lvm-thin":       []bool{false, true, true, false, false, false},
	"nfs":            []bool{true, true, true, true, true, true},
	"smb":            []bool{true, true, true, true, true, true},
	"glusterfs":      []bool{true, false, true, true, true, true},
	"iscsi":          []bool{false, false, true, false, false, false},
	"cephfs":         []bool{true, false, false, true, true, true},
	"rbd":            []bool{false, true, true, false, false, false},
	"zfs-over-iscsi": []bool{false, false, true, false, false, false},
	"zfs":            []bool{false, true, true, false, false, false},
	"pbs":            []bool{true, false, false, false, false, false},
}

type ConfigStorageContent struct {
	Backup    *bool `json:"backup,omitempty"`    //backup
	Iso       *bool `json:"iso,omitempty"`       //iso
	Template  *bool `json:"template,omitempty"`  //vztmpl
	DiskImage *bool `json:"diskimage,omitempty"` //images
	Container *bool `json:"container,omitempty"` //rootdir
	Snippets  *bool `json:"snippets,omitempty"`  //snippets
}

func (c *ConfigStorageContent) MapStorageContent(array []bool) (list string) {
	if c != nil {
		for i, e := range []interface{}{c.Backup, c.Container, c.DiskImage, c.Iso, c.Snippets, c.Template} {
			if e.(*bool) != nil {
				if *e.(*bool) && array[i] {
					list = AddToList(list, storageContentTypesAPI[i])
				}
			}
		}
	}
	if list == "" {
		return "none"
	}
	return
}

func (c *ConfigStorageContent) Validate(storageType string) error {
	// iscsi is the only storage type which can have "none" as content.
	if storageType == "iscsi" {
		return nil
	}
	array := storageContentTypes[storageType].([]bool)
	contentList := c.MapStorageContent(array)
	if contentList != "" {
		return nil
	}
	var list string
	for i, e := range array {
		if e {
			list = AddToList(list, storageContentTypesStruct[i])
		}
	}
	return fmt.Errorf("error at least one of the keys (content:{ %s }) must be true", list)
}

type ConfigStorageBackupRetention struct {
	Last    *int `json:"last,omitempty"`
	Hourly  *int `json:"hourly,omitempty"`
	Daily   *int `json:"daily,omitempty"`
	Monthly *int `json:"monthly,omitempty"`
	Weekly  *int `json:"weekly,omitempty"`
	Yearly  *int `json:"yearly,omitempty"`
}

func (b *ConfigStorageBackupRetention) MapStorageBackupRetention() string {
	if !b.AllNil() {
		return "keep-daily=" + strconv.Itoa(*b.Daily) +
			",keep-hourly=" + strconv.Itoa(*b.Hourly) +
			",keep-last=" + strconv.Itoa(*b.Last) +
			",keep-monthly=" + strconv.Itoa(*b.Monthly) +
			",keep-weekly=" + strconv.Itoa(*b.Weekly) +
			",keep-yearly=" + strconv.Itoa(*b.Yearly)
	}
	return "keep-all=1"
}

func (b *ConfigStorageBackupRetention) Validate() (err error) {
	if b == nil {
		return nil
	}
	if !b.AllNil() {
		text := []string{"last", "hourly", "daily", "weekly", "monthly", "yearly"}
		for i, e := range []*int{b.Last, b.Hourly, b.Daily, b.Weekly, b.Monthly, b.Yearly} {
			if e == nil {
				return ErrorKeyNotSet("backupretention:{ " + text[i] + " }")
			}
			err = ValidateIntGreater(0, *e, "backupretention:{ "+text[i]+" }")
			if err != nil {
				return
			}
		}
	}
	return nil
}

func (b *ConfigStorageBackupRetention) AllNil() bool {
	check := true
	for _, e := range []*int{b.Last, b.Hourly, b.Daily, b.Weekly, b.Monthly, b.Yearly} {
		if e != nil {
			check = false
		}
	}
	return check
}

// Storage Types
type ConfigStorageDirectory struct {
	Path          string  `json:"path"`
	Preallocation *string `json:"preallocation,omitempty"`
	Shared        bool    `json:"shared"`
}

func (directory *ConfigStorageDirectory) SetDefaults() {
	if directory.Preallocation == nil {
		directory.Preallocation = PointerString("metadata")
	}
}

type ConfigStorageLVM struct {
	VGname string `json:"vgname"`
	Shared bool   `json:"shared,omitempty"`
}

type ConfigStorageLVMThin struct {
	VGname   string `json:"vgname"`
	Thinpool string `json:"thinpool"`
}

type ConfigStorageNFS struct {
	Server        string  `json:"server"`
	Export        string  `json:"export"`
	Preallocation *string `json:"preallocation,omitempty"`
	Version       *string `json:"version,omitempty"`
}

func (nfs *ConfigStorageNFS) SetDefaults() {
	if nfs.Preallocation == nil {
		nfs.Preallocation = PointerString("metadata")
	}
}

type ConfigStorageSMB struct {
	Username      string  `json:"username"`
	Share         string  `json:"share"`
	Preallocation *string `json:"preallocation,omitempty"`
	Domain        string  `json:"domain"`
	Server        string  `json:"server"`
	Password      *string `json:"password,omitempty"`
	Version       *string `json:"version,omitempty"`
}

func (smb *ConfigStorageSMB) SetDefaults() {
	if smb.Preallocation == nil {
		smb.Preallocation = PointerString("metadata")
	}
}

type ConfigStorageGlusterFS struct {
	Server1       string  `json:"server1"`
	Server2       string  `json:"server2,omitempty"`
	Preallocation *string `json:"preallocation,omitempty"`
	Volume        string  `json:"volume"`
}

func (glusterfs *ConfigStorageGlusterFS) SetDefaults() {
	if glusterfs.Preallocation == nil {
		glusterfs.Preallocation = PointerString("metadata")
	}
}

type ConfigStorageISCSI struct {
	Portal string `json:"portal"`
	Target string `json:"target"`
}

type ConfigStorageCephFS struct {
	Monitors  []string `json:"monitors"`
	SecretKey *string  `json:"secret-key,omitempty"`
	Username  string   `json:"username"`
	FSname    string   `json:"fs-name"`
}

type ConfigStorageRBD struct {
	Pool      string   `json:"pool"`
	Monitors  []string `json:"monitors"`
	Username  string   `json:"username"`
	Keyring   *string  `json:"keyring,omitempty"`
	Namespace string   `json:"namespace"`
	KRBD      bool     `json:"krbd"`
}

type ConfigStorageZFSoverISCSI struct {
	Portal        string                             `json:"portal"`
	Pool          string                             `json:"pool"`
	Blocksize     *string                            `json:"blocksize"`
	Target        string                             `json:"target"`
	ISCSIprovider string                             `json:"iscsiprovider"`
	Thinprovision bool                               `json:"thinprovision"`
	Comstar       *ConfigStorageZFSoverISCSI_Comstar `json:"comstar,omitempty"`
	Istgt         *ConfigStorageZFSoverISCSI_istgt   `json:"istgt,omitempty"`
	LIO           *ConfigStorageZFSoverISCSI_LIO     `json:"lio,omitempty"`
}

func (zfsoveriscsi *ConfigStorageZFSoverISCSI) SetDefaults() {
	if zfsoveriscsi.Blocksize == nil {
		zfsoveriscsi.Blocksize = PointerString("4k")
	}
}

func (zfsoveriscsi *ConfigStorageZFSoverISCSI) RemapToAPI() {
	if zfsoveriscsi.ISCSIprovider == "lio" {
		zfsoveriscsi.ISCSIprovider = "LIO"
	}
}

func (zfsoveriscsi *ConfigStorageZFSoverISCSI) RemapFromAPI() {
	if zfsoveriscsi.ISCSIprovider == "LIO" {
		zfsoveriscsi.ISCSIprovider = "lio"
	}
}

type ConfigStorageZFSoverISCSI_Comstar struct {
	TargetGroup string `json:"target-group"`
	HostGroup   string `json:"host-group"`
	Writecache  bool   `json:"writecache"`
}
type ConfigStorageZFSoverISCSI_istgt struct {
	Writecache bool `json:"writecache"`
}
type ConfigStorageZFSoverISCSI_LIO struct {
	TargetPortalGroup string `json:"targetportal-group"`
}

type ConfigStorageZFS struct {
	Pool          string  `json:"pool"`
	Blocksize     *string `json:"blocksize,omitempty"`
	Thinprovision bool    `json:"thinprovision,omitempty"`
}

func (zfs *ConfigStorageZFS) SetDefaults() {
	if zfs.Blocksize == nil {
		zfs.Blocksize = PointerString("8k")
	}
}

type ConfigStoragePBS struct {
	Server      string  `json:"server"`
	Datastore   string  `json:"datastore"`
	Username    string  `json:"username"`
	Password    *string `json:"password,omitempty"`
	Fingerprint string  `json:"fingerprint,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Namespace   string  `json:"namespace,omitempty"`
}

func (pbs *ConfigStoragePBS) SetDefaults() {
	if pbs.Port == nil {
		pbs.Port = PointerInt(8007)
	}
}

// Storage options for the Proxmox API
type ConfigStorage struct {
	ID              string                        `json:"id"`
	Enable          bool                          `json:"enable"`
	Nodes           []string                      `json:"nodes,omitempty"`
	Type            string                        `json:"type"`
	Directory       *ConfigStorageDirectory       `json:"directory,omitempty"`
	LVM             *ConfigStorageLVM             `json:"lvm,omitempty"`
	LVMThin         *ConfigStorageLVMThin         `json:"lvm-thin,omitempty"`
	NFS             *ConfigStorageNFS             `json:"nfs,omitempty"`
	SMB             *ConfigStorageSMB             `json:"smb,omitempty"`
	GlusterFS       *ConfigStorageGlusterFS       `json:"glusterfs,omitempty"`
	ISCSI           *ConfigStorageISCSI           `json:"iscsi,omitempty"`
	CephFS          *ConfigStorageCephFS          `json:"cephfs,omitempty"`
	RBD             *ConfigStorageRBD             `json:"rbd,omitempty"`
	ZFSoverISCSI    *ConfigStorageZFSoverISCSI    `json:"zfs-over-iscsi,omitempty"`
	ZFS             *ConfigStorageZFS             `json:"zfs,omitempty"`
	PBS             *ConfigStoragePBS             `json:"pbs,omitempty"`
	Content         *ConfigStorageContent         `json:"content,omitempty"`
	BackupRetention *ConfigStorageBackupRetention `json:"backupretention,omitempty"`
}

func (config *ConfigStorage) SetDefaults() {
	if config.Directory != nil {
		config.Directory.SetDefaults()
	}
	if config.NFS != nil {
		config.NFS.SetDefaults()
	}
	if config.SMB != nil {
		config.SMB.SetDefaults()
	}
	if config.GlusterFS != nil {
		config.GlusterFS.SetDefaults()
	}
	if config.ZFSoverISCSI != nil {
		config.ZFSoverISCSI.SetDefaults()
	}
	if config.ZFS != nil {
		config.ZFS.SetDefaults()
	}
	if config.PBS != nil {
		config.PBS.SetDefaults()
	}
}

func (config *ConfigStorage) RemapToAPI() {
	switch config.Type {
	case "directory":
		config.Type = "dir"
	case "lvm-thin":
		config.Type = "lvmthin"
	case "smb":
		config.Type = "cifs"
	case "zfs-over-iscsi":
		config.Type = "zfs"
	case "zfs":
		config.Type = "zfspool"
	}
}

func (config *ConfigStorage) RemapFromAPI() {
	switch config.Type {
	case "dir":
		config.Type = "directory"
	case "lvmthin":
		config.Type = "lvm-thin"
	case "cifs":
		config.Type = "smb"
	case "zfs":
		config.Type = "zfs-over-iscsi"
	case "zfspool":
		config.Type = "zfs"
	}
}

func (newConfig *ConfigStorage) Validate(id string, create bool, client *Client) (err error) {
	exists, err := client.CheckStorageExistance(id)
	if err != nil {
		return
	}

	if exists && create {
		return ErrorItemExists(id, "storage")
	}
	if !exists && !create {
		return ErrorItemNotExists(id, "storage")
	}

	err = ValidateStringInArray([]string{"directory", "lvm", "lvm-thin", "nfs", "smb", "glusterfs", "iscsi", "cephfs", "rbd", "zfs-over-iscsi", "zfs", "pbs"}, newConfig.Type, "type")
	if err != nil {
		return
	}

	var currentConfig *ConfigStorage
	if exists {
		currentConfig, err = NewConfigStorageFromApi(id, client)
		if err != nil {
			return
		}
		err = ValidateStringsEqual(newConfig.Type, currentConfig.Type, "type")
		if err != nil {
			return
		}
	}

	switch newConfig.Type {
	case "directory":
		if exists && newConfig.Directory != nil {
			err = ValidateStringsEqual(newConfig.Directory.Path, currentConfig.Directory.Path, "path")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.Directory == nil {
				return ErrorKeyEmpty("directory")
			} else {
				err = ValidateFilePath(newConfig.Directory.Path, "path")
				if err != nil {
					return
				}
			}
		}
	case "lvm":
		if exists && newConfig.LVM != nil {
			err = ValidateStringsEqual(newConfig.LVM.VGname, currentConfig.LVM.VGname, "lvm:{ vgname }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.LVM == nil {
				return ErrorKeyEmpty("lvm")
			} else {
				if newConfig.LVM.VGname == "" {
					return ErrorKeyEmpty("lvm:{ vgname }")
				}
			}
		}
	case "lvm-thin":
		if exists && newConfig.LVMThin != nil {
			err = ValidateStringsEqual(newConfig.LVMThin.VGname, currentConfig.LVMThin.VGname, "lvm-thin:{ vgname }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.LVMThin.Thinpool, currentConfig.LVMThin.Thinpool, "lvm-thin:{ thinpool }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.LVMThin == nil {
				return ErrorKeyEmpty("lvm-thin")
			} else {
				if newConfig.LVMThin.VGname == "" {
					return ErrorKeyEmpty("lvm-thin:{ vgname }")
				}
				if newConfig.LVMThin.Thinpool == "" {
					return ErrorKeyEmpty("lvm-thin:{ thinpool }")
				}
			}
		}
	case "nfs":
		if exists && newConfig.NFS != nil {
			err = ValidateStringsEqual(newConfig.NFS.Export, currentConfig.NFS.Export, "nfs:{ export }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.NFS.Server, currentConfig.NFS.Server, "nfs:{ server }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.NFS == nil {
				return ErrorKeyEmpty("nfs")
			} else {
				err = ValidateStringNotEmpty(newConfig.NFS.Server, "nfs:{ server }")
				if err != nil {
					return
				}
				err = ValidateFilePath(newConfig.NFS.Export, "nfs:{ export }")
				if err != nil {
					return
				}
			}
		}
		if newConfig.NFS != nil {
			if newConfig.NFS.Version != nil {
				err = ValidateStringInArray([]string{"3", "4", "4.1", "4.2"}, *newConfig.NFS.Version, "nfs:{ version }")
				if err != nil {
					return
				}
			}
			if newConfig.NFS.Preallocation != nil {
				err = ValidateStringNotEmpty(*newConfig.NFS.Preallocation, "nfs:{ preallocation }")
				if err != nil {
					return
				}
			}
		}
	case "smb":
		if exists && newConfig.SMB != nil {
			err = ValidateStringsEqual(newConfig.SMB.Server, currentConfig.SMB.Server, "smb:{ server }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.SMB.Share, currentConfig.SMB.Share, "smb:{ share }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.SMB == nil {
				return ErrorKeyEmpty("smb")
			} else {
				err = ValidateStringNotEmpty(newConfig.SMB.Server, "smb:{ server }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.SMB.Share, "smb:{ share }")
				if err != nil {
					return
				}
			}
		}
		if newConfig.SMB != nil {
			if newConfig.SMB.Version != nil {
				err = ValidateStringInArray([]string{"2.0", "2.1", "3", "3.0", "3.11"}, *newConfig.SMB.Version, "smb:{ version }")
				if err != nil {
					return
				}
			}
			if newConfig.SMB.Preallocation != nil {
				err = ValidateStringNotEmpty(*newConfig.SMB.Preallocation, "smb:{ preallocation }")
				if err != nil {
					return
				}
			}
		}
	case "glusterfs":
		if exists && newConfig.GlusterFS != nil {
			err = ValidateStringsEqual(newConfig.GlusterFS.Volume, currentConfig.GlusterFS.Volume, "glusterfs:{ volume }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.GlusterFS == nil {
				return ErrorKeyEmpty("glusterfs")
			} else {
				err = ValidateStringNotEmpty(newConfig.GlusterFS.Server1, "glusterfs:{ server1 }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.GlusterFS.Volume, "glusterfs:{ volume }")
				if err != nil {
					return
				}
			}
		}
		if newConfig.GlusterFS != nil {
			err = ValidateStringNotEmpty(newConfig.GlusterFS.Server1, "glusterfs:{ server1 }")
			if err != nil {
				return
			}
			if newConfig.GlusterFS.Preallocation != nil {
				err = ValidateStringNotEmpty(*newConfig.GlusterFS.Preallocation, "glusterfs:{ preallocation }")
				if err != nil {
					return
				}
			}
		}
	case "iscsi":
		if exists && newConfig.ISCSI != nil {
			err = ValidateStringsEqual(newConfig.ISCSI.Portal, currentConfig.ISCSI.Portal, "iscsi:{ portal }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.ISCSI.Target, currentConfig.ISCSI.Target, "iscsi:{ target }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.ISCSI == nil {
				return ErrorKeyEmpty("iscsi")
			} else {
				err = ValidateStringNotEmpty(newConfig.ISCSI.Portal, "iscsi:{ portal }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.ISCSI.Target, "iscsi:{ target }")
				if err != nil {
					return
				}
			}
		}
	case "cephfs":
		if !exists && newConfig.CephFS == nil {
			return ErrorKeyEmpty("cephfs")
		}
		if newConfig.CephFS != nil {
			err = ValidateArrayNotEmpty(newConfig.CephFS.Monitors, "cephfs:{ monitors }")
			if err != nil {
				return
			}
		}
	case "rbd":
		if !exists && newConfig.RBD == nil {
			return ErrorKeyEmpty("rbd")
		}
		if newConfig.RBD != nil {
			err = ValidateArrayNotEmpty(newConfig.RBD.Monitors, "rbd:{ monitors }")
			if err != nil {
				return
			}
		}
	case "zfs-over-iscsi":
		if exists && newConfig.ZFSoverISCSI != nil {
			err = ValidateStringsEqual(newConfig.ZFSoverISCSI.ISCSIprovider, currentConfig.ZFSoverISCSI.ISCSIprovider, "zfs-over-iscsi:{ iscsiprovider }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.ZFSoverISCSI.Portal, currentConfig.ZFSoverISCSI.Portal, "zfs-over-iscsi:{ portal }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.ZFSoverISCSI.Target, currentConfig.ZFSoverISCSI.Target, "zfs-over-iscsi:{ target }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.ZFSoverISCSI.Pool, currentConfig.ZFSoverISCSI.Pool, "zfs-over-iscsi:{ pool }")
			if err != nil {
				return
			}
			// err = ValidateStringsEqual(*newConfig.ZFSoverISCSI.Blocksize, *currentConfig.ZFSoverISCSI.Blocksize, "zfs-over-iscsi:{ blocksize }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.ZFSoverISCSI == nil {
				return ErrorKeyEmpty("zfs-over-iscsi")
			} else {
				err = ValidateStringInArray([]string{"comstar", "istgt", "lio", "iet"}, newConfig.ZFSoverISCSI.ISCSIprovider, "zfs-over-iscsi:{ iscsiprovider }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.ZFSoverISCSI.Portal, "zfs-over-iscsi:{ portal }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.ZFSoverISCSI.Pool, "zfs-over-iscsi:{ pool }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.ZFSoverISCSI.Target, "zfs-over-iscsi:{ target }")
				if err != nil {
					return
				}
			}
		}
		switch newConfig.ZFSoverISCSI.ISCSIprovider {
		case "comstar":
			if exists && newConfig.ZFSoverISCSI.Comstar != nil {
				err = ValidateStringsEqual(newConfig.ZFSoverISCSI.Comstar.HostGroup, currentConfig.ZFSoverISCSI.Comstar.HostGroup, "zfs-over-iscsi:{ comstar:{ host-group } }")
				if err != nil {
					return
				}
				err = ValidateStringsEqual(newConfig.ZFSoverISCSI.Comstar.TargetGroup, currentConfig.ZFSoverISCSI.Comstar.TargetGroup, "zfs-over-iscsi:{ comstar:{ target-group } }")
				if err != nil {
					return
				}
			} else if !exists && newConfig.ZFSoverISCSI.Comstar == nil {
				return ErrorKeyEmpty("zfs-over-iscsi:{ comstar }")
			}
		case "istgt":
			if !exists && newConfig.ZFSoverISCSI.Istgt == nil {
				return ErrorKeyEmpty("zfs-over-iscsi:{ istgt }")
			}
		case "lio":
			if !exists && newConfig.ZFSoverISCSI.LIO == nil {
				return ErrorKeyEmpty("zfs-over-iscsi:{ lio }")
			} else {
				err = ValidateStringNotEmpty(newConfig.ZFSoverISCSI.LIO.TargetPortalGroup, "zfs-over-iscsi:{ lio:{ targetportal-group } }")
				if err != nil {
					return
				}
			}
		}
	case "zfs":
		if exists && newConfig.ZFS != nil {
			err = ValidateStringsEqual(newConfig.ZFS.Pool, currentConfig.ZFS.Pool, "zfs:{ pool }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.ZFS == nil {
				return ErrorKeyEmpty("zfs")
			} else {
				err = ValidateStringNotEmpty(newConfig.ZFS.Pool, "zfs:{ pool }")
				if err != nil {
					return
				}
			}
		}
		if newConfig.ZFS != nil {
			if newConfig.ZFS.Blocksize != nil {
				err = ValidateStringNotEmpty(*newConfig.ZFS.Blocksize, "zfs:{ blocksize }")
				if err != nil {
					return
				}
			}
		}
	case "pbs":
		if exists && newConfig.PBS != nil {
			err = ValidateStringsEqual(newConfig.PBS.Server, currentConfig.PBS.Server, "pbs:{ server }")
			if err != nil {
				return
			}
			err = ValidateStringsEqual(newConfig.PBS.Datastore, currentConfig.PBS.Datastore, "pbs:{ datastore }")
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.PBS == nil {
				return ErrorKeyEmpty("pbs")
			} else {
				err = ValidateStringNotEmpty(newConfig.PBS.Server, "pbs:{ server }")
				if err != nil {
					return
				}
				err = ValidateStringNotEmpty(newConfig.PBS.Datastore, "pbs:{ datastore }")
				if err != nil {
					return
				}
				if newConfig.PBS.Password == nil {
					return ErrorKeyNotSet("pbs:{ password }")
				}
			}
		}
		if newConfig.PBS != nil {
			if newConfig.PBS.Port != nil {
				err = ValidateIntInRange(1, 65536, *newConfig.PBS.Port, "pbs:{ port }")
				if err != nil {
					return
				}
			}
			err = ValidateStringNotEmpty(newConfig.PBS.Username, "pbs:{ username }")
			if err != nil {
				return
			}
		}
	}
	if !inArray([]string{"pbs", "zfs-over-iscsi"}, newConfig.Type) {
		// pbs has a hardcoded content type, of type backup.
		// zfs-over-iscsi has a hardcoded content type, of type diskimage.
		if exists && newConfig.Content != nil {
			err = newConfig.Content.Validate(newConfig.Type)
			if err != nil {
				return
			}
		} else if !exists {
			if newConfig.Content == nil {
				return ErrorKeyEmpty("content")
			} else {
				err = newConfig.Content.Validate(newConfig.Type)
				if err != nil {
					return
				}
			}
		}
	}
	return newConfig.BackupRetention.Validate()
}

func (config *ConfigStorage) mapToApiValues(create bool) (params map[string]interface{}) {
	var deletions string
	params = map[string]interface{}{
		"storage": config.ID,
		"disable": BoolInvert(config.Enable),
		"nodes":   ArrayToCSV(config.Nodes),
	}

	switch config.Type {
	case "directory":
		if config.Directory != nil {
			config.Directory.SetDefaults()
			params["shared"] = config.Directory.Shared
			params["preallocation"] = *config.Directory.Preallocation
			if create {
				params["path"] = config.Directory.Path
			}
		}
	case "lvm":
		if config.LVM != nil {
			params["shared"] = config.LVM.Shared
			if create {
				params["vgname"] = config.LVM.VGname
			}
		}
	case "lvm-thin":
		if config.LVMThin != nil {
			if create {
				params["thinpool"] = config.LVMThin.Thinpool
				params["vgname"] = config.LVMThin.VGname
			}
		}
	case "nfs":
		if config.NFS != nil {
			config.NFS.SetDefaults()
			if config.NFS.Version != nil {
				params["options"] = "vers=" + *config.NFS.Version
			} else {
				deletions = AddToList(deletions, "options")
			}
			if create {
				params["server"] = config.NFS.Server
				params["export"] = config.NFS.Export
			}
			params["preallocation"] = *config.NFS.Preallocation
		}
	case "smb":
		if config.SMB != nil {
			config.SMB.SetDefaults()
			params["domain"] = config.SMB.Domain
			params["username"] = config.SMB.Username
			if create {
				params["share"] = config.SMB.Share
				params["server"] = config.SMB.Server
			}
			if config.SMB.Password != nil {
				params["password"] = *config.SMB.Password
			}
			if config.SMB.Version != nil {
				params["smbversion"] = *config.SMB.Version
			} else {
				deletions = AddToList(deletions, "smbversion")
			}
			params["preallocation"] = *config.SMB.Preallocation
		}
	case "glusterfs":
		if config.GlusterFS != nil {
			config.GlusterFS.SetDefaults()
			params["server"] = config.GlusterFS.Server1
			if config.GlusterFS.Server2 != "" {
				params["server2"] = config.GlusterFS.Server2
			} else if !create {
				deletions = AddToList(deletions, "server2")
			}
			if create {
				params["volume"] = config.GlusterFS.Volume
			}
			params["preallocation"] = *config.GlusterFS.Preallocation
		}
	case "iscsi":
		if create {
			params["portal"] = config.ISCSI.Portal
			params["target"] = config.ISCSI.Target
		}
	case "cephfs":
		if config.CephFS != nil {
			params["monhost"] = ArrayToCSV(config.CephFS.Monitors)
			params["fs-name"] = config.CephFS.FSname
			params["username"] = config.CephFS.Username
			if config.CephFS.SecretKey != nil {
				// not sure if this is the right api parameter
				params["keyring"] = *config.CephFS.SecretKey
			}
		}
	case "rbd":
		if config.RBD != nil {
			params["krbd"] = config.RBD.KRBD
			params["monhost"] = ArrayToCSV(config.RBD.Monitors)
			params["pool"] = config.RBD.Pool
			params["namespace"] = config.RBD.Namespace
			params["username"] = config.RBD.Username
			if config.RBD.Keyring != nil {
				params["keyring"] = *config.RBD.Keyring
			}
		}
	case "zfs-over-iscsi":
		if config.ZFSoverISCSI != nil {
			config.ZFSoverISCSI.SetDefaults()
			params["sparse"] = config.ZFSoverISCSI.Thinprovision
			switch config.ZFSoverISCSI.ISCSIprovider {
			case "comstar":
				if config.ZFSoverISCSI.Comstar != nil {
					params["nowritecache"] = BoolInvert(config.ZFSoverISCSI.Comstar.Writecache)
					if create {
						params["comstar_hg"] = config.ZFSoverISCSI.Comstar.HostGroup
						params["comstar_tg"] = config.ZFSoverISCSI.Comstar.TargetGroup
					}
				}
			case "istgt":
				if config.ZFSoverISCSI.Istgt != nil {
					params["nowritecache"] = BoolInvert(config.ZFSoverISCSI.Istgt.Writecache)
				}
			case "lio":
				if config.ZFSoverISCSI.LIO != nil {
					params["lio_tpg"] = config.ZFSoverISCSI.LIO.TargetPortalGroup
				}
			}
			config.ZFSoverISCSI.RemapToAPI()
			if create {
				params["iscsiprovider"] = config.ZFSoverISCSI.ISCSIprovider
				params["portal"] = config.ZFSoverISCSI.Portal
				params["target"] = config.ZFSoverISCSI.Target
				params["pool"] = config.ZFSoverISCSI.Pool
				params["blocksize"] = *config.ZFSoverISCSI.Blocksize
			}
		}
		config.Content = &ConfigStorageContent{
			DiskImage: PointerBool(true),
		}
	case "zfs":
		if config.ZFS != nil {
			config.ZFS.SetDefaults()
			params["sparse"] = config.ZFS.Thinprovision
			params["blocksize"] = *config.ZFS.Blocksize
			if create {
				params["pool"] = config.ZFS.Pool
			}
		}
	case "pbs":
		if config.PBS != nil {
			config.PBS.SetDefaults()
			params["username"] = config.PBS.Username
			if config.PBS.Fingerprint != "" {
				params["fingerprint"] = config.PBS.Fingerprint
			} else {
				deletions = AddToList(deletions, "fingerprint")
			}
			if config.PBS.Port != nil {
				params["port"] = *config.PBS.Port
			}
			if create {
				params["server"] = config.PBS.Server
				params["datastore"] = config.PBS.Datastore
			}
			if config.PBS.Password != nil {
				params["password"] = *config.PBS.Password
			}
			if config.PBS.Namespace != "" {
				params["namespace"] = strings.TrimLeft(config.PBS.Namespace, "/")
			}
		}
		config.Content = &ConfigStorageContent{
			Backup: PointerBool(true),
		}
	}

	params["content"] = config.Content.MapStorageContent(storageContentTypes[config.Type].([]bool))

	if config.BackupRetention != nil {
		if storageContentTypes[config.Type].([]bool)[0] {
			params["prune-backups"] = config.BackupRetention.MapStorageBackupRetention()
		}
	}

	if create {
		config.RemapToAPI()
		params["type"] = config.Type
	} else if deletions != "" {
		params["delete"] = deletions
	}
	return
}

func (config *ConfigStorage) CreateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, true, client)
	if err != nil {
		return
	}
	return config.Create(id, true, client)
}

func (config *ConfigStorage) Create(id string, errorSupression bool, client *Client) (err error) {
	var enableStorage bool
	if errorSupression && config.Enable {
		config.Enable = false
		enableStorage = true
	}
	config.ID = id
	params := config.mapToApiValues(true)
	err = client.CreateStorage(params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Storage Backend: %v, (params: %v)", err, string(params))
	}
	// if it gets enabled after it has been configured proxmox wont give the error that it can't connect to the storage backend
	if enableStorage {
		err = client.EnableStorage(id)
	}
	return
}

func (config *ConfigStorage) UpdateWithValidate(id string, client *Client) (err error) {
	err = config.Validate(id, false, client)
	if err != nil {
		return
	}
	return config.Update(id, client)
}

func (config *ConfigStorage) Update(id string, client *Client) (err error) {
	config.ID = id
	params := config.mapToApiValues(false)
	err = client.UpdateStorage(id, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Storage Backend: %v, (params: %v)", err, string(params))
	}
	return
}

func NewConfigStorageFromApi(storageid string, client *Client) (config *ConfigStorage, err error) {
	// prepare json map to receive the information from the api
	var rawConfig map[string]interface{}
	rawConfig, err = client.GetStorageConfig(storageid)
	if err != nil {
		return nil, err
	}

	config = new(ConfigStorage)

	config.ID = storageid
	config.Type = rawConfig["type"].(string)

	if _, isSet := rawConfig["nodes"]; isSet {
		config.Nodes = CSVtoArray(rawConfig["nodes"].(string))
	}

	config.RemapFromAPI()

	if _, isSet := rawConfig["disable"]; isSet {
		config.Enable = BoolInvert(Itob(int(rawConfig["disable"].(float64))))
	} else {
		config.Enable = true
	}

	switch config.Type {
	case "directory":
		config.Directory = new(ConfigStorageDirectory)
		config.Directory.Path = rawConfig["path"].(string)
		config.Directory.Shared = Itob(int(rawConfig["shared"].(float64)))
		if _, isSet := rawConfig["preallocation"]; isSet {
			config.Directory.Preallocation = PointerString(rawConfig["preallocation"].(string))
		}
	case "lvm":
		config.LVM = new(ConfigStorageLVM)
		config.LVM.VGname = rawConfig["vgname"].(string)
		config.LVM.Shared = Itob(int(rawConfig["shared"].(float64)))
	case "lvm-thin":
		config.LVMThin = new(ConfigStorageLVMThin)
		config.LVMThin.Thinpool = rawConfig["thinpool"].(string)
		config.LVMThin.VGname = rawConfig["vgname"].(string)
	case "nfs":
		config.NFS = new(ConfigStorageNFS)
		config.NFS.Server = rawConfig["server"].(string)
		config.NFS.Export = rawConfig["export"].(string)
		if _, isSet := rawConfig["options"]; isSet {
			version := strings.Split(rawConfig["options"].(string), "=")
			config.NFS.Version = PointerString(version[1])
		}
		if _, isSet := rawConfig["preallocation"]; isSet {
			config.NFS.Preallocation = PointerString(rawConfig["preallocation"].(string))
		}
	case "smb":
		config.SMB = new(ConfigStorageSMB)
		config.SMB.Server = rawConfig["server"].(string)
		config.SMB.Share = rawConfig["share"].(string)
		if _, isSet := rawConfig["smbversion"]; isSet {
			smbVersion := rawConfig["smbversion"].(string)
			if smbVersion == "default" {
				config.SMB.Version = nil
			} else {
				config.SMB.Version = PointerString(smbVersion)
			}
		}
		if _, isSet := rawConfig["domain"]; isSet {
			config.SMB.Domain = rawConfig["domain"].(string)
		}
		if _, isSet := rawConfig["username"]; isSet {
			config.SMB.Username = rawConfig["username"].(string)
		}
		if _, isSet := rawConfig["preallocation"]; isSet {
			config.SMB.Preallocation = PointerString(rawConfig["preallocation"].(string))
		}
	case "glusterfs":
		config.GlusterFS = new(ConfigStorageGlusterFS)
		config.GlusterFS.Server1 = rawConfig["server"].(string)
		config.GlusterFS.Volume = rawConfig["volume"].(string)
		if _, isSet := rawConfig["server2"]; isSet {
			config.GlusterFS.Server2 = rawConfig["server2"].(string)
		}
		if _, isSet := rawConfig["preallocation"]; isSet {
			config.GlusterFS.Preallocation = PointerString(rawConfig["preallocation"].(string))
		}
	case "iscsi":
		config.ISCSI = new(ConfigStorageISCSI)
		config.ISCSI.Portal = rawConfig["portal"].(string)
		config.ISCSI.Target = rawConfig["target"].(string)
	case "cephfs":
		config.CephFS = new(ConfigStorageCephFS)
		config.CephFS.Monitors = CSVtoArray(rawConfig["monhost"].(string))
		if _, isSet := rawConfig["fs-name"]; isSet {
			config.CephFS.FSname = rawConfig["fs-name"].(string)
		}
		if _, isSet := rawConfig["username"]; isSet {
			config.CephFS.Username = rawConfig["username"].(string)
		}
	case "rbd":
		config.RBD = new(ConfigStorageRBD)
		config.RBD.KRBD = Itob(int(rawConfig["krbd"].(float64)))
		config.RBD.Monitors = CSVtoArray(rawConfig["monhost"].(string))
		config.RBD.Pool = rawConfig["pool"].(string)
		if _, isSet := rawConfig["namespace"]; isSet {
			config.RBD.Namespace = rawConfig["namespace"].(string)
		}
		if _, isSet := rawConfig["username"]; isSet {
			config.RBD.Username = rawConfig["username"].(string)
		}
	case "zfs-over-iscsi":
		config.ZFSoverISCSI = new(ConfigStorageZFSoverISCSI)
		config.ZFSoverISCSI.Blocksize = PointerString(rawConfig["blocksize"].(string))
		config.ZFSoverISCSI.ISCSIprovider = rawConfig["iscsiprovider"].(string)
		config.ZFSoverISCSI.RemapFromAPI()
		switch config.ZFSoverISCSI.ISCSIprovider {
		case "comstar":
			config.ZFSoverISCSI.Comstar = new(ConfigStorageZFSoverISCSI_Comstar)
			if _, isSet := rawConfig["comstar_hg"]; isSet {
				config.ZFSoverISCSI.Comstar.Writecache = BoolInvert(Itob(int(rawConfig["nowritecache"].(float64))))
			} else {
				config.ZFSoverISCSI.Comstar.Writecache = true
			}
			if _, isSet := rawConfig["comstar_hg"]; isSet {
				config.ZFSoverISCSI.Comstar.HostGroup = rawConfig["comstar_hg"].(string)
			}
			if _, isSet := rawConfig["comstar_tg"]; isSet {
				config.ZFSoverISCSI.Comstar.TargetGroup = rawConfig["comstar_tg"].(string)
			}
		case "istgt":
			config.ZFSoverISCSI.Istgt = new(ConfigStorageZFSoverISCSI_istgt)
			config.ZFSoverISCSI.Istgt.Writecache = BoolInvert(Itob(int(rawConfig["nowritecache"].(float64))))
		case "lio":
			config.ZFSoverISCSI.LIO = new(ConfigStorageZFSoverISCSI_LIO)
			config.ZFSoverISCSI.LIO.TargetPortalGroup = rawConfig["lio_tpg"].(string)
		}
		config.ZFSoverISCSI.Pool = rawConfig["pool"].(string)
		config.ZFSoverISCSI.Portal = rawConfig["portal"].(string)
		config.ZFSoverISCSI.Target = rawConfig["target"].(string)
		config.ZFSoverISCSI.Thinprovision = Itob(int(rawConfig["sparse"].(float64)))
	case "zfs":
		config.ZFS = new(ConfigStorageZFS)
		config.ZFS.Pool = rawConfig["pool"].(string)
		config.ZFS.Thinprovision = Itob(int(rawConfig["sparse"].(float64)))
		if _, isSet := rawConfig["blocksize"]; isSet {
			config.ZFS.Blocksize = PointerString(rawConfig["blocksize"].(string))
		}
	case "pbs":
		config.PBS = new(ConfigStoragePBS)
		config.PBS.Datastore = rawConfig["datastore"].(string)
		config.PBS.Server = rawConfig["server"].(string)
		config.PBS.Username = rawConfig["username"].(string)
		if _, isSet := rawConfig["port"]; isSet {
			config.PBS.Port = PointerInt(int(rawConfig["port"].(float64)))
		}
		if _, isSet := rawConfig["fingerprint"]; isSet {
			config.PBS.Fingerprint = rawConfig["fingerprint"].(string)
		}
		if _, isSet := rawConfig["namespace"]; isSet {
			config.PBS.Namespace = rawConfig["namespace"].(string)
		}
	}
	config.SetDefaults()
	if _, isSet := rawConfig["content"]; isSet {
		content := rawConfig["content"].(string)
		if content != "none" {
			contentArray := CSVtoArray(content)
			config.Content = new(ConfigStorageContent)
			if storageContentTypes[config.Type].([]bool)[0] {
				config.Content.Backup = PointerBool(inArray(contentArray, storageContentTypesAPI[0]))
			}
			if storageContentTypes[config.Type].([]bool)[1] {
				config.Content.Container = PointerBool(inArray(contentArray, storageContentTypesAPI[1]))
			}
			if storageContentTypes[config.Type].([]bool)[2] {
				config.Content.DiskImage = PointerBool(inArray(contentArray, storageContentTypesAPI[2]))
			}
			if storageContentTypes[config.Type].([]bool)[3] {
				config.Content.Iso = PointerBool(inArray(contentArray, storageContentTypesAPI[3]))
			}
			if storageContentTypes[config.Type].([]bool)[4] {
				config.Content.Snippets = PointerBool(inArray(contentArray, storageContentTypesAPI[4]))
			}
			if storageContentTypes[config.Type].([]bool)[5] {
				config.Content.Template = PointerBool(inArray(contentArray, storageContentTypesAPI[5]))
			}
		} else {
			// Edge cases
			if config.Type == "iscsi" {
				config.Content = new(ConfigStorageContent)
				config.Content.DiskImage = PointerBool(false)
			}
		}
	}
	if _, isSet := rawConfig["prune-backups"]; isSet {
		prune := CSVtoArray(rawConfig["prune-backups"].(string))
		if !inArray(prune, "keep-all=1") {
			retentionSettings := make(map[string]int)
			for _, e := range prune {
				a := strings.Split(e, "=")
				retentionSettings[a[0]], _ = strconv.Atoi(a[1])
			}
			config.BackupRetention = new(ConfigStorageBackupRetention)
			config.BackupRetention.Daily = PointerInt(retentionSettings["keep-daily"])
			config.BackupRetention.Hourly = PointerInt(retentionSettings["keep-hourly"])
			config.BackupRetention.Last = PointerInt(retentionSettings["keep-last"])
			config.BackupRetention.Monthly = PointerInt(retentionSettings["keep-monthly"])
			config.BackupRetention.Weekly = PointerInt(retentionSettings["keep-weekly"])
			config.BackupRetention.Yearly = PointerInt(retentionSettings["keep-yearly"])
		}
	}
	return
}

func NewConfigStorageFromJson(input []byte) (config *ConfigStorage, err error) {
	config = &ConfigStorage{}
	err = json.Unmarshal([]byte(input), config)
	if err != nil {
		log.Fatal(err)
	}
	config.SetDefaults()
	return
}
