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

func (t Tags) mapToApiCreate() string {
	if len(t) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for i := range t {
		builder.WriteString(comma + t[i].String())
	}
	return builder.String()[3:]
}

// Tags not converted to lowercase during Qemu VM creation
// https://bugzilla.proxmox.com/show_bug.cgi?id=7549
func (t Tags) mapToApiCreateLower() string {
	if len(t) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for i := range t {
		builder.WriteString(comma + strings.ToLower(t[i].String()))
	}
	return builder.String()[3:]
}

func (t Tags) mapToApiUpdate(current Tags) (string, bool) {
	if len(t) != len(current) {
		return t.String(), true
	}
	sort.Sort(t)
	sort.Sort(current)
	newTags := t.mapToApiCreate()
	if newTags == current.mapToApiCreate() {
		return "", false
	}
	return newTags, true
}

func (t *Tags) mapToSDK(tags string) {
	// Handle Proxmox API bug: sometimes returns " " (whitespace) for VMs with no tags
	trimmed := strings.TrimSpace(tags)
	if trimmed == "" {
		return
	}
	seperators := countSeperator(trimmed)
	tagsTyped := make(Tags, seperators+1)
	var entry, lastSeperator int
	for i := 0; i < len(trimmed); i++ {
		switch trimmed[i] {
		case ',', ';':
			tagsTyped[entry] = Tag(trimmed[lastSeperator:i])
			entry++
			i++
			lastSeperator = i
		}
	}
	tagsTyped[entry] = Tag(trimmed[lastSeperator:])
	*t = tagsTyped
}

func (t Tags) String() string { // String is for fmt.Stringer.
	if len(t) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for i := range t {
		builder.WriteString("," + t[i].String())
	}
	return builder.String()[1:]
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
