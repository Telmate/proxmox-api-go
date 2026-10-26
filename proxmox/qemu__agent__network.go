package proxmox

import (
	"context"
	"net"
	"strings"
)

func (c *qemuAgentClient) NetworkInterfaces(ctx context.Context, vmr VmRef) (RawAgentNetworkInterfaces, GuestAgentState, error) {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	raw, err := ref.checkQemu(ctx, c.api)
	if err != nil {
		return nil, GuestAgentStateUnknown, err
	}
	if raw != nil {
		if raw.GetState() != PowerStateRunning {
			return nil, GuestAgentStateVmNotRunning, nil
		}
	}
	return c.NetworkInterfacesNoCheck(ctx, ref.legacy())
}

func (c *qemuAgentClient) NetworkInterfacesNoCheck(ctx context.Context, vmr VmRef) (RawAgentNetworkInterfaces, GuestAgentState, error) {
	ref := QemuRef{ID: vmr.vmId, Node: vmr.node}.generalize()
	params, err := c.api.getAgent(ctx, ref.ID, ref.Node, "network-get-interfaces", "guest agent", "data")
	if apiErr, ok := err.(*ApiError); ok {
		if strings.HasPrefix(apiErr.Message, "QEMU guest agent is not running") {
			return nil, GuestAgentStateNotRunning, nil
		}
		if strings.HasPrefix(apiErr.Message, "VM "+ref.ID.String()+" is not running") {
			return nil, GuestAgentStateVmNotRunning, nil
		}
	}
	return &rawAgentNetworkInterfaces{a: params.([]any)}, GuestAgentStateRunning, nil
}

// GuestAgentState is an Enum.
//
//	const(
//		GuestAgentStateRunning
//		GuestAgentStateNotRunning
//		GuestAgentStateVmNotRunning
//	)
type GuestAgentState int8

const (
	GuestAgentStateUnknown GuestAgentState = iota
	GuestAgentStateRunning
	GuestAgentStateNotRunning
	GuestAgentStateVmNotRunning
)

// Deprecated: use QemuAgentInterface.NetworkInterfaces() instead.
func (vmr *VmRef) GetAgentInformation(ctx context.Context, c *Client) (RawAgentNetworkInterfaces, GuestAgentState, error) {
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return nil, GuestAgentStateUnknown, err
	}
	return c.New().QemuAgent.NetworkInterfaces(ctx, *vmr)
}

type (
	RawAgentNetworkInterfaces interface {
		Get() []AgentNetworkInterface
		SelectMacAddress(address net.HardwareAddr) (RawAgentNetworkInterface, bool)
		SelectName(name string) (RawAgentNetworkInterface, bool)
	}
	rawAgentNetworkInterfaces struct {
		a []any
	}
)

func (raw *rawAgentNetworkInterfaces) Get() []AgentNetworkInterface {
	if len(raw.a) == 0 {
		return nil
	}
	agentInterfaces := make([]AgentNetworkInterface, len(raw.a))
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		agentInterfaces[i] = AgentNetworkInterface{
			IpAddresses: agentMapToSdkIpAddresses(iFace),
			MacAddress:  agentMapToSdkMacAddress(iFace),
			Name:        agentMapToSdkName(iFace),
			Statistics:  agentMapToSdkStatistics(iFace)}
	}
	return agentInterfaces
}

func (raw *rawAgentNetworkInterfaces) SelectMacAddress(address net.HardwareAddr) (RawAgentNetworkInterface, bool) {
	if len(raw.a) == 0 {
		return nil, false
	}
	addressString := address.String()
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		if v, isSet := iFace[agentApiKeyMaxAddress]; isSet {
			if strings.EqualFold(addressString, v.(string)) {
				return &rawAgentNetworkInterface{a: iFace}, true
			}
		}
	}
	return nil, false
}

func (raw *rawAgentNetworkInterfaces) SelectName(name string) (RawAgentNetworkInterface, bool) {
	if len(raw.a) == 0 {
		return nil, false
	}
	for i := range raw.a {
		iFace := raw.a[i].(map[string]any)
		if v, isSet := iFace[agentApiKeyName]; isSet {
			if name == v.(string) {
				return &rawAgentNetworkInterface{a: iFace}, true
			}
		}
	}
	return nil, false
}

type (
	RawAgentNetworkInterface interface {
		Get() AgentNetworkInterface
		GetIpAddresses() []net.IP
		GetMacAddress() net.HardwareAddr
		GetName() string
		GetStatistics() *AgentInterfaceStatistics
	}
	rawAgentNetworkInterface struct {
		a map[string]any
	}
)

func (raw *rawAgentNetworkInterface) Get() AgentNetworkInterface {
	return AgentNetworkInterface{
		IpAddresses: raw.GetIpAddresses(),
		MacAddress:  raw.GetMacAddress(),
		Name:        raw.GetName(),
		Statistics:  raw.GetStatistics()}
}

func (raw *rawAgentNetworkInterface) GetIpAddresses() []net.IP {
	return agentMapToSdkIpAddresses(raw.a)
}

func (raw *rawAgentNetworkInterface) GetMacAddress() net.HardwareAddr {
	return agentMapToSdkMacAddress(raw.a)
}

func (raw *rawAgentNetworkInterface) GetName() string {
	return agentMapToSdkName(raw.a)
}

func (raw *rawAgentNetworkInterface) GetStatistics() *AgentInterfaceStatistics {
	return agentMapToSdkStatistics(raw.a)
}

func agentMapToSdkMacAddress(params map[string]any) net.HardwareAddr {
	if v, isSet := params[agentApiKeyMaxAddress]; isSet {
		mac, _ := net.ParseMAC(v.(string))
		return mac
	}
	return nil
}

func agentMapToSdkIpAddresses(params map[string]any) []net.IP {
	if v, isSet := params[agentApiKeyIpAddresses]; isSet {
		RawIPs := v.([]any)
		ips := make([]net.IP, len(RawIPs))
		for i := range RawIPs {
			ip := RawIPs[i].(map[string]any)
			ips[i] = net.ParseIP(ip["ip-address"].(string))
		}
		return ips
	}
	return nil
}

func agentMapToSdkName(params map[string]any) string {
	if v, isSet := params[agentApiKeyName]; isSet {
		return v.(string)
	}
	return ""
}

func agentMapToSdkStatistics(params map[string]any) *AgentInterfaceStatistics {
	if v, isSet := params[agentApiKeyStatistics]; isSet {
		stats := v.(map[string]any)
		return &AgentInterfaceStatistics{
			RxBytes:   uint(stats["rx-bytes"].(float64)),
			RxDropped: uint(stats["rx-dropped"].(float64)),
			RxErrors:  uint(stats["rx-errs"].(float64)),
			RxPackets: uint(stats["rx-packets"].(float64)),
			TxBytes:   uint(stats["tx-bytes"].(float64)),
			TxDropped: uint(stats["tx-dropped"].(float64)),
			TxErrors:  uint(stats["tx-errs"].(float64)),
			TxPackets: uint(stats["tx-packets"].(float64))}
	}
	return nil
}

const (
	agentApiKeyResult      string = "result"
	agentApiKeyIpAddresses string = "ip-addresses"
	agentApiKeyMaxAddress  string = "hardware-address"
	agentApiKeyName        string = "name"
	agentApiKeyStatistics  string = "statistics"
)

type AgentNetworkInterface struct {
	MacAddress  net.HardwareAddr
	IpAddresses []net.IP
	Name        string
	Statistics  *AgentInterfaceStatistics
}

type AgentInterfaceStatistics struct {
	RxBytes   uint
	RxDropped uint
	RxErrors  uint
	RxPackets uint
	TxBytes   uint
	TxDropped uint
	TxErrors  uint
	TxPackets uint
}
