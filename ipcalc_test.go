package main

import (
	"net"
	"testing"
)

func TestGetNetworkInfo(t *testing.T) {
	var validTests = []struct {
		actual, expected IPv4Calc
	}{
		{
			IPv4Calc{Address: "192.168.0.100", Cidr: 24},
			IPv4Calc{Address: "192.168.0.100", Cidr: 24, Netmask: "255.255.255.0",
				NetworkCidr: "192.168.0.0/24", Network: "192.168.0.0",
				HostMin: "192.168.0.1", HostMax: "192.168.0.254", Broadcast: "192.168.0.255",
				HostsAvailable: 254, HostsTotal: 256},
		},
		{
			IPv4Calc{Address: "192.168.0.100", Cidr: 31},
			IPv4Calc{Address: "192.168.0.100", Cidr: 31, Netmask: "255.255.255.254",
				NetworkCidr: "192.168.0.100/31", Network: "192.168.0.100",
				HostMin: "192.168.0.100", HostMax: "192.168.0.101", Broadcast: "",
				HostsAvailable: 2, HostsTotal: 2},
		},
		{
			IPv4Calc{Address: "192.168.0.100", Cidr: 32},
			IPv4Calc{Address: "192.168.0.100", Cidr: 32, Netmask: "255.255.255.255",
				NetworkCidr: "192.168.0.100/32", Network: "192.168.0.100",
				HostMin: "192.168.0.100", HostMax: "192.168.0.100", Broadcast: "",
				HostsAvailable: 1, HostsTotal: 1},
		},
	}

	for _, tt := range validTests {
		tt.actual.getNetworkInfo()
		if tt.actual != tt.expected {
			t.Errorf("Network info invalid, wanted %+v, got %+v", tt.expected, tt.actual)
		}
	}

	var invalidTests = []struct {
		ipCalc IPv4Calc
	}{
		{IPv4Calc{Address: "192.168.0.100", Cidr: 44}},
		{IPv4Calc{Address: "192.168.0.1000", Cidr: 32}},
		{IPv4Calc{Address: "192.168.0.1000", Cidr: 44}},
	}

	for _, tt := range invalidTests {
		err := tt.ipCalc.getNetworkInfo()
		if err == nil {
			t.Errorf("it should have failed as cidr: %d, or/and ip: %s, are invalid", tt.ipCalc.Cidr, tt.ipCalc.Address)
		}
	}
}

func TestNewIPv4Calc(t *testing.T) {
	var ipCalc interface{}

	ipCalc = newIPv4Calc("192.168.0.100", 24)
	switch ipCalc.(type) {
	case *IPv4Calc:
		return
	default:
		t.Errorf("returned struct is not from expected type %T, got %T", IPv4Calc{}, ipCalc)
	}
}

func TestNextIPv4(t *testing.T) {
	var tests = []struct {
		ip, nextIP string
		inc        uint
	}{
		{"192.168.0.0", "192.168.0.100", 100},
		{"192.168.0.0", "192.168.0.255", 255},
		{"192.168.0.0", "192.168.1.254", 510},
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		result := nextIPv4(ip, tt.inc)
		if result != tt.nextIP {
			t.Errorf("wanted %s, got %s", tt.nextIP, result)
		}
	}
}
