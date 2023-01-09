package proxmox

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type test_Content_ContentType struct {
	ApiValue ContentType
	err      error
}

// test if the ContentType value is properly converted to API values
func Test_Content_ContentType_toApiValue(t *testing.T) {
	input := test_Content_ContentType_Input()
	output := test_Content_ContentType_Output()
	for i := range input {
		require.Equal(t, output[i].ApiValue, input[i].toApiValue())
	}
}

// test if the ContentType value is properly converted to API values and test if its a valid enum
func Test_Content_ContentType_toApiValueAndValidate(t *testing.T) {
	input := test_Content_ContentType_Input()
	output := test_Content_ContentType_Output()
	for i := range input {
		api, err := input[i].toApiValueAndValidate()
		require.Equal(t, output[i].ApiValue, api)
		require.Equal(t, output[i].err, err)
	}
}

// test if the ContentType value is a valid enum
func Test_Content_ContentType_Validate(t *testing.T) {
	input := test_Content_ContentType_Input()
	output := test_Content_ContentType_Output()
	for i := range input {
		require.Equal(t, output[i].err, input[i].Validate())
	}
}

// input data for testing
func test_Content_ContentType_Input() []ContentType {
	return []ContentType{
		ContentType_Backup,
		ContentType_Container,
		ContentType_DiskImage,
		ContentType_Iso,
		ContentType_Snippets,
		ContentType_Template,
		"invalid input",
		contentType_Backup_ApiValue,
		contentType_Container_ApiValue,
		contentType_DiskImage_ApiValue,
		contentType_Iso_ApiValue,
		contentType_Snippets_ApiValue,
		contentType_Template_ApiValue,
		"",
	}
}

// result data for testing
func test_Content_ContentType_Output() []test_Content_ContentType {
	var c ContentType
	testArray := []test_Content_ContentType{
		{
			ApiValue: contentType_Backup_ApiValue,
			err:      nil,
		},
		{
			ApiValue: contentType_Container_ApiValue,
			err:      nil,
		},
		{
			ApiValue: contentType_DiskImage_ApiValue,
			err:      nil,
		},
		{
			ApiValue: contentType_Iso_ApiValue,
			err:      nil,
		},
		{
			ApiValue: contentType_Snippets_ApiValue,
			err:      nil,
		},
		{
			ApiValue: contentType_Template_ApiValue,
			err:      nil,
		},
		{
			ApiValue: "",
			err:      errors.New("value should be one of (" + c.enumList() + ")"),
		},
	}
	return append(testArray, testArray...)
}

// test the formatting of the file object into a single string
func Test_Content_File_format(t *testing.T) {
	input := []Content_File{
		{
			Storage:     "Local",
			ContentType: "vztmpl",
			FilePath:    "debian-11-standard_11.0-1_amd64.tar.gz",
		},
		{
			Storage:     "local",
			ContentType: "vztmpl",
			FilePath:    "/ubuntu-22.10-standard_22.10-1_amd64.tar.zst",
		},
	}
	output := []string{"/Local:vztmpl/debian-11-standard_11.0-1_amd64.tar.gz",
		"/local:vztmpl/ubuntu-22.10-standard_22.10-1_amd64.tar.zst"}
	for i := range input {
		require.Equal(t, output[i], input[i].format())
	}
}

// test if the existence of the file wil be detected
func Test_Content_CheckFileExistence(t *testing.T) {
	fileList := func() []Content_FileProperties {
		return []Content_FileProperties{
			{
				Name: "aaaC",
			},
			{
				Name: "aaaB",
			},
			{
				Name: "aaaA",
			},
		}
	}
	data := []struct {
		FileName       string
		Existence      bool
		FileProperties []Content_FileProperties
	}{
		{
			FileName:       "aaaA",
			Existence:      true,
			FileProperties: fileList(),
		},
		{
			FileName:       "aaaB",
			Existence:      true,
			FileProperties: fileList(),
		},
		{
			FileName:       "",
			Existence:      false,
			FileProperties: fileList(),
		},
		{
			FileName:       "aaaA",
			Existence:      false,
			FileProperties: []Content_FileProperties{},
		},
		{
			FileName:       "aaaX",
			Existence:      false,
			FileProperties: fileList(),
		},
	}
	for _, e := range data {
		require.Equal(t, e.Existence, CheckFileExistence(e.FileName, &e.FileProperties))
	}
}

// test the conversion from a volumeID to a file path
func Test_Content_createFilesList(t *testing.T) {
	input := [][]interface{}{
		{
			map[string]interface{}{
				"ctime":  float64(1671032208),
				"format": "txz",
				"volid":  "local:vztmpl/alpine-3.16-default_20220622_amd64.tar.xz",
				"size":   float64(2540360),
			},
			map[string]interface{}{
				"ctime":  float64(1671032191),
				"format": "txz",
				"volid":  "local:vztmpl/centos-8-default_20201210_amd64.tar.xz",
				"size":   float64(99098368),
			},
			map[string]interface{}{
				"ctime":  float64(1671032200),
				"format": "tgz",
				"volid":  "local:vztmpl/debian-10-standard_10.7-1_amd64.tar.gz",
				"size":   float64(231060971),
			},
		}, {
			map[string]interface{}{
				"ctime":  float64(1665838226),
				"format": "txz",
				"volid":  "local:vztmpl/root-fs.tar.xz",
				"size":   float64(77551540),
			},
		},
	}
	output := []*[]Content_FileProperties{
		{
			{
				Name:         "alpine-3.16-default_20220622_amd64.tar.xz",
				CreationTime: time.UnixMilli(1671032208 * 1000),
				Format:       "txz",
				Size:         2540360,
			}, {
				Name:         "centos-8-default_20201210_amd64.tar.xz",
				CreationTime: time.UnixMilli(1671032191 * 1000),
				Format:       "txz",
				Size:         99098368,
			}, {
				Name:         "debian-10-standard_10.7-1_amd64.tar.gz",
				CreationTime: time.UnixMilli(1671032200 * 1000),
				Format:       "tgz",
				Size:         231060971,
			}}, {{
			Name:         "root-fs.tar.xz",
			CreationTime: time.UnixMilli(1665838226 * 1000),
			Format:       "txz",
			Size:         77551540,
		}}}
	for i := range input {
		require.Equal(t, output[i], createFilesList(input[i]))
	}
}

// test the conversion from a volumeID to a file path
func Test_Content_volumeIdToFileName(t *testing.T) {
	input := []string{"local:vztmpl/alpine-3.16-default_20220622_amd64.tar.xz"}
	output := []string{"alpine-3.16-default_20220622_amd64.tar.xz"}
	for i := range input {
		require.Equal(t, output[i], volumeIdToFileName(input[i]))
	}
}
