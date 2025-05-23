package proxmox

import (
	"errors"
	"regexp"
	"strings"
)

type Tags []Tag

const (
	Tags_Error_Duplicate string = "duplicate tag found"
)

func (t Tags) mapToApi() string {
	if len(t) == 0 {
		return ""
	}
	var tags string
	for _, e := range t {
		tags += ";" + string(e)
	}
	return tags[1:]
}

func (Tags) mapToSDK(tags string) Tags {
	tmpTags := strings.Split(tags, ";")
	typedTags := make(Tags, len(tmpTags))
	for i, e := range tmpTags {
		typedTags[i] = Tag(e)
	}
	return typedTags
}

func (t Tags) Validate() error {
	if len(t) == 0 {
		return nil
	}
	for i, e := range t {
		if err := e.Validate(); err != nil {
			return err
		}
		for j := i + 1; j < len(t); j++ {
			if e == t[j] {
				return errors.New(Tags_Error_Duplicate)
			}
		}
	}
	return nil
}

type Tag string

var (
	regexTag = regexp.MustCompile(`^[a-zA-Z0-9_][a-zA-Z0-9-._]*$`)
)

const (
	Tag_Error_Invalid   string = "tag may not start with -. and may only include the following characters: abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._"
	Tag_Error_MaxLength string = "tag may only be 124 characters"
	Tag_Error_Empty     string = "tag may not be empty"
)

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
