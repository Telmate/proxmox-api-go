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

// Return an error if the one of the values is empty.
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

type TemplateItem struct {
	Architecture   string
	Description    string
	Headline       string
	InfoPage       string
	Location       string
	ManageURL      string
	OS             string
	Package        string
	Section        string
	SHA512Checksum string
	Source         string
	Template       string
	Type           string
	Version        string
}

// Map values from the API to the TemplateItem struct.
func createTemplateList(templateList []interface{}) *[]TemplateItem {
	templates := make([]TemplateItem, len(templateList))
	for i := range templateList {
		itemMap := templateList[i].(map[string]interface{})
		tmp := TemplateItem{}
		if _, isSet := itemMap["architecture"]; isSet {
			tmp.Architecture = itemMap["architecture"].(string)
		}
		if _, isSet := itemMap["description"]; isSet {
			tmp.Description = itemMap["description"].(string)
		}
		if _, isSet := itemMap["headline"]; isSet {
			tmp.Headline = itemMap["headline"].(string)
		}
		if _, isSet := itemMap["infopage"]; isSet {
			tmp.InfoPage = itemMap["infopage"].(string)
		}
		if _, isSet := itemMap["location"]; isSet {
			tmp.Location = itemMap["location"].(string)
		}
		if _, isSet := itemMap["manageurl"]; isSet {
			tmp.ManageURL = itemMap["manageurl"].(string)
		}
		if _, isSet := itemMap["os"]; isSet {
			tmp.OS = itemMap["os"].(string)
		}
		if _, isSet := itemMap["package"]; isSet {
			tmp.Package = itemMap["package"].(string)
		}
		if _, isSet := itemMap["section"]; isSet {
			tmp.Section = itemMap["section"].(string)
		}
		if _, isSet := itemMap["sha512sum"]; isSet {
			tmp.SHA512Checksum = itemMap["sha512sum"].(string)
		}
		if _, isSet := itemMap["source"]; isSet {
			tmp.Source = itemMap["source"].(string)
		}
		if _, isSet := itemMap["template"]; isSet {
			tmp.Template = itemMap["template"].(string)
		}
		if _, isSet := itemMap["type"]; isSet {
			tmp.Type = itemMap["type"].(string)
		}
		if _, isSet := itemMap["version"]; isSet {
			tmp.Version = itemMap["version"].(string)
		}
		templates[i] = tmp
	}
	return &templates
}

// Download an LXC template.
func DownloadLxcTemplate(client *Client, content ConfigContent_Template) (err error) {
	_, err = client.PostWithTask(content.mapToApiValues(), "/nodes/"+content.Node+"/aplinfo")
	return
}

// List all LXC templates available for download.
func ListTemplates(client *Client, node string) (templateList *[]TemplateItem, err error) {
	tmpList, err := client.GetItemListInterfaceArray("/nodes/" + node + "/aplinfo")
	if err != nil {
		return
	}
	templateList = createTemplateList(tmpList)
	return
}
