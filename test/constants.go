package test

import (
	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
)

const (
	DownloadedLXCTemplate = `alpine-3.21-default_20241217_amd64.tar.xz`
	TemplateStorage       = `local`
	FirstNode             = `pve`
	GuestStorage          = `local-zfs`
	QemuTemplateID        = 9000
	ApiURL                = `https://127.0.0.1:8006/api2/json`
	UserID                = `root@pam`
	Password              = `Enter123!`
)

var FeatureFlags = pveSDK.FeatureFlags{
	AsyncTask:          true,
	PanicOnInvalidTask: true,
}
