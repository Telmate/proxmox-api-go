package proxmox

type ConfigContent_Template struct {
	Node     string
	Storage  string
	Template string
}

func (content ConfigContent_Template) mapToApiValues() map[string]interface{} {
	return map[string]interface{}{
		"storage":  content.Storage,
		"template": content.Template,
	}
}

func DownloadLxcTemplate(client *Client, content ConfigContent_Template) (err error) {
	_, err = client.PostWithTask(content.mapToApiValues(), "/nodes/"+content.Node+"/aplinfo")
	return
}
