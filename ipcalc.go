package main

import (
	"fmt"
	"math"
	"net"
	"path"
	"strconv"
	"strings"
)

// A Calculate interface is used for IPv4 and IPv6
type Calculate interface {
	getNetworkInfo() error
}

type IPv6Calc struct {
	Address        string `json:"address"`
	Cidr           int    `json:"cidr"`
	Netmask        string `json:"netmask"`
	NetworkCidr    string `json:"networkCidr"`
	Network        string `json:"network"`
	HostsTotal     string `json:"hostsTotal"`
	HostsAvailable string `json:"hostsAvailable"`
	//HostMin        string `json:"hostMin"`
	//HostMax        string `json:"hostMax"`
}

// A IPv4Calc defines all the info for a network/IP
type IPv4Calc struct {
	Address        string `json:"address"`
	Broadcast      string `json:"broadcast,omitempty"`
	Cidr           int    `json:"cidr"`
	Netmask        string `json:"netmask"`
	NetworkCidr    string `json:"networkCidr"`
	Network        string `json:"network"`
	HostMin        string `json:"hostMin"`
	HostMax        string `json:"hostMax"`
	HostsAvailable int    `json:"hostsAvailable"`
	HostsTotal     int    `json:"hostsTotal"`
}

type errorJSON struct {
	Error string `json:"error"`
}

func calculate(calc Calculate) error {
	return calc.getNetworkInfo()
}

func newIPv4Calc(ip string, cidr int) *IPv4Calc {
	return &IPv4Calc{Address: ip, Cidr: cidr}
}

func newIPv6Calc(ip string, cidr int) *IPv6Calc {
	return &IPv6Calc{Address: ip, Cidr: cidr}
}

func (i *IPv6Calc) getNetworkInfo() error {
	_, ipNet, err := net.ParseCIDR(path.Join(i.Address, strconv.Itoa(i.Cidr)))
	if err != nil {
		return fmt.Errorf("something went wrong during cidr parsing: %v", err)
	}

	i.Netmask = net.IP(ipNet.Mask).String()
	totalFl := math.Pow(2, float64(128-i.Cidr))
	i.HostsTotal = fmt.Sprintf("%.0f", totalFl)
	i.HostsAvailable = i.HostsTotal
	i.NetworkCidr = ipNet.String()
	i.Network = strings.Split(i.NetworkCidr, "/")[0]

	return nil
}

func (i *IPv4Calc) getNetworkInfo() error {
	_, ipNet, err := net.ParseCIDR(path.Join(i.Address, strconv.Itoa(i.Cidr)))
	if err != nil {
		return fmt.Errorf("something went wrong during cidr parsing: %v", err)
	}

	// Available for all cidr
	i.Netmask = net.IP(ipNet.Mask).String()
	i.HostsTotal = 1 << (32 - i.Cidr)
	i.NetworkCidr = ipNet.String()
	i.Network = strings.Split(i.NetworkCidr, "/")[0]

	switch i.Cidr {
	case 32:
		i.HostMin = i.Address
		i.HostMax = i.Address
		i.HostsAvailable = 1
	case 31:
		i.HostMin = i.Network
		i.HostMax = nextIPv4(ipNet.IP, uint(i.HostsTotal-1))
		i.HostsAvailable = 2
	default:
		i.HostMin = nextIPv4(ipNet.IP, uint(1))
		i.HostMax = nextIPv4(ipNet.IP, uint(i.HostsTotal-2))
		i.Broadcast = nextIPv4(ipNet.IP, uint(i.HostsTotal-1))
		i.HostsAvailable = i.HostsTotal - 2
	}

	return nil
}

// nextIP will get the next ip from an IP address
func nextIPv4(ip net.IP, inc uint) string {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)

	return net.IPv4(v0, v1, v2, v3).String()
}
