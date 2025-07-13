package proxmox

import (
	"errors"
	"net"
)

type IPv4Address string

const IPv4Address_Error_Invalid = "ipv4Address is not a valid ipv4 address"

func (ip IPv4Address) String() string { return string(ip) } // String is for fmt.Stringer.

func (ip IPv4Address) Validate() error {
	if ip == "" {
		return nil
	}
	if net.ParseIP(string(ip)) == nil {
		return errors.New(IPv4Address_Error_Invalid)
	}
	if !isIPv4(string(ip)) {
		return errors.New(IPv4Address_Error_Invalid)
	}
	return nil
}

type IPv4CIDR string

const IPv4CIDR_Error_Invalid = "ipv4CIDR is not a valid ipv4 address"

func (cidr IPv4CIDR) String() string { return string(cidr) } // String is for fmt.Stringer.

func (cidr IPv4CIDR) Validate() error {
	if cidr == "" {
		return nil
	}
	ip, _, err := net.ParseCIDR(string(cidr))
	if err != nil {
		return errors.New(IPv4CIDR_Error_Invalid)
	}
	if !isIPv4(ip.String()) {
		return errors.New(IPv4CIDR_Error_Invalid)
	}
	return err
}

type IPv6Address string

const IPv6Address_Error_Invalid = "ipv6Address is not a valid ipv6 address"

func (ip IPv6Address) String() string { return string(ip) } // String is for fmt.Stringer.

func (ip IPv6Address) Validate() error {
	if ip == "" {
		return nil
	}
	if net.ParseIP(string(ip)) == nil {
		return errors.New(IPv6Address_Error_Invalid)
	}
	if !isIPv6(string(ip)) {
		return errors.New(IPv6Address_Error_Invalid)
	}
	return nil
}

type IPv6CIDR string

const IPv6CIDR_Error_Invalid = "ipv6CIDR is not a valid ipv6 address"

func (cidr IPv6CIDR) String() string { return string(cidr) } // String is for fmt.Stringer.

func (cidr IPv6CIDR) Validate() error {
	if cidr == "" {
		return nil
	}
	ip, _, err := net.ParseCIDR(string(cidr))
	if err != nil {
		return errors.New(IPv6CIDR_Error_Invalid)
	}
	if !isIPv6(ip.String()) {
		return errors.New(IPv6CIDR_Error_Invalid)
	}
	return nil
}
