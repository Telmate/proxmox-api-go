package proxmox

import (
	"net"
	"strings"
)

type (
	RawLxcInfoNetworkInterfaces interface {
		Get() []LxcInfoNetworkInterface
		SelectMacAddress(address net.HardwareAddr) (RawLxcInfoNetworkInterface, bool)
		SelectName(name LxcNetworkName) (RawLxcInfoNetworkInterface, bool)
	}
	rawLxcInfoNetworkInterfaces struct {
		a []any
	}
)

var _ RawLxcInfoNetworkInterfaces = (*rawLxcInfoNetworkInterfaces)(nil)

func (raw *rawLxcInfoNetworkInterfaces) Get() []LxcInfoNetworkInterface {
	if len(raw.a) == 0 {
		return nil
	}
	agentInterfaces := make([]LxcInfoNetworkInterface, len(raw.a))
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		agentInterfaces[i] = LxcInfoNetworkInterface{
			IpAddresses: agentMapToSdkIpAddresses(iFace),
			MacAddress:  agentMapToSdkMacAddress(iFace),
			Name:        LxcNetworkName(agentMapToSdkName(iFace))}
	}
	return agentInterfaces
}

func (raw *rawLxcInfoNetworkInterfaces) SelectMacAddress(address net.HardwareAddr) (RawLxcInfoNetworkInterface, bool) {
	if len(raw.a) == 0 {
		return nil, false
	}
	addressString := address.String()
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		if v, isSet := iFace["hardware-address"]; isSet {
			if strings.EqualFold(addressString, v.(string)) {
				return &rawLxcInfoNetworkInterface{a: iFace}, true
			}
		}
	}
	return nil, false
}

func (raw *rawLxcInfoNetworkInterfaces) SelectName(name LxcNetworkName) (RawLxcInfoNetworkInterface, bool) {
	if len(raw.a) == 0 {
		return nil, false
	}
	tmpName := name.String()
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		if v, isSet := iFace[agentApiKeyName]; isSet {
			if tmpName == v.(string) {
				return &rawLxcInfoNetworkInterface{a: iFace}, true
			}
		}
	}
	return nil, false
}

type (
	RawLxcInfoNetworkInterface interface {
		Get() LxcInfoNetworkInterface
		GetIpAddresses() []net.IP
		GetMacAddress() net.HardwareAddr
		GetName() LxcNetworkName
	}
	rawLxcInfoNetworkInterface struct {
		a map[string]any
	}
)

var _ RawLxcInfoNetworkInterface = (*rawLxcInfoNetworkInterface)(nil)

func (raw *rawLxcInfoNetworkInterface) Get() LxcInfoNetworkInterface {
	return LxcInfoNetworkInterface{
		IpAddresses: raw.GetIpAddresses(),
		MacAddress:  raw.GetMacAddress(),
		Name:        raw.GetName()}
}

func (raw *rawLxcInfoNetworkInterface) GetIpAddresses() []net.IP {
	return agentMapToSdkIpAddresses(raw.a)
}

func (raw *rawLxcInfoNetworkInterface) GetMacAddress() net.HardwareAddr {
	return agentMapToSdkMacAddress(raw.a)
}

func (raw *rawLxcInfoNetworkInterface) GetName() LxcNetworkName {
	return LxcNetworkName(agentMapToSdkName(raw.a))
}

type LxcInfoNetworkInterface struct {
	IpAddresses []net.IP
	MacAddress  net.HardwareAddr
	Name        LxcNetworkName
}
