package proxmox

import (
	"errors"
)

type ConfigContent_Template struct {
	Node     string
	Storage  string
	Template string
}

func (content ConfigContent_Template) error(text string) error {
	return errors.New("the value of (" + text + ") may not be empty")
}

func (content ConfigContent_Template) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"storage":  content.Storage,
		"template": content.Template,
	}
}

func (content ConfigContent_Template) Validate() (err error) {
	if content.Node == "" {
		return content.error("Node")
	}
	if content.Storage == "" {
		return content.error("Storage")
	}
	if content.Template == "" {
		return content.error("Template")
	}
	return
}

func DownloadLxcTemplate(client *Client, content ConfigContent_Template) (err error) {
	_, err = client.PostWithTask(content.mapToApiValues(), "/nodes/"+content.Node+"/aplinfo")
	return
}
