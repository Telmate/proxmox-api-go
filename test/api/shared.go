package api_test

import (
	"crypto/tls"

	pveSDK "github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/Telmate/proxmox-api-go/test"
)

func NewClient() (*pveSDK.Client, error) {
	return pveSDK.NewClient(test.ApiURL, nil, "", &tls.Config{InsecureSkipVerify: true}, "", 1000, nil, test.FeatureFlags)
}
