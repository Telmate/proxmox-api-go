package proxmox

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/parse"
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

func (flags CpuFlags) mapToApi(current *CpuFlags, b *strings.Builder) {
	const numberOfFlags = 12
	flagNames := [numberOfFlags]string{
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

	flagValues := [numberOfFlags]*TriBool{
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

	var currentValues [numberOfFlags]*TriBool
	if current != nil {
		currentValues = [numberOfFlags]*TriBool{
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
	}

	for i, value := range flagValues {
		if value != nil {
			switch *value {
			case TriBoolTrue:
				b.WriteString(semicolon + plus)
				b.WriteString(flagNames[i])
			case TriBoolFalse:
				b.WriteString(semicolon + "-")
				b.WriteString(flagNames[i])
			}
		} else if currentValues[i] != nil {
			switch *currentValues[i] {
			case TriBoolTrue:
				b.WriteString(semicolon + plus)
				b.WriteString(flagNames[i])
			case TriBoolFalse:
				b.WriteString(semicolon + "-")
				b.WriteString(flagNames[i])
			}
		}
	}
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

func (limit CpuLimit) String() string { return strconv.FormatUint(uint64(limit), 10) } // For fmt.Stringer

func (limit CpuLimit) Validate() error {
	if limit > 128 {
		return errors.New(CpuLimit_Error_Maximum)
	}
	return nil
}

type CpuType string // enum

var cpuTypeTableV7 = map[string]string{
	string(CpuType_AmdAthlon):                         string(CpuType_AmdAthlon),
	string(CpuType_AmdPhenom):                         string(CpuType_AmdPhenom),
	string(CpuType_Intel486):                          string(CpuType_Intel486),
	string(CpuType_IntelCore2Duo):                     string(CpuType_IntelCore2Duo),
	string(CpuType_IntelCoreDuo):                      string(CpuType_IntelCoreDuo),
	string(CpuType_IntelPentium):                      string(CpuType_IntelPentium),
	string(CpuType_IntelPentium2):                     string(CpuType_IntelPentium2),
	string(CpuType_IntelPentium3):                     string(CpuType_IntelPentium3),
	string(CpuType_QemuKvm32):                         string(CpuType_QemuKvm32),
	string(CpuType_QemuKvm64):                         string(CpuType_QemuKvm64),
	string(CpuType_QemuMax):                           string(CpuType_QemuMax),
	string(CpuType_Qemu32):                            string(CpuType_Qemu32),
	string(CpuType_Qemu64):                            string(CpuType_Qemu64),
	string(CpuType_Host):                              string(CpuType_Host),
	string(cpuType_AmdEPYC_Lower):                     string(CpuType_AmdEPYC),
	string(cpuType_AmdEPYCIBPB_Lower):                 string(CpuType_AmdEPYCIBPB),
	string(cpuType_AmdEPYCMilan_Lower):                string(CpuType_AmdEPYCMilan),
	string(cpuType_AmdEPYCRome_Lower):                 string(CpuType_AmdEPYCRome),
	string(cpuType_AmdOpteronG1_Lower):                string(CpuType_AmdOpteronG1),
	string(cpuType_AmdOpteronG2_Lower):                string(CpuType_AmdOpteronG2),
	string(cpuType_AmdOpteronG3_Lower):                string(CpuType_AmdOpteronG3),
	string(cpuType_AmdOpteronG4_Lower):                string(CpuType_AmdOpteronG4),
	string(cpuType_AmdOpteronG5_Lower):                string(CpuType_AmdOpteronG5),
	string(cpuType_IntelBroadwell_Lower):              string(CpuType_IntelBroadwell),
	string(cpuType_IntelBroadwellIBRS_Lower):          string(CpuType_IntelBroadwellIBRS),
	string(cpuType_IntelBroadwellNoTSX_Lower):         string(CpuType_IntelBroadwellNoTSX),
	string(cpuType_IntelBroadwellNoTSXIBRS_Lower):     string(CpuType_IntelBroadwellNoTSXIBRS),
	string(cpuType_IntelCascadelakeServer_Lower):      string(CpuType_IntelCascadelakeServer),
	string(cpuType_IntelCascadelakeServerNoTSX_Lower): string(CpuType_IntelCascadelakeServerNoTSX),
	string(cpuType_IntelConroe_Lower):                 string(CpuType_IntelConroe),
	string(cpuType_IntelHaswell_Lower):                string(CpuType_IntelHaswell),
	string(cpuType_IntelHaswellIBRS_Lower):            string(CpuType_IntelHaswellIBRS),
	string(cpuType_IntelHaswellNoTSX_Lower):           string(CpuType_IntelHaswellNoTSX),
	string(cpuType_IntelHaswellNoTSXIBRS_Lower):       string(CpuType_IntelHaswellNoTSXIBRS),
	string(cpuType_IntelIcelakeClient_Lower):          string(CpuType_IntelIcelakeClient),
	string(cpuType_IntelIcelakeClientNoTSX_Lower):     string(CpuType_IntelIcelakeClientNoTSX),
	string(cpuType_IntelIcelakeServer_Lower):          string(CpuType_IntelIcelakeServer),
	string(cpuType_IntelIcelakeServerNoTSX_Lower):     string(CpuType_IntelIcelakeServerNoTSX),
	string(cpuType_IntelIvybridge_Lower):              string(CpuType_IntelIvybridge),
	string(cpuType_IntelIvybridgeIBRS_Lower):          string(CpuType_IntelIvybridgeIBRS),
	string(cpuType_IntelKnightsmill_Lower):            string(CpuType_IntelKnightsmill),
	string(cpuType_IntelNehalem_Lower):                string(CpuType_IntelNehalem),
	string(cpuType_IntelNehalemIBRS_Lower):            string(CpuType_IntelNehalemIBRS),
	string(cpuType_IntelPenrym_Lower):                 string(CpuType_IntelPenrym),
	string(cpuType_IntelSandyBridge_Lower):            string(CpuType_IntelSandyBridge),
	string(cpuType_IntelSandybridgeIBRS_Lower):        string(CpuType_IntelSandybridgeIBRS),
	string(cpuType_IntelSkylakeClient_Lower):          string(CpuType_IntelSkylakeClient),
	string(cpuType_IntelSkylakeClientIBRS_Lower):      string(CpuType_IntelSkylakeClientIBRS),
	string(cpuType_IntelSkylakeClientNoTSXIBRS_Lower): string(CpuType_IntelSkylakeClientNoTSXIBRS),
	string(cpuType_IntelSkylakeServer_Lower):          string(CpuType_IntelSkylakeServer),
	string(cpuType_IntelSkylakeServerIBRS_Lower):      string(CpuType_IntelSkylakeServerIBRS),
	string(cpuType_IntelSkylakeServerNoTSXIBRS_Lower): string(CpuType_IntelSkylakeServerNoTSXIBRS),
	string(cpuType_IntelWestmere_Lower):               string(CpuType_IntelWestmere),
	string(cpuType_IntelWestmereIBRS_Lower):           string(CpuType_IntelWestmereIBRS),
	string(cpuType_ArmCortexA57_Lower):                string(CpuType_ArmCortexA57),
	string(cpuType_ArmCortexA72_Lower):                string(CpuType_ArmCortexA72),
}

var cpuTypeTableV8 = map[string]string{
	string(cpuType_IntelCascadelakeServerV2_Lower): string(CpuType_IntelCascadelakeServerV2),
	string(cpuType_IntelCascadelakeServerV4_Lower): string(CpuType_IntelCascadelakeServerV4),
	string(cpuType_IntelCascadelakeServerV5_Lower): string(CpuType_IntelCascadelakeServerV5),
	string(cpuType_IntelCooperlake_Lower):          string(CpuType_IntelCooperlake),
	string(cpuType_IntelCooperlakeV2_Lower):        string(CpuType_IntelCooperlakeV2),
	string(cpuType_AmdEPYCMilanV2_Lower):           string(CpuType_AmdEPYCMilanV2),
	string(cpuType_AmdEPYCRomeV2_Lower):            string(CpuType_AmdEPYCRomeV2),
	string(cpuType_AmdEPYCV3_Lower):                string(CpuType_AmdEPYCV3),
	string(cpuType_AmdEPYCGenoa_Lower):             string(CpuType_AmdEPYCGenoa),
	string(cpuType_AmdEPYCGenoaV2_Lower):           string(CpuType_AmdEPYCGenoaV2),
	string(cpuType_IntelIcelakeServerV3_Lower):     string(CpuType_IntelIcelakeServerV3),
	string(cpuType_IntelIcelakeServerV4_Lower):     string(CpuType_IntelIcelakeServerV4),
	string(cpuType_IntelIcelakeServerV5_Lower):     string(CpuType_IntelIcelakeServerV5),
	string(cpuType_IntelIcelakeServerV6_Lower):     string(CpuType_IntelIcelakeServerV6),
	string(cpuType_IntelSapphireRapids_Lower):      string(CpuType_IntelSapphireRapids),
	string(cpuType_IntelSkylakeClientV4_Lower):     string(CpuType_IntelSkylakeClientV4),
	string(cpuType_IntelSkylakeServerV4_Lower):     string(CpuType_IntelSkylakeServerV4),
	string(cpuType_IntelSkylakeServerV5_Lower):     string(CpuType_IntelSkylakeServerV5),
	string(cpuType_X86_64_v2_Lower):                string(CpuType_X86_64_v2),
	string(cpuType_X86_64_v2_AES_Lower):            string(CpuType_X86_64_v2_AES),
	string(cpuType_X86_64_v3_Lower):                string(CpuType_X86_64_v3),
	string(cpuType_X86_64_v4_Lower):                string(CpuType_X86_64_v4),
}

var cpuTypeTableV9 = map[string]string{
	string(cpuType_AmdEPYCTurin_Lower): string(CpuType_AmdEPYCTurin),
}

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
	CpuType_AmdEPYCMilanV2                    CpuType = "EPYC-Milan-v2"
	cpuType_AmdEPYCMilanV2_Lower              CpuType = "epycmilanv2"
	CpuType_AmdEPYCRome                       CpuType = "EPYC-Rome"
	cpuType_AmdEPYCRome_Lower                 CpuType = "epycrome"
	CpuType_AmdEPYCRomeV2                     CpuType = "EPYC-Rome-v2"
	cpuType_AmdEPYCRomeV2_Lower               CpuType = "epycromev2"
	CpuType_AmdEPYCV3                         CpuType = "EPYC-v3"
	cpuType_AmdEPYCV3_Lower                   CpuType = "epycv3"
	CpuType_AmdEPYCGenoa                      CpuType = "EPYC-Genoa"
	cpuType_AmdEPYCGenoa_Lower                CpuType = "epycgenoa"
	CpuType_AmdEPYCGenoaV2                    CpuType = "EPYC-Genoa-v2"
	cpuType_AmdEPYCGenoaV2_Lower              CpuType = "epycgenoav2"
	CpuType_AmdEPYCTurin                      CpuType = "EPYC-Turin"
	cpuType_AmdEPYCTurin_Lower                CpuType = "epycturin"
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
	CpuType_ArmCortexA57                      CpuType = "cortex-a57"
	cpuType_ArmCortexA57_Lower                CpuType = "cortexa57"
	CpuType_ArmCortexA72                      CpuType = "cortex-a72"
	cpuType_ArmCortexA72_Lower                CpuType = "cortexa72"
)

func (CpuType) Error(version Version) error {
	length := len(cpuTypeTableV7)
	if version.Major >= 8 { // v8
		length += len(cpuTypeTableV8)
		if version.Major >= 9 { //v9
			length += len(cpuTypeTableV9)
		}
	}
	cpus := make([]string, length)
	offset := 0
	for v := range cpuTypeTableV7 {
		cpus[offset] = string(v)
		offset++
	}
	if version.Major >= 8 {
		for v := range cpuTypeTableV8 {
			cpus[offset] = string(v)
			offset++
		}
		if version.Major >= 9 {
			for v := range cpuTypeTableV9 {
				cpus[offset] = string(v)
				offset++
			}
		}
	}
	slices.Sort(cpus)
	return errors.New("cpuType can only be one of the following values: " + strings.Join(cpus, ", "))
}

func (cpu CpuType) mapToApi(version Version) string {
	cpuString := string(cpu)
	if strings.HasPrefix(cpuString, "custom-") {
		return cpuString
	}
	cpuNormalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(cpuString, "_", ""), "-", ""))
	if v, ok := cpuTypeTableV7[cpuNormalized]; ok {
		return v
	}
	if version.Major >= 8 {
		if v, ok := cpuTypeTableV8[cpuNormalized]; ok {
			return v
		}
		if version.Major >= 9 {
			if v, ok := cpuTypeTableV9[cpuNormalized]; ok {
				return v
			}
		}
	}
	return ""
}

func (cpu CpuType) Validate(version Version) error {
	if cpu == "" || cpu.mapToApi(version) != "" {
		return nil
	}
	return CpuType("").Error(version)
}

type QemuCpuUnits uint32 // min value 0 is unset, max value of 262144

const CpuUnits_Error_Maximum string = "maximum value of QemuCpuUnits is 262144"

func (units QemuCpuUnits) String() string { return strconv.FormatUint(uint64(units), 10) } // For fmt.Stringer

func (units QemuCpuUnits) Validate() error {
	if units > 262144 {
		return errors.New(CpuUnits_Error_Maximum)
	}
	return nil
}

type CpuVirtualCores uint16 // min value 0 is unset, max value 512. is QemuCpuCores * CpuSockets

func (cores CpuVirtualCores) Error() error {
	return errors.New("CpuVirtualCores may have a maximum of " + strconv.FormatInt(int64(cores), 10))
}

func (cores CpuVirtualCores) String() string { return strconv.FormatUint(uint64(cores), 10) } // For fmt.Stringer

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
	Cores        *QemuCpuCores    `json:"cores,omitempty"` // Required during creation, never nil when returned
	Flags        *CpuFlags        `json:"flags,omitempty"`
	Limit        *CpuLimit        `json:"limit,omitempty"`
	Numa         *bool            `json:"numa,omitempty"`    // Never nil when returned
	Sockets      *QemuCpuSockets  `json:"sockets,omitempty"` // Never nil when returned
	Type         *CpuType         `json:"type,omitempty"`
	Units        *QemuCpuUnits    `json:"units,omitempty"`
	VirtualCores *CpuVirtualCores `json:"vcores,omitempty"`
}

const (
	QemuCPU_Error_CoresRequired string = "cores is required"
)

func (cpu QemuCPU) mapToApiCreate(version Version, b *strings.Builder) {
	if cpu.Affinity != nil && len(*cpu.Affinity) > 0 {
		b.WriteString("&" + qemuApiKeyCpuAffinity + "=")
		affinity := make([]uint, len(*cpu.Affinity))
		copy(affinity, *cpu.Affinity)
		qemuCpuMapToApiAffinity(affinity, b)
	}
	if cpu.Cores != nil {
		b.WriteString("&" + qemuApiKeyCpuCores + "=")
		b.WriteString(cpu.Cores.String())
	}
	if cpu.Limit != nil && *cpu.Limit != 0 {
		b.WriteString("&" + qemuApiKeyCpuLimit + "=")
		b.WriteString(cpu.Limit.String())
	}
	if cpu.Numa != nil {
		b.WriteString("&" + qemuApiKeyCpuNuma + "=")
		b.WriteRune(bTOr(*cpu.Numa))
	}
	if cpu.Sockets != nil {
		b.WriteString("&" + qemuApiKeyCpuSockets + "=")
		b.WriteString(cpu.Sockets.String())
	}
	if cpu.Flags != nil || cpu.Type != nil {
		var cpuType string
		var tmpBuilder strings.Builder
		if cpu.Type != nil {
			cpuType = cpu.Type.mapToApi(version)
		}
		if cpu.Flags != nil {
			cpu.Flags.mapToApi(nil, &tmpBuilder)
		}
		if cpuType != "" || tmpBuilder.Len() > 0 {
			b.WriteString("&" + qemuApiKeyCpuType + "=")
			if cpuType != "" {
				b.WriteString(cpuType)
			}
			if tmpBuilder.Len() > 3 {
				b.WriteString(comma + "flags" + equal)
				b.WriteString(tmpBuilder.String()[3:])
			}
		}
	}
	if cpu.Units != nil && *cpu.Units != 0 {
		b.WriteString("&" + qemuApiKeyCpuUnits + "=")
		b.WriteString(cpu.Units.String())
	}
	if cpu.VirtualCores != nil && *cpu.VirtualCores != 0 {
		b.WriteString("&" + qemuApiKeyCpuVirtual + "=")
		b.WriteString(cpu.VirtualCores.String())
	}
}

func (cpu QemuCPU) mapToApiUpdate(current QemuCPU, version Version, b, delete *strings.Builder) {
	if cpu.Affinity != nil {
		if current.Affinity == nil {
			if len(*cpu.Affinity) > 0 { // create
				b.WriteString("&" + qemuApiKeyCpuAffinity + "=")
				affinity := make([]uint, len(*cpu.Affinity))
				copy(affinity, *cpu.Affinity)
				qemuCpuMapToApiAffinity(affinity, b)
			}
		} else if len(*current.Affinity) > 0 {
			if len(*cpu.Affinity) == 0 { // delete
				delete.WriteString("," + qemuApiKeyCpuAffinity)
			} else { // update
				var tmpBuilder strings.Builder
				qemuCpuMapToApiAffinity(*current.Affinity, &tmpBuilder)
				currentAffinity := tmpBuilder.String()
				tmpBuilder = strings.Builder{}
				affinity := make([]uint, len(*cpu.Affinity))
				copy(affinity, *cpu.Affinity)
				qemuCpuMapToApiAffinity(affinity, &tmpBuilder)
				newAffinity := tmpBuilder.String()
				if newAffinity != currentAffinity {
					b.WriteString("&" + qemuApiKeyCpuAffinity + "=")
					b.WriteString(newAffinity)
				}
			}
		}
	}
	if cpu.Cores != nil && *cpu.Cores != *current.Cores {
		b.WriteString("&" + qemuApiKeyCpuCores + "=")
		b.WriteString(cpu.Cores.String())
	}
	if cpu.Limit != nil {
		if current.Limit == nil {
			if *cpu.Limit != 0 {
				b.WriteString("&" + qemuApiKeyCpuLimit + "=")
				b.WriteString(cpu.Limit.String())
			}
		} else if *cpu.Limit != *current.Limit {
			if *cpu.Limit == 0 {
				delete.WriteString("," + qemuApiKeyCpuLimit)
			} else {
				b.WriteString("&" + qemuApiKeyCpuLimit + "=")
				b.WriteString(cpu.Limit.String())
			}
		}
	}
	if cpu.Numa != nil && *cpu.Numa != *current.Numa {
		b.WriteString("&" + qemuApiKeyCpuNuma + "=")
		b.WriteRune(bTOr(*cpu.Numa))
	}
	if cpu.Sockets != nil && *cpu.Sockets != *current.Sockets {
		b.WriteString("&" + qemuApiKeyCpuSockets + "=")
		b.WriteString(cpu.Sockets.String())
	}
	if cpu.Flags != nil || cpu.Type != nil {
		var changes bool
		var cpuType, flags string
		if cpu.Type != nil {
			cpuType = cpu.Type.mapToApi(version)
			if current.Type != nil {
				if cpuType != string(*current.Type) {
					changes = true
				}
			} else if *cpu.Type != "" {
				changes = true
			}
		} else if current.Type != nil {
			cpuType = string(*current.Type)
		}
		if cpu.Flags != nil {
			var tmpBuilder strings.Builder
			if current.Flags != nil { // update
				current.Flags.mapToApi(nil, &tmpBuilder)
				flags = tmpBuilder.String()
				tmpBuilder = strings.Builder{}
				cpu.Flags.mapToApi(current.Flags, &tmpBuilder)
				newFlags := tmpBuilder.String()
				if newFlags != flags {
					flags = newFlags
					changes = true
				}
			} else { // create
				cpu.Flags.mapToApi(nil, &tmpBuilder)
				flags = tmpBuilder.String()
				changes = true
			}
		} else if changes && current.Flags != nil {
			var tmpBuilder strings.Builder
			current.Flags.mapToApi(nil, &tmpBuilder)
			flags = tmpBuilder.String()
		}
		if changes {
			if cpuType != "" || flags != "" {
				b.WriteString("&" + qemuApiKeyCpuType + "=")
				if cpuType != "" {
					b.WriteString(cpuType)
				}
				if len(flags) > 3 {
					b.WriteString(comma + "flags" + equal)
					b.WriteString(flags[3:])
				}
			} else {
				delete.WriteString("," + qemuApiKeyCpuType)
			}
		}
	}
	if cpu.Units != nil {
		if current.Units == nil {
			if *cpu.Units != 0 {
				b.WriteString("&" + qemuApiKeyCpuUnits + "=")
				b.WriteString(cpu.Units.String())
			}
		} else if *cpu.Units != *current.Units {
			if *cpu.Units == 0 {
				delete.WriteString("," + qemuApiKeyCpuUnits)
			} else {
				b.WriteString("&" + qemuApiKeyCpuUnits + "=")
				b.WriteString(cpu.Units.String())
			}
		}
	}
	if cpu.VirtualCores != nil {
		if current.VirtualCores == nil {
			if *cpu.VirtualCores != 0 {
				b.WriteString("&" + qemuApiKeyCpuVirtual + "=")
				b.WriteString(cpu.VirtualCores.String())
			}
		} else if *cpu.VirtualCores != *current.VirtualCores {
			if *cpu.VirtualCores == 0 {
				delete.WriteString("," + qemuApiKeyCpuVirtual)
			} else {
				b.WriteString("&" + qemuApiKeyCpuVirtual + "=")
				b.WriteString(cpu.VirtualCores.String())
			}
		}
	}
}

func qemuCpuMapToApiAffinity(affinity []uint, b *strings.Builder) {
	slices.Sort(affinity)
	rangeStart, rangeEnd := affinity[0], affinity[0]
	for i := 1; i < len(affinity); i++ {
		if affinity[i] == affinity[i-1] {
			continue
		}
		if affinity[i] == rangeEnd+1 {
			// Continue the range
			rangeEnd = affinity[i]
			continue
		}
		// Close the current range and start a new range
		b.WriteString(strconv.Itoa(int(rangeStart)))
		if rangeStart == rangeEnd {
			b.WriteString(comma)
		} else {
			b.WriteRune('-')
			b.WriteString(strconv.Itoa(int(rangeEnd)))
			b.WriteString(comma)
		}
		rangeStart, rangeEnd = affinity[i], affinity[i]
	}
	b.WriteString(strconv.Itoa(int(rangeStart)))
	// Append the last range
	if rangeStart != rangeEnd {
		b.WriteRune('-')
		b.WriteString(strconv.Itoa(int(rangeEnd)))
	}
}

func (raw *rawConfigQemu) GetCPU() QemuCPU {
	cpu := QemuCPU{
		Cores:   new(QemuCpuCores(1)),
		Numa:    new(false),
		Sockets: new(QemuCpuSockets(1))}
	if v, isSet := raw.a[qemuApiKeyCpuAffinity]; isSet {
		if v.(string) != "" {
			cpu.Affinity = new(QemuCPU{}.mapToSdkAffinity(v.(string)))
		} else {
			cpu.Affinity = new(make([]uint, 0))
		}
	}
	if v, isSet := raw.a[qemuApiKeyCpuCores]; isSet {
		*cpu.Cores = QemuCpuCores(v.(float64))
	}
	if v, isSet := raw.a[qemuApiKeyCpuType]; isSet {
		cpuParams := strings.SplitN(v.(string), ",", 2)
		cpu.Type = new((CpuType)(cpuParams[0]))
		if len(cpuParams) > 1 && len(cpuParams[1]) > 6 {
			// `flags=` is the 6 characters bieng removed from the start of the string
			cpu.Flags = CpuFlags{}.mapToSDK(strings.Split(cpuParams[1][6:], ";"))
		}
	}
	if v, isSet := raw.a[qemuApiKeyCpuLimit]; isSet {
		tmp, _ := parse.Uint(v)
		cpu.Limit = new(CpuLimit(tmp))
	}
	if v, isSet := raw.a[qemuApiKeyCpuUnits]; isSet {
		cpu.Units = new(QemuCpuUnits((v.(float64))))
	}
	if v, isSet := raw.a[qemuApiKeyCpuNuma]; isSet {
		*cpu.Numa = v.(float64) == 1
	}
	if v, isSet := raw.a[qemuApiKeyCpuSockets]; isSet {
		*cpu.Sockets = QemuCpuSockets(v.(float64))
	}
	if v, isSet := raw.a[qemuApiKeyCpuVirtual]; isSet {
		cpu.VirtualCores = new(CpuVirtualCores((v.(float64))))
	}
	return cpu
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

func (cores QemuCpuCores) String() string { return strconv.FormatUint(uint64(cores), 10) } // For fmt.Stringer

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

func (sockets QemuCpuSockets) String() string { return strconv.FormatUint(uint64(sockets), 10) } // For fmt.Stringer

func (sockets QemuCpuSockets) Validate() error {
	if sockets < 1 {
		return errors.New(QemuCpuSockets_Error_LowerBound)
	}
	if sockets > 4 {
		return errors.New(QemuCpuSockets_Error_UpperBound)
	}
	return nil
}
