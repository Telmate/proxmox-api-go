package proxmox

import (
	"context"
	"net"
	"strconv"
)

func (vmr *VmRef) GetAgentInformation(ctx context.Context, c *Client, statistics bool) ([]AgentNetworkInterface, error) {
	if err := c.CheckVmRef(ctx, vmr); err != nil {
		return nil, err
	}
	vmid := strconv.FormatInt(int64(vmr.vmId), 10)
	params, err := c.GetItemConfigMapStringInterface(ctx,
		"/nodes/"+vmr.node.String()+"/qemu/"+vmid+"/agent/network-get-interfaces", "guest agent", "data",
		"500 QEMU guest agent is not running",
		"500 VM "+vmid+" is not running")
	if err != nil {
		return nil, err
	}
	return AgentNetworkInterface{}.mapToSDK(params, statistics), nil
}

type AgentNetworkInterface struct {
	MacAddress  net.HardwareAddr
	IpAddresses []net.IP
	Name        string
	Statistics  *AgentInterfaceStatistics
}

func (AgentNetworkInterface) mapToSDK(params map[string]interface{}, statistics bool) []AgentNetworkInterface {
	var interfaces []interface{}
	if v, isSet := params["result"]; isSet {
		interfaces = v.([]interface{})
	}
	if len(interfaces) == 0 {
		return nil
	}
	agentInterfaces := make([]AgentNetworkInterface, len(interfaces))
	for i, e := range interfaces {
		iFace := e.(map[string]interface{})
		agentInterfaces[i] = AgentNetworkInterface{}
		if v, isSet := iFace["hardware-address"]; isSet {
			agentInterfaces[i].MacAddress, _ = net.ParseMAC(v.(string))
		}
		if v, isSet := iFace["ip-addresses"]; isSet {
			ips := v.([]interface{})
			agentInterfaces[i].IpAddresses = make([]net.IP, len(ips))
			for ii, ee := range ips {
				ip := ee.(map[string]interface{})
				agentInterfaces[i].IpAddresses[ii], _, _ = net.ParseCIDR(ip["ip-address"].(string) + "/" + strconv.FormatInt(int64(ip["prefix"].(float64)), 10))
			}
		}
		if v, isSet := iFace["name"]; isSet {
			agentInterfaces[i].Name = v.(string)
		}
		if statistics {
			if v, isSet := iFace["statistics"]; isSet {
				stats := v.(map[string]interface{})
				agentInterfaces[i].Statistics = &AgentInterfaceStatistics{
					RxBytes:   uint(stats["rx-bytes"].(float64)),
					RxDropped: uint(stats["rx-dropped"].(float64)),
					RxErrors:  uint(stats["rx-errs"].(float64)),
					RxPackets: uint(stats["rx-packets"].(float64)),
					TxBytes:   uint(stats["tx-bytes"].(float64)),
					TxDropped: uint(stats["tx-dropped"].(float64)),
					TxErrors:  uint(stats["tx-errs"].(float64)),
					TxPackets: uint(stats["tx-packets"].(float64))}
			}
		}
	}
	return agentInterfaces
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
