package proxmox

import (
	"errors"
)

type ConfigContent_Iso struct {
	Checksum          string
	ChecksumAlgorithm string
	DownloadUrl       string
	Filename          string
	Node              string
	Storage           string
}

func (content ConfigContent_Iso) error(text string) error {
	return errors.New("the value of (" + text + ") may not be empty")
}

func (content ConfigContent_Iso) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"checksum-algorithm": content.ChecksumAlgorithm,
		"checksum":           content.Checksum,
		"content":            "iso",
		"filename":           content.Filename,
		"storage":            content.Storage,
		"url":                content.DownloadUrl,
	}
}

// Return an error if the one of the values is empty.
func (content ConfigContent_Iso) Validate() (err error) {
	if content.Node == "" {
		return content.error("Node")
	}
	if content.Storage == "" {
		return content.error("Storage")
	}
	if content.DownloadUrl == "" {
		return content.error("URL")
	}
	if content.Filename == "" {
		return content.error("Filename")
	}
	return
}

// Download an Iso file from a given URL.
// https://pve.proxmox.com/pve-docs/api-viewer/#/nodes/{node}/storage/{storage}/download-url
func DownloadIsoFromUrl(client *Client, content ConfigContent_Iso) (err error) {
	_, err = client.PostWithTask(content.mapToApiValues(), "/nodes/"+content.Node+"/storage/"+content.Storage+"/download-url")
	if err != nil {
		return err
	}
	return nil
}
