package proxmox

import (
	"errors"
	"strings"
)

const (
	permissionErrorPrefixShort = "Permission check failed"
	permissionErrorPrefixLong  = permissionErrorPrefixShort + " ("
)

type PermissionError struct {
	err        error
	Path       *string
	Privileges *Privileges
}

var errPermission = errors.New("incorrect permission")

func newPermissionError(errMsg string) error {
	if len(errMsg) < len(permissionErrorPrefixShort) { // "Permission check failed"
		return nil
	}
	if !strings.EqualFold(errMsg[:len(permissionErrorPrefixShort)], permissionErrorPrefixShort) {
		return nil
	}
	if len(errMsg) <= len(permissionErrorPrefixLong) { // "Permission check failed (/pool/Test_Pool_Exists_No_Permission, Pool.Audit)"
		return &PermissionError{err: errPermission}
	}
	index := strings.IndexByte(errMsg[len(permissionErrorPrefixLong):], ',')
	path := errMsg[len(permissionErrorPrefixLong) : len(permissionErrorPrefixLong)+index]
	var p Privileges
	p.parse(errMsg[index+2+len(permissionErrorPrefixLong) : len(errMsg)-1])
	return &PermissionError{
		Privileges: &p,
		Path:       &path,
		err:        errPermission}
}

func (e *PermissionError) Error() string {
	if e.Path != nil && e.Privileges != nil {
		return "missing permission: " + e.Privileges.String() + " on: " + *e.Path
	}
	return "missing permission"
}

func (e *PermissionError) Unwrap() error { return e.err }
