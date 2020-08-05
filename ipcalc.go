package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type IpCalc struct {
	Address        string `json:"address"`
	Cidr           int    `json:"cidr"`
	Netmask        string `json:"netmask"`
	NetworkCidr    string `json:"networkCidr"`
	Network        string `json:"network"`
	HostMin        string `json:"hostMin"`
	HostMax        string `json:"hostMax"`
	Broadcast      string `json:"broadcast,omitempty"`
	HostsAvailable int    `json:"hostsAvailable"`
	HostsTotal     int    `json:"hostsTotal"`
}

type ipCalcError struct {
	Error string `json:"error"`
}

func NewIpCalc() *IpCalc {
	return &IpCalc{}
}

func (i *IpCalc) getNetworkInfo() error {
	_, ipNet, err := net.ParseCIDR(i.Address + "/" + strconv.Itoa(i.Cidr))
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
		i.HostMax = nextIP(ipNet.IP, uint(i.HostsTotal-1))
		i.HostsAvailable = 2
	default:
		i.HostMin = nextIP(ipNet.IP, uint(1))
		i.HostMax = nextIP(ipNet.IP, uint(i.HostsTotal-2))
		i.Broadcast = nextIP(ipNet.IP, uint(i.HostsTotal-1))
		i.HostsAvailable = i.HostsTotal - 2
	}
	return nil
}

// nextIP will get the next ip from an IP address
func nextIP(ip net.IP, inc uint) string {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)

	return net.IPv4(v0, v1, v2, v3).String()
}

func getIpCalc(w http.ResponseWriter, r *http.Request) {
	ipCalc := NewIpCalc()

	// populate r.Form
	r.ParseForm()

	// ip and cidr must be provided
	if len(r.Form) != 2 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong number of arguments: %v", len(r.Form)))
		return
	}

	// Get query parameters
	for k, v := range r.Form {
		val := v[0]
		switch k {
		case "ip":
			// ParseIP returns nil if the IP is wrong
			if ip := net.ParseIP(val); ip == nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid ip: %v", val))
				return
			}
			ipCalc.Address = val
		case "cidr":
			valInt, _ := strconv.Atoi(val)
			if net.CIDRMask(valInt, 32) == nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid or too low (<8) cidr: %v", val))
				return
			}
			ipCalc.Cidr = valInt
		default:
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong argument provided: %v", k))
			return
		}
	}

	err := ipCalc.getNetworkInfo()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, ipCalc)
	return
}
