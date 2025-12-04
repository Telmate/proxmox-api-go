package proxmox

import (
	"errors"
	"regexp"
	"sort"
	"strings"
)

type Tags []Tag

const (
	Tags_Error_Duplicate string = "duplicate tag found"
)

func (t Tags) Len() int           { return len(t) }           // Len is for sort.Interface.
func (t Tags) Less(i, j int) bool { return t[i] < t[j] }      // Less is for sort.Interface.
func (t Tags) Swap(i, j int)      { t[i], t[j] = t[j], t[i] } // Swap is for sort.Interface.

func (new Tags) mapToApiCreate() string {
	return new.String()
}

func (new Tags) mapToApiUpdate(current *Tags) (string, bool) {
	if current != nil {
		sort.Sort(new)
		sort.Sort(*current)
		newTags := new.String()
		if newTags == current.String() {
			return "", false
		}
		return newTags, true
	}
	return new.String(), true
}

func (Tags) mapToSDK(tags string) Tags {
	// Handle Proxmox API bug: sometimes returns " " (whitespace) for VMs with no tags
	trimmed := strings.TrimSpace(tags)
	if trimmed == "" {
		return Tags{}
	}
	
	tmpTags := strings.Split(trimmed, ";")
	typedTags := make(Tags, 0, len(tmpTags))
	for _, tag := range tmpTags {
		// Trim whitespace from each tag and skip empty tags
		if trimmedTag := strings.TrimSpace(tag); trimmedTag != "" {
			typedTags = append(typedTags, Tag(trimmedTag))
		}
	}
	return typedTags
}

func (t Tags) String() string { // String is for fmt.Stringer.
	if len(t) == 0 {
		return ""
	}
	var tags string
	for _, e := range t {
		tags += ";" + e.String()
	}
	return tags[1:]
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

func (t Tag) String() string { return string(t) } // String is for fmt.Stringer.

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
