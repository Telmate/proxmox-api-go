package proxmox

import (
	"crypto"
	"errors"
	"net"
	"net/netip"
	"testing"

	"github.com/Telmate/proxmox-api-go/internal/util"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_pool"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_qemu"
	"github.com/Telmate/proxmox-api-go/test/data/test_data_tag"
	"github.com/stretchr/testify/require"
)

func Test_ConfigQemu_mapToAPI(t *testing.T) {
	cloudInitCustom := func() *CloudInitCustom {
		return &CloudInitCustom{
			Meta: &CloudInitSnippet{
				Storage:  "local-zfs",
				FilePath: "ci-meta.yml"},
			Network: &CloudInitSnippet{
				Storage:  "local-lvm",
				FilePath: "ci-network.yml"},
			User: &CloudInitSnippet{
				Storage:  "folder",
				FilePath: "ci-user.yml"},
			Vendor: &CloudInitSnippet{
				Storage:  "local",
				FilePath: "snippets/ci-custom.yml"}}
	}
	cloudInitNetworkConfig := func() CloudInitNetworkConfig {
		return CloudInitNetworkConfig{
			IPv4: &CloudInitIPv4Config{
				Address: util.Pointer(IPv4CIDR("192.168.56.30/24")),
				Gateway: util.Pointer(IPv4Address("192.168.56.1"))},
			IPv6: &CloudInitIPv6Config{
				Address: util.Pointer(IPv6CIDR("2001:0db8:abcd::/48")),
				Gateway: util.Pointer(IPv6Address("2001:0db8:abcd::1"))}}
	}
	parseIP := func(rawIP string) (ip netip.Addr) {
		ip, _ = netip.ParseAddr(rawIP)
		return
	}
	parseMAC := func(rawMAC string) (mac net.HardwareAddr) {
		mac, _ = net.ParseMAC(rawMAC)
		return
	}
	networkInterface := func() QemuNetworkInterface {
		return QemuNetworkInterface{
			Bridge:        util.Pointer("vmbr0"),
			Connected:     util.Pointer(false),
			Firewall:      util.Pointer(true),
			MAC:           util.Pointer(parseMAC("52:54:00:12:34:56")),
			MTU:           util.Pointer(QemuMTU{Value: 1500}),
			Model:         util.Pointer(QemuNetworkModel("virtio")),
			MultiQueue:    util.Pointer(QemuNetworkQueue(5)),
			RateLimitKBps: util.Pointer(QemuNetworkRate(45)),
			NativeVlan:    util.Pointer(Vlan(23)),
			TaggedVlans:   util.Pointer(Vlans{12, 23, 45})}
	}
	format_Raw := QemuDiskFormat_Raw
	float10 := QemuDiskBandwidthMBpsLimitConcurrent(10.3)
	float45 := QemuDiskBandwidthMBpsLimitConcurrent(45.23)
	float79 := QemuDiskBandwidthMBpsLimitBurst(79.23)
	float99 := QemuDiskBandwidthMBpsLimitBurst(99.20)
	uint1 := uint(1)
	uint23 := QemuDiskBandwidthIopsLimitConcurrent(23)
	uint34 := QemuDiskBandwidthIopsLimitConcurrent(34)
	uint78 := QemuDiskBandwidthIopsLimitBurst(78)
	uint89 := QemuDiskBandwidthIopsLimitBurst(89)
	update_CloudInit := func() *QemuCloudInitDisk {
		return &QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "test"}
	}
	ideBase := func() *QemuIdeStorage {
		return &QemuIdeStorage{Disk: &QemuIdeDisk{Format: QemuDiskFormat_Raw, Id: 23, SizeInKibibytes: 10, Storage: "test"}}
	}
	sataBase := func() *QemuSataStorage {
		return &QemuSataStorage{Disk: &QemuSataDisk{Format: QemuDiskFormat_Raw, Id: 23, SizeInKibibytes: 10, Storage: "test"}}
	}
	scsiBase := func() *QemuScsiStorage {
		return &QemuScsiStorage{Disk: &QemuScsiDisk{Format: QemuDiskFormat_Raw, Id: 23, SizeInKibibytes: 10, Storage: "test"}}
	}
	virtioBase := func() *QemuVirtIOStorage {
		return &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: QemuDiskFormat_Raw, Id: 23, SizeInKibibytes: 10, Storage: "test"}}
	}
	type test struct {
		name          string
		config        *ConfigQemu
		currentConfig ConfigQemu
		version       Version
		reboot        bool
		output        map[string]interface{}
	}
	tests := []struct {
		category     string
		create       []test
		createUpdate []test // value of currentConfig wil be used for update and ignored for create
		update       []test
	}{
		{category: `Agent`,
			create: []test{
				{name: `Agent=nil`,
					config: &ConfigQemu{},
					output: map[string]interface{}{}},
				{name: `Agent Full`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{
						Enable: util.Pointer(true),
						Type:   util.Pointer(QemuGuestAgentType_VirtIO),
						Freeze: util.Pointer(true),
						FsTrim: util.Pointer(true)}},
					output: map[string]interface{}{"agent": "1,freeze-fs-on-backup=1,fstrim_cloned_disks=1,type=virtio"}},
				{name: `Agent.Enable`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(true)}},
					output: map[string]interface{}{"agent": "1"}},
				{name: `Agent.Type=""`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_None)}},
					output: map[string]interface{}{"agent": "0"}},
				{name: `Agent.Type="virtio"`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}},
					output: map[string]interface{}{"agent": "0,type=virtio"}},
				{name: `Agent.Freeze`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{Freeze: util.Pointer(true)}},
					output: map[string]interface{}{"agent": "0,freeze-fs-on-backup=1"}},
				{name: `Agent.FsTrim`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{FsTrim: util.Pointer(true)}},
					output: map[string]interface{}{"agent": "0,fstrim_cloned_disks=1"}}},
			update: []test{
				{name: `Agent !nil nil`,
					config: &ConfigQemu{Agent: &QemuGuestAgent{}},
					output: map[string]interface{}{"agent": "0"}},
				{name: `Agent nil !nil`,
					config: &ConfigQemu{},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{
						Enable: util.Pointer(true),
						Type:   util.Pointer(QemuGuestAgentType_VirtIO),
						Freeze: util.Pointer(true),
						FsTrim: util.Pointer(true)}},
					output: map[string]interface{}{}},
				{name: `Agent nil nil `,
					config: &ConfigQemu{},
					output: map[string]interface{}{}},
				{name: `Agent.Enable !nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(true)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "1"}},
				{name: `Agent.Enable nil !nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(true)}},
					output:        map[string]interface{}{"agent": "1"}},
				{name: `Agent.Enable nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0"}},
				{name: `Agent.Type !nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0,type=virtio"}},
				{name: `Agent.Type "" !nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_None)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0"}},
				{name: `Agent.Type "" nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_None)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}},
					output:        map[string]interface{}{"agent": "0"}},
				{name: `Agent.Type nil !nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}},
					output:        map[string]interface{}{"agent": "0,type=virtio"}},
				{name: `Agent.Type nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0"}},
				{name: `Agent.Freeze !nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{Freeze: util.Pointer(false)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0,freeze-fs-on-backup=0"}},
				{name: `Agent.Freeze nil !nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{Freeze: util.Pointer(true)}},
					output:        map[string]interface{}{"agent": "0,freeze-fs-on-backup=1"}},
				{name: `Agent.Freeze nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0"}},
				{name: `Agent.FsTrim !nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{FsTrim: util.Pointer(false)}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0,fstrim_cloned_disks=0"}},
				{name: `Agent.FsTrim nil !nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{FsTrim: util.Pointer(true)}},
					output:        map[string]interface{}{"agent": "0,fstrim_cloned_disks=1"}},
				{name: `Agent.FsTrim nil nil`,
					config:        &ConfigQemu{Agent: &QemuGuestAgent{}},
					currentConfig: ConfigQemu{Agent: &QemuGuestAgent{}},
					output:        map[string]interface{}{"agent": "0"}}}},
		{category: `CPU`,
			create: []test{
				{name: `Affinity empty`,
					config: &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{})}},
					output: map[string]interface{}{}},
				{name: `Flags AES`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AES: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+aes"}},
				{name: `Flags AmdNoSSB`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AmdNoSSB: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-amd-no-ssb"}},
				{name: `Flags AmdSSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{AmdSSBD: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+amd-ssbd"}},
				{name: `Flags HvEvmcs`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{HvEvmcs: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-hv-evmcs"}},
				{name: `Flags HvTlbFlush`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{HvTlbFlush: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+hv-tlbflush"}},
				{name: `Flags Ibpb`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{Ibpb: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-ibpb"}},
				{name: `Flags MdClear`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{MdClear: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+md-clear"}},
				{name: `Flags PCID`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{PCID: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-pcid"}},
				{name: `Flags Pdpe1GB`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{Pdpe1GB: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+pdpe1gb"}},
				{name: `Flags SSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{SSBD: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-ssbd"}},
				{name: `Flags SpecCtrl`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{SpecCtrl: util.Pointer(TriBoolTrue)}}},
					output: map[string]interface{}{"cpu": ",flags=+spec-ctrl"}},
				{name: `Flags VirtSSBD`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{VirtSSBD: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=-virt-ssbd"}},
				{name: `Flags mixed`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        util.Pointer(TriBoolTrue),
						AmdNoSSB:   util.Pointer(TriBoolFalse),
						AmdSSBD:    util.Pointer(TriBoolTrue),
						HvEvmcs:    util.Pointer(TriBoolNone),
						HvTlbFlush: util.Pointer(TriBoolTrue),
						MdClear:    util.Pointer(TriBoolTrue),
						PCID:       util.Pointer(TriBoolFalse),
						Pdpe1GB:    util.Pointer(TriBoolNone)}}},
					output: map[string]interface{}{"cpu": ",flags=+aes;-amd-no-ssb;+amd-ssbd;+hv-tlbflush;+md-clear;-pcid"}},
				{name: `Flags all nil`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}},
					output: map[string]interface{}{}},
				{name: `Flags all none`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        util.Pointer(TriBoolNone),
						AmdNoSSB:   util.Pointer(TriBoolNone),
						AmdSSBD:    util.Pointer(TriBoolNone),
						HvEvmcs:    util.Pointer(TriBoolNone),
						HvTlbFlush: util.Pointer(TriBoolNone),
						MdClear:    util.Pointer(TriBoolNone),
						PCID:       util.Pointer(TriBoolNone),
						Pdpe1GB:    util.Pointer(TriBoolNone),
						SSBD:       util.Pointer(TriBoolNone),
						SpecCtrl:   util.Pointer(TriBoolNone),
						VirtSSBD:   util.Pointer(TriBoolNone)}}},
					output: map[string]interface{}{}},
				{name: `Flags all none & Type ""`,
					config: &ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							AES:        util.Pointer(TriBoolNone),
							AmdNoSSB:   util.Pointer(TriBoolNone),
							AmdSSBD:    util.Pointer(TriBoolNone),
							HvEvmcs:    util.Pointer(TriBoolNone),
							HvTlbFlush: util.Pointer(TriBoolNone),
							MdClear:    util.Pointer(TriBoolNone),
							PCID:       util.Pointer(TriBoolNone),
							Pdpe1GB:    util.Pointer(TriBoolNone),
							SSBD:       util.Pointer(TriBoolNone),
							SpecCtrl:   util.Pointer(TriBoolNone),
							VirtSSBD:   util.Pointer(TriBoolNone)},
						Type: util.Pointer(CpuType(""))}},
					output: map[string]interface{}{}},
				{name: `Limit`,
					config: &ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(0))}},
					output: map[string]interface{}{}},
				{name: `Units 0`,
					config: &ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(0))}},
					output: map[string]interface{}{}},
				{name: `VirtualCores 0`,
					config: &ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(0))}},
					output: map[string]interface{}{}}},
			createUpdate: []test{
				{name: `Affinity consecutive`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{0, 0, 1, 2, 2, 3})}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{0, 1, 2})}},
					output:        map[string]interface{}{"affinity": "0-3"}},
				{name: `Affinity singular`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{2})}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{0, 1, 2})}},
					output:        map[string]interface{}{"affinity": "2"}},
				{name: `Affinity mixed`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{5, 0, 4, 2, 9, 3, 2, 11, 7, 2, 12, 4, 13})}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{0, 1, 2})}},
					output:        map[string]interface{}{"affinity": "0,2-5,7,9,11-13"}},
				{name: `Cores`,
					config:        &ConfigQemu{CPU: &QemuCPU{Cores: util.Pointer(QemuCpuCores(1))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Cores: util.Pointer(QemuCpuCores(2))}},
					output:        map[string]interface{}{"cores": 1}},
				{name: `Limit`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(50))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(100))}},
					output:        map[string]interface{}{"cpulimit": 50}},
				{name: `Numa`,
					config:        &ConfigQemu{CPU: &QemuCPU{Numa: util.Pointer(true)}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Numa: util.Pointer(false)}},
					output:        map[string]interface{}{"numa": 1}},
				{name: `Sockets`,
					config:        &ConfigQemu{CPU: &QemuCPU{Sockets: util.Pointer(QemuCpuSockets(3))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Sockets: util.Pointer(QemuCpuSockets(2))}},
					output:        map[string]interface{}{"sockets": 3}},
				{name: `Type lower`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(cpuType_X86_64_v2_AES_Lower)}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_Host)}},
					version:       Version{}.max(),
					output:        map[string]interface{}{"cpu": string(CpuType_X86_64_v2_AES)}},
				{name: `Type normal`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_X86_64_v2_AES)}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_Host)}},
					version:       Version{}.max(),
					output:        map[string]interface{}{"cpu": string(CpuType_X86_64_v2_AES)}},
				{name: `Type weird`,
					config:        &ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType("X_-8-_6_-6-4---V_-2-aE--s__"))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_Host)}},
					version:       Version{}.max(),
					output:        map[string]interface{}{"cpu": string(CpuType_X86_64_v2_AES)}},
				{name: `Units 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(100))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(200))}},
					output:        map[string]interface{}{"cpuunits": 100}},
				{name: `VirtualCores`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(4))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(12))}},
					output:        map[string]interface{}{"vcpus": 4}},
			},
			update: []test{
				{name: `Affinity empty`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{})}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{0, 1, 2})}},
					output:        map[string]interface{}{"affinity": ""}},
				{name: `Affinity empty no current`,
					config:        &ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{})}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{}},
					output:        map[string]interface{}{}},
				{name: `Flags nil`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:  util.Pointer(TriBoolTrue),
						PCID: util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=+aes;-pcid"}},
				{name: `Flags`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        util.Pointer(TriBoolTrue),
						AmdNoSSB:   util.Pointer(TriBoolNone),
						HvTlbFlush: util.Pointer(TriBoolTrue),
						Ibpb:       util.Pointer(TriBoolNone),
						MdClear:    util.Pointer(TriBoolFalse),
						PCID:       util.Pointer(TriBoolTrue),
						SpecCtrl:   util.Pointer(TriBoolFalse),
						VirtSSBD:   util.Pointer(TriBoolFalse)}}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB:   util.Pointer(TriBoolTrue),
						HvEvmcs:    util.Pointer(TriBoolFalse),
						HvTlbFlush: util.Pointer(TriBoolFalse),
						Ibpb:       util.Pointer(TriBoolTrue),
						MdClear:    util.Pointer(TriBoolTrue),
						SpecCtrl:   util.Pointer(TriBoolFalse)}}},
					output: map[string]interface{}{"cpu": ",flags=+aes;-hv-evmcs;+hv-tlbflush;-md-clear;+pcid;-spec-ctrl;-virt-ssbd"}},
				{name: `Flags all none`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AES:        util.Pointer(TriBoolNone),
						AmdNoSSB:   util.Pointer(TriBoolNone),
						AmdSSBD:    util.Pointer(TriBoolNone),
						HvEvmcs:    util.Pointer(TriBoolNone),
						HvTlbFlush: util.Pointer(TriBoolNone),
						MdClear:    util.Pointer(TriBoolNone),
						PCID:       util.Pointer(TriBoolNone),
						Pdpe1GB:    util.Pointer(TriBoolNone),
						SSBD:       util.Pointer(TriBoolNone),
						SpecCtrl:   util.Pointer(TriBoolNone),
						VirtSSBD:   util.Pointer(TriBoolNone)}}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							AES:        util.Pointer(TriBoolTrue),
							AmdNoSSB:   util.Pointer(TriBoolTrue),
							AmdSSBD:    util.Pointer(TriBoolTrue),
							HvEvmcs:    util.Pointer(TriBoolTrue),
							HvTlbFlush: util.Pointer(TriBoolTrue),
							MdClear:    util.Pointer(TriBoolTrue),
							PCID:       util.Pointer(TriBoolTrue),
							Pdpe1GB:    util.Pointer(TriBoolTrue),
							SSBD:       util.Pointer(TriBoolTrue),
							SpecCtrl:   util.Pointer(TriBoolTrue),
							VirtSSBD:   util.Pointer(TriBoolTrue)},
						Type: util.Pointer(CpuType_Host)}},
					output: map[string]interface{}{"cpu": "host,flags="}},
				{name: `Flags & Type, update Flags`,
					config: &ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
						AmdNoSSB: util.Pointer(TriBoolTrue)}}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							HvEvmcs:    util.Pointer(TriBoolFalse),
							HvTlbFlush: util.Pointer(TriBoolFalse),
							Ibpb:       util.Pointer(TriBoolTrue),
							MdClear:    util.Pointer(TriBoolTrue),
							SpecCtrl:   util.Pointer(TriBoolFalse)},
						Type: util.Pointer(CpuType_Host)}},
					output: map[string]interface{}{"cpu": "host,flags=+amd-no-ssb;-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-spec-ctrl"}},
				{name: `Flags & Type, update Type`,
					config: &ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_X86_64_v2_AES)}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							HvEvmcs:    util.Pointer(TriBoolFalse),
							HvTlbFlush: util.Pointer(TriBoolFalse),
							Ibpb:       util.Pointer(TriBoolTrue),
							MdClear:    util.Pointer(TriBoolTrue),
							SpecCtrl:   util.Pointer(TriBoolFalse)},
						Type: util.Pointer(CpuType_Host)}},
					version: Version{}.max(),
					output:  map[string]interface{}{"cpu": "x86-64-v2-AES,flags=-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-spec-ctrl"}},
				{name: `Limit 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(0))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(100))}},
					output:        map[string]interface{}{"delete": "cpulimit"}},
				{name: `Limit 0 no current`,
					config:        &ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(0))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{}},
					output:        map[string]interface{}{}},
				{name: `Units 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(0))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(100))}},
					output:        map[string]interface{}{"delete": "cpuunits"}},
				{name: `VirtualCores 0`,
					config:        &ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(0))}},
					currentConfig: ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(4))}},
					output:        map[string]interface{}{"delete": "vcpus"}},
			}},
		{category: `CloudInit`, // Create CloudInit no need for update as update and create behave the same. will be changed in the future
			createUpdate: []test{
				{name: `CloudInit=nil`,
					config: &ConfigQemu{},
					output: map[string]interface{}{}},
				{name: `CloudInit DNS NameServers`,
					config: &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{
						NameServers: &[]netip.Addr{parseIP("9.9.9.9")}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{
						NameServers: &[]netip.Addr{parseIP("8.8.8.8")}}}},
					output: map[string]interface{}{"nameserver": "9.9.9.9"}},
				{name: `CloudInit DNS SearchDomain`,
					config:        &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{SearchDomain: util.Pointer("example.com")}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{SearchDomain: util.Pointer("example.org")}}},
					output:        map[string]interface{}{"searchdomain": "example.com"}},
				{name: `CloudInit PublicSSHkeys`,
					config:        &ConfigQemu{CloudInit: &CloudInit{PublicSSHkeys: util.Pointer(test_data_qemu.PublicKey_Decoded_Input())}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{PublicSSHkeys: util.Pointer([]crypto.PublicKey{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+0roY6F4yzq5RfA6V2+8gOgKlLOg9RtB1uGyTYvOMU6wxWUXVZP44+XozNxXZK4/MfPjCZLomqv78RlAedIQbqU8l6J9fdrrsRt6NknusE36UqD4HGPLX3Wn7svjSyNRfrjlk5BrBQ26rglLGlRSeD/xWvQ+5jLzzdo5NczszGkE9IQtrmKye7Gq7NQeGkHb1h0yGH7nMQ48WJ6ZKv1JG+GzFb8n4Qoei3zK9zpWxF+0AzF5u/zzCRZ4yU7FtfHgGRBDPze8oe3nVe+aO8MBH2dy8G/BRMXBdjWrSkaT9ZyeaT0k9SMjsCr9DQzUtVSOeqZZokpNU1dVglI+HU0vN test-key"})}},
					output:        map[string]interface{}{"sshkeys": test_data_qemu.PublicKey_Encoded_Output()}},
				{name: `CloudInit UpgradePackages v7`,
					version:       Version{Major: 7, Minor: 255, Patch: 255},
					config:        &ConfigQemu{CloudInit: &CloudInit{UpgradePackages: util.Pointer(false)}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{UpgradePackages: util.Pointer(true)}}, // this is only possible with user error when using the advanced features
					output:        map[string]interface{}{}},
				{name: `CloudInit UpgradePackages v8`,
					version:       Version{Major: 8},
					config:        &ConfigQemu{CloudInit: &CloudInit{UpgradePackages: util.Pointer(false)}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{UpgradePackages: util.Pointer(true)}},
					output:        map[string]interface{}{"ciupgrade": 0}},
				{name: `CloudInit Username`,
					config:        &ConfigQemu{CloudInit: &CloudInit{Username: util.Pointer("root")}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Username: util.Pointer("admin")}},
					output:        map[string]interface{}{"ciuser": "root"}},
				{name: `CloudInit UserPassword`,
					config:        &ConfigQemu{CloudInit: &CloudInit{UserPassword: util.Pointer("Enter123!")}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{UserPassword: util.Pointer("Abc123!")}},
					output:        map[string]interface{}{"cipassword": "Enter123!"}},
			},
			create: []test{
				{name: `CloudInit Full v7`,
					version: Version{Major: 7, Minor: 255, Patch: 255},
					config: &ConfigQemu{CloudInit: &CloudInit{
						Custom: &CloudInitCustom{
							Meta: &CloudInitSnippet{
								Storage:  "local-zfs",
								FilePath: "ci-meta.yml"},
							Network: &CloudInitSnippet{
								Storage:  "local-lvm",
								FilePath: "ci-network.yml"},
							User: &CloudInitSnippet{
								Storage:  "folder",
								FilePath: "ci-user.yml"},
							Vendor: &CloudInitSnippet{
								Storage:  "local",
								FilePath: "snippets/ci-custom.yml"}},
						DNS: &GuestDNS{
							SearchDomain: util.Pointer("example.com"),
							NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
						NetworkInterfaces: CloudInitNetworkInterfaces{
							QemuNetworkInterfaceID0: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{DHCP: true},
								IPv6: &CloudInitIPv6Config{DHCP: true}},
							QemuNetworkInterfaceID19: CloudInitNetworkConfig{},
							QemuNetworkInterfaceID31: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}},
						PublicSSHkeys:   util.Pointer(test_data_qemu.PublicKey_Decoded_Input()),
						UpgradePackages: util.Pointer(false),
						UserPassword:    util.Pointer("Enter123!"),
						Username:        util.Pointer("root")}},
					output: map[string]interface{}{
						"cicustom":     "meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml",
						"searchdomain": "example.com",
						"nameserver":   "1.1.1.1 8.8.8.8 9.9.9.9",
						"ipconfig0":    "ip=dhcp,ip6=dhcp",
						"ipconfig31":   "ip=10.20.4.7/22",
						"sshkeys":      test_data_qemu.PublicKey_Encoded_Output(),
						"cipassword":   "Enter123!",
						"ciuser":       "root"}},
				{name: `CloudInit Full v8`,
					version: Version{Major: 8},
					config: &ConfigQemu{CloudInit: &CloudInit{
						Custom: &CloudInitCustom{
							Meta: &CloudInitSnippet{
								Storage:  "local-zfs",
								FilePath: "ci-meta.yml"},
							Network: &CloudInitSnippet{
								Storage:  "local-lvm",
								FilePath: "ci-network.yml"},
							User: &CloudInitSnippet{
								Storage:  "folder",
								FilePath: "ci-user.yml"},
							Vendor: &CloudInitSnippet{
								Storage:  "local",
								FilePath: "snippets/ci-custom.yml"}},
						DNS: &GuestDNS{
							SearchDomain: util.Pointer("example.com"),
							NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
						NetworkInterfaces: CloudInitNetworkInterfaces{
							QemuNetworkInterfaceID0: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{DHCP: true},
								IPv6: &CloudInitIPv6Config{DHCP: true}},
							QemuNetworkInterfaceID19: CloudInitNetworkConfig{},
							QemuNetworkInterfaceID31: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}},
						PublicSSHkeys:   util.Pointer(test_data_qemu.PublicKey_Decoded_Input()),
						UpgradePackages: util.Pointer(true),
						UserPassword:    util.Pointer("Enter123!"),
						Username:        util.Pointer("root")}},
					output: map[string]interface{}{
						"cicustom":     "meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml",
						"searchdomain": "example.com",
						"nameserver":   "1.1.1.1 8.8.8.8 9.9.9.9",
						"ipconfig0":    "ip=dhcp,ip6=dhcp",
						"ipconfig31":   "ip=10.20.4.7/22",
						"sshkeys":      test_data_qemu.PublicKey_Encoded_Output(),
						"ciupgrade":    1,
						"cipassword":   "Enter123!",
						"ciuser":       "root"}},
				{name: `CloudInit Custom Network`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Network: &CloudInitSnippet{
							Storage:  "local",
							FilePath: "ci-network.yml"}}}},
					output: map[string]interface{}{"cicustom": "network=local:ci-network.yml"}},
				{name: `CloudInit Custom User`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						User: &CloudInitSnippet{
							Storage:  "file",
							FilePath: "abcd.yml"}}}},
					output: map[string]interface{}{"cicustom": "user=file:abcd.yml"}},
				{name: `CloudInit Custom Vendor`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Vendor: &CloudInitSnippet{
							Storage:  "local",
							FilePath: "vendor-ci"}}}},
					output: map[string]interface{}{"cicustom": "vendor=local:vendor-ci"}},
				{name: `CloudInit Custom Meta`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Meta: &CloudInitSnippet{
							Storage:  "local-zfs",
							FilePath: "ci-meta.yml"}}}},
					output: map[string]interface{}{"cicustom": "meta=local-zfs:ci-meta.yml"}},
				{name: `CloudInit DNS NameServers empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{
						NameServers: &[]netip.Addr{}}}},
					output: map[string]interface{}{}},
				{name: `CloudInit DNS SearchDomain empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{SearchDomain: util.Pointer("")}}},
					output: map[string]interface{}{}},
				{name: `CloudInit NetworkInterfaces`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID1: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true},
							IPv6: &CloudInitIPv6Config{DHCP: true}},
						QemuNetworkInterfaceID20: CloudInitNetworkConfig{},
						QemuNetworkInterfaceID30: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}}}},
					output: map[string]interface{}{
						"ipconfig1":  "ip=dhcp,ip6=dhcp",
						"ipconfig30": "ip=10.20.4.7/22"}},
				{name: `CloudInit PublicSSHkeys empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{PublicSSHkeys: util.Pointer([]crypto.PublicKey{})}},
					output: map[string]interface{}{}},
				{name: `CloudInit Username empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{Username: util.Pointer("")}},
					output: map[string]interface{}{}},
				{name: `CloudInit UserPassword empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{UserPassword: util.Pointer("")}},
					output: map[string]interface{}{}}},
			update: []test{
				{name: `CloudInit Custom clear`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Meta:    &CloudInitSnippet{},
						Network: &CloudInitSnippet{},
						User:    &CloudInitSnippet{},
						Vendor:  &CloudInitSnippet{}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Custom: cloudInitCustom()}},
					output:        map[string]interface{}{"cicustom": ""}},
				{name: `CloudInit Custom Network`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Network: &CloudInitSnippet{
							Storage:  "newStorage",
							FilePath: "new.yml"}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Custom: cloudInitCustom()}},
					output:        map[string]interface{}{"cicustom": "meta=local-zfs:ci-meta.yml,network=newStorage:new.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml"}},
				{name: `CloudInit Custom User`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						User: &CloudInitSnippet{
							Storage:  "newStorage",
							FilePath: "new.yml"}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Custom: cloudInitCustom()}},
					output:        map[string]interface{}{"cicustom": "meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=newStorage:new.yml,vendor=local:snippets/ci-custom.yml"}},
				{name: `CloudInit Custom Vendor`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Vendor: &CloudInitSnippet{
							Storage:  "newStorage",
							FilePath: "new.yml"}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Custom: cloudInitCustom()}},
					output:        map[string]interface{}{"cicustom": "meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=newStorage:new.yml"}},
				{name: `CloudInit Custom Meta`,
					config: &ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{
						Meta: &CloudInitSnippet{
							Storage:  "newStorage",
							FilePath: "new.yml"}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Custom: cloudInitCustom()}},
					output:        map[string]interface{}{"cicustom": "meta=newStorage:new.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml"}},
				{name: `CloudInit DNS NameServers empty`,
					config: &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{
						NameServers: &[]netip.Addr{}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{
						NameServers: &[]netip.Addr{parseIP("8.8.8.8")}}}},
					output: map[string]interface{}{"delete": "nameserver"}},
				{name: `CloudInit DNS SearchDomain empty`,
					config:        &ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{SearchDomain: util.Pointer("")}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{DNS: &GuestDNS{SearchDomain: util.Pointer("example.org")}}},
					output:        map[string]interface{}{"delete": "searchdomain"}},
				{name: `CloudInit NetworkInterfaces Ipv4.Address update`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.1.10/24"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig0": "ip=192.168.1.10/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv4.Address remove`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID1: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID1: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig1": "gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv4.DHCP set`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID2: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID2: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig2": "ip=dhcp,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv4.Gateway update`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID3: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.1.1"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID3: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig3": "ip=192.168.56.30/24,gw=192.168.1.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv4.Gateway overwrite Ipv4.DHCP`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID4: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.1.1"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID4: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true}}}}},
					output: map[string]interface{}{"ipconfig4": "gw=192.168.1.1"}},
				{name: `CloudInit NetworkInterfaces Ipv4.Gateway remove`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID5: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID5: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig5": "ip=192.168.56.30/24,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Address update`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID6: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/48"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID6: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig6": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:85a3::/48,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Address remove`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID7: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID7: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig7": "ip=192.168.56.30/24,gw=192.168.56.1,gw6=2001:0db8:abcd::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.DHCP set`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID8: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{DHCP: true}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID8: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig8": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=dhcp"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Gateway update`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID9: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID9: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig9": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:85a3::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Gateway overwrite Ipv6.DHCP`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID10: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID10: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{DHCP: true}}}}},
					output: map[string]interface{}{"ipconfig10": "gw6=2001:0db8:85a3::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Gateway overwrite Ipv6.SLAAC`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID11: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID11: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{SLAAC: true}}}}},
					output: map[string]interface{}{"ipconfig11": "gw6=2001:0db8:85a3::1"}},
				{name: `CloudInit NetworkInterfaces Ipv6.Gateway remove`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID12: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID12: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig12": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48"}},
				{name: `CloudInit NetworkInterfaces Ipv6.SLAAC set`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID13: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{SLAAC: true}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID13: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"ipconfig13": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=auto"}},
				{name: `CloudInit NetworkInterfaces delete existing interface`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID14: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("")),
								Gateway: util.Pointer(IPv4Address(""))},
							IPv6: &CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("")),
								Gateway: util.Pointer(IPv6Address(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID14: cloudInitNetworkConfig()}}},
					output: map[string]interface{}{"delete": "ipconfig14"}},
				{name: `CloudInit NetworkInterfaces delete non-existing interface`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID20: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("")),
								Gateway: util.Pointer(IPv4Address(""))},
							IPv6: &CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("")),
								Gateway: util.Pointer(IPv6Address(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{}}},
					output:        map[string]interface{}{}},
				{name: `CloudInit NetworkInterfaces no updates`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID29: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID30: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true},
							IPv6: &CloudInitIPv6Config{DHCP: true}}}}},
					output: map[string]interface{}{}},
				{name: `CloudInit NetworkInterfaces full`,
					config: &ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.1.10/24"))}},
						QemuNetworkInterfaceID1: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))}},
						QemuNetworkInterfaceID2: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true}},
						QemuNetworkInterfaceID3: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.1.1"))}},
						QemuNetworkInterfaceID4: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.1.1"))}},
						QemuNetworkInterfaceID5: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))}},
						QemuNetworkInterfaceID6: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/48"))}},
						QemuNetworkInterfaceID7: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))}},
						QemuNetworkInterfaceID8: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{DHCP: true}},
						QemuNetworkInterfaceID9: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}},
						QemuNetworkInterfaceID10: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}},
						QemuNetworkInterfaceID11: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::1"))}},
						QemuNetworkInterfaceID12: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))}},
						QemuNetworkInterfaceID13: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{SLAAC: true}},
						QemuNetworkInterfaceID14: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("")),
								Gateway: util.Pointer(IPv4Address(""))},
							IPv6: &CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("")),
								Gateway: util.Pointer(IPv6Address(""))}},
						QemuNetworkInterfaceID20: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("")),
								Gateway: util.Pointer(IPv4Address(""))},
							IPv6: &CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("")),
								Gateway: util.Pointer(IPv6Address(""))}}}}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
						QemuNetworkInterfaceID0: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID1: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID2: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID3: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID4: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true}},
						QemuNetworkInterfaceID5: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID6: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID7: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID8: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID9: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID10: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{DHCP: true}},
						QemuNetworkInterfaceID11: CloudInitNetworkConfig{
							IPv6: &CloudInitIPv6Config{SLAAC: true}},
						QemuNetworkInterfaceID12: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID13: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID14: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID29: cloudInitNetworkConfig(),
						QemuNetworkInterfaceID30: CloudInitNetworkConfig{
							IPv4: &CloudInitIPv4Config{DHCP: true},
							IPv6: &CloudInitIPv6Config{DHCP: true}}}}},
					output: map[string]interface{}{
						"ipconfig0":  "ip=192.168.1.10/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1",
						"ipconfig1":  "gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1",
						"ipconfig2":  "ip=dhcp,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1",
						"ipconfig3":  "ip=192.168.56.30/24,gw=192.168.1.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1",
						"ipconfig4":  "gw=192.168.1.1",
						"ipconfig5":  "ip=192.168.56.30/24,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1",
						"ipconfig6":  "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:85a3::/48,gw6=2001:0db8:abcd::1",
						"ipconfig7":  "ip=192.168.56.30/24,gw=192.168.56.1,gw6=2001:0db8:abcd::1",
						"ipconfig8":  "ip=192.168.56.30/24,gw=192.168.56.1,ip6=dhcp",
						"ipconfig9":  "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:85a3::1",
						"ipconfig10": "gw6=2001:0db8:85a3::1",
						"ipconfig11": "gw6=2001:0db8:85a3::1",
						"ipconfig12": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48",
						"ipconfig13": "ip=192.168.56.30/24,gw=192.168.56.1,ip6=auto",
						"delete":     "ipconfig14"}},
				{name: `CloudInit PublicSSHkeys empty`,
					config:        &ConfigQemu{CloudInit: &CloudInit{PublicSSHkeys: util.Pointer([]crypto.PublicKey{})}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{PublicSSHkeys: util.Pointer([]crypto.PublicKey{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+0roY6F4yzq5RfA6V2+8gOgKlLOg9RtB1uGyTYvOMU6wxWUXVZP44+XozNxXZK4/MfPjCZLomqv78RlAedIQbqU8l6J9fdrrsRt6NknusE36UqD4HGPLX3Wn7svjSyNRfrjlk5BrBQ26rglLGlRSeD/xWvQ+5jLzzdo5NczszGkE9IQtrmKye7Gq7NQeGkHb1h0yGH7nMQ48WJ6ZKv1JG+GzFb8n4Qoei3zK9zpWxF+0AzF5u/zzCRZ4yU7FtfHgGRBDPze8oe3nVe+aO8MBH2dy8G/BRMXBdjWrSkaT9ZyeaT0k9SMjsCr9DQzUtVSOeqZZokpNU1dVglI+HU0vN test-key"})}},
					output:        map[string]interface{}{"delete": "sshkeys"}},
				{name: `CloudInit Username empty`,
					config:        &ConfigQemu{CloudInit: &CloudInit{Username: util.Pointer("")}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{Username: util.Pointer("admin")}},
					output:        map[string]interface{}{"delete": "ciuser"}},
				{name: `CloudInit UserPassword empty`,
					config:        &ConfigQemu{CloudInit: &CloudInit{UserPassword: util.Pointer("")}},
					currentConfig: ConfigQemu{CloudInit: &CloudInit{UserPassword: util.Pointer("Abc123!")}},
					output:        map[string]interface{}{"delete": "cipassword"}}}},
		{category: `Description`,
			create: []test{
				{name: `Description empty`,
					config: &ConfigQemu{Description: util.Pointer("")},
					output: map[string]interface{}{}}},
			createUpdate: []test{
				{name: `Description set`,
					config:        &ConfigQemu{Description: util.Pointer("test description")},
					currentConfig: ConfigQemu{Description: util.Pointer("old description")},
					output:        map[string]interface{}{"description": "test description"}}},
			update: []test{
				{name: `Description empty`,
					config:        &ConfigQemu{Description: util.Pointer("")},
					currentConfig: ConfigQemu{Description: util.Pointer("old description")},
					output:        map[string]interface{}{"description": ""}}}},
		{category: `Disks.Ide`,
			update: []test{
				{name: `Disk.Ide.Disk_X DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{}}}},
					output:        map[string]interface{}{"delete": "ide0"}}}},
		{category: `Disks.Ide.CdRom`,
			create: []test{
				{name: `Disks.Ide.Disk_X.CdRom none`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}},
					output: map[string]interface{}{"ide0": "none,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom.Iso`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output: map[string]interface{}{"ide1": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom.Passthrough`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output: map[string]interface{}{"ide2": "cdrom,media=cdrom"}}},
			update: []test{
				{name: `Disks.Ide.Disk_X.CdRom CHANGE ISO TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"ide1": "none,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom CHANGE ISO TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"ide2": "cdrom,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom CHANGE None TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"ide3": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom CHANGE None TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"ide0": "cdrom,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom CHANGE Passthrough TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"ide1": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom CHANGE Passthrough TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"ide2": "none,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{}}}},
					output:        map[string]interface{}{"delete": "ide3"}},
				{name: `Disks.Ide.Disk_X.CdRom SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.CdRom.Iso.File CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test2.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"ide1": "Test:iso/test2.iso,media=cdrom"}},
				{name: `Disks.Ide.Disk_X.CdRom.Iso.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "NewStorage"}}}}}},
					output:        map[string]interface{}{"ide2": "NewStorage:iso/test.iso,media=cdrom"}}}},
		{category: `Disks.Ide.CloudInit`,
			create: []test{
				{name: `Disks.Ide.Disk_X.CloudInit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					output: map[string]interface{}{"ide1": "Test:cloudinit,format=raw"}}},
			update: []test{
				{name: `Disks.Ide.Disk_X.CloudInit DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{}}}},
					output:        map[string]interface{}{"delete": "ide3"}},
				{name: `Disks.Ide.Disk_X.CloudInit SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: update_CloudInit()}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.CloudInit.Format CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Qcow2, Storage: "Test"}}}}},
					output:        map[string]interface{}{"ide1": "Test:cloudinit,format=qcow2"}},
				{name: `Disks.Ide.Disk_X.CloudInit.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "NewStorage"}}}}},
					output:        map[string]interface{}{"ide2": "NewStorage:cloudinit,format=raw"}}}},
		{category: `Disks.Ide.Disk`,
			create: []test{
				{name: `Disks.Ide.Disk_X.Disk All`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						EmulateSSD:      true,
						Format:          format_Raw,
						Replicate:       true,
						Serial:          "558485ef-478",
						SizeInKibibytes: 33554432,
						Storage:         "Test",
						WorldWideName:   "0x5000D31000C9876F"}}}}},
					output: map[string]interface{}{"ide0": "Test:32,aio=native,cache=directsync,discard=on,format=raw,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,serial=558485ef-478,ssd=1,wwn=0x5000D31000C9876F"}},
				{name: `Disks.Ide.Disk_X.Disk Create Gibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 33554432,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"ide0": "Test:32,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk Create Kibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 33554433,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"ide0": "Test:0.001,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{AsyncIO: QemuDiskAsyncIO_Native}}}}},
					output: map[string]interface{}{"ide1": ",aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Backup: true}}}}},
					output: map[string]interface{}{"ide2": ",replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Cache: QemuDiskCache_DirectSync}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,cache=directsync,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Discard: true}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{EmulateSSD: true}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Ide.Disk_X.Disk.Format`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: format_Raw}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Replicate: true}}}}},
					output: map[string]interface{}{"ide1": ",backup=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Serial: "558485ef-478"}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,replicate=0,serial=558485ef-478"}},
				{name: `Disks.Ide.Disk_X.Disk.Size`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{SizeInKibibytes: 32}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Storage`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Storage: "Test"}}}}},
					output: map[string]interface{}{"ide0": "Test:0,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{WorldWideName: "0x5001234000F876AB"}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0,wwn=0x5001234000F876AB"}}},
			update: []test{
				{name: `Disks.Ide.Disk_X.Disk CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide3": "test:0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide3": "test:100/base-100-disk-1.raw/0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide3": "test:vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							syntax:          diskSyntaxVolume,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide3": "test:base-100-disk-1/vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: ideBase()}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{}}}},
					output:        map[string]interface{}{"delete": "ide0"}},
				{name: `Disks.Ide.Disk_X.Disk MIGRATE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"ide1": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk MIGRATE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"ide1": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk MIGRATE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"ide1": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk MIGRATE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"ide1": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Gibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Gibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Gibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Gibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Kibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Kibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Kibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE DOWN Kibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE UP File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE UP File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE UP Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk RESIZE UP Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: ideBase()}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk.Format CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide1": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Format CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"ide1": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Disk.Format CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Ide.Disk_X.Disk.Format CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.Ide.Passthrough`,
			create: []test{
				{name: `Disks.Ide.Disk_X.Passthrough All`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:         QemuDiskCache_Unsafe,
						Discard:       true,
						EmulateSSD:    true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						Serial:        "test-serial_757465-gdg",
						WorldWideName: "0x500CBA2000D76543"}}}}},
					output: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,cache=unsafe,discard=on,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,serial=test-serial_757465-gdg,ssd=1,wwn=0x500CBA2000D76543"}},
				{name: `Disks.Ide.Disk_X.Passthrough.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{AsyncIO: QemuDiskAsyncIO_Threads}}}}},
					output: map[string]interface{}{"ide1": ",aio=threads,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Backup: true}}}}},
					output: map[string]interface{}{"ide2": ",replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"ide1": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Cache: QemuDiskCache_Unsafe}}}}},
					output: map[string]interface{}{"ide2": ",backup=0,cache=unsafe,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Discard: true}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{EmulateSSD: true}}}}},
					output: map[string]interface{}{"ide0": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Ide.Disk_X.Passthrough.File`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{File: "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}}}}},
					output: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Replicate: true}}}}},
					output: map[string]interface{}{"ide2": ",backup=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Serial: "test-serial_757465-gdg"}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0,serial=test-serial_757465-gdg"}},
				{name: `Disks.Ide.Disk_X.Passthrough.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{WorldWideName: "0x500FED1000B65432"}}}}},
					output: map[string]interface{}{"ide3": ",backup=0,replicate=0,wwn=0x500FED1000B65432"}}},
			update: []test{
				{name: `Disks.Ide.Disk_X.Passthrough CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						AsyncIO: QemuDiskAsyncIO_Native,
						File:    "/dev/disk/sda"}}}}},
					output: map[string]interface{}{"ide0": "/dev/disk/sda,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Ide.Disk_X.Passthrough SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						File: "/dev/disk/sda"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.Sata`,
			update: []test{
				{name: `Disks.Sata.Disk_X DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{}}}},
					output:        map[string]interface{}{"delete": "sata0"}}}},
		{category: `Disks.Sata.CdRom`,
			create: []test{
				{name: `Disks.Sata.Disk_X.CdRom none`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CdRom: &QemuCdRom{}}}}},
					output: map[string]interface{}{"sata0": "none,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom.Iso`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output: map[string]interface{}{"sata1": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom.Passthrough`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output: map[string]interface{}{"sata2": "cdrom,media=cdrom"}}},
			update: []test{
				{name: `Disks.Sata.Disk_X.CdRom CHANGE ISO TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"sata1": "none,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom CHANGE ISO TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"sata2": "cdrom,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom CHANGE None TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"sata3": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom CHANGE None TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"sata4": "cdrom,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom CHANGE Passthrough TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"sata5": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom CHANGE Passthrough TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"sata0": "none,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{}}}},
					output:        map[string]interface{}{"delete": "sata1"}},
				{name: `Disks.Sata.Disk_X.CdRom SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.CdRom.Iso.File CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test2.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"sata3": "Test:iso/test2.iso,media=cdrom"}},
				{name: `Disks.Sata.Disk_X.CdRom.Iso.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "NewStorage"}}}}}},
					output:        map[string]interface{}{"sata4": "NewStorage:iso/test.iso,media=cdrom"}}}},
		{category: `Disks.Sata.CloudInit`,
			create: []test{
				{name: `Disks.Sata.Disk_X.CloudInit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					output: map[string]interface{}{"sata1": "Test:cloudinit,format=raw"}}},
			update: []test{
				{name: `Disks.Sata.Disk_X.CloudInit DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{}}}},
					output:        map[string]interface{}{"delete": "sata5"}},
				{name: `Disks.Sata.Disk_X.CloudInit SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: update_CloudInit()}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.CloudInit.Format CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Qcow2, Storage: "Test"}}}}},
					output:        map[string]interface{}{"sata1": "Test:cloudinit,format=qcow2"}},
				{name: `Disks.Sata.Disk_X.CloudInit.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "NewStorage"}}}}},
					output:        map[string]interface{}{"sata2": "NewStorage:cloudinit,format=raw"}}}},
		{category: `Disks.Sata.Disk`,
			create: []test{
				{name: `Disks.Sata.Disk_X.Disk ALL`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:           QemuDiskCache_Unsafe,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Replicate:       true,
						Serial:          "ab_C-12_3",
						SizeInKibibytes: 16777216,
						Storage:         "Test",
						WorldWideName:   "0x5009876000A321DC"}}}}},
					output: map[string]interface{}{"sata0": "Test:16,aio=native,cache=unsafe,discard=on,format=qcow2,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,serial=ab_C-12_3,ssd=1,wwn=0x5009876000A321DC"}},
				{name: `Disks.Sata.Disk_X.Disk Create Gibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 16777216,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"sata0": "Test:16,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk Create Kibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 16777217,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"sata0": "Test:0.001,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{AsyncIO: QemuDiskAsyncIO_Native}}}}},
					output: map[string]interface{}{"sata0": ",aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Backup: true}}}}},
					output: map[string]interface{}{"sata1": ",replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Cache: QemuDiskCache_DirectSync}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,cache=directsync,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Discard: true}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{EmulateSSD: true}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Sata.Disk_X.Disk.Format`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Format: format_Raw}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,format=raw,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Replicate: true}}}}},
					output: map[string]interface{}{"sata2": ",backup=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Serial: "558485ef-478"}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,replicate=0,serial=558485ef-478"}},
				{name: `Disks.Sata.Disk_X.Disk.Size`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{SizeInKibibytes: 32}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Storage`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Storage: "Test"}}}}},
					output: map[string]interface{}{"sata5": "Test:0,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{WorldWideName: "0x500DCBA500E23456"}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,replicate=0,wwn=0x500DCBA500E23456"}}},
			update: []test{
				{name: `Disks.Sata.Disk_X.Disk CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:100/base-100-disk-1.raw/0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:base-100-disk-1/vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: sataBase()}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{}}}},
					output:        map[string]interface{}{"delete": "sata4"}},
				{name: `Disks.Sata.Disk_X.Disk MIGRATE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"sata5": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk MIGRATE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"sata5": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk MIGRATE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"sata5": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk MIGRATE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"sata5": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Gibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Gibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Gibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:9,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Gibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:9,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Kibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Kibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Kibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:0.001,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE DOWN Kibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata0": "test:0.001,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE UP File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE UP File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE UP Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk RESIZE UP Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: sataBase()}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk.Format CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Format CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"sata3": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Disk.Format CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Sata.Disk_X.Disk.Format CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.Sata.Passthrough`,
			create: []test{
				{name: `Disks.Sata.Disk_X.Passthrough All`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:         QemuDiskCache_Unsafe,
						Discard:       true,
						EmulateSSD:    true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						Serial:        "test-serial_757465-gdg",
						WorldWideName: "0x5007892000C4321A"}}}}},
					output: map[string]interface{}{"sata0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,cache=unsafe,discard=on,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,serial=test-serial_757465-gdg,ssd=1,wwn=0x5007892000C4321A"}},
				{name: `Disks.Sata.Disk_X.Passthrough.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{AsyncIO: QemuDiskAsyncIO_Threads}}}}},
					output: map[string]interface{}{"sata1": ",aio=threads,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Backup: true}}}}},
					output: map[string]interface{}{"sata2": ",replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"sata3": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"sata4": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Cache: QemuDiskCache_Unsafe}}}}},
					output: map[string]interface{}{"sata0": ",backup=0,cache=unsafe,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Discard: true}}}}},
					output: map[string]interface{}{"sata1": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{EmulateSSD: true}}}}},
					output: map[string]interface{}{"sata2": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Sata.Disk_X.Passthrough.File`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{File: "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}}}}},
					output: map[string]interface{}{"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Replicate: true}}}}},
					output: map[string]interface{}{"sata4": ",backup=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Serial: "test-serial_757465-gdg"}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,replicate=0,serial=test-serial_757465-gdg"}},
				{name: `Disks.Sata.Disk_X.Passthrough.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{WorldWideName: "0x5001ABE000987654"}}}}},
					output: map[string]interface{}{"sata5": ",backup=0,replicate=0,wwn=0x5001ABE000987654"}}},
			update: []test{
				{name: `Disks.Sata.Disk_X.Passthrough CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						AsyncIO: QemuDiskAsyncIO_Native,
						File:    "/dev/disk/sda"}}}}},
					output: map[string]interface{}{"sata0": "/dev/disk/sda,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Sata.Disk_X.Passthrough SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						File: "/dev/disk/sda"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.Scsi`,
			update: []test{
				{name: `Disks.Scsi.Disk_X DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{}}}},
					output:        map[string]interface{}{"delete": "scsi0"}}}},
		{category: `Disks.Scsi.CdRom`,
			create: []test{
				{name: `Disks.Scsi.CdRom none`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}},
					output: map[string]interface{}{"scsi0": "none,media=cdrom"}},
				{name: `Disks.Scsi.CdRom.Iso`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output: map[string]interface{}{"scsi1": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Scsi.CdRom.Passthrough`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output: map[string]interface{}{"scsi2": "cdrom,media=cdrom"}}},
			update: []test{
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE ISO TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"scsi1": "none,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE ISO TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"scsi2": "cdrom,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE None TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"scsi3": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE None TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"scsi4": "cdrom,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE Passthrough TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"scsi5": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom CHANGE Passthrough TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"scsi6": "none,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{}}}},
					output:        map[string]interface{}{"delete": "scsi7"}},
				{name: `Disks.Scsi.Disk_X.CdRom SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.CdRom.Iso.File CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test2.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"scsi9": "Test:iso/test2.iso,media=cdrom"}},
				{name: `Disks.Scsi.Disk_X.CdRom.Iso.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "NewStorage"}}}}}},
					output:        map[string]interface{}{"scsi10": "NewStorage:iso/test.iso,media=cdrom"}}}},
		{category: `Disks.Scsi.CloudInit`,
			create: []test{
				{name: `Disks.Scsi.CloudInit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					output: map[string]interface{}{"scsi1": "Test:cloudinit,format=raw"}}},
			update: []test{
				{name: `Disks.Scsi.Disk_X.CloudInit DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{}}}},
					output:        map[string]interface{}{"delete": "scsi11"}},
				{name: `Disks.Scsi.Disk_X.CloudInit SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{CloudInit: update_CloudInit()}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.CloudInit.Format CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Qcow2, Storage: "Test"}}}}},
					output:        map[string]interface{}{"scsi13": "Test:cloudinit,format=qcow2"}},
				{name: `Disks.Scsi.Disk_X.CloudInit.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "NewStorage"}}}}},
					output:        map[string]interface{}{"scsi14": "NewStorage:cloudinit,format=raw"}}}},
		{category: `Disks.Scsi.Disk`,
			create: []test{
				{name: `Disks.Scsi.Disk_X.Disk All`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						EmulateSSD:      true,
						Format:          format_Raw,
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       true,
						Serial:          "558485ef-478",
						SizeInKibibytes: 76546048,
						Storage:         "Test",
						WorldWideName:   "0x500D567800BAC321"}}}}},
					output: map[string]interface{}{"scsi0": "Test:73,aio=native,cache=directsync,discard=on,format=raw,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,iothread=1,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,ro=1,serial=558485ef-478,ssd=1,wwn=0x500D567800BAC321"}},
				{name: `Disks.Scsi.Disk_X.Disk Create Gibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
						SizeInKibibytes: 76546048,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"scsi0": "Test:73,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk Create Kibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
						SizeInKibibytes: 76546049,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"scsi0": "Test:0.001,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{AsyncIO: QemuDiskAsyncIO_Native}}}}},
					output: map[string]interface{}{"scsi1": ",aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Backup: true}}}}},
					output: map[string]interface{}{"scsi2": ",replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"scsi3": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi11": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi12": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"scsi13": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"scsi13": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"scsi14": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi15": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"scsi16": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"scsi16": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"scsi17": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"scsi4": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi5": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"scsi6": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"scsi7": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi8": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"scsi9": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"scsi10": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{Cache: QemuDiskCache_DirectSync}}}}},
					output: map[string]interface{}{"scsi18": ",backup=0,cache=directsync,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{Discard: true}}}}},
					output: map[string]interface{}{"scsi19": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_20: &QemuScsiStorage{Disk: &QemuScsiDisk{EmulateSSD: true}}}}},
					output: map[string]interface{}{"scsi20": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Scsi.Disk_X.Disk.Format`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_20: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: format_Raw}}}}},
					output: map[string]interface{}{"scsi20": ",backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.IOThread`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{IOThread: true}}}}},
					output: map[string]interface{}{"scsi21": ",backup=0,iothread=1,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.ReadOnly`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_22: &QemuScsiStorage{Disk: &QemuScsiDisk{ReadOnly: true}}}}},
					output: map[string]interface{}{"scsi22": ",backup=0,replicate=0,ro=1"}},
				{name: `Disks.Scsi.Disk_X.Disk.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_23: &QemuScsiStorage{Disk: &QemuScsiDisk{Replicate: true}}}}},
					output: map[string]interface{}{"scsi23": ",backup=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_24: &QemuScsiStorage{Disk: &QemuScsiDisk{Serial: "558485ef-478"}}}}},
					output: map[string]interface{}{"scsi24": ",backup=0,replicate=0,serial=558485ef-478"}},
				{name: `Disks.Scsi.Disk_X.Disk.Size`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_25: &QemuScsiStorage{Disk: &QemuScsiDisk{SizeInKibibytes: 32}}}}},
					output: map[string]interface{}{"scsi25": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Storage`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_26: &QemuScsiStorage{Disk: &QemuScsiDisk{Storage: "Test"}}}}},
					output: map[string]interface{}{"scsi26": "Test:0,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_27: &QemuScsiStorage{Disk: &QemuScsiDisk{WorldWideName: "0x500EF32100D76589"}}}}},
					output: map[string]interface{}{"scsi27": ",backup=0,replicate=0,wwn=0x500EF32100D76589"}}},
			update: []test{
				{name: `Disks.Scsi.Disk_X.Disk CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi15": "test:0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi15": "test:100/base-100-disk-1.raw/0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi15": "test:vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi15": "test:base-100-disk-1/vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: scsiBase()}}},
					config:        &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{}}}},
					output:        map[string]interface{}{"delete": "scsi16"}},
				{name: `Disks.Scsi.Disk_X.Disk MIGRATE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"scsi17": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk MIGRATE File Linked Clone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"scsi17": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk MIGRATE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"scsi17": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk MIGRATE Volume Linked Clone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"scsi17": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Gibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Gibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Gibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Gibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Kibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Kibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Kibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE DOWN Kibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi18": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE UP File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE UP File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE UP Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk RESIZE UP Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_20: scsiBase()}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_20: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk.Format CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi21": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Format CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"scsi21": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Disk.Format CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.Scsi.Disk_X.Disk.Format CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.Scsi.Passthrough`,
			create: []test{
				{name: `Disks.Scsi.Disk_X.Passthrough All`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:         QemuDiskCache_Unsafe,
						Discard:       true,
						EmulateSSD:    true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:      true,
						ReadOnly:      true,
						Replicate:     true,
						Serial:        "test-serial_757465-gdg",
						WorldWideName: "0x500BCA3000F09876"}}}}},
					output: map[string]interface{}{"scsi0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,cache=unsafe,discard=on,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,iothread=1,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,ro=1,serial=test-serial_757465-gdg,ssd=1,wwn=0x500BCA3000F09876"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{AsyncIO: QemuDiskAsyncIO_Threads}}}}},
					output: map[string]interface{}{"scsi1": ",aio=threads,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Backup: true}}}}},
					output: map[string]interface{}{"scsi2": ",replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"scsi3": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"scsi11": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi12": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"scsi13": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"scsi13": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"scsi14": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi15": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"scsi16": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"scsi16": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"scsi17": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"scsi4": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi5": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"scsi6": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"scsi7": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"scsi8": ",backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"scsi9": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"scsi10": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Cache: QemuDiskCache_Unsafe}}}}},
					output: map[string]interface{}{"scsi18": ",backup=0,cache=unsafe,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Discard: true}}}}},
					output: map[string]interface{}{"scsi19": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.EmulateSSD`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_20: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{EmulateSSD: true}}}}},
					output: map[string]interface{}{"scsi20": ",backup=0,replicate=0,ssd=1"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.File`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_21: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{File: "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}}}}},
					output: map[string]interface{}{"scsi21": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.IOThread`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_22: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{IOThread: true}}}}},
					output: map[string]interface{}{"scsi22": ",backup=0,iothread=1,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.ReadOnly`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_23: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{ReadOnly: true}}}}},
					output: map[string]interface{}{"scsi23": ",backup=0,replicate=0,ro=1"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_24: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Replicate: true}}}}},
					output: map[string]interface{}{"scsi24": ",backup=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_25: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Serial: "test-serial_757465-gdg"}}}}},
					output: map[string]interface{}{"scsi25": ",backup=0,replicate=0,serial=test-serial_757465-gdg"}},
				{name: `Disks.Scsi.Disk_X.Passthrough.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_25: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{WorldWideName: "0x5004DC0100E239C7"}}}}},
					output: map[string]interface{}{"scsi25": ",backup=0,replicate=0,wwn=0x5004DC0100E239C7"}}},
			update: []test{
				{name: `Disks.Scsi.Disk_X.Passthrough CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						AsyncIO: QemuDiskAsyncIO_Native,
						File:    "/dev/disk/sda"}}}}},
					output: map[string]interface{}{"scsi0": "/dev/disk/sda,aio=native,backup=0,replicate=0"}},
				{name: `Disks.Scsi.Disk_X.Passthrough SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						File: "/dev/disk/sda"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.VirtIO`,
			update: []test{
				{name: `Disks.VirtIO.Disk_X DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{}}}},
					output:        map[string]interface{}{"delete": "virtio0"}}}},
		{category: `Disks.VirtIO.CdRom`,
			create: []test{
				{name: `Disks.VirtIO.Disk_X.CdRom none`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}},
					output: map[string]interface{}{"virtio0": "none,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom.Iso`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output: map[string]interface{}{"virtio1": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom.Passthrough`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output: map[string]interface{}{"virtio2": "cdrom,media=cdrom"}}},
			update: []test{
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE ISO TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"virtio1": "none,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE ISO TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"virtio2": "cdrom,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE None TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"virtio3": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE None TO Passthrough`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{"virtio4": "cdrom,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE Passthrough TO ISO`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"virtio5": "Test:iso/test.iso,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom CHANGE Passthrough TO None`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}},
					output:        map[string]interface{}{"virtio6": "none,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{}}}},
					output:        map[string]interface{}{"delete": "virtio7"}},
				{name: `Disks.VirtIO.Disk_X.CdRom SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.CdRom.Iso.File CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test2.iso", Storage: "Test"}}}}}},
					output:        map[string]interface{}{"virtio9": "Test:iso/test2.iso,media=cdrom"}},
				{name: `Disks.VirtIO.Disk_X.CdRom.Iso.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "Test"}}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test.iso", Storage: "NewStorage"}}}}}},
					output:        map[string]interface{}{"virtio10": "NewStorage:iso/test.iso,media=cdrom"}}}},
		{category: `Disks.VirtIO.CloudInit`,
			create: []test{
				{name: `Disks.VirtIO.Disk_X.CloudInit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					output: map[string]interface{}{"virtio1": "Test:cloudinit,format=raw"}}},
			update: []test{
				{name: `Disks.VirtIO.Disk_X.CloudInit DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{}}}},
					output:        map[string]interface{}{"delete": "virtio11"}},
				{name: `Disks.VirtIO.Disk_X.CloudInit SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{CloudInit: update_CloudInit()}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{CloudInit: update_CloudInit()}}}},
					output:        map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.CloudInit.Format CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Qcow2, Storage: "Test"}}}}},
					output:        map[string]interface{}{"virtio13": "Test:cloudinit,format=qcow2"}},
				{name: `Disks.VirtIO.Disk_X.CloudInit.Storage CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "Test"}}}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: format_Raw, Storage: "NewStorage"}}}}},
					output:        map[string]interface{}{"virtio14": "NewStorage:cloudinit,format=raw"}}}},
		{category: `Disks.VirtIO.Disk`,
			create: []test{
				{name: `Disks.VirtIO.Disk_X.Disk All`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						Format:          format_Raw,
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       true,
						Serial:          "558485ef-478",
						SizeInKibibytes: 8238661632,
						Storage:         "Test",
						WorldWideName:   "0x500A7B0800F345D2"}}}}},
					output: map[string]interface{}{"virtio0": "Test:7857,aio=native,cache=directsync,discard=on,format=raw,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,iothread=1,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,ro=1,serial=558485ef-478,wwn=0x500A7B0800F345D2"}},
				{name: `Disks.VirtIO.Disk_X.Disk Create Gibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 8238661632,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"virtio0": "Test:7857,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk Create Kibibyte`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 8238661633,
						Storage:         "Test"}}}}},
					output: map[string]interface{}{"virtio0": "Test:0.001,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{AsyncIO: QemuDiskAsyncIO_Native}}}}},
					output: map[string]interface{}{"virtio1": ",aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Backup: true}}}}},
					output: map[string]interface{}{"virtio2": ",replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"virtio3": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"virtio11": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio12": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"virtio13": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"virtio13": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"virtio14": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio15": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"virtio0": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"virtio0": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"virtio1": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"virtio4": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio5": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 1}}}}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0,iops_rd_max_length=1,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"virtio7": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio8": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"virtio9": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}}}}}}},
					output: map[string]interface{}{"virtio9": ",backup=0,iops_wr_max_length=2,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"virtio10": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Cache: QemuDiskCache_DirectSync}}}}},
					output: map[string]interface{}{"virtio2": ",backup=0,cache=directsync,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Discard: true}}}}},
					output: map[string]interface{}{"virtio3": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Format`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: format_Raw}}}}},
					output: map[string]interface{}{"virtio4": ",backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.IOThread`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{IOThread: true}}}}},
					output: map[string]interface{}{"virtio4": ",backup=0,iothread=1,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.ReadOnly`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{ReadOnly: true}}}}},
					output: map[string]interface{}{"virtio5": ",backup=0,replicate=0,ro=1"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Replicate: true}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Serial: "558485ef-478"}}}}},
					output: map[string]interface{}{"virtio7": ",backup=0,replicate=0,serial=558485ef-478"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Size`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{SizeInKibibytes: 32}}}}},
					output: map[string]interface{}{"virtio8": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Storage`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Storage: "Test"}}}}},
					output: map[string]interface{}{"virtio9": "Test:0,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{WorldWideName: "0x5005FED000B87632"}}}}},
					output: map[string]interface{}{"virtio10": ",backup=0,replicate=0,wwn=0x5005FED000B87632"}}},
			update: []test{
				{name: `Disks.VirtIO.Disk_X.Disk CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio15": "test:0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio15": "test:100/base-100-disk-1.raw/0/vm-0-disk-23.raw,aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_IOuring,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio15": "test:vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:         QemuDiskAsyncIO_Native,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio15": "test:base-100-disk-1/vm-0-disk-23,aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk DELETE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: virtioBase()}}},
					config:        &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{}}}},
					output:        map[string]interface{}{"delete": "virtio0"}},
				{name: `Disks.VirtIO.Disk_X.Disk MIGRATE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"virtio1": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk MIGRATE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"virtio1": "test2:0/vm-0-disk-23.raw,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk MIGRATE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test1",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"virtio1": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk MIGRATE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test1",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 10,
						Storage:         "test2"}}}}},
					output: map[string]interface{}{"virtio1": "test2:vm-0-disk-23,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Gibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Gibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Gibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Id:              23,
						SizeInKibibytes: 9437185,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Gibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437185,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437184,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:9,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Kibibyte File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Kibibyte File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Kibibyte Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Id:              23,
						SizeInKibibytes: 9437186,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE DOWN Kibibyte Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 9437186,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 9437185,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio2": "test:0.001,backup=0,format=raw,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE UP File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE UP File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE UP Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk RESIZE UP Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						SizeInKibibytes: 11,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: virtioBase()}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk.Format CHANGE File`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio5": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Format CHANGE File LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{"virtio5": "test:0/vm-0-disk-23.qcow2,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Disk.Format CHANGE Volume`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Raw,
						Id:              23,
						SizeInKibibytes: 10,
						Storage:         "test",
						syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}},
				{name: `Disks.VirtIO.Disk_X.Disk.Format CHANGE Volume LinkedClone`,
					currentConfig: ConfigQemu{
						LinkedVmId: 100,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							Id:              23,
							LinkedDiskId:    &uint1,
							SizeInKibibytes: 10,
							Storage:         "test",
							syntax:          diskSyntaxVolume}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Format:          QemuDiskFormat_Qcow2,
						SizeInKibibytes: 10,
						Storage:         "test"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Disks.VirtIO.Passthrough`,
			create: []test{
				{name: `Disks.VirtIO.Disk_X.Passthrough All`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: float99, Concurrent: float10},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79, Concurrent: float45}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: uint78, BurstDuration: 3, Concurrent: uint34},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89, BurstDuration: 4, Concurrent: uint23}}},
						Cache:         QemuDiskCache_Unsafe,
						Discard:       true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:      true,
						ReadOnly:      true,
						Replicate:     true,
						Serial:        "test-serial_757465-gdg",
						WorldWideName: "0x500C329500A1EFAB"}}}}},
					output: map[string]interface{}{"virtio0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,cache=unsafe,discard=on,iops_rd=34,iops_rd_max=78,iops_rd_max_length=3,iops_wr=23,iops_wr_max=89,iops_wr_max_length=4,iothread=1,mbps_rd=10.3,mbps_rd_max=99.2,mbps_wr=45.23,mbps_wr_max=79.23,ro=1,serial=test-serial_757465-gdg,wwn=0x500C329500A1EFAB"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.AsyncIO`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{AsyncIO: QemuDiskAsyncIO_Threads}}}}},
					output: map[string]interface{}{"virtio1": ",aio=threads,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Backup`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Backup: true}}}}},
					output: map[string]interface{}{"virtio2": ",replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{}}}}}},
					output: map[string]interface{}{"virtio3": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{}}}}}}},
					output: map[string]interface{}{"virtio11": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio12": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: uint78}}}}}}}},
					output: map[string]interface{}{"virtio13": ",backup=0,iops_rd_max=78,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}}}}}}},
					output: map[string]interface{}{"virtio13": ",backup=0,iops_rd_max_length=3,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint34}}}}}}}},
					output: map[string]interface{}{"virtio14": ",backup=0,iops_rd=34,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio15": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: uint89}}}}}}}},
					output: map[string]interface{}{"virtio0": ",backup=0,iops_wr_max=89,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.BurstDuration`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 4}}}}}}}},
					output: map[string]interface{}{"virtio0": ",backup=0,iops_wr_max_length=4,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.Iops.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: uint23}}}}}}}},
					output: map[string]interface{}{"virtio1": ",backup=0,iops_wr=23,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{}}}}}}},
					output: map[string]interface{}{"virtio4": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio5": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: float99}}}}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0,mbps_rd_max=99.2,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.ReadLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float10}}}}}}}},
					output: map[string]interface{}{"virtio7": ",backup=0,mbps_rd=10.3,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{}}}}}}}},
					output: map[string]interface{}{"virtio8": ",backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Burst`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: float79}}}}}}}},
					output: map[string]interface{}{"virtio9": ",backup=0,mbps_wr_max=79.23,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Bandwidth.MBps.WriteLimit.Concurrent`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: float45}}}}}}}},
					output: map[string]interface{}{"virtio10": ",backup=0,mbps_wr=45.23,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Cache`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Cache: QemuDiskCache_Unsafe}}}}},
					output: map[string]interface{}{"virtio2": ",backup=0,cache=unsafe,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Discard`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Discard: true}}}}},
					output: map[string]interface{}{"virtio3": ",backup=0,discard=on,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.File`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{File: "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"}}}}},
					output: map[string]interface{}{"virtio4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.IOThread`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{IOThread: true}}}}},
					output: map[string]interface{}{"virtio5": ",backup=0,iothread=1,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.ReadOnly`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{ReadOnly: true}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0,replicate=0,ro=1"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Replicate`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Replicate: true}}}}},
					output: map[string]interface{}{"virtio6": ",backup=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.Serial`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Serial: "test-serial_757465-gdg"}}}}},
					output: map[string]interface{}{"virtio7": ",backup=0,replicate=0,serial=test-serial_757465-gdg"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough.WorldWideName`,
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{WorldWideName: "0x500D41A600C67853"}}}}},
					output: map[string]interface{}{"virtio8": ",backup=0,replicate=0,wwn=0x500D41A600C67853"}}},
			update: []test{
				{name: `Disks.VirtIO.Disk_X.Passthrough CHANGE`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						AsyncIO: QemuDiskAsyncIO_Native,
						File:    "/dev/disk/sda"}}}}},
					output: map[string]interface{}{"virtio0": "/dev/disk/sda,aio=native,backup=0,replicate=0"}},
				{name: `Disks.VirtIO.Disk_X.Passthrough SAME`,
					currentConfig: ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						File: "/dev/disk/sda"}}}}},
					config: &ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						File: "/dev/disk/sda"}}}}},
					output: map[string]interface{}{}}}},
		{category: `Iso`,
			create: []test{
				{name: `Iso`,
					config: &ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					output: map[string]interface{}{"ide2": "test:iso/file.iso,media=cdrom"}}},
			update: []test{
				{name: `Iso nil`,
					currentConfig: ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					config:        &ConfigQemu{Iso: nil},
					output:        map[string]interface{}{}},
				{name: `Iso SAME`,
					currentConfig: ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					config:        &ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					output:        map[string]interface{}{"ide2": "test:iso/file.iso,media=cdrom"}},
				{name: `Iso.File`,
					currentConfig: ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					config:        &ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file2.iso"}},
					output:        map[string]interface{}{"ide2": "test:iso/file2.iso,media=cdrom"}},
				{name: `Iso.Storage`,
					currentConfig: ConfigQemu{Iso: &IsoFile{Storage: "test", File: "file.iso"}},
					config:        &ConfigQemu{Iso: &IsoFile{Storage: "NewStorage", File: "file.iso"}},
					output:        map[string]interface{}{"ide2": "NewStorage:iso/file.iso,media=cdrom"}}}},
		{category: `Memory`,
			create: []test{
				{name: `MinimumCapacityMiB`,
					config: &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1024))}},
					output: map[string]interface{}{
						"memory":  QemuMemoryBalloonCapacity(1024),
						"balloon": QemuMemoryBalloonCapacity(1024)}},
				{name: `Shares`,
					config: &ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(40000))}},
					output: map[string]interface{}{"shares": QemuMemoryShares(40000)}},
				{name: `Shares 0`,
					config: &ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(0))}},
					output: map[string]interface{}{}}},
			createUpdate: []test{
				{name: `CapacityMiB`,
					config:        &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(2048))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1024))}},
					output:        map[string]interface{}{"memory": QemuMemoryCapacity(2048)}}},
			update: []test{
				{name: `CapacityMiB smaller then current MinimumCapacityMiB`,
					config:        &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1024))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(2048)), MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2048))}},
					output:        map[string]interface{}{"memory": QemuMemoryCapacity(1024), "balloon": QemuMemoryCapacity(1024), "delete": "shares"}},
				{name: `CapacityMiB smaller then current MinimumCapacityMiB and MinimumCapacityMiB lowered`,
					config:        &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1024)), MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(512))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(2048)), MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2048))}},
					output:        map[string]interface{}{"memory": QemuMemoryCapacity(1024), "balloon": QemuMemoryBalloonCapacity(512)}},
				{name: `MinimumCapacityMiB`,
					config:        &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1024))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2048))}},
					output:        map[string]interface{}{"balloon": QemuMemoryBalloonCapacity(1024)}},
				{name: `MinimumCapacityMiB 0`,
					config:        &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(0))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1024))}},
					output:        map[string]interface{}{"balloon": QemuMemoryBalloonCapacity(0), "delete": "shares"}},
				{name: `Shares`,
					config:        &ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(40000))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(20000))}},
					output:        map[string]interface{}{"shares": QemuMemoryShares(40000)}},
				{name: `Shares 0`,
					config:        &ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(0))}},
					currentConfig: ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(20000))}},
					output:        map[string]interface{}{"delete": "shares"}}}},
		{category: `Networks`,
			create: []test{
				{name: `Delete`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
						Bridge:        util.Pointer("vmbr0"),
						Connected:     util.Pointer(true),
						Delete:        true,
						Firewall:      util.Pointer(true),
						MAC:           util.Pointer(net.HardwareAddr("00:11:22:33:44:55")),
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						MultiQueue:    util.Pointer(QemuNetworkQueue(4)),
						RateLimitKBps: util.Pointer(QemuNetworkRate(45)),
						NativeVlan:    util.Pointer(Vlan(23))}}},
					output: map[string]interface{}{}}},
			createUpdate: []test{
				{name: `Bridge`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Bridge: util.Pointer("vmbr0")}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Bridge: util.Pointer("vmbr1")}}},
					output:        map[string]interface{}{"net0": ",bridge=vmbr0"}},
				{name: `Connected true`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{Connected: util.Pointer(true)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{Connected: util.Pointer(false)}}},
					output:        map[string]interface{}{"net1": ""}},
				{name: `Connected false`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{Connected: util.Pointer(false)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{Connected: util.Pointer(true)}}},
					output:        map[string]interface{}{"net2": ",link_down=1"}},
				{name: `Firewall true`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{Firewall: util.Pointer(true)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{Firewall: util.Pointer(false)}}},
					output:        map[string]interface{}{"net3": ",firewall=1"}},
				{name: `Firewall false`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{Firewall: util.Pointer(false)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{Firewall: util.Pointer(true)}}},
					output:        map[string]interface{}{"net4": ""}},
				{name: `MAC`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
						Model: util.Pointer(QemuNetworkModelE1000),
						MAC:   util.Pointer(net.HardwareAddr(parseMAC("BC:11:22:33:44:55")))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
						Model: util.Pointer(QemuNetworkModelVirtIO),
						MAC:   util.Pointer(net.HardwareAddr(parseMAC("bc:11:22:33:44:56")))}}},
					output: map[string]interface{}{"net5": "e1000=BC:11:22:33:44:55"}},
				{name: `MTU.Inherit model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
						Model: util.Pointer(QemuNetworkModelVirtIO),
						MTU:   util.Pointer(QemuMTU{Inherit: true})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{Value: MTU(1500)})}}},
					output:        map[string]interface{}{"net6": "virtio,mtu=1"}},
				{name: `MTU.Value model=none`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{Value: MTU(1400)})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{Value: MTU(1500)})}}},
					output:        map[string]interface{}{"net7": ""}},
				{name: `MTU.Value model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
						Model: util.Pointer(QemuNetworkModelVirtIO),
						MTU:   util.Pointer(QemuMTU{Value: MTU(1400)})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{Value: MTU(1500)})}}},
					output:        map[string]interface{}{"net7": "virtio,mtu=1400"}},
				{name: `MTU.Value=0 model=virtio`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
						Model: util.Pointer(QemuNetworkModelVirtIO),
						MTU:   util.Pointer(QemuMTU{Value: MTU(0)})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{})}}},
					output:        map[string]interface{}{"net8": "virtio"}},
				{name: `Model`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{Model: util.Pointer(qemuNetworkModelE100082544gc_Lower)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{Model: util.Pointer(QemuNetworkModelVirtIO)}}},
					output:        map[string]interface{}{"net9": "e1000-82544gc"}},
				{name: `Model invalid`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{Model: util.Pointer(QemuNetworkModel("gibberish"))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{Model: util.Pointer(QemuNetworkModelVirtIO)}}},
					output:        map[string]interface{}{"net10": ""}},
				{name: `MultiQueue set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(4))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(2))}}},
					output:        map[string]interface{}{"net11": ",queues=4"}},
				{name: `MultiQueue unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(0))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(2))}}},
					output:        map[string]interface{}{"net12": ""}},
				{name: `RateLimitKBps 0`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(0))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net13": ""}},
				{name: `RateLimitKBps 0.007`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(7))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net13": ",rate=0.007"}},
				{name: `RateLimitKBps 0.07`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(70))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net14": ",rate=0.07"}},
				{name: `RateLimitKBps 0.7`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(700))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net15": ",rate=0.7"}},
				{name: `RateLimitKBps 7`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(7000))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net16": ",rate=7"}},
				{name: `RateLimitKBps 7.546`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(7546))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net17": ",rate=7.546"}},
				{name: `RateLimitKBps 734.546`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(734546))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(5))}}},
					output:        map[string]interface{}{"net18": ",rate=734.546"}},
				{name: `NativeVlan unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(0))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(2))}}},
					output:        map[string]interface{}{"net19": ""}},
				{name: `NativeVlan set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(83))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(2))}}},
					output:        map[string]interface{}{"net20": ",tag=83"}},
				{name: `TaggedVlans set`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{10, 43, 23})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{12, 56})}}},
					output:        map[string]interface{}{"net21": ",trunks=10;43;23"}},
				{name: `TaggedVlans unset`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{12, 56})}}},
					output:        map[string]interface{}{"net22": ""}}},
			update: []test{
				{name: `Bridge replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: QemuNetworkInterface{Bridge: util.Pointer("vmbr45")}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: networkInterface()}},
					output:        map[string]interface{}{"net31": "virtio=52:54:00:12:34:56,bridge=vmbr45,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `Connected replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID30: QemuNetworkInterface{Connected: util.Pointer(true)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID30: networkInterface()}},
					output:        map[string]interface{}{"net30": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `Firewall replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID29: QemuNetworkInterface{Firewall: util.Pointer(false)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID29: networkInterface()}},
					output:        map[string]interface{}{"net29": "virtio=52:54:00:12:34:56,bridge=vmbr0,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `MAC replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{MAC: util.Pointer(net.HardwareAddr(parseMAC("BC:24:11:C2:75:20")))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: networkInterface()}},
					output:        map[string]interface{}{"net28": "virtio=BC:24:11:C2:75:20,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `MAC replace binary match`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{MAC: util.Pointer(net.HardwareAddr(parseMAC("BC:24:11:C2:75:20")))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{
						MAC: util.Pointer(net.HardwareAddr(parseMAC("bc:24:11:C2:75:20"))),
						mac: "bc:24:11:C2:75:20"}}},
					output: map[string]interface{}{"net28": "=bc:24:11:C2:75:20"}},
				{name: `MAC no update mixed case`,
					config: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID28: QemuNetworkInterface{
						MAC: util.Pointer(net.HardwareAddr(parseMAC("bc:24:11:C2:75:20"))),
						mac: "bc:24:11:C2:75:20"}}},
					output: map[string]interface{}{"net28": "=bc:24:11:C2:75:20"}},
				{name: `MTU replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID27: QemuNetworkInterface{MTU: util.Pointer(QemuMTU{Value: MTU(1400)})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID27: networkInterface()}},
					output:        map[string]interface{}{"net27": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1400,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `Model replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID26: QemuNetworkInterface{Model: util.Pointer(qemuNetworkModelE100082544gc_Lower)}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID26: networkInterface()}},
					output:        map[string]interface{}{"net26": "e1000-82544gc=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,queues=5,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `MultiQueue replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID25: QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(4))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID25: networkInterface()}},
					output:        map[string]interface{}{"net25": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=4,rate=0.045,tag=23,trunks=12;23;45"}},
				{name: `RateLimitKBps replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID24: QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(539))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID24: networkInterface()}},
					output:        map[string]interface{}{"net24": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.539,tag=23,trunks=12;23;45"}},
				{name: `NaitiveVlan replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID23: QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(0))}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID23: networkInterface()}},
					output:        map[string]interface{}{"net23": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,trunks=12;23;45"}},
				{name: `TaggedVlans replace`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{10, 70, 18})}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID22: networkInterface()}},
					output:        map[string]interface{}{"net22": "virtio=52:54:00:12:34:56,bridge=vmbr0,firewall=1,link_down=1,mtu=1500,queues=5,rate=0.045,tag=23,trunks=10;70;18"}},
				{name: `Delete`,
					config:        &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{Delete: true}}},
					currentConfig: ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{}}},
					output:        map[string]interface{}{"delete": "net21"}}}},
		{category: `PciDevices`,
			create: []test{
				{name: `Delete`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Delete: true}}},
					output: map[string]interface{}{}}},
			createUpdate: []test{
				{name: `Mapping.DeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Mapping: &QemuPciMapping{
							DeviceID: util.Pointer(PciDeviceID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{Mapping: &QemuPciMapping{
							DeviceID: util.Pointer(PciDeviceID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci15": "mapping=,rombar=0,device-id=0x8086"}},
				{name: `Mapping.ID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							ID: util.Pointer(ResourceMappingPciID("aaaaa"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{Mapping: &QemuPciMapping{
							ID: util.Pointer(ResourceMappingPciID("bbbbb"))}}}},
					output: map[string]interface{}{"hostpci14": "mapping=aaaaa,rombar=0"}},
				{name: `Mapping.Pci`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{Mapping: &QemuPciMapping{
							PCIe: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{Mapping: &QemuPciMapping{
							PCIe: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci13": "mapping=,pcie=1,rombar=0"}},
				{name: `Mapping.PrimaryGPU`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{Mapping: &QemuPciMapping{
							PrimaryGPU: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{Mapping: &QemuPciMapping{
							PrimaryGPU: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci12": "mapping=,x-vga=1,rombar=0"}},
				{name: `Mapping.ROMbar`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{Mapping: &QemuPciMapping{
							ROMbar: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{Mapping: &QemuPciMapping{
							ROMbar: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci11": "mapping="}},
				{name: `Mapping.SubDeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{Mapping: &QemuPciMapping{
							SubDeviceID: util.Pointer(PciSubDeviceID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{Mapping: &QemuPciMapping{
							SubDeviceID: util.Pointer(PciSubDeviceID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci10": "mapping=,rombar=0,sub-device-id=0x8086"}},
				{name: `Mapping.SubVendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{Mapping: &QemuPciMapping{
							SubVendorID: util.Pointer(PciSubVendorID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{Mapping: &QemuPciMapping{
							SubVendorID: util.Pointer(PciSubVendorID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci9": "mapping=,rombar=0,sub-vendor-id=0x8086"}},
				{name: `Mapping.VendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{Mapping: &QemuPciMapping{
							VendorID: util.Pointer(PciVendorID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{Mapping: &QemuPciMapping{
							VendorID: util.Pointer(PciVendorID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci8": "mapping=,rombar=0,vendor-id=0x8086"}},
				{name: `Raw.DeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{Raw: &QemuPciRaw{
							DeviceID: util.Pointer(PciDeviceID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{Raw: &QemuPciRaw{
							DeviceID: util.Pointer(PciDeviceID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci0": ",rombar=0,device-id=0x8086"}},
				{name: `Raw.ID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{Raw: &QemuPciRaw{
							ID: util.Pointer(PciID("0000:00:00.0"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{Raw: &QemuPciRaw{
							ID: util.Pointer(PciID("0000:00:00.1"))}}}},
					output: map[string]interface{}{"hostpci1": "0000:00:00.0,rombar=0"}},
				{name: `Raw.Pci`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							PCIe: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{Raw: &QemuPciRaw{
							PCIe: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci2": ",pcie=1,rombar=0"}},
				{name: `Raw.PrimaryGPU`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{Raw: &QemuPciRaw{
							PrimaryGPU: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{Raw: &QemuPciRaw{
							PrimaryGPU: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci3": ",x-vga=1,rombar=0"}},
				{name: `Raw.ROMbar`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{Raw: &QemuPciRaw{
							ROMbar: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{Raw: &QemuPciRaw{
							ROMbar: util.Pointer(false)}}}},
					output: map[string]interface{}{"hostpci4": ""}},
				{name: `Raw.SubDeviceID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Raw: &QemuPciRaw{
							SubDeviceID: util.Pointer(PciSubDeviceID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Raw: &QemuPciRaw{
							SubDeviceID: util.Pointer(PciSubDeviceID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci5": ",rombar=0,sub-device-id=0x8086"}},
				{name: `Raw.SubVendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{Raw: &QemuPciRaw{
							SubVendorID: util.Pointer(PciSubVendorID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{Raw: &QemuPciRaw{
							SubVendorID: util.Pointer(PciSubVendorID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci6": ",rombar=0,sub-vendor-id=0x8086"}},
				{name: `Raw.VendorID`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: util.Pointer(PciVendorID("8086"))}}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{Raw: &QemuPciRaw{
							VendorID: util.Pointer(PciVendorID("0x8000"))}}}},
					output: map[string]interface{}{"hostpci7": ",rombar=0,vendor-id=0x8086"}}},
			update: []test{
				{name: `Delete`,
					config: &ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{Delete: true}}},
					currentConfig: ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{}}},
					output: map[string]interface{}{"delete": "hostpci5"}}},
		},
		{category: `Serials`,
			createUpdate: []test{
				{name: `delete non existing`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{Delete: true},
						SerialID2: SerialInterface{Delete: true}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Socket: true},
						SerialID3: SerialInterface{Path: "/dev/tty2"}}},
					output: map[string]interface{}{}},
				{name: `add`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Path: "/dev/tty6"},
						SerialID3: SerialInterface{Socket: true}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{},
						SerialID2: SerialInterface{}}},
					output: map[string]interface{}{
						"serial1": "/dev/tty6",
						"serial3": "socket"}}},
			update: []test{
				{name: `existing socket no change`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{Socket: true}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{Socket: true}}},
					output: map[string]interface{}{}},
				{name: `existing path no change`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Path: "/dev/tty3"}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Path: "/dev/tty3"}}},
					output: map[string]interface{}{}},
				{name: `existing path to path`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID2: SerialInterface{Path: "/dev/tty3"}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID2: SerialInterface{Path: "/dev/tty7"}}},
					output: map[string]interface{}{"serial2": "/dev/tty3"}},
				{name: `existing socket to path`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID3: SerialInterface{Path: "/dev/tty2"}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID3: SerialInterface{Socket: true}}},
					output: map[string]interface{}{"serial3": "/dev/tty2"}},
				{name: `existing path to socket`,
					config: &ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Socket: true}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID1: SerialInterface{Path: "/dev/tty7"}}},
					output: map[string]interface{}{"serial1": "socket"}},
				{name: `delete existing`,
					config: &ConfigQemu{Serials: SerialInterfaces{SerialID2: SerialInterface{Delete: true}}},
					currentConfig: ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{Socket: true},
						SerialID2: SerialInterface{Path: "/dev/tty78"}}},
					output: map[string]interface{}{"delete": "serial2"}}},
		},
		{category: `Tags`,
			createUpdate: []test{
				{name: `Tags Empty`,
					currentConfig: ConfigQemu{Tags: util.Pointer([]Tag{"tag5", "tag6"})},
					config:        &ConfigQemu{Tags: util.Pointer([]Tag{})},
					output:        map[string]interface{}{"tags": string("")}},
				{name: `Tags Full`,
					currentConfig: ConfigQemu{Tags: util.Pointer([]Tag{"tag5", "tag6"})},
					config:        &ConfigQemu{Tags: util.Pointer([]Tag{"tag1", "tag2"})},
					output:        map[string]interface{}{"tags": string("tag1;tag2")}}}},
		{category: `TPM`,
			create: []test{
				{name: `TPM`,
					config: &ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion_2_0)}},
					output: map[string]interface{}{"tpmstate0": "test:1,version=v2.0"}}},
			update: []test{
				{name: `TPM`,
					config:        &ConfigQemu{TPM: &TpmState{Storage: "aaaa", Version: util.Pointer(TpmVersion_1_2)}},
					currentConfig: ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion_2_0)}},
					output:        map[string]interface{}{}},
				{name: `TPM Delete`,
					config: &ConfigQemu{TPM: &TpmState{Delete: true}},
					output: map[string]interface{}{"delete": "tpmstate0"}},
				{name: `TPM Delete Full`,
					config: &ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion_2_0), Delete: true}},
					output: map[string]interface{}{"delete": "tpmstate0"}}}},
		{category: `USBs`,
			create: []test{
				{name: `Delete`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Delete: true}}},
					output: map[string]interface{}{}},
			},
			createUpdate: []test{
				{name: `Device all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("1234:5678")),
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					output: map[string]interface{}{"usb0": "host=1234:5678,usb3=1"}},
				{name: `Device.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("abcd:35fe")),
							USB3: util.Pointer(false)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					output: map[string]interface{}{"usb1": "host=abcd:35fe"}},
				{name: `Device.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID: util.Pointer(UsbDeviceID("8235:95af"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					output: map[string]interface{}{"usb1": "host=8235:95af"}},
				{name: `Mapping all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("test")),
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					output: map[string]interface{}{"usb1": "mapping=test,usb3=1"}},
				{name: `Mapping.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("test")),
							USB3: util.Pointer(false)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					output: map[string]interface{}{"usb1": "mapping=test"}},
				{name: `Mapping.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID: util.Pointer(ResourceMappingUsbID("test"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{}}}},
					output: map[string]interface{}{"usb1": "mapping=test"}},
				{name: `Port all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("1-2")),
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					output: map[string]interface{}{"usb2": "host=1-2,usb3=1"}},
				{name: `Port.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("1-2")),
							USB3: util.Pointer(false)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					output: map[string]interface{}{"usb2": "host=1-2"}},
				{name: `Port.USB3 nil`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID: util.Pointer(UsbPortID("1-2"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{}}}},
					output: map[string]interface{}{"usb2": "host=1-2"}},
				{name: `Spice all`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Spice: &QemuUsbSpice{
							USB3: true}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Device: &QemuUsbDevice{}}}},
					output: map[string]interface{}{"usb3": "spice,usb3=1"}},
				{name: `Spice.USB3 false`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Spice: &QemuUsbSpice{
							USB3: false}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Device: &QemuUsbDevice{}}}},
					output: map[string]interface{}{"usb3": "spice"}},
			},
			update: []test{
				{name: `Delete`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Delete: true}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
					output: map[string]interface{}{"delete": "usb0"}},
				{name: `Device.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID: util.Pointer(UsbDeviceID("1234:5678"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("abcd:35fe")),
							USB3: util.Pointer(true)}}}},
					output: map[string]interface{}{"usb1": "host=1234:5678,usb3=1"}},
				{name: `Device.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("abcd:35fe")),
							USB3: util.Pointer(false)}}}},
					output: map[string]interface{}{"usb1": "host=abcd:35fe,usb3=1"}},
				{name: `Mapping.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID: util.Pointer(ResourceMappingUsbID("test"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("test2")),
							USB3: util.Pointer(true)}}}},
					output: map[string]interface{}{"usb1": "mapping=test,usb3=1"}},
				{name: `Mapping.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("test2")),
							USB3: util.Pointer(false)}}}},
					output: map[string]interface{}{"usb1": "mapping=test2,usb3=1"}},
				{name: `Port.ID update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID: util.Pointer(UsbPortID("1-2"))}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("2-3")),
							USB3: util.Pointer(true)}}}},
					output: map[string]interface{}{"usb1": "host=1-2,usb3=1"}},
				{name: `Port.USB3 update`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							USB3: util.Pointer(true)}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("2-3")),
							USB3: util.Pointer(false)}}}},
					output: map[string]interface{}{"usb1": "host=2-3,usb3=1"}},
				{name: `Spice`,
					config: &ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Spice: &QemuUsbSpice{
							USB3: false}}}},
					currentConfig: ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{}}}},
					output: map[string]interface{}{"usb1": "spice"}},
			},
		},
	}
	for _, test := range tests {
		for _, subTest := range append(test.create, test.createUpdate...) {
			name := test.category + "/Create/" + subTest.name
			t.Run(name, func(*testing.T) {
				reboot, tmpParams, _ := subTest.config.mapToAPI(ConfigQemu{}, subTest.version)
				require.Equal(t, subTest.output, tmpParams, name)
				require.Equal(t, false, reboot, name)
			})
		}
		for _, subTest := range append(test.update, test.createUpdate...) {
			name := test.category + "/Update/" + subTest.name
			t.Run(name, func(*testing.T) {
				reboot, tmpParams, _ := subTest.config.mapToAPI(subTest.currentConfig, subTest.version)
				require.Equal(t, subTest.output, tmpParams, name)
				require.Equal(t, subTest.reboot, reboot, name)
			})
		}
	}
}

func Test_ConfigQemu_mapToStruct(t *testing.T) {
	baseConfig := func(config ConfigQemu) *ConfigQemu {
		if config.CPU == nil {
			config.CPU = &QemuCPU{}
		}
		if config.Memory == nil {
			config.Memory = &QemuMemory{}
		}
		return &config
	}
	parseIP := func(rawIP string) (ip netip.Addr) {
		ip, _ = netip.ParseAddr(rawIP)
		return
	}
	parseMAC := func(rawMAC string) (mac net.HardwareAddr) {
		mac, _ = net.ParseMAC(rawMAC)
		return
	}
	uint1 := uint(1)
	uint2 := uint(2)
	uint31 := uint(31)
	uint47 := uint(47)
	uint53 := uint(53)
	type test struct {
		name   string
		input  map[string]interface{}
		vmr    *VmRef
		output *ConfigQemu
		err    error
	}
	tests := []struct {
		category string
		tests    []test
	}{
		// TODO add test cases for all other items of ConfigQemu{}
		{category: `Agent`,
			tests: []test{
				{name: `ALL`,
					input: map[string]interface{}{"agent": string("1,freeze-fs-on-backup=1,fstrim_cloned_disks=1,type=virtio")},
					output: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{
						Enable: util.Pointer(true),
						Freeze: util.Pointer(true),
						FsTrim: util.Pointer(true),
						Type:   util.Pointer(QemuGuestAgentType_VirtIO)}})},
				{name: `Enabled`,
					input:  map[string]interface{}{"agent": string("1")},
					output: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(true)}})},
				{name: `Freeze`,
					input:  map[string]interface{}{"agent": string("0,freeze-fs-on-backup=1")},
					output: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(false), Freeze: util.Pointer(true)}})},
				{name: `FsTrim`,
					input:  map[string]interface{}{"agent": string("0,fstrim_cloned_disks=1")},
					output: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(false), FsTrim: util.Pointer(true)}})},
				{name: `Type`,
					input:  map[string]interface{}{"agent": string("1,type=virtio")},
					output: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Enable: util.Pointer(true), Type: util.Pointer(QemuGuestAgentType_VirtIO)}})}}},
		{category: `CPU`,
			tests: []test{
				{name: `all`,
					input: map[string]interface{}{
						"cores":    float64(10),
						"cpulimit": float64(35),
						"cpuunits": float64(1234),
						"numa":     float64(0),
						"sockets":  float64(4),
						"vcpus":    float64(40),
						"cpu":      string("host,flags=-aes;+amd-no-ssb;-amd-ssbd;+hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;-pcid;-pdpe1gb;-ssbd;+spec-ctrl;+virt-ssbd")},
					output: baseConfig(ConfigQemu{
						CPU: &QemuCPU{
							Cores: util.Pointer(QemuCpuCores(10)),
							Flags: &CpuFlags{
								AES:        util.Pointer(TriBoolFalse),
								AmdNoSSB:   util.Pointer(TriBoolTrue),
								AmdSSBD:    util.Pointer(TriBoolFalse),
								HvEvmcs:    util.Pointer(TriBoolTrue),
								HvTlbFlush: util.Pointer(TriBoolFalse),
								Ibpb:       util.Pointer(TriBoolTrue),
								MdClear:    util.Pointer(TriBoolTrue),
								PCID:       util.Pointer(TriBoolFalse),
								Pdpe1GB:    util.Pointer(TriBoolFalse),
								SSBD:       util.Pointer(TriBoolFalse),
								SpecCtrl:   util.Pointer(TriBoolTrue),
								VirtSSBD:   util.Pointer(TriBoolTrue)},
							Limit:        util.Pointer(CpuLimit(35)),
							Numa:         util.Pointer(false),
							Sockets:      util.Pointer(QemuCpuSockets(4)),
							Type:         util.Pointer(CpuType_Host),
							Units:        util.Pointer(CpuUnits(1234)),
							VirtualCores: util.Pointer(CpuVirtualCores(40))}})},
				{name: `affinity consecutive`,
					input:  map[string]interface{}{"affinity": "2-4"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{2, 3, 4})}})},
				{name: `affinity empty`,
					input:  map[string]interface{}{"affinity": ""},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{})}})},
				{name: `affinity mixed`,
					input:  map[string]interface{}{"affinity": "2,4-6,8,10,12-15"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{2, 4, 5, 6, 8, 10, 12, 13, 14, 15})}})},
				{name: `affinity singular`,
					input:  map[string]interface{}{"affinity": "2"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Affinity: util.Pointer([]uint{2})}})},
				{name: `cores`,
					input:  map[string]interface{}{"cores": float64(1)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Cores: util.Pointer(QemuCpuCores(1))}})},
				{name: `cpu flag aes`,
					input: map[string]interface{}{"cpu": ",flags=+aes"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{AES: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag amd-no-ssb`,
					input: map[string]interface{}{"cpu": ",flags=-amd-no-ssb"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{AmdNoSSB: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag amd-ssbd`,
					input: map[string]interface{}{"cpu": ",flags=+amd-ssbd"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{AmdSSBD: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag hv-evmcs`,
					input: map[string]interface{}{"cpu": ",flags=-hv-evmcs"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{HvEvmcs: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag hv-tlbflush`,
					input: map[string]interface{}{"cpu": ",flags=+hv-tlbflush"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{HvTlbFlush: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag ibpb`,
					input: map[string]interface{}{"cpu": ",flags=-ibpb"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{Ibpb: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag md-clear`,
					input: map[string]interface{}{"cpu": ",flags=+md-clear"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{MdClear: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag pcid`,
					input: map[string]interface{}{"cpu": ",flags=-pcid"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{PCID: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag pdpe1gb`,
					input: map[string]interface{}{"cpu": ",flags=+pdpe1gb"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{Pdpe1GB: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag ssbd`,
					input: map[string]interface{}{"cpu": ",flags=-ssbd"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{SSBD: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag spec-ctrl`,
					input: map[string]interface{}{"cpu": ",flags=+spec-ctrl"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{SpecCtrl: util.Pointer(TriBoolTrue)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flag virt-ssbd`,
					input: map[string]interface{}{"cpu": ",flags=-virt-ssbd"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{VirtSSBD: util.Pointer(TriBoolFalse)},
						Type:  util.Pointer(CpuType(""))}})},
				{name: `cpu flags multiple`,
					input: map[string]interface{}{"cpu": ",flags=-aes;+amd-no-ssb;-amd-ssbd;-hv-evmcs;-hv-tlbflush;+ibpb;+md-clear;+pcid;-virt-ssbd"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							AES:        util.Pointer(TriBoolFalse),
							AmdNoSSB:   util.Pointer(TriBoolTrue),
							AmdSSBD:    util.Pointer(TriBoolFalse),
							HvEvmcs:    util.Pointer(TriBoolFalse),
							HvTlbFlush: util.Pointer(TriBoolFalse),
							Ibpb:       util.Pointer(TriBoolTrue),
							MdClear:    util.Pointer(TriBoolTrue),
							PCID:       util.Pointer(TriBoolTrue),
							VirtSSBD:   util.Pointer(TriBoolFalse)},
						Type: util.Pointer(CpuType(""))}})},
				{name: `cpu model only, no flags`,
					input:  map[string]interface{}{"cpu": string(CpuType_X86_64_v2_AES)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType("x86-64-v2-AES"))}})},
				{name: `cpu with flags`,
					input: map[string]interface{}{"cpu": "x86-64-v2-AES,flags=+spec-ctrl;-md-clear"},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{
						Flags: &CpuFlags{
							MdClear:  util.Pointer(TriBoolFalse),
							SpecCtrl: util.Pointer(TriBoolTrue)},
						Type: util.Pointer(CpuType_X86_64_v2_AES)}})},
				{name: `cpulimit float64`,
					input:  map[string]interface{}{"cpulimit": float64(10)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(10))}})},
				{name: `cpulimit string`,
					input:  map[string]interface{}{"cpulimit": string("25")},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(25))}})},
				{name: `cpuunits`,
					input:  map[string]interface{}{"cpuunits": float64(1000)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(1000))}})},
				{name: `numa true`,
					input:  map[string]interface{}{"numa": float64(1)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Numa: util.Pointer(true)}})},
				{name: `numa false`,
					input:  map[string]interface{}{"numa": float64(0)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Numa: util.Pointer(false)}})},
				{name: `sockets`,
					input:  map[string]interface{}{"sockets": float64(1)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{Sockets: util.Pointer(QemuCpuSockets(1))}})},
				{name: `vcpus`,
					input:  map[string]interface{}{"vcpus": float64(1)},
					output: baseConfig(ConfigQemu{CPU: &QemuCPU{VirtualCores: util.Pointer(CpuVirtualCores(1))}})}}},
		{category: `CloudInit`,
			tests: []test{
				{name: `ALL`,
					input: map[string]interface{}{
						"cicustom":     string("meta=local-zfs:ci-meta.yml,network=local-lvm:ci-network.yml,user=folder:ci-user.yml,vendor=local:snippets/ci-custom.yml"),
						"searchdomain": string("example.com"),
						"nameserver":   string("1.1.1.1 8.8.8.8 9.9.9.9"),
						"ipconfig0":    string("ip=dhcp,ip6=dhcp"),
						"ipconfig19":   string(""),
						"ipconfig31":   string("ip=10.20.4.7/22"),
						"sshkeys":      test_data_qemu.PublicKey_Encoded_Input(),
						"ciupgrade":    float64(1),
						"cipassword":   string("Enter123!"),
						"ciuser":       string("root")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						Custom: &CloudInitCustom{
							Meta: &CloudInitSnippet{
								FilePath: "ci-meta.yml",
								Storage:  "local-zfs"},
							Network: &CloudInitSnippet{
								FilePath: "ci-network.yml",
								Storage:  "local-lvm"},
							User: &CloudInitSnippet{
								FilePath: "ci-user.yml",
								Storage:  "folder"},
							Vendor: &CloudInitSnippet{
								FilePath: "snippets/ci-custom.yml",
								Storage:  "local"}},
						DNS: &GuestDNS{
							SearchDomain: util.Pointer("example.com"),
							NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
						NetworkInterfaces: CloudInitNetworkInterfaces{
							QemuNetworkInterfaceID0: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{DHCP: true},
								IPv6: &CloudInitIPv6Config{DHCP: true}},
							QemuNetworkInterfaceID31: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}},
						PublicSSHkeys:   util.Pointer(test_data_qemu.PublicKey_Decoded_Output()),
						UpgradePackages: util.Pointer(true),
						UserPassword:    util.Pointer("Enter123!"),
						Username:        util.Pointer("root")}})},
				{name: `Custom Meta`,
					input: map[string]interface{}{"cicustom": string("meta=local-zfs:ci-meta.yml")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						Custom:            &CloudInitCustom{Meta: &CloudInitSnippet{FilePath: "ci-meta.yml", Storage: "local-zfs"}},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `Custom Network`,
					input: map[string]interface{}{"cicustom": string("network=local-lvm:ci-network.yml")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						Custom:            &CloudInitCustom{Network: &CloudInitSnippet{FilePath: "ci-network.yml", Storage: "local-lvm"}},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `Custom User`,
					input: map[string]interface{}{
						"cicustom": string("user=folder:ci-user.yml")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						Custom:            &CloudInitCustom{User: &CloudInitSnippet{FilePath: "ci-user.yml", Storage: "folder"}},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `Custom Vendor`,
					input: map[string]interface{}{"cicustom": string("vendor=local:snippets/ci-custom.yml")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						Custom:            &CloudInitCustom{Vendor: &CloudInitSnippet{FilePath: "snippets/ci-custom.yml", Storage: "local"}},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `DNS SearchDomain`,
					input: map[string]interface{}{"searchdomain": string("example.com")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						DNS: &GuestDNS{
							SearchDomain: util.Pointer("example.com"),
							NameServers:  util.Pointer(uninitializedArray[netip.Addr]())},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `DNS SearchDomain empty`,
					input:  map[string]interface{}{"searchdomain": string(" ")},
					output: baseConfig(ConfigQemu{})},
				{name: `DNS NameServers`,
					input: map[string]interface{}{"nameserver": string("1.1.1.1 8.8.8.8 9.9.9.9")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						DNS: &GuestDNS{
							SearchDomain: util.Pointer(""),
							NameServers:  &[]netip.Addr{parseIP("1.1.1.1"), parseIP("8.8.8.8"), parseIP("9.9.9.9")}},
						NetworkInterfaces: CloudInitNetworkInterfaces{}}})},
				{name: `NetworkInterfaces`,
					input: map[string]interface{}{
						"ipconfig0":  string("ip=dhcp,ip6=dhcp"),
						"ipconfig1":  string("ip6=auto"),
						"ipconfig2":  string("ip=192.168.1.10/24,gw=192.168.56.1,ip6=2001:0db8:abcd::/48,gw6=2001:0db8:abcd::1"),
						"ipconfig19": string(""),
						"ipconfig20": string(" "), // this single space is on porpuse to test if it is ignored
						"ipconfig31": string("ip=10.20.4.7/22")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						NetworkInterfaces: CloudInitNetworkInterfaces{
							QemuNetworkInterfaceID0: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{DHCP: true},
								IPv6: &CloudInitIPv6Config{DHCP: true}},
							QemuNetworkInterfaceID1: CloudInitNetworkConfig{
								IPv6: &CloudInitIPv6Config{SLAAC: true}},
							QemuNetworkInterfaceID2: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{
									Address: util.Pointer(IPv4CIDR("192.168.1.10/24")),
									Gateway: util.Pointer(IPv4Address("192.168.56.1"))},
								IPv6: &CloudInitIPv6Config{
									Address: util.Pointer(IPv6CIDR("2001:0db8:abcd::/48")),
									Gateway: util.Pointer(IPv6Address("2001:0db8:abcd::1"))}},
							QemuNetworkInterfaceID31: CloudInitNetworkConfig{
								IPv4: &CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("10.20.4.7/22"))}}}}})},
				{name: `PublicSSHkeys`,
					input: map[string]interface{}{"sshkeys": test_data_qemu.PublicKey_Encoded_Input()},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						NetworkInterfaces: CloudInitNetworkInterfaces{},
						PublicSSHkeys:     util.Pointer(test_data_qemu.PublicKey_Decoded_Output())}})},
				{name: `UpgradePackages`,
					input: map[string]interface{}{"ciupgrade": float64(0)},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						NetworkInterfaces: CloudInitNetworkInterfaces{},
						UpgradePackages:   util.Pointer(false)}})},
				{name: `UserPassword`,
					input: map[string]interface{}{"cipassword": string("Enter123!")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						NetworkInterfaces: CloudInitNetworkInterfaces{},
						UserPassword:      util.Pointer("Enter123!")}})},
				{name: `Username`,
					input: map[string]interface{}{"ciuser": string("root")},
					output: baseConfig(ConfigQemu{CloudInit: &CloudInit{
						NetworkInterfaces: CloudInitNetworkInterfaces{},
						Username:          util.Pointer("root")}})},
				{name: `Username empty`,
					input:  map[string]interface{}{"ciuser": string(" ")},
					output: baseConfig(ConfigQemu{})}}},
		{category: `Description`,
			tests: []test{
				{input: map[string]interface{}{"description": string("test description")},
					output: baseConfig(ConfigQemu{Description: util.Pointer("test description")})}}},
		{category: `Disks Ide CdRom`,
			tests: []test{
				{name: `none`,
					input:  map[string]interface{}{"ide1": "none,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{}}}}})},
				{name: `passthrough`,
					input:  map[string]interface{}{"ide2": "cdrom,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
				{name: `iso`,
					input: map[string]interface{}{"ide3": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
						File:            "debian-11.0.0-amd64-netinst.iso",
						Storage:         "local",
						SizeInKibibytes: "377M"}}}}}})}}},
		{category: `Disks Ide CloudInit`,
			tests: []test{
				{name: `file`,
					input: map[string]interface{}{"ide0": "Test:100/vm-100-cloudinit.raw,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})},
				{name: `lvm`,
					input: map[string]interface{}{"ide3": "Test:vm-100-cloudinit,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})}}},
		{category: `Disks Ide Disk`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `ALL`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,aio=io_uring,backup=0,cache=writethrough,discard=on,iops_rd=12,iops_rd_max=13,iops_rd_max_length=4,iops_wr=15,iops_wr_max=14,iops_wr_max_length=5,mbps_rd=1.46,mbps_rd_max=3.57,mbps_wr=2.68,mbps_wr_max=4.55,replicate=0,serial=disk-9763,size=1032G,ssd=1,wwn=0x500F753600A987E1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO: QemuDiskAsyncIO_IOuring,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.57, Concurrent: 1.46},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55, Concurrent: 2.68}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 12},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14, BurstDuration: 5, Concurrent: 15}}},
						Cache:           QemuDiskCache_WriteThrough,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint53,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1082130432,
						Storage:         "test2",
						WorldWideName:   "0x500F753600A987E1"}}}}})},
				{name: `ALL LinkedClone`,
					input: map[string]interface{}{"ide1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-53.qcow2,aio=io_uring,backup=0,cache=writethrough,discard=on,iops_rd=12,iops_rd_max=13,iops_rd_max_length=4,iops_wr=15,iops_wr_max=14,iops_wr_max_length=5,mbps_rd=1.46,mbps_rd_max=3.57,mbps_wr=2.68,mbps_wr_max=4.55,replicate=0,serial=disk-9763,size=1032G,ssd=1,wwn=0x500679CE00B1DAF4"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							AsyncIO: QemuDiskAsyncIO_IOuring,
							Backup:  false,
							Bandwidth: QemuDiskBandwidth{
								MBps: QemuDiskBandwidthMBps{
									ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.57, Concurrent: 1.46},
									WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55, Concurrent: 2.68}},
								Iops: QemuDiskBandwidthIops{
									ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 12},
									WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14, BurstDuration: 5, Concurrent: 15}}},
							Cache:           QemuDiskCache_WriteThrough,
							Discard:         true,
							EmulateSSD:      true,
							Format:          QemuDiskFormat_Qcow2,
							Id:              uint53,
							LinkedDiskId:    &uint1,
							Replicate:       false,
							Serial:          "disk-9763",
							SizeInKibibytes: 1082130432,
							Storage:         "test2",
							WorldWideName:   "0x500679CE00B1DAF4"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"ide2": "test2:100/vm-100-disk-53.qcow2,aio=io_uring"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						AsyncIO:   QemuDiskAsyncIO_IOuring,
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"ide3": "test2:100/vm-100-disk-53.qcow2,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    false,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,cache=writethrough"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Cache:     QemuDiskCache_WriteThrough,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Discard:   true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"ide2": "test2:100/vm-100-disk-53.qcow2,iops_rd=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 12}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"ide3": "test2:100/vm-100-disk-53.qcow2,iops_rd_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"ide3": "test2:100/vm-100-disk-53.qcow2,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,iops_wr=15"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 15}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,iops_wr_max=14"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 14}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"ide2": "test2:100/vm-100-disk-53.qcow2,mbps_rd=1.46"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.46}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"ide3": "test2:100/vm-100-disk-53.qcow2,mbps_rd_max=3.57"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.57}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,mbps_wr=2.68"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.68}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,mbps_wr_max=4.55"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.55}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"ide2": "test2:100/vm-100-disk-53.qcow2,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: false,
						Storage:   "test2"}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"ide3": "test2:100/vm-100-disk-53.qcow2,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint53,
						Replicate: true,
						Serial:    "disk-9763",
						Storage:   "test2"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint53,
						Replicate:       true,
						SizeInKibibytes: 1082130432,
						Storage:         "test2"}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint53,
						Replicate:       true,
						SizeInKibibytes: 1032,
						Storage:         "test2"}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,size=1032M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint53,
						Replicate:       true,
						SizeInKibibytes: 1056768,
						Storage:         "test2"}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"ide0": "test2:100/vm-100-disk-53.qcow2,size=2T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint53,
						Replicate:       true,
						SizeInKibibytes: 2147483648,
						Storage:         "test2"}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:     true,
						EmulateSSD: true,
						Format:     QemuDiskFormat_Qcow2,
						Id:         uint53,
						Replicate:  true,
						Storage:    "test2"}}}}})},
				{name: `syntax fileSyntaxVolume`,
					input: map[string]interface{}{"ide2": "test:vm-100-disk-2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Raw,
						Id:        uint2,
						Replicate: true,
						Storage:   "test",
						syntax:    diskSyntaxVolume}}}}})},
				{name: `syntax fileSyntaxVolume LinkedClone`,
					input: map[string]interface{}{"ide3": "test:base-110-disk-1/vm-100-disk-2"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Backup:       true,
							Format:       QemuDiskFormat_Raw,
							Id:           uint2,
							LinkedDiskId: &uint1,
							Replicate:    true,
							Storage:      "test",
							syntax:       diskSyntaxVolume}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"ide1": "test2:100/vm-100-disk-53.qcow2,wwn=0x500DB82100C6FA59"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
						Backup:        true,
						Format:        QemuDiskFormat_Qcow2,
						Id:            uint53,
						Replicate:     true,
						Storage:       "test2",
						WorldWideName: "0x500DB82100C6FA59"}}}}})}}},
		{category: `Disks Ide Passthrough`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `All`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,backup=0,cache=unsafe,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=1G,ssd=1,wwn=0x500CBE4300D978A6"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_Unsafe,
						Discard:         true,
						EmulateSSD:      true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1048576,
						WorldWideName:   "0x500CBE4300D978A6"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						AsyncIO:   QemuDiskAsyncIO_Threads,
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    false,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=unsafe"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Cache:     QemuDiskCache_Unsafe,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Discard:   true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"ide2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: false}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"ide3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true,
						Serial:    "disk-9763"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 10485760}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 10}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 10240}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"ide0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=10T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 10737418240}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:     true,
						EmulateSSD: true,
						File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:  true}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"ide1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5005AC1200F643B8"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
						Backup:        true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						WorldWideName: "0x5005AC1200F643B8"}}}}})}}},
		{category: `Disks Sata CdRom`,
			tests: []test{
				{name: `none`,
					input:  map[string]interface{}{"sata5": "none,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{CdRom: &QemuCdRom{}}}}})},
				{name: `passthrough`,
					input:  map[string]interface{}{"sata4": "cdrom,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
				{name: `iso`,
					input: map[string]interface{}{"sata3": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
						File:            "debian-11.0.0-amd64-netinst.iso",
						Storage:         "local",
						SizeInKibibytes: "377M"}}}}}})}}},
		{category: `Disks Sata CloudInit`,
			tests: []test{
				{name: `file`,
					input: map[string]interface{}{"sata4": "Test:100/vm-100-cloudinit.raw,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})},
				{name: `lvm`,
					input: map[string]interface{}{"sata0": "Test:vm-100-cloudinit,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})}}},
		{category: `Disks Sata Disk`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"sata0": "test2:100/vm-100-disk-47.qcow2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `ALL`,
					input: map[string]interface{}{"sata1": "test2:100/vm-100-disk-47.qcow2,aio=native,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=32G,ssd=1,wwn=0x500DFA8900E3C641"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_None,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint47,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test2",
						WorldWideName:   "0x500DFA8900E3C641"}}}}})},
				{name: `ALL LinkedClone`,
					input: map[string]interface{}{"sata1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-47.qcow2,aio=native,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=32G,ssd=1,wwn=0x5003B97600A8F2D4"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
							AsyncIO: QemuDiskAsyncIO_Native,
							Backup:  false,
							Bandwidth: QemuDiskBandwidth{
								MBps: QemuDiskBandwidthMBps{
									ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
									WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
								Iops: QemuDiskBandwidthIops{
									ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
									WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
							Cache:           QemuDiskCache_None,
							Discard:         true,
							EmulateSSD:      true,
							Format:          QemuDiskFormat_Qcow2,
							Id:              uint47,
							LinkedDiskId:    &uint1,
							Replicate:       false,
							Serial:          "disk-9763",
							SizeInKibibytes: 33554432,
							Storage:         "test2",
							WorldWideName:   "0x5003B97600A8F2D4"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"sata2": "test2:100/vm-100-disk-47.qcow2,aio=native"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
						AsyncIO:   QemuDiskAsyncIO_Native,
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"sata3": "test2:100/vm-100-disk-47.qcow2,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    false,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,cache=none"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Cache:     QemuDiskCache_None,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"sata5": "test2:100/vm-100-disk-47.qcow2,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Discard:   true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"sata0": "test2:100/vm-100-disk-47.qcow2,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"sata1": "test2:100/vm-100-disk-47.qcow2,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"sata1": "test2:100/vm-100-disk-47.qcow2,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"sata2": "test2:100/vm-100-disk-47.qcow2,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"sata3": "test2:100/vm-100-disk-47.qcow2,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"sata3": "test2:100/vm-100-disk-47.qcow2,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"sata5": "test2:100/vm-100-disk-47.qcow2,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"sata0": "test2:100/vm-100-disk-47.qcow2,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup: true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"sata1": "test2:100/vm-100-disk-47.qcow2,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"sata2": "test2:100/vm-100-disk-47.qcow2,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: false,
						Storage:   "test2"}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"sata3": "test2:100/vm-100-disk-47.qcow2,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint47,
						Replicate: true,
						Serial:    "disk-9763",
						Storage:   "test2"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,size=32G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint47,
						Replicate:       true,
						SizeInKibibytes: 33554432,
						Storage:         "test2"}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,size=32K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint47,
						Replicate:       true,
						SizeInKibibytes: 32,
						Storage:         "test2"}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,size=32M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint47,
						Replicate:       true,
						SizeInKibibytes: 32768,
						Storage:         "test2"}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"sata4": "test2:100/vm-100-disk-47.qcow2,size=3T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint47,
						Replicate:       true,
						SizeInKibibytes: 3221225472,
						Storage:         "test2"}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"sata5": "test2:100/vm-100-disk-47.qcow2,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:     true,
						EmulateSSD: true,
						Format:     QemuDiskFormat_Qcow2,
						Id:         uint47,
						Replicate:  true,
						Storage:    "test2"}}}}})},
				{name: `syntax fileSyntaxVolume`,
					input: map[string]interface{}{"sata0": "test:vm-100-disk-2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Raw,
						Id:        uint2,
						Replicate: true,
						Storage:   "test",
						syntax:    diskSyntaxVolume}}}}})},
				{name: `syntax fileSyntaxVolume LinkedClone`,
					input: map[string]interface{}{"sata1": "test:base-110-disk-1/vm-100-disk-2"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
							Backup:       true,
							Format:       QemuDiskFormat_Raw,
							Id:           uint2,
							LinkedDiskId: &uint1,
							Replicate:    true,
							Storage:      "test",
							syntax:       diskSyntaxVolume}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"sata5": "test2:100/vm-100-disk-47.qcow2,wwn=0x5001E48A00D567C9"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{
						Backup:        true,
						Format:        QemuDiskFormat_Qcow2,
						Id:            uint47,
						Replicate:     true,
						Storage:       "test2",
						WorldWideName: "0x5001E48A00D567C9"}}}}})}}},
		{category: `Disks Sata Passthrough`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `All`,
					input: map[string]interface{}{"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=5,iops_wr=11,iops_wr_max=13,iops_wr_max_length=4,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,serial=disk-9763,size=1G,ssd=1,wwn=500E9FBC00F2A6D3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						AsyncIO: QemuDiskAsyncIO_IOuring,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 5, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 4, Concurrent: 11}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						EmulateSSD:      true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1048576,
						WorldWideName:   "500E9FBC00F2A6D3"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=io_uring"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						AsyncIO:   QemuDiskAsyncIO_IOuring,
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    false,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=directsync"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Cache:     QemuDiskCache_DirectSync,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Discard:   true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"sata0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"sata0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"sata1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"sata2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: false}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"sata3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true,
						Serial:    "disk-9763"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=3G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 3145728}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=5789K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 5789}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=45M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 46080}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"sata4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=7T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 7516192768}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:     true,
						EmulateSSD: true,
						File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:  true}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"sata5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5004D2EF00C8B57A"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
						Backup:        true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						WorldWideName: "0x5004D2EF00C8B57A"}}}}})}}},
		{category: `Disks Scsi CdRom`,
			tests: []test{
				{name: `none`,
					input:  map[string]interface{}{"scsi30": "none,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_30: &QemuScsiStorage{CdRom: &QemuCdRom{}}}}})},
				{name: `passthrough`,
					input:  map[string]interface{}{"scsi29": "cdrom,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_29: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
				{name: `iso`,
					input: map[string]interface{}{"scsi28": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_28: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
						File:            "debian-11.0.0-amd64-netinst.iso",
						Storage:         "local",
						SizeInKibibytes: "377M"}}}}}})}}},
		{category: `Disks Scsi CloudInit`,
			tests: []test{
				{name: `file`,
					input: map[string]interface{}{"scsi0": "Test:100/vm-100-cloudinit.raw,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})},
				{name: `lvm`,
					input: map[string]interface{}{"scsi23": "Test:vm-100-cloudinit,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_23: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})}}},
		{category: `Disks Scsi Disk`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"scsi0": "test:100/vm-100-disk-2.qcow2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `ALL`,
					input: map[string]interface{}{"scsi1": "test:100/vm-100-disk-2.qcow2,aio=threads,backup=0,cache=writeback,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,ssd=1,wwn=0x500AF18700E9CD25"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_WriteBack,
						Discard:         true,
						EmulateSSD:      true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint2,
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test",
						WorldWideName:   "0x500AF18700E9CD25"}}}}})},
				{name: `ALL LinkedClone`,
					input: map[string]interface{}{"scsi1": "test:110/base-110-disk-1.qcow2/100/vm-100-disk-2.qcow2,aio=threads,backup=0,cache=writeback,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,ssd=1,wwn=0x500879DC00F3BE6A"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{
							AsyncIO: QemuDiskAsyncIO_Threads,
							Backup:  false,
							Bandwidth: QemuDiskBandwidth{
								MBps: QemuDiskBandwidthMBps{
									ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
									WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
								Iops: QemuDiskBandwidthIops{
									ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
									WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
							Cache:           QemuDiskCache_WriteBack,
							Discard:         true,
							EmulateSSD:      true,
							Format:          QemuDiskFormat_Qcow2,
							Id:              uint2,
							IOThread:        true,
							LinkedDiskId:    &uint1,
							ReadOnly:        true,
							Replicate:       false,
							Serial:          "disk-9763",
							SizeInKibibytes: 33554432,
							Storage:         "test",
							WorldWideName:   "0x500879DC00F3BE6A"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"scsi2": "test:100/vm-100-disk-2.qcow2,aio=threads"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{
						AsyncIO:   QemuDiskAsyncIO_Threads,
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"scsi3": "test:100/vm-100-disk-2.qcow2,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    false,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"scsi4": "test:100/vm-100-disk-2.qcow2,cache=writeback"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Cache:     QemuDiskCache_WriteBack,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"scsi5": "test:100/vm-100-disk-2.qcow2,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Discard:   true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"scsi6": "test:100/vm-100-disk-2.qcow2,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"scsi7": "test:100/vm-100-disk-2.qcow2,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"scsi7": "test:100/vm-100-disk-2.qcow2,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"scsi8": "test:100/vm-100-disk-2.qcow2,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"scsi9": "test:100/vm-100-disk-2.qcow2,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"scsi9": "test:100/vm-100-disk-2.qcow2,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `iothread`,
					input: map[string]interface{}{"scsi10": "test:100/vm-100-disk-2.qcow2,iothread=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						IOThread:  true,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"scsi11": "test:100/vm-100-disk-2.qcow2,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"scsi12": "test:100/vm-100-disk-2.qcow2,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"scsi13": "test:100/vm-100-disk-2.qcow2,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"scsi14": "test:100/vm-100-disk-2.qcow2,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"scsi15": "test:100/vm-100-disk-2.qcow2,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: false,
						Storage:   "test"}}}}})},
				{name: `ro`,
					input: map[string]interface{}{"scsi16": "test:100/vm-100-disk-2.qcow2,ro=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						ReadOnly:  true,
						Replicate: true,
						Storage:   "test"}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"scsi17": "test:100/vm-100-disk-2.qcow2,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint2,
						Replicate: true,
						Serial:    "disk-9763",
						Storage:   "test"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"scsi18": "test:100/vm-100-disk-2.qcow2,size=32G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint2,
						Replicate:       true,
						SizeInKibibytes: 33554432,
						Storage:         "test"}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"scsi18": "test:100/vm-100-disk-2.qcow2,size=32K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint2,
						Replicate:       true,
						SizeInKibibytes: 32,
						Storage:         "test"}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"scsi18": "test:100/vm-100-disk-2.qcow2,size=32M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint2,
						Replicate:       true,
						SizeInKibibytes: 32768,
						Storage:         "test"}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"scsi18": "test:100/vm-100-disk-2.qcow2,size=4T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint2,
						Replicate:       true,
						SizeInKibibytes: 4294967296,
						Storage:         "test"}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"scsi19": "test:100/vm-100-disk-2.qcow2,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:     true,
						EmulateSSD: true,
						Format:     QemuDiskFormat_Qcow2,
						Id:         uint2,
						Replicate:  true,
						Storage:    "test"}}}}})},
				{name: `syntax fileSyntaxVolume`,
					input: map[string]interface{}{"scsi6": "test:vm-100-disk-2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:    true,
						Format:    QemuDiskFormat_Raw,
						Id:        uint2,
						Replicate: true,
						Storage:   "test",
						syntax:    diskSyntaxVolume}}}}})},
				{name: `syntax fileSyntaxVolume LinkedClone`,
					input: map[string]interface{}{"scsi7": "test:base-110-disk-1/vm-100-disk-2"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Backup:       true,
							Format:       QemuDiskFormat_Raw,
							Id:           uint2,
							LinkedDiskId: &uint1,
							Replicate:    true,
							Storage:      "test",
							syntax:       diskSyntaxVolume}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"scsi19": "test:100/vm-100-disk-2.qcow2,wwn=0x500E265400A1F3D7"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Disk: &QemuScsiDisk{
						Backup:        true,
						Format:        QemuDiskFormat_Qcow2,
						Id:            uint2,
						Replicate:     true,
						Storage:       "test",
						WorldWideName: "0x500E265400A1F3D7"}}}}})}}},
		{category: `Disks Scsi Passthrough`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"scsi0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `All`,
					input: map[string]interface{}{"scsi1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads,backup=0,cache=none,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=1G,ssd=1,wwn=500CB15600D8FE32"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						AsyncIO: QemuDiskAsyncIO_Threads,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_None,
						Discard:         true,
						EmulateSSD:      true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1048576,
						WorldWideName:   "500CB15600D8FE32"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"scsi2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=threads"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						AsyncIO:   QemuDiskAsyncIO_Threads,
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"scsi3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    false,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"scsi4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=none"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Cache:     QemuDiskCache_None,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"scsi5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Discard:   true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"scsi6": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"scsi7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"scsi7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"scsi8": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"scsi9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"scsi9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iothread`,
					input: map[string]interface{}{"scsi10": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iothread=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:  true,
						Replicate: true}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"scsi11": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"scsi12": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"scsi13": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"scsi14": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"scsi15": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: false}}}}})},
				{name: `ro`,
					input: map[string]interface{}{"scsi16": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ro=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						ReadOnly:  true,
						Replicate: true}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"scsi17": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_17: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true,
						Serial:    "disk-9763"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1048576}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1024}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"scsi18": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_18: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1073741824}}}}})},
				{name: `ssd`,
					input: map[string]interface{}{"scsi19": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ssd=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:     true,
						EmulateSSD: true,
						File:       "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:  true}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"scsi19": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5009A4FC00B7C613"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_19: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
						Backup:        true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						WorldWideName: "0x5009A4FC00B7C613"}}}}})}}},
		{category: `Disks VirtIO CdRom`,
			tests: []test{
				{name: `none`,
					input:  map[string]interface{}{"virtio11": "none,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{CdRom: &QemuCdRom{}}}}})},
				{name: `passthrough`,
					input:  map[string]interface{}{"virtio10": "cdrom,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{CdRom: &QemuCdRom{Passthrough: true}}}}})},
				{name: `iso`,
					input: map[string]interface{}{"virtio9": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{
						File:            "debian-11.0.0-amd64-netinst.iso",
						Storage:         "local",
						SizeInKibibytes: "377M"}}}}}})}}},
		{category: `Disks VirtIO CloudInit`,
			tests: []test{
				{name: `file`,
					input: map[string]interface{}{"virtio0": "Test:100/vm-100-cloudinit.raw,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})},
				{name: `lvm`,
					input: map[string]interface{}{"virtio7": "Test:vm-100-cloudinit,media=cdrom"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{
						Format:  QemuDiskFormat_Raw,
						Storage: "Test"}}}}})}}},
		{category: `Disks VirtIO Disk`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"virtio0": "test2:100/vm-100-disk-31.qcow2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `ALL`,
					input: map[string]interface{}{"virtio1": "test2:100/vm-100-disk-31.qcow2,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=2,iops_wr=11,iops_wr_max=13,iops_wr_max_length=3,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,wwn=0x50015B3900F8EAD2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO: QemuDiskAsyncIO_IOuring,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 2, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 3, Concurrent: 11}}},
						Cache:           QemuDiskCache_DirectSync,
						Discard:         true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint31,
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 33554432,
						Storage:         "test2",
						WorldWideName:   "0x50015B3900F8EAD2"}}}}})},
				{name: `ALL LinkedClone`,
					input: map[string]interface{}{"virtio1": "test2:110/base-110-disk-1.qcow2/100/vm-100-disk-31.qcow2,aio=io_uring,backup=0,cache=directsync,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=2,iops_wr=11,iops_wr_max=13,iops_wr_max_length=3,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=32G,wwn=0x500FA2D000C69587"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							AsyncIO: QemuDiskAsyncIO_IOuring,
							Backup:  false,
							Bandwidth: QemuDiskBandwidth{
								MBps: QemuDiskBandwidthMBps{
									ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
									WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
								Iops: QemuDiskBandwidthIops{
									ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 2, Concurrent: 10},
									WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 3, Concurrent: 11}}},
							Cache:           QemuDiskCache_DirectSync,
							Discard:         true,
							Format:          QemuDiskFormat_Qcow2,
							Id:              uint31,
							IOThread:        true,
							LinkedDiskId:    &uint1,
							ReadOnly:        true,
							Replicate:       false,
							Serial:          "disk-9763",
							SizeInKibibytes: 33554432,
							Storage:         "test2",
							WorldWideName:   "0x500FA2D000C69587"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,aio=io_uring"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						AsyncIO:   QemuDiskAsyncIO_IOuring,
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"virtio3": "test2:100/vm-100-disk-31.qcow2,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    false,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"virtio4": "test2:100/vm-100-disk-31.qcow2,cache=directsync"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Cache:     QemuDiskCache_DirectSync,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"virtio5": "test2:100/vm-100-disk-31.qcow2,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Discard:   true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"virtio6": "test2:100/vm-100-disk-31.qcow2,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"virtio7": "test2:100/vm-100-disk-31.qcow2,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"virtio7": "test2:100/vm-100-disk-31.qcow2,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"virtio8": "test2:100/vm-100-disk-31.qcow2,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"virtio9": "test2:100/vm-100-disk-31.qcow2,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"virtio9": "test2:100/vm-100-disk-31.qcow2,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `iothread`,
					input: map[string]interface{}{"virtio10": "test2:100/vm-100-disk-31.qcow2,iothread=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						IOThread:  true,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"virtio11": "test2:100/vm-100-disk-31.qcow2,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup: true,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"virtio12": "test2:100/vm-100-disk-31.qcow2,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"virtio13": "test2:100/vm-100-disk-31.qcow2,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"virtio14": "test2:100/vm-100-disk-31.qcow2,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"virtio15": "test2:100/vm-100-disk-31.qcow2,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: false,
						Storage:   "test2"}}}}})},
				{name: `ro`,
					input: map[string]interface{}{"virtio0": "test2:100/vm-100-disk-31.qcow2,ro=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						ReadOnly:  true,
						Replicate: true,
						Storage:   "test2"}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"virtio1": "test2:100/vm-100-disk-31.qcow2,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Qcow2,
						Id:        uint31,
						Replicate: true,
						Serial:    "disk-9763",
						Storage:   "test2"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint31,
						Replicate:       true,
						SizeInKibibytes: 33554432,
						Storage:         "test2"}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint31,
						Replicate:       true,
						SizeInKibibytes: 32,
						Storage:         "test2"}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,size=32M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint31,
						Replicate:       true,
						SizeInKibibytes: 32768,
						Storage:         "test2"}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,size=5T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:          true,
						Format:          QemuDiskFormat_Qcow2,
						Id:              uint31,
						Replicate:       true,
						SizeInKibibytes: 5368709120,
						Storage:         "test2"}}}}})},
				{name: `syntax fileSyntaxVolume`,
					input: map[string]interface{}{"virtio7": "test:vm-100-disk-2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:    true,
						Format:    QemuDiskFormat_Raw,
						Id:        uint2,
						Replicate: true,
						Storage:   "test",
						syntax:    diskSyntaxVolume}}}}})},
				{name: `syntax fileSyntaxVolume LinkedClone`,
					input: map[string]interface{}{"virtio8": "test:base-110-disk-1/vm-100-disk-2"},
					output: baseConfig(ConfigQemu{
						LinkedVmId: 110,
						Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Backup:       true,
							Format:       QemuDiskFormat_Raw,
							Id:           uint2,
							LinkedDiskId: &uint1,
							Replicate:    true,
							Storage:      "test",
							syntax:       diskSyntaxVolume}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"virtio2": "test2:100/vm-100-disk-31.qcow2,wwn=0x500D3FAB00B4E672"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
						Backup:        true,
						Format:        QemuDiskFormat_Qcow2,
						Id:            uint31,
						Replicate:     true,
						Storage:       "test2",
						WorldWideName: "0x500D3FAB00B4E672"}}}}})}}},
		{category: `Disks VirtIO Passthrough`,
			tests: []test{
				{name: ``,
					input: map[string]interface{}{"virtio0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `ALL`,
					input: map[string]interface{}{"virtio1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=native,backup=0,cache=unsafe,discard=on,iops_rd=10,iops_rd_max=12,iops_rd_max_length=4,iops_wr=11,iops_wr_max=13,iops_wr_max_length=5,iothread=1,mbps_rd=1.51,mbps_rd_max=3.51,mbps_wr=2.51,mbps_wr_max=4.51,replicate=0,ro=1,serial=disk-9763,size=1G,wwn=0x500B6ED600F1C945"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						AsyncIO: QemuDiskAsyncIO_Native,
						Backup:  false,
						Bandwidth: QemuDiskBandwidth{
							MBps: QemuDiskBandwidthMBps{
								ReadLimit:  QemuDiskBandwidthMBpsLimit{Burst: 3.51, Concurrent: 1.51},
								WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51, Concurrent: 2.51}},
							Iops: QemuDiskBandwidthIops{
								ReadLimit:  QemuDiskBandwidthIopsLimit{Burst: 12, BurstDuration: 4, Concurrent: 10},
								WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13, BurstDuration: 5, Concurrent: 11}}},
						Cache:           QemuDiskCache_Unsafe,
						Discard:         true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:        true,
						ReadOnly:        true,
						Replicate:       false,
						Serial:          "disk-9763",
						SizeInKibibytes: 1048576,
						WorldWideName:   "0x500B6ED600F1C945"}}}}})},
				{name: `aio`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,aio=native"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						AsyncIO:   QemuDiskAsyncIO_Native,
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `backup`,
					input: map[string]interface{}{"virtio3": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,backup=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    false,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `cache`,
					input: map[string]interface{}{"virtio4": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,cache=unsafe"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Cache:     QemuDiskCache_Unsafe,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `discard`,
					input: map[string]interface{}{"virtio5": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,discard=on"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Discard:   true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd`,
					input: map[string]interface{}{"virtio6": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd=10"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 10}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max`,
					input: map[string]interface{}{"virtio7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max=12"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 12}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_rd_max_length`,
					input: map[string]interface{}{"virtio7": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_rd_max_length=2"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 2}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr`,
					input: map[string]interface{}{"virtio8": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr=11"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 11}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max`,
					input: map[string]interface{}{"virtio9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max=13"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 13}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iops_wr_max_length`,
					input: map[string]interface{}{"virtio9": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iops_wr_max_length=3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{BurstDuration: 3}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `iothread`,
					input: map[string]interface{}{"virtio10": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,iothread=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						IOThread:  true,
						Replicate: true}}}}})},
				{name: `mbps_rd`,
					input: map[string]interface{}{"virtio11": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd=1.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 1.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_rd_max`,
					input: map[string]interface{}{"virtio12": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_rd_max=3.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 3.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr`,
					input: map[string]interface{}{"virtio13": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr=2.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 2.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `mbps_wr_max`,
					input: map[string]interface{}{"virtio14": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,mbps_wr_max=4.51"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 4.51}}},
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true}}}}})},
				{name: `replicate`,
					input: map[string]interface{}{"virtio15": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,replicate=0"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: false}}}}})},
				{name: `ro`,
					input: map[string]interface{}{"virtio0": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,ro=1"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						ReadOnly:  true,
						Replicate: true}}}}})},
				{name: `serial`,
					input: map[string]interface{}{"virtio1": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,serial=disk-9763"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:    true,
						File:      "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate: true,
						Serial:    "disk-9763"}}}}})},
				{name: `size G`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1G"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1048576}}}}})},
				{name: `size K`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1K"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1}}}}})},
				{name: `size M`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1M"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1024}}}}})},
				{name: `size T`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,size=1T"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:          true,
						File:            "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:       true,
						SizeInKibibytes: 1073741824}}}}})},
				{name: `wwn`,
					input: map[string]interface{}{"virtio2": "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8,wwn=0x5008FA6500D9C8B3"},
					output: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
						Backup:        true,
						File:          "/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_drive-scsi8",
						Replicate:     true,
						WorldWideName: "0x5008FA6500D9C8B3"}}}}})}}},
		{category: `EFIDisk`,
			tests: []test{
				{name: `All`,
					input: map[string]interface{}{"efidisk0": "local-lvm:vm-1000-disk-0,efitype=2m,size=4M"},
					output: baseConfig(ConfigQemu{EFIDisk: map[string]interface{}{
						"efitype": "2m",
						"size":    "4M",
						"storage": "local-lvm",
						"file":    "vm-1000-disk-0",
						"volume":  "local-lvm:vm-1000-disk-0"}})}}},
		{category: `Iso`,
			tests: []test{
				{name: `All`,
					input: map[string]interface{}{"ide2": "local:iso/debian-11.0.0-amd64-netinst.iso,media=cdrom,size=377M"},
					output: baseConfig(ConfigQemu{
						Iso: &IsoFile{
							File:            "debian-11.0.0-amd64-netinst.iso",
							Storage:         "local",
							SizeInKibibytes: "377M"},
						Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{
							Iso: &IsoFile{
								File:            "debian-11.0.0-amd64-netinst.iso",
								Storage:         "local",
								SizeInKibibytes: "377M"}}}}}})}}},
		{category: `Memory`,
			tests: []test{
				{name: `All float64`,
					input: map[string]interface{}{
						"memory":  float64(1024),
						"balloon": float64(512),
						"shares":  float64(50)},
					output: baseConfig(ConfigQemu{Memory: &QemuMemory{
						CapacityMiB:        util.Pointer(QemuMemoryCapacity(1024)),
						MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(512)),
						Shares:             util.Pointer(QemuMemoryShares(50))}})},
				{name: `All string`,
					input: map[string]interface{}{
						"memory":  "1024",
						"balloon": "512",
						"shares":  "50"},
					output: baseConfig(ConfigQemu{Memory: &QemuMemory{
						CapacityMiB:        util.Pointer(QemuMemoryCapacity(1024)),
						MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(512)),
						Shares:             util.Pointer(QemuMemoryShares(50))}})},
				{name: `memory`,
					input:  map[string]interface{}{"memory": float64(2000)},
					output: baseConfig(ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(2000))}})},
				{name: `balloon`,
					input:  map[string]interface{}{"balloon": float64(1000)},
					output: baseConfig(ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}})},
				{name: `shares`,
					input:  map[string]interface{}{"shares": float64(100)},
					output: baseConfig(ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(100))}})}}},
		{category: `Networks`,
			tests: []test{
				{name: `all e1000`,
					input: map[string]interface{}{"net0": "e1000=BC:24:11:E1:BB:5d,bridge=vmbr0,mtu=1395,firewall=1,link_down=1,queues=23,rate=1.53,tag=12,trunks=34;18;25"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
						Bridge:    util.Pointer("vmbr0"),
						Connected: util.Pointer(false),
						Firewall:  util.Pointer(true),
						MAC:       util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:       "BC:24:11:E1:BB:5d",
						// MTU is only supported for virtio
						Model:         util.Pointer(QemuNetworkModelE1000),
						MultiQueue:    util.Pointer(QemuNetworkQueue(23)),
						RateLimitKBps: util.Pointer(QemuNetworkRate(1530)),
						NativeVlan:    util.Pointer(Vlan(12)),
						TaggedVlans:   util.Pointer(Vlans{34, 18, 25})}}})},
				{name: `all virtio`,
					input: map[string]interface{}{"net31": "virtio=BC:24:11:E1:BB:5D,bridge=vmbr0,mtu=1395,firewall=1,link_down=1,queues=23,rate=1.53,tag=12,trunks=34;18;25"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID31: QemuNetworkInterface{
						Bridge:        util.Pointer("vmbr0"),
						Connected:     util.Pointer(false),
						Firewall:      util.Pointer(true),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						MTU:           util.Pointer(QemuMTU{Value: 1395}),
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						MultiQueue:    util.Pointer(QemuNetworkQueue(23)),
						RateLimitKBps: util.Pointer(QemuNetworkRate(1530)),
						NativeVlan:    util.Pointer(Vlan(12)),
						TaggedVlans:   util.Pointer(Vlans{34, 18, 25})}}})},
				{name: `Bridge`,
					input: map[string]interface{}{"net2": "virtio=BC:24:11:E1:BB:5D,bridge=vmbr0"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{
						Bridge:      util.Pointer("vmbr0"),
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Model and Mac`,
					input: map[string]interface{}{"net3": "virtio=BC:24:11:E1:BB:5D"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Connected false`,
					input: map[string]interface{}{"net4": "virtio=BC:24:11:E1:BB:5D,link_down=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{
						Connected:   util.Pointer(false),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Connected true`,
					input: map[string]interface{}{"net5": "virtio=BC:24:11:E1:BB:5D,link_down=0"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Connected unset`,
					input: map[string]interface{}{"net6": "virtio=BC:24:11:E1:BB:5D"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Firwall true`,
					input: map[string]interface{}{"net7": "virtio=BC:24:11:E1:BB:5D,firewall=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(true),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Firwall false`,
					input: map[string]interface{}{"net8": "virtio=BC:24:11:E1:BB:5D,firewall=0"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `Firwall unset`,
					input: map[string]interface{}{"net9": "virtio=BC:24:11:E1:BB:5D"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `MTU value`,
					input: map[string]interface{}{"net10": "virtio=BC:24:11:E1:BB:5D,mtu=1500"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						MTU:         &QemuMTU{Value: 1500},
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `MTU inherit`,
					input: map[string]interface{}{"net11": "virtio=BC:24:11:E1:BB:5D,mtu=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						MTU:         &QemuMTU{Inherit: true},
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `MultiQueue disable`,
					input: map[string]interface{}{"net12": "virtio=BC:24:11:E1:BB:5D"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `MultiQueue enable`,
					input: map[string]interface{}{"net0": "virtio=BC:24:11:E1:BB:5D,queues=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MultiQueue:  util.Pointer(QemuNetworkQueue(1)),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps disable`,
					input: map[string]interface{}{"net13": "virtio=BC:24:11:E1:BB:5D"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 0.001`,
					input: map[string]interface{}{"net14": "virtio=BC:24:11:E1:BB:5D,rate=0.001"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(1)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 0.01`,
					input: map[string]interface{}{"net15": "virtio=BC:24:11:E1:BB:5D,rate=0.010"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(10)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 0.1`,
					input: map[string]interface{}{"net16": "virtio=BC:24:11:E1:BB:5D,rate=0.1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(100)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 1`,
					input: map[string]interface{}{"net17": "virtio=BC:24:11:E1:BB:5D,rate=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(1000)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 1.264`,
					input: map[string]interface{}{"net18": "virtio=BC:24:11:E1:BB:5D,rate=1.264"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID18: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(1264)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `RateLimitKBps 15.264`,
					input: map[string]interface{}{"net19": "virtio=BC:24:11:E1:BB:5D,rate=15.264"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID19: QemuNetworkInterface{
						Connected:     util.Pointer(true),
						Firewall:      util.Pointer(false),
						MAC:           util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:           "BC:24:11:E1:BB:5D",
						Model:         util.Pointer(QemuNetworkModelVirtIO),
						RateLimitKBps: util.Pointer(QemuNetworkRate(15264)),
						TaggedVlans:   util.Pointer(Vlans{})}}})},
				{name: `NaitiveVlan`,
					input: map[string]interface{}{"net20": "virtio=BC:24:11:E1:BB:5D,tag=1"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID20: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						NativeVlan:  util.Pointer(Vlan(1)),
						TaggedVlans: util.Pointer(Vlans{})}}})},
				{name: `TaggedVlans`,
					input: map[string]interface{}{"net21": "virtio=BC:24:11:E1:BB:5D,trunks=1;63;21"},
					output: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID21: QemuNetworkInterface{
						Connected:   util.Pointer(true),
						Firewall:    util.Pointer(false),
						MAC:         util.Pointer(parseMAC("BC:24:11:E1:BB:5D")),
						mac:         "BC:24:11:E1:BB:5D",
						Model:       util.Pointer(QemuNetworkModelVirtIO),
						TaggedVlans: util.Pointer(Vlans{1, 63, 21})}}})},
			},
		},
		{category: `Node`,
			tests: []test{
				{name: `vmr nil`,
					output: baseConfig(ConfigQemu{})},
				{name: `vmr empty`,
					vmr:    &VmRef{node: ""},
					output: baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(""))})},
				{name: `vmr populated`,
					vmr:    &VmRef{node: "test"},
					output: baseConfig(ConfigQemu{Node: "test", Pool: util.Pointer(PoolName(""))})}}},
		{category: `Pool`,
			tests: []test{
				{name: `vmr nil`,
					output: baseConfig(ConfigQemu{})},
				{name: `vmr empty`,
					vmr:    &VmRef{pool: ""},
					output: baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(""))})},
				{name: `vmr populated`,
					vmr:    &VmRef{pool: "test"},
					output: baseConfig(ConfigQemu{Pool: util.Pointer(PoolName("test"))})}}},
		{category: `PciDevices`,
			tests: []test{
				{name: `Mapping all`,
					input: map[string]interface{}{"hostpci0": "mapping=abc,device-id=0xa97f,pcie=1,x-vga=1,rombar=0,sub-device-id=0x61a4,sub-vendor-id=0x98f1,vendor-id=0x4003"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID0: QemuPci{
							Mapping: &QemuPciMapping{
								DeviceID:    util.Pointer(PciDeviceID("0xa97f")),
								ID:          util.Pointer(ResourceMappingPciID("abc")),
								PCIe:        util.Pointer(true),
								PrimaryGPU:  util.Pointer(true),
								ROMbar:      util.Pointer(false),
								SubDeviceID: util.Pointer(PciSubDeviceID("0x61a4")),
								SubVendorID: util.Pointer(PciSubVendorID("0x98f1")),
								VendorID:    util.Pointer(PciVendorID("0x4003"))}}}})},
				{name: `Mapping.DeviceID`,
					input: map[string]interface{}{"hostpci1": "mapping=abc,device-id=0xa97f"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID1: QemuPci{
							Mapping: &QemuPciMapping{
								DeviceID:   util.Pointer(PciDeviceID("0xa97f")),
								ID:         util.Pointer(ResourceMappingPciID("abc")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Mapping.ID`,
					input: map[string]interface{}{"hostpci2": "mapping=xyz"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID2: QemuPci{
							Mapping: &QemuPciMapping{
								ID:         util.Pointer(ResourceMappingPciID("xyz")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Mapping.Pci`,
					input: map[string]interface{}{"hostpci3": "mapping=abc,pcie=1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID3: QemuPci{
							Mapping: &QemuPciMapping{
								ID:         util.Pointer(ResourceMappingPciID("abc")),
								PCIe:       util.Pointer(true),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Mapping.PrimaryGPU`,
					input: map[string]interface{}{"hostpci4": "mapping=abc,x-vga=1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID4: QemuPci{
							Mapping: &QemuPciMapping{
								ID:         util.Pointer(ResourceMappingPciID("abc")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(true),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Mapping.ROMbar`,
					input: map[string]interface{}{"hostpci5": "mapping=abc,rombar=0"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID5: QemuPci{
							Mapping: &QemuPciMapping{
								ID:         util.Pointer(ResourceMappingPciID("abc")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(false)}}}})},
				{name: `Mapping.SubDeviceID`,
					input: map[string]interface{}{"hostpci6": "mapping=abc,sub-device-id=0x61a4"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID6: QemuPci{
							Mapping: &QemuPciMapping{
								ID:          util.Pointer(ResourceMappingPciID("abc")),
								PCIe:        util.Pointer(false),
								PrimaryGPU:  util.Pointer(false),
								ROMbar:      util.Pointer(true),
								SubDeviceID: util.Pointer(PciSubDeviceID("0x61a4"))}}}})},
				{name: `Mapping.SubVendorID`,
					input: map[string]interface{}{"hostpci7": "mapping=abc,sub-vendor-id=0x98f1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{
							Mapping: &QemuPciMapping{
								ID:          util.Pointer(ResourceMappingPciID("abc")),
								PCIe:        util.Pointer(false),
								PrimaryGPU:  util.Pointer(false),
								ROMbar:      util.Pointer(true),
								SubVendorID: util.Pointer(PciSubVendorID("0x98f1"))}}}})},
				{name: `Mapping.VendorID`,
					input: map[string]interface{}{"hostpci8": "vendor-id=0x4003,mapping=abc"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{
							Mapping: &QemuPciMapping{
								ID:         util.Pointer(ResourceMappingPciID("abc")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true),
								VendorID:   util.Pointer(PciVendorID("0x4003"))}}}})},
				{name: `Raw all`,
					input: map[string]interface{}{"hostpci15": "0000:02:05.7,device-id=0xa97f,pcie=1,x-vga=1,rombar=0,sub-device-id=0x61a4,sub-vendor-id=0x98f1,vendor-id=0x4003"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID15: QemuPci{
							Raw: &QemuPciRaw{
								DeviceID:    util.Pointer(PciDeviceID("0xa97f")),
								ID:          util.Pointer(PciID("0000:02:05.7")),
								PCIe:        util.Pointer(true),
								PrimaryGPU:  util.Pointer(true),
								ROMbar:      util.Pointer(false),
								SubDeviceID: util.Pointer(PciSubDeviceID("0x61a4")),
								SubVendorID: util.Pointer(PciSubVendorID("0x98f1")),
								VendorID:    util.Pointer(PciVendorID("0x4003"))}}}})},
				{name: `Raw.DeviceID`,
					input: map[string]interface{}{"hostpci14": "0000:02:05.7,device-id=0xa97f"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID14: QemuPci{
							Raw: &QemuPciRaw{
								DeviceID:   util.Pointer(PciDeviceID("0xa97f")),
								ID:         util.Pointer(PciID("0000:02:05.7")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Raw.ID`,
					input: map[string]interface{}{"hostpci13": "0001:43:86.5"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID13: QemuPci{
							Raw: &QemuPciRaw{
								ID:         util.Pointer(PciID("0001:43:86.5")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Raw.Pci`,
					input: map[string]interface{}{"hostpci12": "0000:02:05.7,pcie=1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID12: QemuPci{
							Raw: &QemuPciRaw{
								ID:         util.Pointer(PciID("0000:02:05.7")),
								PCIe:       util.Pointer(true),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Raw.PrimaryGPU`,
					input: map[string]interface{}{"hostpci11": "0000:02:05.7,x-vga=1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID11: QemuPci{
							Raw: &QemuPciRaw{
								ID:         util.Pointer(PciID("0000:02:05.7")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(true),
								ROMbar:     util.Pointer(true)}}}})},
				{name: `Raw.ROMbar`,
					input: map[string]interface{}{"hostpci10": "0000:02:05.7,rombar=0"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID10: QemuPci{
							Raw: &QemuPciRaw{
								ID:         util.Pointer(PciID("0000:02:05.7")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(false)}}}})},
				{name: `Raw.SubDeviceID`,
					input: map[string]interface{}{"hostpci9": "0000:02:05.7,sub-device-id=0x61a4"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID9: QemuPci{
							Raw: &QemuPciRaw{
								ID:          util.Pointer(PciID("0000:02:05.7")),
								PCIe:        util.Pointer(false),
								PrimaryGPU:  util.Pointer(false),
								ROMbar:      util.Pointer(true),
								SubDeviceID: util.Pointer(PciSubDeviceID("0x61a4"))}}}})},
				{name: `Raw.SubVendorID`,
					input: map[string]interface{}{"hostpci8": "0000:02:05.7,sub-vendor-id=0x98f1"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID8: QemuPci{
							Raw: &QemuPciRaw{
								ID:          util.Pointer(PciID("0000:02:05.7")),
								PCIe:        util.Pointer(false),
								PrimaryGPU:  util.Pointer(false),
								ROMbar:      util.Pointer(true),
								SubVendorID: util.Pointer(PciSubVendorID("0x98f1"))}}}})},
				{name: `Raw.VendorID`,
					input: map[string]interface{}{"hostpci7": "0000:02:05.7,vendor-id=0x4003"},
					output: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
						QemuPciID7: QemuPci{
							Raw: &QemuPciRaw{
								ID:         util.Pointer(PciID("0000:02:05.7")),
								PCIe:       util.Pointer(false),
								PrimaryGPU: util.Pointer(false),
								ROMbar:     util.Pointer(true),
								VendorID:   util.Pointer(PciVendorID("0x4003"))}}}})},
			},
		},
		{category: `Serials`,
			tests: []test{
				{name: `All`,
					input: map[string]interface{}{
						"serial2": "/dev/tty1",
						"serial0": "socket",
						"serial3": "/dev/tty3",
						"serial1": "socket"},
					output: baseConfig(ConfigQemu{Serials: SerialInterfaces{
						SerialID0: SerialInterface{Socket: true},
						SerialID1: SerialInterface{Socket: true},
						SerialID2: SerialInterface{Path: "/dev/tty1"},
						SerialID3: SerialInterface{Path: "/dev/tty3"}}})},
				{name: `single path`,
					input:  map[string]interface{}{"serial3": "/dev/test/tty7"},
					output: baseConfig(ConfigQemu{Serials: SerialInterfaces{SerialID3: SerialInterface{Path: "/dev/test/tty7"}}})},
				{name: `single socket`,
					input:  map[string]interface{}{"serial2": "socket"},
					output: baseConfig(ConfigQemu{Serials: SerialInterfaces{SerialID2: SerialInterface{Socket: true}}})}}},
		{category: `TPM`,
			tests: []test{
				{name: `All`,
					input:  map[string]interface{}{"tpmstate0": string("local-lvm:vm-101-disk-0,size=4M,version=v2.0")},
					output: baseConfig(ConfigQemu{TPM: &TpmState{Storage: "local-lvm", Version: util.Pointer(TpmVersion("v2.0"))}})}}},
		{category: `USBs`,
			tests: []test{
				{name: `Device`,
					input: map[string]interface{}{
						"usb0": "host=1234:5678"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("1234:5678")),
							USB3: util.Pointer(false)}}}})},
				{name: `Device usb3`,
					input: map[string]interface{}{
						"usb1": "host=1234:5678,usb3=1"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Device: &QemuUsbDevice{
							ID:   util.Pointer(UsbDeviceID("1234:5678")),
							USB3: util.Pointer(true)}}}})},
				{name: `Port`,
					input: map[string]interface{}{"usb2": "host=1-2"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("1-2")),
							USB3: util.Pointer(false)}}}})},
				{name: `Port usb3`,
					input: map[string]interface{}{"usb3": "host=2-4,usb3=1"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID3: QemuUSB{Port: &QemuUsbPort{
							ID:   util.Pointer(UsbPortID("2-4")),
							USB3: util.Pointer(true)}}}})},
				{name: `mapping`,
					input: map[string]interface{}{"usb4": "mapping=abcde"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID4: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("abcde")),
							USB3: util.Pointer(false)}}}})},
				{name: `mapping usb3`,
					input: map[string]interface{}{"usb0": "mapping=testmapping,usb3=1"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
							ID:   util.Pointer(ResourceMappingUsbID("testmapping")),
							USB3: util.Pointer(true)}}}})},
				{name: `spice`,
					input: map[string]interface{}{"usb1": "spice"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID1: QemuUSB{Spice: &QemuUsbSpice{USB3: false}}}})},
				{name: `spice usb3`,
					input: map[string]interface{}{"usb2": "spice,usb3=1"},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{
						QemuUsbID2: QemuUSB{Spice: &QemuUsbSpice{USB3: true}}}})},
				{name: `code coverage`,
					input:  map[string]interface{}{"usb3": ""},
					output: baseConfig(ConfigQemu{USBs: QemuUSBs{QemuUsbID3: QemuUSB{}}})}}},
		{category: `VmID`,
			tests: []test{
				{name: `vmr nil`,
					output: baseConfig(ConfigQemu{})},
				{name: `vmr empty`,
					vmr:    &VmRef{vmId: 0},
					output: baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(""))})},
				{name: `vmr populated`,
					vmr:    &VmRef{vmId: 100},
					output: baseConfig(ConfigQemu{VmID: 100, Pool: util.Pointer(PoolName(""))})}}},
	}
	for _, test := range tests {
		for _, subTest := range test.tests {
			name := test.category
			if len(test.tests) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				output, err := ConfigQemu{}.mapToStruct(subTest.vmr, subTest.input)
				if err != nil {
					require.Equal(t, subTest.err, err, name)
				} else {
					require.Equal(t, subTest.output, output, name)
				}
			})
		}
	}
}

func Test_ConfigQemu_Validate(t *testing.T) {
	BandwidthValid0 := func() QemuDiskBandwidth {
		return QemuDiskBandwidth{
			MBps: QemuDiskBandwidthMBps{
				ReadLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      0,
					Concurrent: 0},
				WriteLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      0,
					Concurrent: 0}},
			Iops: QemuDiskBandwidthIops{
				ReadLimit: QemuDiskBandwidthIopsLimit{
					Burst:      0,
					Concurrent: 0},
				WriteLimit: QemuDiskBandwidthIopsLimit{
					Burst:      0,
					Concurrent: 0}}}
	}
	BandwidthValid1 := func() QemuDiskBandwidth {
		return QemuDiskBandwidth{
			MBps: QemuDiskBandwidthMBps{
				ReadLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      1,
					Concurrent: 1},
				WriteLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      1,
					Concurrent: 1}},
			Iops: QemuDiskBandwidthIops{
				ReadLimit: QemuDiskBandwidthIopsLimit{
					Burst:      10,
					Concurrent: 10},
				WriteLimit: QemuDiskBandwidthIopsLimit{
					Burst:      10,
					Concurrent: 10}}}
	}
	BandwidthValid2 := func() QemuDiskBandwidth {
		return QemuDiskBandwidth{
			MBps: QemuDiskBandwidthMBps{
				ReadLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      1,
					Concurrent: 0},
				WriteLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      1,
					Concurrent: 0}},
			Iops: QemuDiskBandwidthIops{
				ReadLimit: QemuDiskBandwidthIopsLimit{
					Burst:      10,
					Concurrent: 0},
				WriteLimit: QemuDiskBandwidthIopsLimit{
					Burst:      10,
					Concurrent: 0}}}
	}
	BandwidthValid3 := func() QemuDiskBandwidth {
		return QemuDiskBandwidth{
			MBps: QemuDiskBandwidthMBps{
				ReadLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      0,
					Concurrent: 1},
				WriteLimit: QemuDiskBandwidthMBpsLimit{
					Burst:      0,
					Concurrent: 1}},
			Iops: QemuDiskBandwidthIops{
				ReadLimit: QemuDiskBandwidthIopsLimit{
					Burst:      0,
					Concurrent: 10},
				WriteLimit: QemuDiskBandwidthIopsLimit{
					Burst:      0,
					Concurrent: 10}}}
	}
	baseConfig := func(config ConfigQemu) ConfigQemu {
		if config.CPU == nil {
			config.CPU = &QemuCPU{Cores: util.Pointer(QemuCpuCores(1))}
		} else if config.CPU.Cores == nil {
			config.CPU.Cores = util.Pointer(QemuCpuCores(1))
		}
		if config.Memory == nil {
			config.Memory = &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1024))}
		}
		return config
	}
	baseNetwork := func(id QemuNetworkInterfaceID, config QemuNetworkInterface) QemuNetworkInterfaces {
		if config.Bridge == nil {
			config.Bridge = util.Pointer("vmbr0")
		}
		if config.Model == nil {
			config.Model = util.Pointer(QemuNetworkModelVirtIO)
		}
		return QemuNetworkInterfaces{id: config}
	}
	validCloudInit := QemuCloudInitDisk{Format: QemuDiskFormat_Raw, Storage: "Test"}
	validTags := func() []Tag {
		array := test_data_tag.Tag_Legal()
		tags := make([]Tag, len(array))
		for i, e := range array {
			tags[i] = Tag(e)
		}
		return tags
	}
	type test struct {
		name    string
		input   ConfigQemu
		current *ConfigQemu
		err     error
		version Version
	}
	type testType struct {
		create       []test
		createUpdate []test // value of currentConfig wil be used for update and ignored for create
		update       []test
	}
	tests := []struct {
		category string
		valid    testType
		invalid  testType
	}{
		{category: `Agent`,
			valid: testType{
				createUpdate: []test{
					{input: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType("isa"))}}),
						current: &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}}}}},
			invalid: testType{
				createUpdate: []test{
					{input: baseConfig(ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType("test"))}}),
						current: &ConfigQemu{Agent: &QemuGuestAgent{Type: util.Pointer(QemuGuestAgentType_VirtIO)}},
						err:     errors.New(QemuGuestAgentType_Error_Invalid)}}}},
		{category: `CloudInit`,
			valid: testType{
				createUpdate: []test{
					{name: `All v7`,
						version: Version{Major: 7, Minor: 255, Patch: 255},
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{
							Custom: &CloudInitCustom{
								Meta:    &CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())},
								Network: &CloudInitSnippet{FilePath: ""},
								User:    &CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())},
								Vendor:  &CloudInitSnippet{FilePath: ""}},
							NetworkInterfaces: CloudInitNetworkInterfaces{
								QemuNetworkInterfaceID0: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1/24"))})},
								QemuNetworkInterfaceID1: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))})},
								QemuNetworkInterfaceID2: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{
										Address: util.Pointer(IPv4CIDR("")),
										DHCP:    true})},
								QemuNetworkInterfaceID3: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{
										Gateway: util.Pointer(IPv4Address("")),
										DHCP:    true})},
								QemuNetworkInterfaceID4: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1"))})},
								QemuNetworkInterfaceID5: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))})},
								QemuNetworkInterfaceID9: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64"))})},
								QemuNetworkInterfaceID10: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))})},
								QemuNetworkInterfaceID11: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Address: util.Pointer(IPv6CIDR("")),
										DHCP:    true})},
								QemuNetworkInterfaceID12: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Gateway: util.Pointer(IPv6Address("")),
										DHCP:    true})},
								QemuNetworkInterfaceID13: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})},
								QemuNetworkInterfaceID14: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))})},
								QemuNetworkInterfaceID15: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Address: util.Pointer(IPv6CIDR("")),
										SLAAC:   true})},
								QemuNetworkInterfaceID16: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Gateway: util.Pointer(IPv6Address("")),
										SLAAC:   true})}},
							UpgradePackages: util.Pointer(false)}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}}},
					{name: `All v8`,
						version: Version{Major: 8},
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{
							Custom: &CloudInitCustom{
								Meta:    &CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())},
								Network: &CloudInitSnippet{FilePath: ""},
								User:    &CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Legal())},
								Vendor:  &CloudInitSnippet{FilePath: ""}},
							NetworkInterfaces: CloudInitNetworkInterfaces{
								QemuNetworkInterfaceID0: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1/24"))})},
								QemuNetworkInterfaceID1: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR(""))})},
								QemuNetworkInterfaceID2: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{
										Address: util.Pointer(IPv4CIDR("")),
										DHCP:    true})},
								QemuNetworkInterfaceID3: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{
										Gateway: util.Pointer(IPv4Address("")),
										DHCP:    true})},
								QemuNetworkInterfaceID4: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1"))})},
								QemuNetworkInterfaceID5: CloudInitNetworkConfig{
									IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address(""))})},
								QemuNetworkInterfaceID9: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64"))})},
								QemuNetworkInterfaceID10: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR(""))})},
								QemuNetworkInterfaceID11: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Address: util.Pointer(IPv6CIDR("")),
										DHCP:    true})},
								QemuNetworkInterfaceID12: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Gateway: util.Pointer(IPv6Address("")),
										DHCP:    true})},
								QemuNetworkInterfaceID13: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})},
								QemuNetworkInterfaceID14: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address(""))})},
								QemuNetworkInterfaceID15: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Address: util.Pointer(IPv6CIDR("")),
										SLAAC:   true})},
								QemuNetworkInterfaceID16: CloudInitNetworkConfig{
									IPv6: util.Pointer(CloudInitIPv6Config{
										Gateway: util.Pointer(IPv6Address("")),
										SLAAC:   true})}},
							UpgradePackages: util.Pointer(true)}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(CloudInit_Error_UpgradePackagesPre8)`,
						version: Version{Major: 7, Minor: 255, Patch: 255},
						input:   baseConfig(ConfigQemu{CloudInit: &CloudInit{UpgradePackages: util.Pointer(true)}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInit_Error_UpgradePackagesPre8)},
					{name: `errors.New(CloudInitSnippetPath_Error_InvalidCharacters)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{Meta: &CloudInitSnippet{
							FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Character_Illegal()[0])}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitSnippetPath_Error_InvalidCharacters)},
					{name: `errors.New(CloudInitSnippetPath_Error_InvalidPath)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{Network: &CloudInitSnippet{
							FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_InvalidPath())}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitSnippetPath_Error_InvalidPath)},
					{name: `errors.New(CloudInitSnippetPath_Error_MaxLength)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{User: &CloudInitSnippet{
							FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Max_Illegal())}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitSnippetPath_Error_MaxLength)},
					{name: `errors.New(CloudInitSnippetPath_Error_Relative)`,
						input:   baseConfig(ConfigQemu{CloudInit: &CloudInit{Custom: &CloudInitCustom{Vendor: &CloudInitSnippet{FilePath: CloudInitSnippetPath(test_data_qemu.CloudInitSnippetPath_Relative())}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitSnippetPath_Error_Relative)},
					{name: `errors.New(QemuNetworkInterfaceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{
							32: CloudInitNetworkConfig{}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(QemuNetworkInterfaceID_Error_Invalid)},
					{name: `CloudInitNetworkInterfaces IPv4 Address Mutually exclusive with DHCP`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID5: CloudInitNetworkConfig{
							IPv4: util.Pointer(CloudInitIPv4Config{
								Address: util.Pointer(IPv4CIDR("192.168.45.1/24")),
								DHCP:    true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv4Config_Error_DhcpAddressMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv4 Gateway Mutually exclusive with DHCP`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID6: CloudInitNetworkConfig{
							IPv4: util.Pointer(CloudInitIPv4Config{
								Gateway: util.Pointer(IPv4Address("192.168.45.1")),
								DHCP:    true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv4Config_Error_DhcpGatewayMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv4 Address errors.New(IPv4CIDR_Error_Invalid)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID7: CloudInitNetworkConfig{
							IPv4: util.Pointer(CloudInitIPv4Config{Address: util.Pointer(IPv4CIDR("192.168.45.1"))})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(IPv4CIDR_Error_Invalid)},
					{name: `CloudInitNetworkInterfaces IPv4 Gateway errors.New(IPv4Address_Error_Invalid)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID8: CloudInitNetworkConfig{
							IPv4: util.Pointer(CloudInitIPv4Config{Gateway: util.Pointer(IPv4Address("192.168.45.1/24"))})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(IPv4Address_Error_Invalid)},
					{name: `CloudInitNetworkInterfaces IPv6 Address Mutually exclusive with DHCP`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID17: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
								DHCP:    true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv6Config_Error_DhcpAddressMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv6 Address Mutually exclusive with SLAAC`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID18: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{
								Address: util.Pointer(IPv6CIDR("2001:0db8:85a3::/64")),
								SLAAC:   true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv6Config_Error_SlaacAddressMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv6 DHCP Mutually exclusive with SLAAC`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID19: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{
								DHCP:  true,
								SLAAC: true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv6Config_Error_DhcpSlaacMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv6 Gateway Mutually exclusive with DHCP`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID20: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{
								Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
								DHCP:    true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv6Config_Error_DhcpGatewayMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv6 Gateway Mutually exclusive with SLAAC`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID21: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{
								Gateway: util.Pointer(IPv6Address("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc")),
								SLAAC:   true})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(CloudInitIPv6Config_Error_SlaacGatewayMutuallyExclusive)},
					{name: `CloudInitNetworkInterfaces IPv6 Address errors.New(IPv6CIDR_Error_Invalid)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID22: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{Address: util.Pointer(IPv6CIDR("3f6d:5b2a:1e4d:7c91:abcd:1234:5678:9abc"))})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(IPv6CIDR_Error_Invalid)},
					{name: `CloudInitNetworkInterfaces IPv6 Gateway errors.New(IPv6Address_Error_Invalid)`,
						input: baseConfig(ConfigQemu{CloudInit: &CloudInit{NetworkInterfaces: CloudInitNetworkInterfaces{QemuNetworkInterfaceID23: CloudInitNetworkConfig{
							IPv6: util.Pointer(CloudInitIPv6Config{Gateway: util.Pointer(IPv6Address("2001:0db8:85a3::/64"))})}}}}),
						current: &ConfigQemu{CloudInit: &CloudInit{}},
						err:     errors.New(IPv6Address_Error_Invalid)}}}},
		{category: `CPU`,
			valid: testType{
				createUpdate: []test{
					{name: `Cores`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Cores: util.Pointer(QemuCpuCores(1))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Maximum`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{
							Cores:        util.Pointer(QemuCpuCores(128)),
							Sockets:      util.Pointer(QemuCpuSockets(4)),
							VirtualCores: util.Pointer(CpuVirtualCores(512))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Minimum`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{
							Cores:        util.Pointer(QemuCpuCores(128)),
							Sockets:      util.Pointer(QemuCpuSockets(4)),
							VirtualCores: util.Pointer(CpuVirtualCores(0))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Flags all set`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							AES:        util.Pointer(TriBoolFalse),
							AmdNoSSB:   util.Pointer(TriBoolNone),
							AmdSSBD:    util.Pointer(TriBoolTrue),
							HvEvmcs:    util.Pointer(TriBoolFalse),
							HvTlbFlush: util.Pointer(TriBoolNone),
							Ibpb:       util.Pointer(TriBoolTrue),
							MdClear:    util.Pointer(TriBoolFalse),
							PCID:       util.Pointer(TriBoolNone),
							Pdpe1GB:    util.Pointer(TriBoolTrue),
							SSBD:       util.Pointer(TriBoolFalse),
							SpecCtrl:   util.Pointer(TriBoolNone),
							VirtSSBD:   util.Pointer(TriBoolTrue)}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Flags all nil`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Flags mixed`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							AmdNoSSB:   util.Pointer(TriBoolTrue),
							AmdSSBD:    util.Pointer(TriBoolFalse),
							HvTlbFlush: util.Pointer(TriBoolTrue),
							Ibpb:       util.Pointer(TriBoolFalse),
							MdClear:    util.Pointer(TriBoolNone),
							PCID:       util.Pointer(TriBoolTrue),
							Pdpe1GB:    util.Pointer(TriBoolFalse),
							SpecCtrl:   util.Pointer(TriBoolTrue)}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Limit maximum`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(128))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Limit minimum`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(0))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Sockets`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Sockets: util.Pointer(QemuCpuSockets(1))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Type empty`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType(""))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Type host`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType_Host)}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Units Minimum`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(0))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}},
					{name: `Units Maximum`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(262144))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}}},
				update: []test{
					{name: `nothing`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{}}),
						current: &ConfigQemu{CPU: &QemuCPU{}}}}},
			invalid: testType{
				create: []test{
					{name: `erross.New(ConfigQemu_Error_CpuRequired)`,
						err: errors.New(ConfigQemu_Error_CpuRequired)},
					{name: `errors.New(QemuCPU_Error_CoresRequired)`,
						input: ConfigQemu{CPU: &QemuCPU{}},
						err:   errors.New(QemuCPU_Error_CoresRequired)}},
				createUpdate: []test{
					{name: `errors.New(CpuLimit_Error_Maximum)`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Limit: util.Pointer(CpuLimit(129))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(CpuLimit_Error_Maximum)},
					{name: `errors.New(CpuUnits_Error_Maximum)`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Units: util.Pointer(CpuUnits(262145))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(CpuUnits_Error_Maximum)},
					{name: `errors.New(QemuCpuCores_Error_LowerBound)`,
						input:   ConfigQemu{CPU: &QemuCPU{Cores: util.Pointer(QemuCpuCores(0))}},
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(QemuCpuCores_Error_LowerBound)},
					{name: `errors.New(QemuCpuSockets_Error_LowerBound)`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{
							Cores:   util.Pointer(QemuCpuCores(1)),
							Sockets: util.Pointer(QemuCpuSockets(0))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(QemuCpuSockets_Error_LowerBound)},
					{name: `CpuVirtualCores(1).Error() 1 1`,
						input: ConfigQemu{CPU: &QemuCPU{
							Cores:        util.Pointer(QemuCpuCores(1)),
							Sockets:      util.Pointer(QemuCpuSockets(1)),
							VirtualCores: util.Pointer(CpuVirtualCores(2))}},
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     CpuVirtualCores(1).Error()},
					{name: `Invalid AES`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							AES: util.Pointer(TriBool(-2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid AmdNoSSB`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							AmdNoSSB: util.Pointer(TriBool(2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid AmdSSBD`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							AmdSSBD: util.Pointer(TriBool(-27))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid HvEvmcs`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							HvEvmcs: util.Pointer(TriBool(32))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid HvTlbFlush`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							HvTlbFlush: util.Pointer(TriBool(-2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid Ibpb`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							Ibpb: util.Pointer(TriBool(52))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid MdClear`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							MdClear: util.Pointer(TriBool(-52))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid PCID`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							PCID: util.Pointer(TriBool(82))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid Pdpe1GB`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							Pdpe1GB: util.Pointer(TriBool(-2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid SSBD`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							SSBD: util.Pointer(TriBool(3))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid SpecCtrl`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							SpecCtrl: util.Pointer(TriBool(-2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Invalid VirtSSBD`,
						input: baseConfig(ConfigQemu{CPU: &QemuCPU{Flags: &CpuFlags{
							VirtSSBD: util.Pointer(TriBool(2))}}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						err:     errors.New(TriBool_Error_Invalid)},
					{name: `Type`,
						input:   baseConfig(ConfigQemu{CPU: &QemuCPU{Type: util.Pointer(CpuType("invalid"))}}),
						current: &ConfigQemu{CPU: &QemuCPU{}},
						version: Version{}.max(),
						err:     CpuType("").Error(Version{}.max())}}}},
		{category: `Disks`,
			valid: testType{
				create: []test{
					{name: `Empty 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{}})},
					{name: `Empty 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{}},
							Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{}},
							Scsi:   &QemuScsiDisks{Disk_0: &QemuScsiStorage{}},
							VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{}}}})}}},
			invalid: testType{
				create: []test{
					{name: `MutuallyExclusive Ide 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{
							CdRom: &QemuCdRom{},
							Disk:  &QemuIdeDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 2`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{
							CdRom:       &QemuCdRom{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 3`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuIdeDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 4`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 5`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{
							Disk:        &QemuIdeDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 6`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuIdeDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 7`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuIdeDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 8`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{
							CdRom:       &QemuCdRom{},
							Disk:        &QemuIdeDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 9`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Ide 10`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuIdeDisk{},
							Passthrough: &QemuIdePassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{
							CdRom: &QemuCdRom{},
							Disk:  &QemuSataDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 2`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{
							CdRom:       &QemuCdRom{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 3`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuSataDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 4`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 5`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{
							Disk:        &QemuSataDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 6`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuSataDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 7`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuSataDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 8`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{
							CdRom:       &QemuCdRom{},
							Disk:        &QemuSataDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 9`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Sata 10`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuSataDisk{},
							Passthrough: &QemuSataPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{
							CdRom: &QemuCdRom{},
							Disk:  &QemuScsiDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 2`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{
							CdRom:       &QemuCdRom{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 3`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuScsiDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 4`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 5`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{
							Disk:        &QemuScsiDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 6`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuScsiDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 7`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuScsiDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 8`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{
							CdRom:       &QemuCdRom{},
							Disk:        &QemuScsiDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 9`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive Scsi 10`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuScsiDisk{},
							Passthrough: &QemuScsiPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{
							CdRom: &QemuCdRom{},
							Disk:  &QemuVirtIODisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 2`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{
							CdRom:       &QemuCdRom{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 3`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuVirtIODisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 4`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 5`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{
							Disk:        &QemuVirtIODisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 6`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{
							CdRom:     &QemuCdRom{},
							CloudInit: &QemuCloudInitDisk{},
							Disk:      &QemuVirtIODisk{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 7`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuVirtIODisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 8`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{
							CdRom:       &QemuCdRom{},
							Disk:        &QemuVirtIODisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 9`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)},
					{name: `MutuallyExclusive VirtIO 10`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{
							CdRom:       &QemuCdRom{},
							CloudInit:   &QemuCloudInitDisk{},
							Disk:        &QemuVirtIODisk{},
							Passthrough: &QemuVirtIOPassthrough{}}}}}),
						err: errors.New(Error_QemuDisk_MutuallyExclusive)}}}},
		{category: `Disks CdRom`,
			valid: testType{
				create: []test{
					{input: baseConfig(ConfigQemu{Disks: &QemuStorages{
						Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{}}},
						Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}}}},
						Scsi:   &QemuScsiDisks{Disk_0: &QemuScsiStorage{CdRom: &QemuCdRom{Passthrough: true}}},
						VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}}}}}})}}},
			invalid: testType{
				create: []test{
					{name: `Ide errors.New(Error_IsoFile_File) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Ide errors.New(Error_IsoFile_File) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{Storage: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Ide errors.New(Error_IsoFile_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_Storage)},
					{name: `Ide errors.New(Error_QemuCdRom_MutuallyExclusive)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}, Passthrough: true}}}}}),
						err:   errors.New(Error_QemuCdRom_MutuallyExclusive)},
					{name: `Sata errors.New(Error_IsoFile_File) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Sata errors.New(Error_IsoFile_File) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{Storage: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Sata errors.New(Error_IsoFile_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_Storage)},
					{name: `Sata errors.New(Error_QemuCdRom_MutuallyExclusive)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}, Passthrough: true}}}}}),
						err:   errors.New(Error_QemuCdRom_MutuallyExclusive)},
					{name: `Scsi errors.New(Error_IsoFile_File) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Scsi errors.New(Error_IsoFile_File) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{Storage: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `Scsi errors.New(Error_IsoFile_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_Storage)},
					{name: `Scsi errors.New(Error_QemuCdRom_MutuallyExclusive)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}, Passthrough: true}}}}}),
						err:   errors.New(Error_QemuCdRom_MutuallyExclusive)},
					{name: `VirtIO errors.New(Error_IsoFile_File) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `VirtIO errors.New(Error_IsoFile_File) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{Storage: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_File)},
					{name: `VirtIO errors.New(Error_IsoFile_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test"}}}}}}),
						err:   errors.New(Error_IsoFile_Storage)},
					{name: `VirtIO errors.New(Error_QemuCdRom_MutuallyExclusive)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{CdRom: &QemuCdRom{Iso: &IsoFile{File: "test", Storage: "test"}, Passthrough: true}}}}}),
						err:   errors.New(Error_QemuCdRom_MutuallyExclusive)}}}},
		{category: `Disks CloudInit`,
			valid: testType{
				create: []test{
					{name: `Ide`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: &validCloudInit}}}})},
					{name: `Sata`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: &validCloudInit}}}})},
					{name: `Scsi`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CloudInit: &validCloudInit}}}})},
					{name: `VirtIO`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CloudInit: &validCloudInit}}}})}}},
			invalid: testType{
				create: []test{
					{name: `Duplicate errors.New(Error_QemuCloudInitDisk_OnlyOne)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							Ide:    &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: &validCloudInit}},
							Sata:   &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: &validCloudInit}},
							Scsi:   &QemuScsiDisks{Disk_0: &QemuScsiStorage{CloudInit: &validCloudInit}},
							VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CloudInit: &validCloudInit}}}}),
						err: errors.New(Error_QemuCloudInitDisk_OnlyOne)},
					{name: `Duplicate Ide errors.New(Error_QemuCloudInitDisk_OnlyOne)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							Ide: &QemuIdeDisks{
								Disk_0: &QemuIdeStorage{CloudInit: &validCloudInit},
								Disk_1: &QemuIdeStorage{CloudInit: &validCloudInit}}}}),
						err: errors.New(Error_QemuCloudInitDisk_OnlyOne)},
					{name: `Duplicate Sata errors.New(Error_QemuCloudInitDisk_OnlyOne)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							Sata: &QemuSataDisks{
								Disk_0: &QemuSataStorage{CloudInit: &validCloudInit},
								Disk_1: &QemuSataStorage{CloudInit: &validCloudInit}}}}),
						err: errors.New(Error_QemuCloudInitDisk_OnlyOne)},
					{name: `Duplicate Scsi errors.New(Error_QemuCloudInitDisk_OnlyOne)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							Scsi: &QemuScsiDisks{
								Disk_0: &QemuScsiStorage{CloudInit: &validCloudInit},
								Disk_1: &QemuScsiStorage{CloudInit: &validCloudInit}}}}),
						err: errors.New(Error_QemuCloudInitDisk_OnlyOne)},
					{name: `Duplicate VirtIO errors.New(Error_QemuCloudInitDisk_OnlyOne)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{
							VirtIO: &QemuVirtIODisks{
								Disk_0: &QemuVirtIOStorage{CloudInit: &validCloudInit},
								Disk_1: &QemuVirtIOStorage{CloudInit: &validCloudInit}}}}),
						err: errors.New(Error_QemuCloudInitDisk_OnlyOne)},
					{name: `Ide QemuDiskFormat("").Error() 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Ide QemuDiskFormat("").Error() 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Storage: "test"}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Ide errors.New(Error_QemuCloudInitDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw}}}}}),
						err:   errors.New(Error_QemuCloudInitDisk_Storage)},
					{name: `Sata QemuDiskFormat("").Error() 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Sata QemuDiskFormat("").Error() 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Storage: "test"}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Sata errors.New(Error_QemuCloudInitDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw}}}}}),
						err:   errors.New(Error_QemuCloudInitDisk_Storage)},
					{name: `Scsi QemuDiskFormat("").Error() 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Scsi QemuDiskFormat("").Error() 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Storage: "test"}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Scsi errors.New(Error_QemuCloudInitDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw}}}}}),
						err:   errors.New(Error_QemuCloudInitDisk_Storage)},
					{name: `VirtIO QemuDiskFormat("").Error() 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `VirtIO QemuDiskFormat("").Error() 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Storage: "test"}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `VirtIO errors.New(Error_QemuCloudInitDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{CloudInit: &QemuCloudInitDisk{Format: QemuDiskFormat_Raw}}}}}),
						err:   errors.New(Error_QemuCloudInitDisk_Storage)}}}},
		{category: `Disks Disk`,
			valid: testType{
				create: []test{
					{input: baseConfig(ConfigQemu{Disks: &QemuStorages{
						Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
							AsyncIO:         QemuDiskAsyncIO_IOuring,
							Bandwidth:       BandwidthValid0(),
							Cache:           QemuDiskCache_DirectSync,
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 5748543,
							Storage:         "test",
							WorldWideName:   "0x500A1B2C3D4E5F60"}}},
						Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							AsyncIO:         QemuDiskAsyncIO_Native,
							Bandwidth:       BandwidthValid1(),
							Cache:           QemuDiskCache_None,
							Format:          QemuDiskFormat_Cow,
							SizeInKibibytes: 4097,
							Storage:         "test",
							WorldWideName:   "0x500F123456789ABC"}}},
						Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{
							AsyncIO:         QemuDiskAsyncIO_Threads,
							Bandwidth:       BandwidthValid2(),
							Cache:           QemuDiskCache_WriteBack,
							Format:          QemuDiskFormat_Qcow2,
							SizeInKibibytes: 9475478,
							Storage:         "test",
							WorldWideName:   "0x5009876543210DEF"}}},
						VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							AsyncIO:         "",
							Bandwidth:       BandwidthValid3(),
							Cache:           "",
							Format:          QemuDiskFormat_Vmdk,
							SizeInKibibytes: 18742,
							Storage:         "test",
							WorldWideName:   "0x500C0D0E0F101112"}}}}})}}},
			invalid: testType{
				create: []test{
					{name: `Ide QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Ide QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Ide QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Ide QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{Format: ""}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Ide errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format: QemuDiskFormat_Raw,
							Serial: "!@^$^&$^&"}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Ide errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format: QemuDiskFormat_Raw,
							Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Ide errors.New(QemuDiskSize_Error_Minimum)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 4096}}}}}),
						err: errors.New(QemuDiskSize_Error_Minimum)},
					{name: `Ide errors.New(Error_QemuDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: qemuDiskSize_Minimum,
							Storage:         ""}}}}}),
						err: errors.New(Error_QemuDisk_Storage)},
					{name: `Ide errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Disk: &QemuIdeDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 32,
							Storage:         "Test",
							WorldWideName:   "0xG123456789ABCDE"}}}}}),
						err: errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `Sata QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Sata QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Sata QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Disk: &QemuSataDisk{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Sata QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Disk: &QemuSataDisk{Format: ""}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Sata errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Format: QemuDiskFormat_Raw,
							Serial: "!@^$^&$^&"}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Sata errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Disk: &QemuSataDisk{
							Format: QemuDiskFormat_Raw,
							Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Sata errors.New(QemuDiskSize_Error_Minimum)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 4096}}}}}),
						err: errors.New(QemuDiskSize_Error_Minimum)},
					{name: `Sata errors.New(Error_QemuDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: qemuDiskSize_Minimum,
							Storage:         ""}}}}}),
						err: errors.New(Error_QemuDisk_Storage)},
					{name: `Sata errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Disk: &QemuSataDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 32,
							Storage:         "Test",
							WorldWideName:   "500A1B2C3D4E5F60"}}}}}),
						err: errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `Scsi QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Disk: &QemuScsiDisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Scsi QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Disk: &QemuScsiDisk{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Disk: &QemuScsiDisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Scsi QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Disk: &QemuScsiDisk{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Scsi QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Disk: &QemuScsiDisk{Format: ""}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `Scsi errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format: QemuDiskFormat_Raw,
							Serial: "!@^$^&$^&"}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Scsi errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format: QemuDiskFormat_Raw,
							Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Scsi errors.New(QemuDiskSize_Error_Minimum)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_14: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 0}}}}}),
						err: errors.New(QemuDiskSize_Error_Minimum)},
					{name: `Scsi errors.New(Error_QemuDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_15: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: qemuDiskSize_Minimum,
							Storage:         ""}}}}}),
						err: errors.New(Error_QemuDisk_Storage)},
					{name: `Scsi errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_16: &QemuScsiStorage{Disk: &QemuScsiDisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 32,
							Storage:         "Test",
							WorldWideName:   "0x5009876543210DEFG"}}}}}),
						err: errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `VirtIO QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `VirtIO QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `VirtIO QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `VirtIO QemuDiskFormat("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{Format: ""}}}}}),
						err:   QemuDiskFormat("").Error()},
					{name: `VirtIO errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format: QemuDiskFormat_Raw,
							Serial: "!@^$^&$^&"}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `VirtIO errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format: QemuDiskFormat_Raw,
							Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err: errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `VirtIO errors.New(QemuDiskSize_Error_Minimum)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_14: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 1024}}}}}),
						err: errors.New(QemuDiskSize_Error_Minimum)},
					{name: `VirtIO errors.New(Error_QemuDisk_Storage)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_15: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: qemuDiskSize_Minimum,
							Storage:         ""}}}}}),
						err: errors.New(Error_QemuDisk_Storage)},
					{name: `VirtIO errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Disk: &QemuVirtIODisk{
							Format:          QemuDiskFormat_Raw,
							SizeInKibibytes: 32,
							Storage:         "Test",
							WorldWideName:   "500C0D0E0F10111"}}}}}),
						err: errors.New(Error_QemuWorldWideName_Invalid)}}}},
		{category: `Disks Passthrough`,
			valid: testType{
				create: []test{
					{input: baseConfig(ConfigQemu{Disks: &QemuStorages{
						Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{
							AsyncIO:       QemuDiskAsyncIO_IOuring,
							Bandwidth:     BandwidthValid3(),
							Cache:         QemuDiskCache_DirectSync,
							File:          "test",
							WorldWideName: "0x5001A2B3C4D5E6F7"}}},
						Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{
							AsyncIO:       QemuDiskAsyncIO_Native,
							Bandwidth:     BandwidthValid2(),
							Cache:         "",
							File:          "test",
							WorldWideName: "0x500B0A0908070605"}}},
						Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{
							AsyncIO:       QemuDiskAsyncIO_Threads,
							Bandwidth:     BandwidthValid1(),
							Cache:         QemuDiskCache_WriteBack,
							File:          "test",
							WorldWideName: "0x500F1E2D3C4B5A69"}}},
						VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{
							AsyncIO:       "",
							Bandwidth:     BandwidthValid0(),
							Cache:         QemuDiskCache_WriteThrough,
							File:          "test",
							WorldWideName: "0x5004A3B2C1D0E0F1"}}}}})}}},
			invalid: testType{
				create: []test{
					{name: `Ide QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Ide errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Ide QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Ide errors.New(Error_QemuDisk_File)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_2: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{File: ""}}}}}),
						err:   errors.New(Error_QemuDisk_File)},
					{name: `Ide errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_3: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{File: "/dev/disk/by-id/scsi1", Serial: "!@^$^&$^&"}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Ide errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_0: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{File: "/dev/disk/by-id/scsi1", Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Ide errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Ide: &QemuIdeDisks{Disk_1: &QemuIdeStorage{Passthrough: &QemuIdePassthrough{File: "/dev/disk/by-id/scsi1", WorldWideName: "5001A2B3C4D5E6F7"}}}}}),
						err:   errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `Sata QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_2: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Sata errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Sata QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_3: &QemuSataStorage{Passthrough: &QemuSataPassthrough{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Sata errors.New(Error_QemuDisk_File)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_4: &QemuSataStorage{Passthrough: &QemuSataPassthrough{File: ""}}}}}),
						err:   errors.New(Error_QemuDisk_File)},
					{name: `Sata errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_5: &QemuSataStorage{Passthrough: &QemuSataPassthrough{File: "/dev/disk/by-id/scsi1", Serial: "!@^$^&$^&"}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Sata errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_0: &QemuSataStorage{Passthrough: &QemuSataPassthrough{File: "/dev/disk/by-id/scsi1", Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Sata errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Sata: &QemuSataDisks{Disk_1: &QemuSataStorage{Passthrough: &QemuSataPassthrough{File: "/dev/disk/by-id/scsi1", WorldWideName: "0x500B0A09080706050"}}}}}),
						err:   errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `Scsi QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_0: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_5: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_6: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_7: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_8: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_1: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_2: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_3: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `Scsi errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_4: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `Scsi QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_9: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `Scsi errors.New(Error_QemuDisk_File)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_10: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{File: ""}}}}}),
						err:   errors.New(Error_QemuDisk_File)},
					{name: `Scsi errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_11: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{File: "/dev/disk/by-id/scsi1", Serial: "!@^$^&$^&"}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `Scsi errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_12: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{File: "/dev/disk/by-id/scsi1", Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `Scsi errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{Scsi: &QemuScsiDisks{Disk_13: &QemuScsiStorage{Passthrough: &QemuScsiPassthrough{File: "/dev/disk/by-id/scsi1", WorldWideName: "500F1E2D3C4B5A69!"}}}}}),
						err:   errors.New(Error_QemuWorldWideName_Invalid)},
					{name: `VirtIO QemuDiskAsyncIO("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_0: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{AsyncIO: "invalid"}}}}}),
						err:   QemuDiskAsyncIO("").Error()},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_5: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Burst: 9}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_6: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{ReadLimit: QemuDiskBandwidthIopsLimit{Concurrent: 8}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_7: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Burst: 7}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_8: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{Iops: QemuDiskBandwidthIops{WriteLimit: QemuDiskBandwidthIopsLimit{Concurrent: 6}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthIopsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_1: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 0`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_2: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{ReadLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitBurst) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_3: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Burst: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitBurst)},
					{name: `VirtIO errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent) 1`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_4: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Bandwidth: QemuDiskBandwidth{MBps: QemuDiskBandwidthMBps{WriteLimit: QemuDiskBandwidthMBpsLimit{Concurrent: 0.99}}}}}}}}),
						err:   errors.New(Error_QemuDiskBandwidthMBpsLimitConcurrent)},
					{name: `VirtIO QemuDiskCache("").Error()`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_9: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{Cache: "invalid"}}}}}),
						err:   QemuDiskCache("").Error()},
					{name: `VirtIO errors.New(Error_QemuDisk_File)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_10: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{File: ""}}}}}),
						err:   errors.New(Error_QemuDisk_File)},
					{name: `VirtIO errors.New(Error_QemuDiskSerial_IllegalCharacter)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_11: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{File: "/dev/disk/by-id/scsi1", Serial: "!@^$^&$^&"}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalCharacter)},
					{name: `VirtIO errors.New(Error_QemuDiskSerial_IllegalLength)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_12: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{File: "/dev/disk/by-id/scsi1", Serial: QemuDiskSerial(test_data_qemu.QemuDiskSerial_Max_Illegal())}}}}}),
						err:   errors.New(Error_QemuDiskSerial_IllegalLength)},
					{name: `VirtIO errors.New(Error_QemuWorldWideName_Invalid)`,
						input: baseConfig(ConfigQemu{Disks: &QemuStorages{VirtIO: &QemuVirtIODisks{Disk_13: &QemuVirtIOStorage{Passthrough: &QemuVirtIOPassthrough{File: "/dev/disk/by-id/scsi1", WorldWideName: "0x5004A3B2C1D0E0F1#"}}}}}),
						err:   errors.New(Error_QemuWorldWideName_Invalid)}}}},
		{category: `Memory`,
			valid: testType{
				create: []test{
					{name: `CapacityMiB only`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1024))}})},
					{name: `new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}})},
					{name: `new.Shares(qemuMemoryShares_Max) new.CapacityMiB & new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1001)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000)),
							Shares:             util.Pointer(QemuMemoryShares(qemuMemoryShares_Max))}})},
					{name: `new.Shares new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000)),
							Shares:             util.Pointer(QemuMemoryShares(0))}})},
					{name: `new.Shares(0) new.CapacityMiB & new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1001)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000)),
							Shares:             util.Pointer(QemuMemoryShares(0))}})}},
				update: []test{
					{name: `new.CapacityMiB smaller then current.MinimumCapacityMiB`,
						input:   ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1000))}},
						current: &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2000))}}},
					{name: `new.CapacityMiB smaller then current.MinimumCapacityMiB and MinimumCapacityMiB lowered`,
						input:   ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1000)), MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}},
						current: &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2000))}}},
					{name: `new.CapacityMiB == new.MinimumCapacityMiB && new.CapacityMiB > current.CapacityMiB`,
						input: ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1500)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1500))}},
						current: &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1000))}}},
					{name: `new.MinimumCapacityMiB > current.MinimumCapacityMiB && new.MinimumCapacityMiB < new.CapacityMiB`,
						input: ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(3000)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2000))}},
						current: &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1500))}}},
					{name: `new.Shares(0) current.CapacityMiB == current.MinimumCapacityMiB`,
						input: ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(0))}},
						current: &ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1000)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}}}}},
			invalid: testType{
				create: []test{
					{name: `new.MinimumCapacityMiB > new.CapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1000)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2000))}}),
						err: errors.New(QemuMemory_Error_MinimumCapacityMiB_GreaterThan_CapacityMiB)},
					{name: `new.Shares(1)`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(1))}}),
						err:   errors.New(QemuMemory_Error_NoMemoryCapacity)},
					{name: `new.Shares(1) when new.CapacityMiB == new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(1))}}),
						err:   errors.New(QemuMemory_Error_NoMemoryCapacity)}},
				createUpdate: []test{
					{name: `new.CapacityMiB(0)`,
						input:   baseConfig(ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(0))}}),
						current: &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(1000))}},
						err:     errors.New(QemuMemoryCapacity_Error_Minimum)},
					{name: `new.CapacityMiB > qemuMemoryCapacity_Max`,
						input:   baseConfig(ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(qemuMemoryCapacity_Max + 1))}}),
						current: &ConfigQemu{Memory: &QemuMemory{CapacityMiB: util.Pointer(QemuMemoryCapacity(qemuMemoryCapacity_Max))}},
						err:     errors.New(QemuMemoryCapacity_Error_Maximum)},
					{name: `new.MinimumCapacityMiB to big`,
						input:   baseConfig(ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(qemuMemoryBalloonCapacity_Max + 1))}}),
						current: &ConfigQemu{Memory: &QemuMemory{MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}},
						err:     errors.New(QemuMemoryBalloonCapacity_Error_Invalid)},
					{name: `new.Shares() too big and new.CapacityMiB & new.MinimumCapacityMiB`,
						input: baseConfig(ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1001)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000)),
							Shares:             util.Pointer(QemuMemoryShares(qemuMemoryShares_Max + 1))}}),
						current: &ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(512)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(256)),
							Shares:             util.Pointer(QemuMemoryShares(1))}},
						err: errors.New(QemuMemoryShares_Error_Invalid)}},
				update: []test{
					{name: `new.Shares(1) when current.CapacityMiB == current.MinimumCapacityMiB`,
						input: ConfigQemu{Memory: &QemuMemory{Shares: util.Pointer(QemuMemoryShares(1))}},
						current: &ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(1000)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1000))}},
						err: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)},
					{name: `new.Shares(1) new.CapacityMiB == current.MinimumCapacityMiB & MinimumCapacityMiB not updated`,
						input: ConfigQemu{Memory: &QemuMemory{
							CapacityMiB: util.Pointer(QemuMemoryCapacity(1024)),
							Shares:      util.Pointer(QemuMemoryShares(1))}},
						current: &ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(2048)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1024))}},
						err: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)},
					{name: `new.Shares(1) new.MinimumCapacityMiB == current.CapacityMiB & CapacityMiB not updated`,
						input: ConfigQemu{Memory: &QemuMemory{
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(2048)),
							Shares:             util.Pointer(QemuMemoryShares(1))}},
						current: &ConfigQemu{Memory: &QemuMemory{
							CapacityMiB:        util.Pointer(QemuMemoryCapacity(2048)),
							MinimumCapacityMiB: util.Pointer(QemuMemoryBalloonCapacity(1024))}},
						err: errors.New(QemuMemory_Error_SharesHasNoEffectWithoutBallooning)}}}},
		{category: `Network`,
			valid: testType{
				createUpdate: []test{
					{name: `Delete`,
						input:   baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{Delete: true}}}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{}}}},
					{name: `MTU inherit model=virtio`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID0,
							QemuNetworkInterface{MTU: &QemuMTU{Inherit: true}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{
							QemuNetworkInterfaceID0: QemuNetworkInterface{}}}},
					{name: `MTU value`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID1,
							QemuNetworkInterface{
								Model: util.Pointer(QemuNetworkModelVirtIO),
								MTU:   &QemuMTU{Value: 1500}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID1: QemuNetworkInterface{}}}},
					{name: `MTU empty e1000`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID2,
							QemuNetworkInterface{
								Model: util.Pointer(QemuNetworkModelE1000),
								MTU:   &QemuMTU{}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{}}}},
					{name: `MTU empty virtio`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID2,
							QemuNetworkInterface{
								Model: util.Pointer(QemuNetworkModelVirtIO),
								MTU:   &QemuMTU{}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID2: QemuNetworkInterface{}}}},
					{name: `Model`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID3,
							QemuNetworkInterface{Model: util.Pointer(QemuNetworkModel("e1000"))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID3: QemuNetworkInterface{}}}},
					{name: `MultiQueue`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID4,
							QemuNetworkInterface{MultiQueue: util.Pointer(QemuNetworkQueue(1))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID4: QemuNetworkInterface{}}}},
					{name: `RateLimitKBps`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID5,
							QemuNetworkInterface{RateLimitKBps: util.Pointer(QemuNetworkRate(10240000))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID5: QemuNetworkInterface{}}}},
					{name: `NativeVlan`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID6,
							QemuNetworkInterface{NativeVlan: util.Pointer(Vlan(56))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID6: QemuNetworkInterface{}}}},
					{name: `TaggedVlans`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID7,
							QemuNetworkInterface{TaggedVlans: util.Pointer(Vlans{0, 4095})})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID7: QemuNetworkInterface{}}}}},
				update: []test{
					{name: `MTU model change`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID0,
							QemuNetworkInterface{Model: util.Pointer(QemuNetworkModelE1000)})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID0: QemuNetworkInterface{
							Model: util.Pointer(QemuNetworkModelVirtIO),
							MTU:   &QemuMTU{Inherit: true}}}}},
					{name: `Update no Bridge && no Model`,
						input:   baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{}}}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID8: QemuNetworkInterface{}}}}}},
			invalid: testType{
				create: []test{
					{name: `no Bridge`,
						input: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID9: QemuNetworkInterface{}}}),
						err:   errors.New(QemuNetworkInterface_Error_BridgeRequired)},
					{name: `no Model`,
						input: baseConfig(ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID10: QemuNetworkInterface{Bridge: util.Pointer("vmbr0")}}}),
						err:   errors.New(QemuNetworkInterface_Error_ModelRequired)}},
				createUpdate: []test{
					{name: `errors.New(MTU_Error_Invalid)`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(
							QemuNetworkInterfaceID11, QemuNetworkInterface{MTU: &QemuMTU{Value: 575}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID11: QemuNetworkInterface{}}},
						err:     errors.New(MTU_Error_Invalid)},
					{name: `errors.New(QemuMTU_Error_Invalid)`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(
							QemuNetworkInterfaceID12, QemuNetworkInterface{MTU: &QemuMTU{Inherit: true, Value: 1500}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{}}},
						err:     errors.New(QemuMTU_Error_Invalid)},
					{name: `errors.New(QemuNetworkInterface_Error_MtuNoEffect) MTU inherit`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(
							QemuNetworkInterfaceID12, QemuNetworkInterface{
								Model: util.Pointer(QemuNetworkModelE100082544gc),
								MTU:   &QemuMTU{Inherit: true}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{}}},
						err:     errors.New(QemuNetworkInterface_Error_MtuNoEffect)},
					{name: `errors.New(QemuNetworkInterface_Error_MtuNoEffect) MTU value`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(
							QemuNetworkInterfaceID12, QemuNetworkInterface{
								Model: util.Pointer(QemuNetworkModelE1000),
								MTU:   &QemuMTU{Value: 1500}})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID12: QemuNetworkInterface{}}},
						err:     errors.New(QemuNetworkInterface_Error_MtuNoEffect)},
					{name: `model`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID13, QemuNetworkInterface{
							Model: util.Pointer(QemuNetworkModel("invalid"))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID13: QemuNetworkInterface{}}},
						err:     QemuNetworkModel("").Error()},
					{name: `errors.New(QemuNetworkQueue_Error_Invalid)`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID14, QemuNetworkInterface{
							MultiQueue: util.Pointer(QemuNetworkQueue(75))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID14: QemuNetworkInterface{}}},
						err:     errors.New(QemuNetworkQueue_Error_Invalid)},
					{name: `errors.New(QemuNetworkRate_Error_Invalid)`,
						input: baseConfig(
							ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID15, QemuNetworkInterface{
								RateLimitKBps: util.Pointer(QemuNetworkRate(10240001))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID15: QemuNetworkInterface{}}},
						err:     errors.New(QemuNetworkRate_Error_Invalid)},
					{name: `NativeVlan errors.New(Vlan_Error_Invalid)`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID16, QemuNetworkInterface{
							NativeVlan: util.Pointer(Vlan(4096))})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID16: QemuNetworkInterface{}}},
						err:     errors.New(Vlan_Error_Invalid)},
					{name: `TaggedVlans errors.New(Vlan_Error_Invalid)`,
						input: baseConfig(ConfigQemu{Networks: baseNetwork(QemuNetworkInterfaceID17, QemuNetworkInterface{
							TaggedVlans: util.Pointer(Vlans{4096})})}),
						current: &ConfigQemu{Networks: QemuNetworkInterfaces{QemuNetworkInterfaceID17: QemuNetworkInterface{}}},
						err:     errors.New(Vlan_Error_Invalid)}}}},
		{category: `PoolName`,
			valid: testType{
				createUpdate: []test{
					{name: `normal`,
						input:   baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(test_data_pool.PoolName_Legal()))}),
						current: &ConfigQemu{Pool: util.Pointer(PoolName("test"))}},
					{name: `empty`,
						input:   baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(""))}),
						current: &ConfigQemu{Pool: util.Pointer(PoolName("test"))}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `Length`,
						input:   baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(test_data_pool.PoolName_Max_Illegal()))}),
						current: &ConfigQemu{Pool: util.Pointer(PoolName("test"))},
						err:     errors.New(PoolName_Error_Length)},
					{name: `Characters`,
						input:   baseConfig(ConfigQemu{Pool: util.Pointer(PoolName(test_data_pool.PoolName_Error_Characters()[0]))}),
						current: &ConfigQemu{Pool: util.Pointer(PoolName("test"))},
						err:     errors.New(PoolName_Error_Characters)}}}},
		{category: `PciDevices`,
			valid: testType{
				createUpdate: []test{
					{name: `Delete`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID15: QemuPci{Delete: true}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}}}},
				update: []test{
					{name: `Mapping`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID14: QemuPci{
								Mapping: &QemuPciMapping{}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID14: QemuPci{
								Mapping: &QemuPciMapping{
									ID: util.Pointer(ResourceMappingPciID("aaa"))}}}}},
					{name: `Raw`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID13: QemuPci{
								Raw: &QemuPciRaw{}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID13: QemuPci{
								Raw: &QemuPciRaw{
									ID: util.Pointer(PciID("0000:00:00"))}}}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(QemuPciID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							20: QemuPci{}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID4: QemuPci{}}},
						err: errors.New(QemuPciID_Error_Invalid)},
					{name: `errors.New(QemuPci_Error_MutualExclusive)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID12: QemuPci{
								Mapping: &QemuPciMapping{
									ID: util.Pointer(ResourceMappingPciID("aaa"))},
								Raw: &QemuPciRaw{
									ID: util.Pointer(PciID("0000:00:00"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID12: QemuPci{}}},
						err:     errors.New(QemuPci_Error_MutualExclusive)},
					{name: `errors.New(QemuPci_Error_MappedID)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID11: QemuPci{
								Mapping: &QemuPciMapping{}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID11: QemuPci{}}},
						err:     errors.New(QemuPci_Error_MappedID)},
					{name: `errors.New(QemuPci_Error_RawID)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID10: QemuPci{
								Raw: &QemuPciRaw{}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID10: QemuPci{}}},
						err:     errors.New(QemuPci_Error_RawID)},
					{name: `errors.New(ResourceMappingPciID_Error_Invalid`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID9: QemuPci{
								Mapping: &QemuPciMapping{
									ID: util.Pointer(ResourceMappingPciID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID9: QemuPci{}}},
						err:     errors.New(ResourceMappingPciID_Error_Invalid)},
					{name: `errors.New(PciID_Error_MaximumFunction)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID8: QemuPci{
								Raw: &QemuPciRaw{ID: util.Pointer(PciID("0000:00:00.8"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID8: QemuPci{}}},
						err:     errors.New(PciID_Error_MaximumFunction)},
					{name: `Mapping errors.New(PciDeviceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID7: QemuPci{
								Mapping: &QemuPciMapping{
									ID:       util.Pointer(ResourceMappingPciID("aaa")),
									DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID7: QemuPci{}}},
						err:     errors.New(PciDeviceID_Error_Invalid)},
					{name: `Mapping errors.New(PciSubDeviceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID6: QemuPci{
								Mapping: &QemuPciMapping{
									ID:          util.Pointer(ResourceMappingPciID("aaa")),
									SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID6: QemuPci{}}},
						err:     errors.New(PciSubDeviceID_Error_Invalid)},
					{name: `Mapping errors.New(PciSubVendorID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID5: QemuPci{
								Mapping: &QemuPciMapping{
									ID:          util.Pointer(ResourceMappingPciID("aaa")),
									SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID5: QemuPci{}}},
						err:     errors.New(PciSubVendorID_Error_Invalid)},
					{name: `Mapping errors.New(PciVendorID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID4: QemuPci{
								Mapping: &QemuPciMapping{
									ID:       util.Pointer(ResourceMappingPciID("aaa")),
									VendorID: util.Pointer(PciVendorID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID4: QemuPci{}}},
						err:     errors.New(PciVendorID_Error_Invalid)},
					{name: `Raw errors.New(PciDeviceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID3: QemuPci{
								Raw: &QemuPciRaw{
									ID:       util.Pointer(PciID("0000:00:00")),
									DeviceID: util.Pointer(PciDeviceID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID3: QemuPci{}}},
						err:     errors.New(PciDeviceID_Error_Invalid)},
					{name: `Raw errors.New(PciSubDeviceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID2: QemuPci{
								Raw: &QemuPciRaw{
									ID:          util.Pointer(PciID("0000:00:00")),
									SubDeviceID: util.Pointer(PciSubDeviceID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID2: QemuPci{}}},
						err:     errors.New(PciSubDeviceID_Error_Invalid)},
					{name: `Raw errors.New(PciSubVendorID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID1: QemuPci{
								Raw: &QemuPciRaw{
									ID:          util.Pointer(PciID("0000:00:00")),
									SubVendorID: util.Pointer(PciSubVendorID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID1: QemuPci{}}},
						err:     errors.New(PciSubVendorID_Error_Invalid)},
					{name: `Raw errors.New(PciVendorID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{PciDevices: QemuPciDevices{
							QemuPciID0: QemuPci{
								Raw: &QemuPciRaw{
									ID:       util.Pointer(PciID("0000:00:00")),
									VendorID: util.Pointer(PciVendorID("a0%^#"))}}}}),
						current: &ConfigQemu{PciDevices: QemuPciDevices{QemuPciID0: QemuPci{}}},
						err:     errors.New(PciVendorID_Error_Invalid)}}}},
		{category: `Serials`,
			valid: testType{
				createUpdate: []test{
					{name: `all`,
						input: baseConfig(ConfigQemu{Serials: SerialInterfaces{
							SerialID0: SerialInterface{Path: "/dev/ttyS0"},
							SerialID1: SerialInterface{Path: "/dev/ttyS1", Delete: true},
							SerialID2: SerialInterface{Socket: true},
							SerialID3: SerialInterface{Socket: true, Delete: true}}}),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID0: SerialInterface{Path: "/dev/ttyS0"}}}},
					{name: `delete`,
						input: baseConfig(ConfigQemu{Serials: SerialInterfaces{
							SerialID3: SerialInterface{Delete: true}}}),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID0: SerialInterface{Path: "/dev/ttyS0"}}}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(SerialID_Errors_Invalid)`,
						input:   baseConfig(ConfigQemu{Serials: SerialInterfaces{4: SerialInterface{}}}),
						err:     errors.New(SerialID_Errors_Invalid),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID0: SerialInterface{Path: "/dev/ttyS0"}}}},
					{name: `errors.New(SerialInterface_Errors_MutualExclusive)`,
						input:   baseConfig(ConfigQemu{Serials: SerialInterfaces{SerialID0: SerialInterface{Path: "/dev/ttyS1", Socket: true}}}),
						err:     errors.New(SerialInterface_Errors_MutualExclusive),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID1: SerialInterface{Path: "/dev/ttyS0"}}}},
					{name: `errors.New(SerialInterface_Errors_Empty)`,
						input:   baseConfig(ConfigQemu{Serials: SerialInterfaces{SerialID1: SerialInterface{}}}),
						err:     errors.New(SerialInterface_Errors_Empty),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID2: SerialInterface{Path: "/dev/ttyS0"}}}},
					{name: `errors.New(SerialPath_Errors_Invalid)`,
						input:   baseConfig(ConfigQemu{Serials: SerialInterfaces{SerialID2: SerialInterface{Path: "invalid"}}}),
						err:     errors.New(SerialPath_Errors_Invalid),
						current: &ConfigQemu{Serials: SerialInterfaces{SerialID3: SerialInterface{Path: "/dev/ttyS0"}}}}}}},
		{category: `Tags`,
			valid: testType{
				create: []test{
					{name: `normal`,
						input:   baseConfig(ConfigQemu{Tags: util.Pointer(validTags())}),
						current: &ConfigQemu{Tags: util.Pointer([]Tag{"a", "b"})}}}},
			invalid: testType{
				createUpdate: []test{
					{name: `errors.New(Tag_Error_Invalid)`,
						input:   baseConfig(ConfigQemu{Tags: util.Pointer([]Tag{Tag(test_data_tag.Tag_Illegal())})}),
						current: &ConfigQemu{Tags: util.Pointer([]Tag{"a", "b"})},
						err:     errors.New(Tag_Error_Invalid)},
					{name: `errors.New(Tag_Error_Duplicate)`,
						input:   baseConfig(ConfigQemu{Tags: util.Pointer([]Tag{Tag(test_data_tag.Tag_Max_Legal()), Tag(test_data_tag.Tag_Max_Legal())})}),
						current: &ConfigQemu{Tags: util.Pointer([]Tag{"a", "b"})},
						err:     errors.New(Tag_Error_Duplicate)},
					{name: `errors.New(Tag_Error_Empty)`,
						input:   baseConfig(ConfigQemu{Tags: util.Pointer([]Tag{Tag(test_data_tag.Tag_Empty())})}),
						current: &ConfigQemu{Tags: util.Pointer([]Tag{"a", "b"})},
						err:     errors.New(Tag_Error_Empty)},
					{name: `errors.New(Tag_Error_MaxLength)`,
						input:   baseConfig(ConfigQemu{Tags: util.Pointer([]Tag{Tag(test_data_tag.Tag_Max_Illegal())})}),
						current: &ConfigQemu{Tags: util.Pointer([]Tag{"a", "b"})},
						err:     errors.New(Tag_Error_MaxLength)}}}},
		{category: `TPM`,
			valid: testType{
				createUpdate: []test{
					{name: `normal`,
						input:   baseConfig(ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion("v2.0"))}}),
						current: &ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion("v1.2"))}}}},
				update: []test{
					{name: `Version=nil`,
						input:   ConfigQemu{TPM: &TpmState{Storage: "test"}},
						current: &ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion("v1.2"))}}}}},
			invalid: testType{
				create: []test{
					{name: `errors.New(TmpState_Error_VersionRequired) Create`,
						input: baseConfig(ConfigQemu{TPM: &TpmState{Storage: "test", Version: nil}}),
						err:   errors.New(TmpState_Error_VersionRequired)}},
				createUpdate: []test{
					{name: `errors.New(storage is required)`,
						input:   baseConfig(ConfigQemu{TPM: &TpmState{Storage: ""}}),
						current: &ConfigQemu{TPM: &TpmState{}},
						err:     errors.New("storage is required")},
					{name: `errors.New(TmpVersion_Error_Invalid) Update`,
						input:   baseConfig(ConfigQemu{TPM: &TpmState{Storage: "test", Version: util.Pointer(TpmVersion(""))}}),
						current: &ConfigQemu{TPM: &TpmState{}},
						err:     errors.New(TpmVersion_Error_Invalid)}}}},
		{category: `USBs`,
			valid: testType{
				createUpdate: []test{
					{name: `delete`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Delete: true}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID: util.Pointer(UsbDeviceID("1234:5678"))}}}}},
					{name: `USBs.Device.ID set/update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID: util.Pointer(UsbDeviceID("5678:1234"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID:   util.Pointer(UsbDeviceID("1234:5678")),
								USB3: util.Pointer(true)}}}}},
					{name: `USBs.Mapped.ID set/update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID: util.Pointer(ResourceMappingUsbID("valid"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID:   util.Pointer(ResourceMappingUsbID("test")),
								USB3: util.Pointer(true)}}}}},
					{name: `USBs.Port.ID set/update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID: util.Pointer(UsbPortID("6-4"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID:   util.Pointer(UsbPortID("1-5")),
								USB3: util.Pointer(true)}}}}},
					{name: `USBs.Spice.USB3 set/update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{
								USB3: true}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{
								USB3: false}}}}}},
				update: []test{
					{name: `USBs.Device to USBs.Mapped`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID: util.Pointer(UsbDeviceID("1234:5678"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID: util.Pointer(ResourceMappingUsbID("test"))}}}}},
					{name: `USBs.Device.USB3 update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								USB3: util.Pointer(true)}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID:   util.Pointer(UsbDeviceID("1234:5678")),
								USB3: util.Pointer(false)}}}}},
					{name: `USBs.Mapped to USBs.Port`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID: util.Pointer(ResourceMappingUsbID("test"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID: util.Pointer(UsbPortID("3-5"))}}}}},
					{name: `USBs.Mapped.USB3 update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								USB3: util.Pointer(true)}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID:   util.Pointer(ResourceMappingUsbID("test")),
								USB3: util.Pointer(false)}}}}},
					{name: `USBs.Port to USBs.Spice`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID: util.Pointer(UsbPortID("2-6"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{}}}}},
					{name: `USBs.Port.USB3 update`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								USB3: util.Pointer(true)}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID:   util.Pointer(UsbPortID("2-6")),
								USB3: util.Pointer(false)}}}}},
					{name: `USBs.Spice to USBs.Device`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Spice: &QemuUsbSpice{}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID: util.Pointer(UsbDeviceID("5678:1234"))}}}}}}},
			invalid: testType{
				create: []test{},
				createUpdate: []test{
					{name: `errors.New(QemuUsbID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							20: QemuUSB{}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
						err: errors.New(QemuUsbID_Error_Invalid)},
					{name: `errors.New(QemuUSB_Error_MutualExclusive)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{
								Device:  &QemuUsbDevice{ID: util.Pointer(UsbDeviceID("1234:5678"))},
								Mapping: &QemuUsbMapping{ID: util.Pointer(ResourceMappingUsbID("test"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{
								Device: &QemuUsbDevice{ID: util.Pointer(UsbDeviceID("1234:5678"))}}}},
						err: errors.New(QemuUSB_Error_MutualExclusive)},
					{name: `errors.New(QemuUSB_Error_DeviceID)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
						err: errors.New(QemuUSB_Error_DeviceID)},
					{name: `errors.New(QemuUSB_Error_MappedID)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}},
						err: errors.New(QemuUSB_Error_MappingID)},
					{name: `errors.New(QemuUSB_Error_PortID)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
						err: errors.New(QemuUSB_Error_PortID)},
					{name: `errors.New(UsbDeviceID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{
								ID: util.Pointer(UsbDeviceID("1234"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Device: &QemuUsbDevice{}}}},
						err: errors.New(UsbDeviceID_Error_Invalid)},
					{name: `errors.New(ResourceMappingUsbID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{
								ID: util.Pointer(ResourceMappingUsbID("Invalid%"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Mapping: &QemuUsbMapping{}}}},
						err: errors.New(ResourceMappingUsbID_Error_Invalid)},
					{name: `errors.New(UsbPortID_Error_Invalid)`,
						input: baseConfig(ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{
								ID: util.Pointer(UsbPortID("2-3-4"))}}}}),
						current: &ConfigQemu{USBs: QemuUSBs{
							QemuUsbID0: QemuUSB{Port: &QemuUsbPort{}}}},
						err: errors.New(UsbPortID_Error_Invalid)}}}},
	}
	for _, test := range tests {
		for _, subTest := range append(test.valid.create, test.valid.createUpdate...) {
			name := test.category + "/Valid/Create"
			if len(test.valid.create)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil, subTest.version), name)
			})
		}
		for _, subTest := range append(test.valid.update, test.valid.createUpdate...) {
			name := test.category + "/Valid/Update"
			if len(test.valid.update)+len(test.valid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current, subTest.version), name)
			})
		}
		for _, subTest := range append(test.invalid.create, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Create"
			if len(test.invalid.create)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.Equal(t, subTest.err, subTest.input.Validate(nil, subTest.version), name)
			})
		}
		for _, subTest := range append(test.invalid.update, test.invalid.createUpdate...) {
			name := test.category + "/Invalid/Update"
			if len(test.invalid.update)+len(test.invalid.createUpdate) > 1 {
				name += "/" + subTest.name
			}
			t.Run(name, func(*testing.T) {
				require.NotNil(t, subTest.current)
				require.Equal(t, subTest.err, subTest.input.Validate(subTest.current, subTest.version), name)
			})
		}
	}
}
