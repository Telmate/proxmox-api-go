package proxmox

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/parse"
	"github.com/Telmate/proxmox-api-go/internal/util"
)

type CpuFlags struct {
	AES        *TriBool `json:"aes,omitempty"`        // Activate AES instruction set for HW acceleration.
	AmdNoSSB   *TriBool `json:"amdnossb,omitempty"`   // Notifies guest OS that host is not vulnerable for Spectre on AMD CPUs.
	AmdSSBD    *TriBool `json:"amdssbd,omitempty"`    // Improves Spectre mitigation performance with AMD CPUs, best used with "VirtSSBD".
	HvEvmcs    *TriBool `json:"hvevmcs,omitempty"`    // Improve performance for nested virtualization. Only supported on Intel CPUs.
	HvTlbFlush *TriBool `json:"hvtlbflush,omitempty"` // Improve performance in overcommitted Windows guests. May lead to guest bluescreens on old CPUs.
	Ibpb       *TriBool `json:"ibpb,omitempty"`       // Allows improved Spectre mitigation with AMD CPUs.
	MdClear    *TriBool `json:"mdclear,omitempty"`    // Required to let the guest OS know if MDS is mitigated correctly.
	PCID       *TriBool `json:"pcid,omitempty"`       // Meltdown fix cost reduction on Westmere, Sandy-, and IvyBridge Intel CPUs.
	Pdpe1GB    *TriBool `json:"pdpe1gb,omitempty"`    // Allow guest OS to use 1GB size pages, if host HW supports it.
	SSBD       *TriBool `json:"ssbd,omitempty"`       // Protection for "Speculative Store Bypass" for Intel models.
	SpecCtrl   *TriBool `json:"specctrl,omitempty"`   // Allows improved Spectre mitigation with Intel CPUs.
	VirtSSBD   *TriBool `json:"cirtssbd,omitempty"`   // Basis for "Speculative Store Bypass" protection for AMD models.
}

func (flags CpuFlags) mapToApi(current *CpuFlags) (string, bool) {
	var builder strings.Builder
	var isSet bool

	flagNames := []string{
		"aes",
		"amd-no-ssb",
		"amd-ssbd",
		"hv-evmcs",
		"hv-tlbflush",
		"ibpb",
		"md-clear",
		"pcid",
		"pdpe1gb",
		"ssbd",
		"spec-ctrl",
		"virt-ssbd"}

	flagValues := []*TriBool{
		flags.AES,
		flags.AmdNoSSB,
		flags.AmdSSBD,
		flags.HvEvmcs,
		flags.HvTlbFlush,
		flags.Ibpb,
		flags.MdClear,
		flags.PCID,
		flags.Pdpe1GB,
		flags.SSBD,
		flags.SpecCtrl,
		flags.VirtSSBD}

	var currentValues []*TriBool
	if current != nil {
		currentValues = []*TriBool{
			current.AES,
			current.AmdNoSSB,
			current.AmdSSBD,
			current.HvEvmcs,
			current.HvTlbFlush,
			current.Ibpb,
			current.MdClear,
			current.PCID,
			current.Pdpe1GB,
			current.SSBD,
			current.SpecCtrl,
			current.VirtSSBD,
		}
	} else {
		currentValues = make([]*TriBool, len(flagValues))
	}

	for i, value := range flagValues {
		if value != nil {
			switch *value {
			case TriBoolTrue:
				builder.WriteString(";+" + flagNames[i])
			case TriBoolFalse:
				builder.WriteString(";-" + flagNames[i])
			}
			isSet = true
		} else if currentValues[i] != nil {
			switch *currentValues[i] {
			case TriBoolTrue:
				builder.WriteString(";+" + flagNames[i])
			case TriBoolFalse:
				builder.WriteString(";-" + flagNames[i])
			}
			isSet = true
		}
	}
	return builder.String(), isSet
}

func (CpuFlags) mapToSDK(flags []string) *CpuFlags {
	flagMap := map[string]rune{}
	for _, e := range flags {
		flagMap[e[1:]] = rune(e[0])
	}
	return &CpuFlags{
		AES:        CpuFlags{}.mapToSdkSubroutine(flagMap, "aes"),
		AmdNoSSB:   CpuFlags{}.mapToSdkSubroutine(flagMap, "amd-no-ssb"),
		AmdSSBD:    CpuFlags{}.mapToSdkSubroutine(flagMap, "amd-ssbd"),
		HvEvmcs:    CpuFlags{}.mapToSdkSubroutine(flagMap, "hv-evmcs"),
		HvTlbFlush: CpuFlags{}.mapToSdkSubroutine(flagMap, "hv-tlbflush"),
		Ibpb:       CpuFlags{}.mapToSdkSubroutine(flagMap, "ibpb"),
		MdClear:    CpuFlags{}.mapToSdkSubroutine(flagMap, "md-clear"),
		PCID:       CpuFlags{}.mapToSdkSubroutine(flagMap, "pcid"),
		Pdpe1GB:    CpuFlags{}.mapToSdkSubroutine(flagMap, "pdpe1gb"),
		SSBD:       CpuFlags{}.mapToSdkSubroutine(flagMap, "ssbd"),
		SpecCtrl:   CpuFlags{}.mapToSdkSubroutine(flagMap, "spec-ctrl"),
		VirtSSBD:   CpuFlags{}.mapToSdkSubroutine(flagMap, "virt-ssbd"),
	}
}

func (CpuFlags) mapToSdkSubroutine(flags map[string]rune, flag string) *TriBool {
	var tmp TriBool
	if v, isSet := flags[flag]; isSet {
		switch v {
		case '+':
			tmp = TriBoolTrue
		case '-':
			tmp = TriBoolFalse
		}
		return &tmp
	}
	return nil
}

func (flags CpuFlags) Validate() (err error) {
	if flags.AES != nil {
		if err = flags.AES.Validate(); err != nil {
			return err
		}
	}
	if flags.AmdNoSSB != nil {
		if err = flags.AmdNoSSB.Validate(); err != nil {
			return err
		}
	}
	if flags.AmdSSBD != nil {
		if err = flags.AmdSSBD.Validate(); err != nil {
			return err
		}
	}
	if flags.HvEvmcs != nil {
		if err = flags.HvEvmcs.Validate(); err != nil {
			return err
		}
	}
	if flags.HvTlbFlush != nil {
		if err = flags.HvTlbFlush.Validate(); err != nil {
			return err
		}
	}
	if flags.Ibpb != nil {
		if err = flags.Ibpb.Validate(); err != nil {
			return err
		}
	}
	if flags.MdClear != nil {
		if err = flags.MdClear.Validate(); err != nil {
			return err
		}
	}
	if flags.PCID != nil {
		if err = flags.PCID.Validate(); err != nil {
			return err
		}
	}
	if flags.Pdpe1GB != nil {
		if err = flags.Pdpe1GB.Validate(); err != nil {
			return err
		}
	}
	if flags.SSBD != nil {
		if err = flags.SSBD.Validate(); err != nil {
			return err
		}
	}
	if flags.SpecCtrl != nil {
		if err = flags.SpecCtrl.Validate(); err != nil {
			return err
		}
	}
	if flags.VirtSSBD != nil {
		if err = flags.VirtSSBD.Validate(); err != nil {
			return err
		}
	}
	return
}

type CpuLimit uint8 // min value 0 is unlimited, max value of 128

const CpuLimit_Error_Maximum string = "maximum value of CpuLimit is 128"

func (limit CpuLimit) Validate() error {
	if limit > 128 {
		return errors.New(CpuLimit_Error_Maximum)
	}
	return nil
}

type CpuType string // enum

const (
	CpuType_Intel486                          CpuType = "486"
	CpuType_AmdAthlon                         CpuType = "athlon"
	CpuType_IntelBroadwell                    CpuType = "Broadwell"
	cpuType_IntelBroadwell_Lower              CpuType = "broadwell"
	CpuType_IntelBroadwellIBRS                CpuType = "Broadwell-IBRS"
	cpuType_IntelBroadwellIBRS_Lower          CpuType = "broadwellibrs"
	CpuType_IntelBroadwellNoTSX               CpuType = "Broadwell-noTSX"
	cpuType_IntelBroadwellNoTSX_Lower         CpuType = "broadwellnotsx"
	CpuType_IntelBroadwellNoTSXIBRS           CpuType = "Broadwell-noTSX-IBRS"
	cpuType_IntelBroadwellNoTSXIBRS_Lower     CpuType = "broadwellnotsxibrs"
	CpuType_IntelCascadelakeServer            CpuType = "Cascadelake-Server"
	cpuType_IntelCascadelakeServer_Lower      CpuType = "cascadelakeserver"
	CpuType_IntelCascadelakeServerNoTSX       CpuType = "Cascadelake-Server-noTSX"
	cpuType_IntelCascadelakeServerNoTSX_Lower CpuType = "cascadelakeservernotsx"
	CpuType_IntelCascadelakeServerV2          CpuType = "Cascadelake-Server-V2"
	cpuType_IntelCascadelakeServerV2_Lower    CpuType = "cascadelakeserverv2"
	CpuType_IntelCascadelakeServerV4          CpuType = "Cascadelake-Server-V4"
	cpuType_IntelCascadelakeServerV4_Lower    CpuType = "cascadelakeserverv4"
	CpuType_IntelCascadelakeServerV5          CpuType = "Cascadelake-Server-V5"
	cpuType_IntelCascadelakeServerV5_Lower    CpuType = "cascadelakeserverv5"
	CpuType_IntelConroe                       CpuType = "Conroe"
	cpuType_IntelConroe_Lower                 CpuType = "conroe"
	CpuType_IntelCooperlake                   CpuType = "Cooperlake"
	cpuType_IntelCooperlake_Lower             CpuType = "cooperlake"
	CpuType_IntelCooperlakeV2                 CpuType = "Cooperlake-V2"
	cpuType_IntelCooperlakeV2_Lower           CpuType = "cooperlakev2"
	CpuType_IntelCore2Duo                     CpuType = "core2duo"
	CpuType_IntelCoreDuo                      CpuType = "coreduo"
	CpuType_AmdEPYC                           CpuType = "EPYC"
	cpuType_AmdEPYC_Lower                     CpuType = "epyc"
	CpuType_AmdEPYCIBPB                       CpuType = "EPYC-IBPB"
	cpuType_AmdEPYCIBPB_Lower                 CpuType = "epycibpb"
	CpuType_AmdEPYCMilan                      CpuType = "EPYC-Milan"
	cpuType_AmdEPYCMilan_Lower                CpuType = "epycmilan"
	CpuType_AmdEPYCRome                       CpuType = "EPYC-Rome"
	cpuType_AmdEPYCRome_Lower                 CpuType = "epycrome"
	CpuType_AmdEPYCRomeV2                     CpuType = "EPYC-Rome-v2"
	cpuType_AmdEPYCRomeV2_Lower               CpuType = "epycromev2"
	CpuType_AmdEPYCV3                         CpuType = "EPYC-v3"
	cpuType_AmdEPYCV3_Lower                   CpuType = "epycv3"
	CpuType_Host                              CpuType = "host"
	CpuType_IntelHaswell                      CpuType = "Haswell"
	cpuType_IntelHaswell_Lower                CpuType = "haswell"
	CpuType_IntelHaswellIBRS                  CpuType = "Haswell-IBRS"
	cpuType_IntelHaswellIBRS_Lower            CpuType = "haswellibrs"
	CpuType_IntelHaswellNoTSX                 CpuType = "Haswell-noTSX"
	cpuType_IntelHaswellNoTSX_Lower           CpuType = "haswellnotsx"
	CpuType_IntelHaswellNoTSXIBRS             CpuType = "Haswell-noTSX-IBRS"
	cpuType_IntelHaswellNoTSXIBRS_Lower       CpuType = "haswellnotsxibrs"
	CpuType_IntelIcelakeClient                CpuType = "Icelake-Client"
	cpuType_IntelIcelakeClient_Lower          CpuType = "icelakeclient"
	CpuType_IntelIcelakeClientNoTSX           CpuType = "Icelake-Client-noTSX"
	cpuType_IntelIcelakeClientNoTSX_Lower     CpuType = "icelakeclientnotsx"
	CpuType_IntelIcelakeServer                CpuType = "Icelake-Server"
	cpuType_IntelIcelakeServer_Lower          CpuType = "icelakeserver"
	CpuType_IntelIcelakeServerNoTSX           CpuType = "Icelake-Server-noTSX"
	cpuType_IntelIcelakeServerNoTSX_Lower     CpuType = "icelakeservernotsx"
	CpuType_IntelIcelakeServerV3              CpuType = "Icelake-Server-v3"
	cpuType_IntelIcelakeServerV3_Lower        CpuType = "icelakeserverv3"
	CpuType_IntelIcelakeServerV4              CpuType = "Icelake-Server-v4"
	cpuType_IntelIcelakeServerV4_Lower        CpuType = "icelakeserverv4"
	CpuType_IntelIcelakeServerV5              CpuType = "Icelake-Server-v5"
	cpuType_IntelIcelakeServerV5_Lower        CpuType = "icelakeserverv5"
	CpuType_IntelIcelakeServerV6              CpuType = "Icelake-Server-v6"
	cpuType_IntelIcelakeServerV6_Lower        CpuType = "icelakeserverv6"
	CpuType_IntelIvybridge                    CpuType = "IvyBridge"
	cpuType_IntelIvybridge_Lower              CpuType = "ivybridge"
	CpuType_IntelIvybridgeIBRS                CpuType = "IvyBridge-IBRS"
	cpuType_IntelIvybridgeIBRS_Lower          CpuType = "ivyBridgeibrs"
	CpuType_IntelKnightsmill                  CpuType = "KnightsMill"
	cpuType_IntelKnightsmill_Lower            CpuType = "knightsmill"
	CpuType_QemuKvm32                         CpuType = "kvm32"
	CpuType_QemuKvm64                         CpuType = "kvm64"
	CpuType_QemuMax                           CpuType = "max"
	CpuType_IntelNehalem                      CpuType = "Nehalem"
	cpuType_IntelNehalem_Lower                CpuType = "nehalem"
	CpuType_IntelNehalemIBRS                  CpuType = "Nehalem-IRBS"
	cpuType_IntelNehalemIBRS_Lower            CpuType = "nehalemibrs"
	CpuType_AmdOpteronG1                      CpuType = "Opteron_G1"
	cpuType_AmdOpteronG1_Lower                CpuType = "opterong1"
	CpuType_AmdOpteronG2                      CpuType = "Opteron_G2"
	cpuType_AmdOpteronG2_Lower                CpuType = "opterong2"
	CpuType_AmdOpteronG3                      CpuType = "Opteron_G3"
	cpuType_AmdOpteronG3_Lower                CpuType = "opterong3"
	CpuType_AmdOpteronG4                      CpuType = "Opteron_G4"
	cpuType_AmdOpteronG4_Lower                CpuType = "opterong4"
	CpuType_AmdOpteronG5                      CpuType = "Opteron_G5"
	cpuType_AmdOpteronG5_Lower                CpuType = "opterong5"
	CpuType_IntelPenrym                       CpuType = "Penrym"
	cpuType_IntelPenrym_Lower                 CpuType = "penrym"
	CpuType_IntelPentium                      CpuType = "pentium"
	CpuType_IntelPentium2                     CpuType = "pentium2"
	CpuType_IntelPentium3                     CpuType = "pentium3"
	CpuType_AmdPhenom                         CpuType = "phenom"
	CpuType_Qemu32                            CpuType = "qemu32"
	CpuType_Qemu64                            CpuType = "qemu64"
	CpuType_IntelSandyBridge                  CpuType = "SandyBridge"
	cpuType_IntelSandyBridge_Lower            CpuType = "sandybridge"
	CpuType_IntelSandybridgeIBRS              CpuType = "SandyBridge-IBRS"
	cpuType_IntelSandybridgeIBRS_Lower        CpuType = "sandybridgeibrs"
	CpuType_IntelSapphireRapids               CpuType = "SapphireRapids"
	cpuType_IntelSapphireRapids_Lower         CpuType = "sapphirerapids"
	CpuType_IntelSkylakeClient                CpuType = "Skylake-Client"
	cpuType_IntelSkylakeClient_Lower          CpuType = "skylakeclient"
	CpuType_IntelSkylakeClientIBRS            CpuType = "Skylake-Client-IBRS"
	cpuType_IntelSkylakeClientIBRS_Lower      CpuType = "skylakeclientibrs"
	CpuType_IntelSkylakeClientNoTSXIBRS       CpuType = "Skylake-Client-noTSX-IBRS"
	cpuType_IntelSkylakeClientNoTSXIBRS_Lower CpuType = "skylakeclientnotsxibrs"
	CpuType_IntelSkylakeClientV4              CpuType = "Skylake-Client-v4"
	cpuType_IntelSkylakeClientV4_Lower        CpuType = "skylakeclientv4"
	CpuType_IntelSkylakeServer                CpuType = "Skylake-Server"
	cpuType_IntelSkylakeServer_Lower          CpuType = "skylakeserver"
	CpuType_IntelSkylakeServerIBRS            CpuType = "Skylake-Server-IBRS"
	cpuType_IntelSkylakeServerIBRS_Lower      CpuType = "skylakeserveribrs"
	CpuType_IntelSkylakeServerNoTSXIBRS       CpuType = "Skylake-Server-noTSX-IBRS"
	cpuType_IntelSkylakeServerNoTSXIBRS_Lower CpuType = "skylakeservernotsxibrs"
	CpuType_IntelSkylakeServerV4              CpuType = "Skylake-Server-v4"
	cpuType_IntelSkylakeServerV4_Lower        CpuType = "skylakeserverv4"
	CpuType_IntelSkylakeServerV5              CpuType = "Skylake-Server-v5"
	cpuType_IntelSkylakeServerV5_Lower        CpuType = "skylakeserverv5"
	CpuType_IntelWestmere                     CpuType = "Westmere"
	cpuType_IntelWestmere_Lower               CpuType = "westmere"
	CpuType_IntelWestmereIBRS                 CpuType = "Westmere-IBRS"
	cpuType_IntelWestmereIBRS_Lower           CpuType = "westmereibrs"
	CpuType_X86_64_v2                         CpuType = "x86-64-v2"
	cpuType_X86_64_v2_Lower                   CpuType = "x8664v2"
	CpuType_X86_64_v2_AES                     CpuType = "x86-64-v2-AES"
	cpuType_X86_64_v2_AES_Lower               CpuType = "x8664v2aes"
	CpuType_X86_64_v3                         CpuType = "x86-64-v3"
	cpuType_X86_64_v3_Lower                   CpuType = "x8664v3"
	CpuType_X86_64_v4                         CpuType = "x86-64-v4"
	cpuType_X86_64_v4_Lower                   CpuType = "x8664v4"
)

func (CpuType) cpuBase() map[CpuType]CpuType {
	return map[CpuType]CpuType{
		CpuType_AmdAthlon:                         CpuType_AmdAthlon,
		CpuType_AmdPhenom:                         CpuType_AmdPhenom,
		CpuType_Intel486:                          CpuType_Intel486,
		CpuType_IntelCore2Duo:                     CpuType_IntelCore2Duo,
		CpuType_IntelCoreDuo:                      CpuType_IntelCoreDuo,
		CpuType_IntelPentium:                      CpuType_IntelPentium,
		CpuType_IntelPentium2:                     CpuType_IntelPentium2,
		CpuType_IntelPentium3:                     CpuType_IntelPentium3,
		CpuType_QemuKvm32:                         CpuType_QemuKvm32,
		CpuType_QemuKvm64:                         CpuType_QemuKvm64,
		CpuType_QemuMax:                           CpuType_QemuMax,
		CpuType_Qemu32:                            CpuType_Qemu32,
		CpuType_Qemu64:                            CpuType_Qemu64,
		CpuType_Host:                              CpuType_Host,
		cpuType_AmdEPYC_Lower:                     CpuType_AmdEPYC,
		cpuType_AmdEPYCIBPB_Lower:                 CpuType_AmdEPYCIBPB,
		cpuType_AmdEPYCMilan_Lower:                CpuType_AmdEPYCMilan,
		cpuType_AmdEPYCRome_Lower:                 CpuType_AmdEPYCRome,
		cpuType_AmdOpteronG1_Lower:                CpuType_AmdOpteronG1,
		cpuType_AmdOpteronG2_Lower:                CpuType_AmdOpteronG2,
		cpuType_AmdOpteronG3_Lower:                CpuType_AmdOpteronG3,
		cpuType_AmdOpteronG4_Lower:                CpuType_AmdOpteronG4,
		cpuType_AmdOpteronG5_Lower:                CpuType_AmdOpteronG5,
		cpuType_IntelBroadwell_Lower:              CpuType_IntelBroadwell,
		cpuType_IntelBroadwellIBRS_Lower:          CpuType_IntelBroadwellIBRS,
		cpuType_IntelBroadwellNoTSX_Lower:         CpuType_IntelBroadwellNoTSX,
		cpuType_IntelBroadwellNoTSXIBRS_Lower:     CpuType_IntelBroadwellNoTSXIBRS,
		cpuType_IntelCascadelakeServer_Lower:      CpuType_IntelCascadelakeServer,
		cpuType_IntelCascadelakeServerNoTSX_Lower: CpuType_IntelCascadelakeServerNoTSX,
		cpuType_IntelConroe_Lower:                 CpuType_IntelConroe,
		cpuType_IntelHaswell_Lower:                CpuType_IntelHaswell,
		cpuType_IntelHaswellIBRS_Lower:            CpuType_IntelHaswellIBRS,
		cpuType_IntelHaswellNoTSX_Lower:           CpuType_IntelHaswellNoTSX,
		cpuType_IntelHaswellNoTSXIBRS_Lower:       CpuType_IntelHaswellNoTSXIBRS,
		cpuType_IntelIcelakeClient_Lower:          CpuType_IntelIcelakeClient,
		cpuType_IntelIcelakeClientNoTSX_Lower:     CpuType_IntelIcelakeClientNoTSX,
		cpuType_IntelIcelakeServer_Lower:          CpuType_IntelIcelakeServer,
		cpuType_IntelIcelakeServerNoTSX_Lower:     CpuType_IntelIcelakeServerNoTSX,
		cpuType_IntelIvybridge_Lower:              CpuType_IntelIvybridge,
		cpuType_IntelIvybridgeIBRS_Lower:          CpuType_IntelIvybridgeIBRS,
		cpuType_IntelKnightsmill_Lower:            CpuType_IntelKnightsmill,
		cpuType_IntelNehalem_Lower:                CpuType_IntelNehalem,
		cpuType_IntelNehalemIBRS_Lower:            CpuType_IntelNehalemIBRS,
		cpuType_IntelPenrym_Lower:                 CpuType_IntelPenrym,
		cpuType_IntelSandyBridge_Lower:            CpuType_IntelSandyBridge,
		cpuType_IntelSandybridgeIBRS_Lower:        CpuType_IntelSandybridgeIBRS,
		cpuType_IntelSkylakeClient_Lower:          CpuType_IntelSkylakeClient,
		cpuType_IntelSkylakeClientIBRS_Lower:      CpuType_IntelSkylakeClientIBRS,
		cpuType_IntelSkylakeClientNoTSXIBRS_Lower: CpuType_IntelSkylakeClientNoTSXIBRS,
		cpuType_IntelSkylakeServer_Lower:          CpuType_IntelSkylakeServer,
		cpuType_IntelSkylakeServerIBRS_Lower:      CpuType_IntelSkylakeServerIBRS,
		cpuType_IntelSkylakeServerNoTSXIBRS_Lower: CpuType_IntelSkylakeServerNoTSXIBRS,
		cpuType_IntelWestmere_Lower:               CpuType_IntelWestmere,
		cpuType_IntelWestmereIBRS_Lower:           CpuType_IntelWestmereIBRS,
	}
}

func (CpuType) cpuV8(cpus map[CpuType]CpuType) {
	cpus[cpuType_IntelCascadelakeServerV2_Lower] = CpuType_IntelCascadelakeServerV2
	cpus[cpuType_IntelCascadelakeServerV4_Lower] = CpuType_IntelCascadelakeServerV4
	cpus[cpuType_IntelCascadelakeServerV5_Lower] = CpuType_IntelCascadelakeServerV5
	cpus[cpuType_IntelCooperlake_Lower] = CpuType_IntelCooperlake
	cpus[cpuType_IntelCooperlakeV2_Lower] = CpuType_IntelCooperlakeV2
	cpus[cpuType_AmdEPYCRomeV2_Lower] = CpuType_AmdEPYCRomeV2
	cpus[cpuType_AmdEPYCV3_Lower] = CpuType_AmdEPYCV3
	cpus[cpuType_IntelIcelakeServerV3_Lower] = CpuType_IntelIcelakeServerV3
	cpus[cpuType_IntelIcelakeServerV4_Lower] = CpuType_IntelIcelakeServerV4
	cpus[cpuType_IntelIcelakeServerV5_Lower] = CpuType_IntelIcelakeServerV5
	cpus[cpuType_IntelIcelakeServerV6_Lower] = CpuType_IntelIcelakeServerV6
	cpus[cpuType_IntelSapphireRapids_Lower] = CpuType_IntelSapphireRapids
	cpus[cpuType_IntelSkylakeClientV4_Lower] = CpuType_IntelSkylakeClientV4
	cpus[cpuType_IntelSkylakeServerV4_Lower] = CpuType_IntelSkylakeServerV4
	cpus[cpuType_IntelSkylakeServerV5_Lower] = CpuType_IntelSkylakeServerV5
	cpus[cpuType_X86_64_v2_Lower] = CpuType_X86_64_v2
	cpus[cpuType_X86_64_v2_AES_Lower] = CpuType_X86_64_v2_AES
	cpus[cpuType_X86_64_v3_Lower] = CpuType_X86_64_v3
	cpus[cpuType_X86_64_v4_Lower] = CpuType_X86_64_v4
}

func (CpuType) Error(version Version) error {
	// v7
	cpus := CpuType("").cpuBase()
	if !version.Smaller(Version{Major: 8}) { // v8
		CpuType("").cpuV8(cpus)
	}
	cpusConverted := make([]string, len(cpus))
	var index int
	for _, e := range cpus {
		cpusConverted[index] = string(e)
		index++
	}
	slices.Sort(cpusConverted)
	return errors.New("cpuType can only be one of the following values: " + strings.Join(cpusConverted, ", "))
}

func (cpu CpuType) mapToApi(version Version) string {
	cpus := CpuType("").cpuBase()
	if !version.Smaller(Version{Major: 8}) {
		cpu.cpuV8(cpus)
	}
	if v, ok := cpus[CpuType(strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(string(cpu), "_", ""), "-", "")))]; ok {
		return string(v)
	}
	return ""
}

func (cpu CpuType) Validate(version Version) error {
	if cpu == "" || cpu.mapToApi(version) != "" {
		return nil
	}
	return CpuType("").Error(version)
}

type CpuUnits uint32 // min value 0 is unset, max value of 262144

const CpuUnits_Error_Maximum string = "maximum value of CpuUnits is 262144"

func (units CpuUnits) Validate() error {
	if units > 262144 {
		return errors.New(CpuUnits_Error_Maximum)
	}
	return nil
}

type CpuVirtualCores uint16 // min value 0 is unset, max value 512. is QemuCpuCores * CpuSockets

func (cores CpuVirtualCores) Error() error {
	return errors.New("CpuVirtualCores may have a maximum of " + strconv.FormatInt(int64(cores), 10))
}

func (vCores CpuVirtualCores) Validate(cores *QemuCpuCores, sockets *QemuCpuSockets, current *QemuCPU) error {
	var usedCores, usedSockets CpuVirtualCores
	if cores != nil {
		usedCores = CpuVirtualCores(*cores)
	} else if current != nil && current.Cores != nil {
		usedCores = CpuVirtualCores(*current.Cores)
	}
	if sockets != nil {
		usedSockets = CpuVirtualCores(*sockets)
	} else if current != nil && current.Sockets != nil {
		usedSockets = CpuVirtualCores(*current.Sockets)
	}
	if vCores > usedCores*usedSockets {
		return (usedCores * usedSockets).Error()
	}
	return nil
}

type QemuCPU struct {
	Affinity     *[]uint          `json:"affinity,omitempty"`
	Cores        *QemuCpuCores    `json:"cores,omitempty"` // Required during creation
	Flags        *CpuFlags        `json:"flags,omitempty"`
	Limit        *CpuLimit        `json:"limit,omitempty"`
	Numa         *bool            `json:"numa,omitempty"`
	Sockets      *QemuCpuSockets  `json:"sockets,omitempty"`
	Type         *CpuType         `json:"type,omitempty"`
	Units        *CpuUnits        `json:"units,omitempty"`
	VirtualCores *CpuVirtualCores `json:"vcores,omitempty"`
}

const (
	QemuCPU_Error_CoresRequired string = "cores is required"
)

func (cpu QemuCPU) mapToApi(current *QemuCPU, params map[string]interface{}, version Version) (delete string) {
	if cpu.Affinity != nil {
		if len(*cpu.Affinity) != 0 {
			params["affinity"] = cpu.mapToApiAffinity(*cpu.Affinity)
		} else if current != nil && current.Affinity != nil {
			params["affinity"] = ""
		}
	}
	if cpu.Cores != nil {
		params["cores"] = int(*cpu.Cores)
	}
	if cpu.Limit != nil {
		if *cpu.Limit != 0 {
			params["cpulimit"] = int(*cpu.Limit)
		} else if current != nil && current.Limit != nil {
			delete += ",cpulimit"
		}
	}
	if cpu.Numa != nil {
		params["numa"] = Btoi(*cpu.Numa)
	}
	if cpu.Sockets != nil {
		params["sockets"] = int(*cpu.Sockets)
	}
	if cpu.Flags != nil || cpu.Type != nil {
		var cpuType, flags string
		var flagsSet bool
		if current == nil { // Create
			if cpu.Flags != nil {
				flags, flagsSet = cpu.Flags.mapToApi(nil)
				if flagsSet && flags == "" {
					flagsSet = false
				}
			}
			if cpu.Type != nil {
				cpuType = cpu.Type.mapToApi(version)
			}
		} else { // Update
			if cpu.Flags != nil {
				flags, flagsSet = cpu.Flags.mapToApi(current.Flags)
			} else {
				flags, flagsSet = CpuFlags{}.mapToApi(current.Flags)
			}
			if cpu.Type != nil {
				cpuType = cpu.Type.mapToApi(version)
			} else if current.Type != nil {
				cpuType = current.Type.mapToApi(version)
			}
		}
		if flagsSet {
			if flags != "" {
				params["cpu"] = cpuType + ",flags=" + flags[1:]
			} else {
				params["cpu"] = cpuType + ",flags="
			}
		} else if cpuType != "" {
			params["cpu"] = cpuType
		}
	}
	if cpu.Units != nil {
		if *cpu.Units != 0 {
			params["cpuunits"] = int(*cpu.Units)
		} else if current != nil {
			delete += ",cpuunits"
		}
	}
	if cpu.VirtualCores != nil {
		if *cpu.VirtualCores != 0 {
			params["vcpus"] = int(*cpu.VirtualCores)
		} else if current != nil && current.VirtualCores != nil {
			delete += ",vcpus"
		}
	}
	return
}

func (QemuCPU) mapToApiAffinity(affinity []uint) string {
	sort.Slice(affinity, func(i, j int) bool {
		return affinity[i] < affinity[j]
	})
	var builder strings.Builder
	rangeStart, rangeEnd := affinity[0], affinity[0]
	for i := 1; i < len(affinity); i++ {
		if affinity[i] == affinity[i-1] {
			continue
		}
		if affinity[i] == rangeEnd+1 {
			// Continue the range
			rangeEnd = affinity[i]
		} else {
			// Close the current range and start a new range
			if rangeStart == rangeEnd {
				builder.WriteString(strconv.Itoa(int(rangeStart)) + ",")
			} else {
				builder.WriteString(strconv.Itoa(int(rangeStart)) + "-" + strconv.Itoa(int(rangeEnd)) + ",")
			}
			rangeStart, rangeEnd = affinity[i], affinity[i]
		}
	}
	// Append the last range
	if rangeStart == rangeEnd {
		builder.WriteString(strconv.Itoa(int(rangeStart)))
	} else {
		builder.WriteString(strconv.Itoa(int(rangeStart)) + "-" + strconv.Itoa(int(rangeEnd)))
	}
	return builder.String()
}

func (QemuCPU) mapToSDK(params map[string]interface{}) *QemuCPU {
	var cpu QemuCPU
	if v, isSet := params["affinity"]; isSet {
		if v.(string) != "" {
			cpu.Affinity = util.Pointer(QemuCPU{}.mapToSdkAffinity(v.(string)))
		} else {
			cpu.Affinity = util.Pointer(make([]uint, 0))
		}
	}
	if v, isSet := params["cores"]; isSet {
		cpu.Cores = util.Pointer(QemuCpuCores(v.(float64)))
	}
	if v, isSet := params["cpu"]; isSet {
		cpuParams := strings.SplitN(v.(string), ",", 2)
		cpu.Type = util.Pointer((CpuType)(cpuParams[0]))
		if len(cpuParams) > 1 && len(cpuParams[1]) > 6 {
			// `flags=` is the 6 characters bieng removed from the start of the string
			cpu.Flags = CpuFlags{}.mapToSDK(strings.Split(cpuParams[1][6:], ";"))
		}
	}
	if v, isSet := params["cpulimit"]; isSet {
		tmp, _ := parse.Uint(v)
		cpu.Limit = util.Pointer(CpuLimit(tmp))
	}
	if v, isSet := params["cpuunits"]; isSet {
		cpu.Units = util.Pointer(CpuUnits((v.(float64))))
	}
	if v, isSet := params["numa"]; isSet {
		cpu.Numa = util.Pointer(v.(float64) == 1)
	}
	if v, isSet := params["sockets"]; isSet {
		cpu.Sockets = util.Pointer(QemuCpuSockets(v.(float64)))
	}
	if value, isSet := params["vcpus"]; isSet {
		cpu.VirtualCores = util.Pointer(CpuVirtualCores((value.(float64))))
	}
	return &cpu
}

func (QemuCPU) mapToSdkAffinity(rawAffinity string) []uint {
	result := make([]uint, 0)
	for _, e := range strings.Split(rawAffinity, ",") {
		if strings.Contains(e, "-") {
			bounds := strings.Split(e, "-")
			start, _ := strconv.Atoi(bounds[0])
			end, _ := strconv.Atoi(bounds[1])
			for i := start; i <= end; i++ {
				result = append(result, uint(i))
			}
		} else {
			num, _ := strconv.Atoi(e)
			result = append(result, uint(num))
		}
	}
	return result
}

func (cpu QemuCPU) Validate(current *QemuCPU, version Version) (err error) {
	if cpu.Cores != nil {
		if err = cpu.Cores.Validate(); err != nil {
			return
		}
	} else if current == nil {
		return errors.New(QemuCPU_Error_CoresRequired)
	}
	if cpu.Flags != nil {
		if err = cpu.Flags.Validate(); err != nil {
			return
		}
	}
	if cpu.Limit != nil {
		if err = cpu.Limit.Validate(); err != nil {
			return
		}
	}
	if cpu.Sockets != nil {
		if err = cpu.Sockets.Validate(); err != nil {
			return
		}
	}
	if cpu.Type != nil {
		if err = cpu.Type.Validate(version); err != nil {
			return
		}
	}
	if cpu.Units != nil {
		if err = cpu.Units.Validate(); err != nil {
			return
		}
	}
	if cpu.VirtualCores != nil {
		if err = cpu.VirtualCores.Validate(cpu.Cores, cpu.Sockets, current); err != nil {
			return
		}
	}
	return
}

type QemuCpuCores uint8 // min value 1, max value of 128

const (
	QemuCpuCores_Error_LowerBound string = "minimum value of QemuCpuCores is 1"
	QemuCpuCores_Error_UpperBound string = "maximum value of QemuCpuCores is 128"
)

func (cores QemuCpuCores) Validate() error {
	if cores < 1 {
		return errors.New(QemuCpuCores_Error_LowerBound)
	}
	if cores > 128 {
		return errors.New(QemuCpuCores_Error_UpperBound)
	}
	return nil
}

type QemuCpuSockets uint8 // min value 1, max value 4

const (
	QemuCpuSockets_Error_LowerBound string = "minimum value of QemuCpuSockets is 1"
	QemuCpuSockets_Error_UpperBound string = "maximum value of QemuCpuSockets is 4"
)

func (sockets QemuCpuSockets) Validate() error {
	if sockets < 1 {
		return errors.New(QemuCpuSockets_Error_LowerBound)
	}
	if sockets > 4 {
		return errors.New(QemuCpuSockets_Error_UpperBound)
	}
	return nil
}
