package proxmox

import "strings"

type HotPlug struct {
	CPU     *bool `json:"cpu,omitempty"`     // Never nil when returned.
	Disk    *bool `json:"disk,omitempty"`    // Never nil when returned.
	Memory  *bool `json:"memory,omitempty"`  // Never nil when returned.
	Network *bool `json:"network,omitempty"` // Never nil when returned.
	USB     *bool `json:"usb,omitempty"`     // Never nil when returned.
}

func (h HotPlug) mapToApi(b *strings.Builder) {
	if h.CPU != nil && *h.CPU {
		b.WriteString(comma + qemuSettingHotPlugCPU)
	}
	if h.Disk != nil && *h.Disk {
		b.WriteString(comma + qemuSettingHotPlugDisk)
	}
	if h.Memory != nil && *h.Memory {
		b.WriteString(comma + qemuSettingHotPlugMemory)
	}
	if h.Network != nil && *h.Network {
		b.WriteString(comma + qemuSettingHotPlugNetwork)
	}
	if h.USB != nil && *h.USB {
		b.WriteString(comma + qemuSettingHotPlugUSB)
	}
}

func (h HotPlug) mapToApiCreate(builder *strings.Builder) {
	var b strings.Builder
	h.mapToApi(&b)
	if b.Len() > 0 {
		tmp := b.String()[len(comma):]
		if tmp == qemuSettingHotPlugDisk+comma+qemuSettingHotPlugNetwork+comma+qemuSettingHotPlugUSB {
			return // It defaults to this value when unset
		}
		builder.WriteString("&" + qemuApiKeyHotPlug + "=")
		builder.WriteString(tmp)
		return
	}
	builder.WriteString("&" + qemuApiKeyHotPlug + "=0")
}

func (h HotPlug) mapToApiUpdate(current *HotPlug, builder *strings.Builder) {
	var combined HotPlug
	var changed bool
	if h.CPU != nil && *h.CPU != *current.CPU {
		combined.CPU = h.CPU
		changed = true
	} else {
		combined.CPU = current.CPU
	}
	if h.Disk != nil && *h.Disk != *current.Disk {
		combined.Disk = h.Disk
		changed = true
	} else {
		combined.Disk = current.Disk
	}
	if h.Memory != nil && *h.Memory != *current.Memory {
		combined.Memory = h.Memory
		changed = true
	} else {
		combined.Memory = current.Memory
	}
	if h.Network != nil && *h.Network != *current.Network {
		combined.Network = h.Network
		changed = true
	} else {
		combined.Network = current.Network
	}
	if h.USB != nil && *h.USB != *current.USB {
		combined.USB = h.USB
		changed = true
	} else {
		combined.USB = current.USB
	}
	if changed {
		builder.WriteString("&" + qemuApiKeyHotPlug + "=")
		var b strings.Builder
		combined.mapToApi(&b)
		if b.Len() > 0 {
			builder.WriteString(b.String()[len(comma):])
		} else {
			builder.WriteByte('0')
		}
	}
}

func (raw *rawConfigQemu) GetHotPlug() HotPlug {
	if v, isSet := raw.a[qemuApiKeyHotPlug]; isSet {
		hotPlug := HotPlug{
			CPU:     new(false),
			Disk:    new(false),
			Memory:  new(false),
			Network: new(false),
			USB:     new(false),
		}
		tmp := v.(string)
		if len(tmp) < 2 {
			return hotPlug
		}
		vv := strings.Split(tmp, ",")
		type listItem struct {
			ptr  *bool
			item byte
		}
		values := [5]listItem{ // They al start with a different letter, therefore we only have to check the first byte
			{ptr: hotPlug.CPU, item: qemuSettingHotPlugCPU[0]},
			{ptr: hotPlug.Disk, item: qemuSettingHotPlugDisk[0]},
			{ptr: hotPlug.Memory, item: qemuSettingHotPlugMemory[0]},
			{ptr: hotPlug.Network, item: qemuSettingHotPlugNetwork[0]},
			{ptr: hotPlug.USB, item: qemuSettingHotPlugUSB[0]}}
		for i := range vv {
			firstByte := vv[i][0]
			for ii := range values {
				if values[ii].item == firstByte {
					*values[ii].ptr = true
					break
				}
			}
		}
		return hotPlug
	}
	return HotPlug{ // Yes it really defaults to this...
		CPU:     new(false),
		Disk:    new(true),
		Memory:  new(false),
		Network: new(true),
		USB:     new(true)}
}

const (
	qemuSettingHotPlugCPU     = "cpu"
	qemuSettingHotPlugDisk    = "disk"
	qemuSettingHotPlugMemory  = "memory"
	qemuSettingHotPlugNetwork = "network"
	qemuSettingHotPlugUSB     = "usb"
)
