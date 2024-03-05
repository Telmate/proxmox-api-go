package proxmox

import (
	"errors"
	"strings"
)

const (
	PermissionErrorPrefix string = "permission error:" // Check if an error starts with this to see if it's a permission error.
)

const (
	key_Privileges_DatastoreAllocate         string = "Datastore.Allocate"
	key_Privileges_DatastoreAllocateSpace    string = "Datastore.AllocateSpace"
	key_Privileges_DatastoreAllocateTemplate string = "Datastore.AllocateTemplate"
	key_Privileges_DatastoreAudit            string = "Datastore.Audit"
	key_Privileges_GroupAllocate             string = "Group.Allocate"
	key_Privileges_PermissionsModify         string = "Permissions.Modify"
	key_Privileges_PoolAllocate              string = "Pool.Allocate"
	key_Privileges_PoolAudit                 string = "Pool.Audit"
	key_Privileges_RealmAllocate             string = "Realm.Allocate"
	key_Privileges_RealmAllocateUser         string = "Realm.AllocateUser"
	key_Privileges_SDNAllocate               string = "SDN.Allocate"
	key_Privileges_SDNAudit                  string = "SDN.Audit"
	key_Privileges_SysAudit                  string = "Sys.Audit"
	key_Privileges_SysConsole                string = "Sys.Console"
	key_Privileges_SysIncoming               string = "Sys.Incoming"
	key_Privileges_SysModify                 string = "Sys.Modify"
	key_Privileges_SysPowerMgmt              string = "Sys.PowerMgmt"
	key_Privileges_SysSyslog                 string = "Sys.Syslog"
	key_Privileges_UserModify                string = "User.Modify"
	key_Privileges_VMAllocate                string = "VM.Allocate"
	key_Privileges_VMAudit                   string = "VM.Audit"
	key_Privileges_VMBackup                  string = "VM.Backup"
	key_Privileges_VMClone                   string = "VM.Clone"
	key_Privileges_VMConfigCDROM             string = "VM.Config.CDROM"
	key_Privileges_VMConfigCPU               string = "VM.Config.CPU"
	key_Privileges_VMConfigCloudinit         string = "VM.Config.Cloudinit"
	key_Privileges_VMConfigDisk              string = "VM.Config.Disk"
	key_Privileges_VMConfigHWType            string = "VM.Config.HWType"
	key_Privileges_VMConfigMemory            string = "VM.Config.Memory"
	key_Privileges_VMConfigNetwork           string = "VM.Config.Network"
	key_Privileges_VMConfigOptions           string = "VM.Config.Options"
	key_Privileges_VMConsole                 string = "VM.Console"
	key_Privileges_VMMigrate                 string = "VM.Migrate"
	key_Privileges_VMMonitor                 string = "VM.Monitor"
	key_Privileges_VMPowerMgmt               string = "VM.PowerMgmt"
	key_Privileges_VMSnapshot                string = "VM.Snapshot"
	key_Privileges_VMSnapshotRollback        string = "VM.Snapshot.Rollback"
)

type Permission struct {
	Category   PermissionCategory
	Item       PermissionItem
	Privileges Privileges
}

// build a list of unique paths from the permissions.
func (Permission) buildPathList(perms []Permission) (paths []permissionPath) {
	paths = make([]permissionPath, 0)
	for _, perm := range perms {
		categoryPath := perm.Category.path()
		var skipPath bool
		for _, path := range paths {
			if path == categoryPath {
				skipPath = true
				break
			}
		}
		if !skipPath {
			paths = append(paths, categoryPath)
		}
		if perm.Item != permissionItemEmpty {
			paths = append(paths, categoryPath.append(perm.Item))
		}
	}
	return
}

// checks if the permissions map contains the required permissions.
func (p Permission) check(permissions map[permissionPath]privileges) error {
	categoryPath := p.Category.path()
	if p.Item != permissionItemEmpty {
		if v, ok := permissions[categoryPath.append(p.Item)]; ok {
			if !v.includes(p.Privileges, privilegeTrue) {
				return Permission{Category: p.Category, Item: p.Item, Privileges: p.Privileges}.error()
			}
		}
	}
	if v, ok := permissions[categoryPath]; ok {
		if !v.includes(p.Privileges, privilegePropagate) {
			return Permission{Category: p.Category, Item: p.Item, Privileges: p.Privileges}.error()
		}
	}
	return nil
}

func (p Permission) error() error {
	return errors.New(PermissionErrorPrefix + " the following privileges (" + p.Privileges.String() + ") are missing from path (" + string(p.Category.path().append(p.Item)) + ")")
}

func (p Permission) Validate() error {
	return p.Category.Validate()
}

type PermissionCategory string // Enum

const (
	PermissionCategory_Root        PermissionCategory = "root"
	permissionCategory_RootPath    PermissionCategory = "/"
	PermissionCategory_Access      PermissionCategory = "access"
	permissionCategory_AccessPath  PermissionCategory = "/access"
	PermissionCategory_Group       PermissionCategory = "group"
	permissionCategory_GroupPath   PermissionCategory = "/access/groups"
	PermissionCategory_Realm       PermissionCategory = "realm"
	permissionCategory_RealmPath   PermissionCategory = "/access/realm"
	PermissionCategory_Node        PermissionCategory = "node"
	permissionCategory_NodePath    PermissionCategory = "/nodes"
	PermissionCategory_Guest       PermissionCategory = "guest"
	permissionCategory_GuestPath   PermissionCategory = "/vms"
	PermissionCategory_Pool        PermissionCategory = "pool"
	permissionCategory_PoolPath    PermissionCategory = "/pool"
	PermissionCategory_Storage     PermissionCategory = "storage"
	permissionCategory_StoragePath PermissionCategory = "/storage"
	PermissionCategory_Zone        PermissionCategory = "zone"
	permissionCategory_ZonePath    PermissionCategory = "/sdn/zones"
)

func (PermissionCategory) Error() error {
	return errors.New("permission category should be one of (" + strings.Join(arrayToStringArray(PermissionCategory("").enumArray()), ",") + ")")
}

func (PermissionCategory) enumArray() []PermissionCategory {
	return []PermissionCategory{PermissionCategory_Root, PermissionCategory_Access, PermissionCategory_Group, PermissionCategory_Realm, PermissionCategory_Node, PermissionCategory_Guest, PermissionCategory_Pool, PermissionCategory_Storage, PermissionCategory_Zone}
}

// returns the path for the category.
// a raw path may be provided, in which case it will be returned as is.
func (c PermissionCategory) path() permissionPath {
	if len(c) > 0 && c[0] == '/' {
		return permissionPath(c)
	}
	switch c {
	case PermissionCategory_Access:
		return "/access"
	case PermissionCategory_Group:
		return "/access/groups"
	case PermissionCategory_Guest:
		return "/vms"
	case PermissionCategory_Node:
		return "/nodes"
	case PermissionCategory_Pool:
		return "/pool"
	case PermissionCategory_Realm:
		return "/access/realm"
	case PermissionCategory_Root:
		return "/"
	case PermissionCategory_Storage:
		return "/storage"
	case PermissionCategory_Zone:
		return "/sdn/zones"
	}
	return ""
}

func (c PermissionCategory) String() string {
	return string(c)
}

func (c PermissionCategory) Validate() error {
	for _, e := range c.enumArray() {
		if c == e {
			return nil
		}
	}
	return PermissionCategory("").Error()
}

type PermissionItem string

const (
	permissionItemEmpty PermissionItem = ""
)

type permissionPath string

func (p permissionPath) append(item PermissionItem) permissionPath {
	if item == "" {
		return p
	}
	if p == "/" {
		return p + permissionPath(item)
	}
	return p + "/" + permissionPath(item)
}

func (permissionPath) mapToSDK(params map[string]interface{}) map[permissionPath]privileges {
	permissions := make(map[permissionPath]privileges)
	for key, e := range params {
		permissions[permissionPath(key)] = privileges{}.mapToSDK(e.(map[string]interface{}))
	}
	return permissions
}

type privilege int8 // Enum

const (
	privilegeFalse     privilege = 0
	privilegeTrue      privilege = 1
	privilegePropagate privilege = 2
)

func (privilege) extract(i interface{}) privilege {
	number, isFloat64 := i.(float64)
	if !isFloat64 {
		return privilegeFalse
	}
	if int(number) == 1 {
		return privilegePropagate
	}
	return privilegeTrue
}

type Privileges struct {
	DatastoreAllocate         bool `json:"Datastore.Allocate,omitempty"`
	DatastoreAllocateSpace    bool `json:"Datastore.AllocateSpace,omitempty"`
	DatastoreAllocateTemplate bool `json:"Datastore.AllocateTemplate,omitempty"`
	DatastoreAudit            bool `json:"Datastore.Audit,omitempty"`
	GroupAllocate             bool `json:"Group.Allocate,omitempty"`
	PermissionsModify         bool `json:"Permissions.Modify,omitempty"`
	PoolAllocate              bool `json:"Pool.Allocate,omitempty"`
	PoolAudit                 bool `json:"Pool.Audit,omitempty"`
	RealmAllocate             bool `json:"Realm.Allocate,omitempty"`
	RealmAllocateUser         bool `json:"Realm.AllocateUser,omitempty"`
	SDNAllocate               bool `json:"SDN.Allocate,omitempty"`
	SDNAudit                  bool `json:"SDN.Audit,omitempty"`
	SysAudit                  bool `json:"Sys.Audit,omitempty"`
	SysConsole                bool `json:"Sys.Console,omitempty"`
	SysIncoming               bool `json:"Sys.Incoming,omitempty"`
	SysModify                 bool `json:"Sys.Modify,omitempty"`
	SysPowerMgmt              bool `json:"Sys.PowerMgmt,omitempty"`
	SysSyslog                 bool `json:"Sys.Syslog,omitempty"`
	UserModify                bool `json:"User.Modify,omitempty"`
	VMAllocate                bool `json:"VM.Allocate,omitempty"`
	VMAudit                   bool `json:"VM.Audit,omitempty"`
	VMBackup                  bool `json:"VM.Backup,omitempty"`
	VMClone                   bool `json:"VM.Clone,omitempty"`
	VMConfigCDROM             bool `json:"VM.Config.CDROM,omitempty"`
	VMConfigCPU               bool `json:"VM.Config.CPU,omitempty"`
	VMConfigCloudinit         bool `json:"VM.Config.Cloudinit,omitempty"`
	VMConfigDisk              bool `json:"VM.Config.Disk,omitempty"`
	VMConfigHWType            bool `json:"VM.Config.HWType,omitempty"`
	VMConfigMemory            bool `json:"VM.Config.Memory,omitempty"`
	VMConfigNetwork           bool `json:"VM.Config.Network,omitempty"`
	VMConfigOptions           bool `json:"VM.Config.Options,omitempty"`
	VMConsole                 bool `json:"VM.Console,omitempty"`
	VMMigrate                 bool `json:"VM.Migrate,omitempty"`
	VMMonitor                 bool `json:"VM.Monitor,omitempty"`
	VMPowerMgmt               bool `json:"VM.PowerMgmt,omitempty"`
	VMSnapshot                bool `json:"VM.Snapshot,omitempty"`
	VMSnapshotRollback        bool `json:"VM.Snapshot.Rollback,omitempty"`
}

func (p Privileges) String() (privileges string) {
	if p.DatastoreAllocate {
		privileges += key_Privileges_DatastoreAllocate + ", "
	}
	if p.DatastoreAllocateSpace {
		privileges += key_Privileges_DatastoreAllocateSpace + ", "
	}
	if p.DatastoreAllocateTemplate {
		privileges += key_Privileges_DatastoreAllocateTemplate + ", "
	}
	if p.DatastoreAudit {
		privileges += key_Privileges_DatastoreAudit + ", "
	}
	if p.GroupAllocate {
		privileges += key_Privileges_GroupAllocate + ", "
	}
	if p.PermissionsModify {
		privileges += key_Privileges_PermissionsModify + ", "
	}
	if p.PoolAllocate {
		privileges += key_Privileges_PoolAllocate + ", "
	}
	if p.PoolAudit {
		privileges += key_Privileges_PoolAudit + ", "
	}
	if p.RealmAllocate {
		privileges += key_Privileges_RealmAllocate + ", "
	}
	if p.RealmAllocateUser {
		privileges += key_Privileges_RealmAllocateUser + ", "
	}
	if p.SDNAllocate {
		privileges += key_Privileges_SDNAllocate + ", "
	}
	if p.SDNAudit {
		privileges += key_Privileges_SDNAudit + ", "
	}
	if p.SysAudit {
		privileges += key_Privileges_SysAudit + ", "
	}
	if p.SysConsole {
		privileges += key_Privileges_SysConsole + ", "
	}
	if p.SysIncoming {
		privileges += key_Privileges_SysIncoming + ", "
	}
	if p.SysModify {
		privileges += key_Privileges_SysModify + ", "
	}
	if p.SysPowerMgmt {
		privileges += key_Privileges_SysPowerMgmt + ", "
	}
	if p.SysSyslog {
		privileges += key_Privileges_SysSyslog + ", "
	}
	if p.UserModify {
		privileges += key_Privileges_UserModify + ", "
	}
	if p.VMAllocate {
		privileges += key_Privileges_VMAllocate + ", "
	}
	if p.VMAudit {
		privileges += key_Privileges_VMAudit + ", "
	}
	if p.VMBackup {
		privileges += key_Privileges_VMBackup + ", "
	}
	if p.VMClone {
		privileges += key_Privileges_VMClone + ", "
	}
	if p.VMConfigCDROM {
		privileges += key_Privileges_VMConfigCDROM + ", "
	}
	if p.VMConfigCPU {
		privileges += key_Privileges_VMConfigCPU + ", "
	}
	if p.VMConfigCloudinit {
		privileges += key_Privileges_VMConfigCloudinit + ", "
	}
	if p.VMConfigDisk {
		privileges += key_Privileges_VMConfigDisk + ", "
	}
	if p.VMConfigHWType {
		privileges += key_Privileges_VMConfigHWType + ", "
	}
	if p.VMConfigMemory {
		privileges += key_Privileges_VMConfigMemory + ", "
	}
	if p.VMConfigNetwork {
		privileges += key_Privileges_VMConfigNetwork + ", "
	}
	if p.VMConfigOptions {
		privileges += key_Privileges_VMConfigOptions + ", "
	}
	if p.VMConsole {
		privileges += key_Privileges_VMConsole + ", "
	}
	if p.VMMigrate {
		privileges += key_Privileges_VMMigrate + ", "
	}
	if p.VMMonitor {
		privileges += key_Privileges_VMMonitor + ", "
	}
	if p.VMPowerMgmt {
		privileges += key_Privileges_VMPowerMgmt + ", "
	}
	if p.VMSnapshot {
		privileges += key_Privileges_VMSnapshot + ", "
	}
	if p.VMSnapshotRollback {
		privileges += key_Privileges_VMSnapshotRollback + ", "
	}
	if privileges != "" {
		privileges = privileges[:len(privileges)-2]
	}
	return
}

// internal struct to map the privileges to the SDK.
type privileges struct {
	DatastoreAllocate         privilege
	DatastoreAllocateSpace    privilege
	DatastoreAllocateTemplate privilege
	DatastoreAudit            privilege
	GroupAllocate             privilege
	PermissionsModify         privilege
	PoolAllocate              privilege
	PoolAudit                 privilege
	RealmAllocate             privilege
	RealmAllocateUser         privilege
	SDNAllocate               privilege
	SDNAudit                  privilege
	SysAudit                  privilege
	SysConsole                privilege
	SysIncoming               privilege
	SysModify                 privilege
	SysPowerMgmt              privilege
	SysSyslog                 privilege
	UserModify                privilege
	VMAllocate                privilege
	VMAudit                   privilege
	VMBackup                  privilege
	VMClone                   privilege
	VMConfigCDROM             privilege
	VMConfigCPU               privilege
	VMConfigCloudinit         privilege
	VMConfigDisk              privilege
	VMConfigHWType            privilege
	VMConfigMemory            privilege
	VMConfigNetwork           privilege
	VMConfigOptions           privilege
	VMConsole                 privilege
	VMMigrate                 privilege
	VMMonitor                 privilege
	VMPowerMgmt               privilege
	VMSnapshot                privilege
	VMSnapshotRollback        privilege
}

func (p privileges) includes(needed Privileges, number privilege) bool {
	if needed.DatastoreAllocate && (p.DatastoreAllocate < number) {
		return false
	}
	if needed.DatastoreAllocateSpace && (p.DatastoreAllocateSpace < number) {
		return false
	}
	if needed.DatastoreAllocateTemplate && (p.DatastoreAllocateTemplate < number) {
		return false
	}
	if needed.DatastoreAudit && (p.DatastoreAudit < number) {
		return false
	}
	if needed.GroupAllocate && (p.GroupAllocate < number) {
		return false
	}
	if needed.PermissionsModify && (p.PermissionsModify < number) {
		return false
	}
	if needed.PoolAllocate && (p.PoolAllocate < number) {
		return false
	}
	if needed.PoolAudit && (p.PoolAudit < number) {
		return false
	}
	if needed.RealmAllocate && (p.RealmAllocate < number) {
		return false
	}
	if needed.RealmAllocateUser && (p.RealmAllocateUser < number) {
		return false
	}
	if needed.SDNAllocate && (p.SDNAllocate < number) {
		return false
	}
	if needed.SDNAudit && (p.SDNAudit < number) {
		return false
	}
	if needed.SysAudit && (p.SysAudit < number) {
		return false
	}
	if needed.SysConsole && (p.SysConsole < number) {
		return false
	}
	if needed.SysIncoming && (p.SysIncoming < number) {
		return false
	}
	if needed.SysModify && (p.SysModify < number) {
		return false
	}
	if needed.SysPowerMgmt && (p.SysPowerMgmt < number) {
		return false
	}
	if needed.SysSyslog && (p.SysSyslog < number) {
		return false
	}
	if needed.UserModify && (p.UserModify < number) {
		return false
	}
	if needed.VMAllocate && (p.VMAllocate < number) {
		return false
	}
	if needed.VMAudit && (p.VMAudit < number) {
		return false
	}
	if needed.VMBackup && (p.VMBackup < number) {
		return false
	}
	if needed.VMClone && (p.VMClone < number) {
		return false
	}
	if needed.VMConfigCDROM && (p.VMConfigCDROM < number) {
		return false
	}
	if needed.VMConfigCPU && (p.VMConfigCPU < number) {
		return false
	}
	if needed.VMConfigCloudinit && (p.VMConfigCloudinit < number) {
		return false
	}
	if needed.VMConfigDisk && (p.VMConfigDisk < number) {
		return false
	}
	if needed.VMConfigHWType && (p.VMConfigHWType < number) {
		return false
	}
	if needed.VMConfigMemory && (p.VMConfigMemory < number) {
		return false
	}
	if needed.VMConfigNetwork && (p.VMConfigNetwork < number) {
		return false
	}
	if needed.VMConfigOptions && (p.VMConfigOptions < number) {
		return false
	}
	if needed.VMConsole && (p.VMConsole < number) {
		return false
	}
	if needed.VMMigrate && (p.VMMigrate < number) {
		return false
	}
	if needed.VMMonitor && (p.VMMonitor < number) {
		return false
	}
	if needed.VMPowerMgmt && (p.VMPowerMgmt < number) {
		return false
	}
	if needed.VMSnapshot && (p.VMSnapshot < number) {
		return false
	}
	if needed.VMSnapshotRollback && (p.VMSnapshotRollback < number) {
		return false
	}
	return true
}

func (privileges) mapToSDK(params map[string]interface{}) (p privileges) {
	if v, isSet := params[key_Privileges_DatastoreAllocate]; isSet {
		p.DatastoreAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_DatastoreAllocateSpace]; isSet {
		p.DatastoreAllocateSpace = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_DatastoreAllocateTemplate]; isSet {
		p.DatastoreAllocateTemplate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_DatastoreAudit]; isSet {
		p.DatastoreAudit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_GroupAllocate]; isSet {
		p.GroupAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_PermissionsModify]; isSet {
		p.PermissionsModify = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_PoolAllocate]; isSet {
		p.PoolAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_PoolAudit]; isSet {
		p.PoolAudit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_RealmAllocate]; isSet {
		p.RealmAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_RealmAllocateUser]; isSet {
		p.RealmAllocateUser = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SDNAllocate]; isSet {
		p.SDNAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SDNAudit]; isSet {
		p.SDNAudit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SysAudit]; isSet {
		p.SysAudit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SysConsole]; isSet {
		p.SysConsole = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SysIncoming]; isSet {
		p.SysIncoming = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SysModify]; isSet {
		p.SysModify = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_SysPowerMgmt]; isSet {
		p.SysPowerMgmt.extract(v)
	}
	if v, isSet := params[key_Privileges_SysSyslog]; isSet {
		p.SysSyslog = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_UserModify]; isSet {
		p.UserModify = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMAllocate]; isSet {
		p.VMAllocate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMAudit]; isSet {
		p.VMAudit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMBackup]; isSet {
		p.VMBackup = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMClone]; isSet {
		p.VMClone = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigCDROM]; isSet {
		p.VMConfigCDROM = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigCPU]; isSet {
		p.VMConfigCPU = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigCloudinit]; isSet {
		p.VMConfigCloudinit = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigDisk]; isSet {
		p.VMConfigDisk = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigHWType]; isSet {
		p.VMConfigHWType = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigMemory]; isSet {
		p.VMConfigMemory = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigNetwork]; isSet {
		p.VMConfigNetwork = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConfigOptions]; isSet {
		p.VMConfigOptions = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMConsole]; isSet {
		p.VMConsole = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMMigrate]; isSet {
		p.VMMigrate = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMMonitor]; isSet {
		p.VMMonitor = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMPowerMgmt]; isSet {
		p.VMPowerMgmt = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMSnapshot]; isSet {
		p.VMSnapshot = privilege(0).extract(v)
	}
	if v, isSet := params[key_Privileges_VMSnapshotRollback]; isSet {
		p.VMSnapshotRollback = privilege(0).extract(v)
	}
	return
}
