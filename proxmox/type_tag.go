package proxmox

import (
	"errors"
	"regexp"
	"strings"
)

type Tag string

var (
	regexTag = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9_-]*$`)
)

const (
	Tag_Error_Invalid   string = "tag may not start with - and may only include the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-"
	Tag_Error_Duplicate string = "duplicate tag found"
	Tag_Error_MaxLength string = "tag may only be 124 characters"
	Tag_Error_Empty     string = "tag may not be empty"
)

func (Tag) mapToApi(tags []Tag) string {
	if len(tags) == 0 {
		return ""
	}
	tagsString := ""
	for _, e := range tags {
		tagsString += string(e) + ";"
	}
	return tagsString[:len(tagsString)-1]
}

func (Tag) mapToSDK(tags string) []Tag {
	tmpTags := strings.Split(tags, ";")
	typedTags := make([]Tag, len(tmpTags))
	for i, e := range tmpTags {
		typedTags[i] = Tag(e)
	}
	return typedTags
}

func (Tag) validate(tags []Tag) error {
	if len(tags) == 0 {
		return nil
	}
	for i, e := range tags {
		if err := e.Validate(); err != nil {
			return err
		}
		for j := i + 1; j < len(tags); j++ {
			if e == tags[j] {
				return errors.New(Tag_Error_Duplicate)
			}
		}
	}
	return nil
}

func (t Tag) Validate() error {
	if len(t) == 0 {
		return errors.New(Tag_Error_Empty)
	}
	if len(t) > 124 {
		return errors.New(Tag_Error_MaxLength)
	}
	if !regexTag.MatchString(string(t)) {
		return errors.New(Tag_Error_Invalid)
	}
	return nil
}
