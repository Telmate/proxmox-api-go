package proxmox

import (
	"errors"
	"strings"
	"time"
)

type ContentType string

const (
	ContentType_Backup             ContentType = "backup"
	ContentType_Container          ContentType = "container"
	ContentType_DiskImage          ContentType = "diskimage"
	ContentType_Iso                ContentType = "iso"
	ContentType_Snippets           ContentType = "snippets"
	ContentType_Template           ContentType = "template"
	contentType_Backup_ApiValue    ContentType = "backup"
	contentType_Container_ApiValue ContentType = "rootdir"
	contentType_DiskImage_ApiValue ContentType = "images"
	contentType_Snippets_ApiValue  ContentType = "snippets"
	contentType_Iso_ApiValue       ContentType = "iso"
	contentType_Template_ApiValue  ContentType = "vztmpl"
)

// Converts the user friendly enum value to a value the proxmox api understands.
func (c ContentType) toApiValue() ContentType {
	switch c {
	case ContentType_Backup:
		return contentType_Backup_ApiValue
	case ContentType_Container, contentType_Container_ApiValue:
		return contentType_Container_ApiValue
	case ContentType_DiskImage, contentType_DiskImage_ApiValue:
		return contentType_DiskImage_ApiValue
	case ContentType_Iso:
		return contentType_Iso_ApiValue
	case ContentType_Snippets:
		return contentType_Snippets_ApiValue
	case ContentType_Template, contentType_Template_ApiValue:
		return contentType_Template_ApiValue
	}
	return ""
}

// Converts the user friendly enum value to a value the proxmox api understands.
// If the input enum value is invalid it will return an error.
func (c ContentType) toApiValueAndValidate() (api ContentType, err error) {
	api = c.toApiValue()
	if api == "" {
		err = errors.New("value should be one of (" + c.enumList() + ")")
	}
	return
}

// Returns a list of all enum options.
func (c ContentType) enumList() string {
	return string(ContentType_Backup) + "," + string(ContentType_Container) + "," + string(ContentType_DiskImage) + "," + string(ContentType_Iso) + "," + string(ContentType_Snippets) + "," + string(ContentType_Template)
}

// Returns an error if the enum value is invalid.
func (c ContentType) Validate() (err error) {
	_, err = c.toApiValueAndValidate()
	return
}

type Content_File struct {
	Storage     string
	ContentType ContentType
	FilePath    string
}

func (c Content_File) format() (fullPath string) {
	return "/" + c.Storage + ":" + string(c.ContentType) + "/" + strings.TrimPrefix(c.FilePath, "/")
}

// Return an error if the required fields are empty
func (c Content_File) Validate() (err error) {
	err = c.ContentType.Validate()
	if err != nil {
		return
	}
	if c.Storage == "" {
		return errors.New("storage may not be empty")
	}
	if c.FilePath == "" {
		return errors.New("filepath may not be empty")
	}
	return
}

type Content_FileProperties struct {
	Name         string    `json:"name"`
	CreationTime time.Time `json:"time"`
	Format       string    `json:"format"`
	Size         uint      `json:"size"`
}

// Check if a file with the specific name exists in the list of files.
func CheckFileExistence(fileName string, files *[]Content_FileProperties) bool {
	for _, e := range *files {
		if e.Name == fileName {
			return true
		}
	}
	return false
}

func createFilesList(fileList []interface{}) *[]Content_FileProperties {
	files := make([]Content_FileProperties, len(fileList))
	for i := range fileList {
		itemMap := fileList[i].(map[string]interface{})
		tmpFile := Content_FileProperties{}
		if _, isSet := itemMap["volid"]; isSet {
			tmpFile.Name = volumeIdToFileName(itemMap["volid"].(string))
		}
		if _, isSet := itemMap["ctime"]; isSet {
			tmpFile.CreationTime = time.UnixMilli(int64(itemMap["ctime"].(float64)) * 1000)
		}
		if _, isSet := itemMap["format"]; isSet {
			tmpFile.Format = itemMap["format"].(string)
		}
		if _, isSet := itemMap["size"]; isSet {
			tmpFile.Size = uint(itemMap["size"].(float64))
		}
		files[i] = tmpFile
	}
	return &files
}

// Deletes te specified file from the specified storage.
func DeleteFile(client *Client, node string, content Content_File) (err error) {
	content.ContentType, err = content.ContentType.toApiValueAndValidate()
	if err != nil {
		return
	}
	_, err = client.DeleteWithTask("/nodes/" + node + "/storage/" + content.Storage + "/content" + content.format())
	return
}

// List all files of the given type in the the specified storage
func ListFiles(client *Client, node, storage string, content ContentType) (files *[]Content_FileProperties, err error) {
	content, err = content.toApiValueAndValidate()
	if err != nil {
		return
	}
	fileList, err := client.GetItemListInterfaceArray("/nodes/" + node + "/storage/" + storage + "/content?content=" + string(content))
	if err != nil {
		return
	}
	return createFilesList(fileList), nil
}

func volumeIdToFileName(volumeId string) string {
	return volumeId[strings.Index(volumeId, "/")+1:]
}
